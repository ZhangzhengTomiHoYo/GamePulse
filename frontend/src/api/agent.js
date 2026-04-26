const buildStreamEndpoint = (conversationId) => `/api/v1/chat/${conversationId}/stream`

const getAuthHeaders = () => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`,
  'Content-Type': 'application/json',
  Accept: 'text/event-stream, application/json, text/plain'
})

const parseMaybeJSON = (value) => {
  if (typeof value !== 'string') return value

  const trimmed = value.trim()
  if (!trimmed) return ''

  try {
    return JSON.parse(trimmed)
  } catch {
    return value
  }
}

const normalizeSources = (rawSources) => {
  if (!Array.isArray(rawSources)) return []

  return rawSources
    .map((item, index) => {
      if (!item || typeof item !== 'object') return null

      const postId = item.post_id ?? item.id ?? item.postId ?? null
      const title = item.title ?? item.post_title ?? `参考帖子 ${index + 1}`
      const excerpt = item.excerpt ?? item.summary ?? item.content ?? item.post_content ?? ''
      const communityName =
        item.community_name ??
        item.community?.name ??
        item.community ??
        '综合'

      return {
        post_id: postId !== null && postId !== undefined && postId !== '' ? String(postId) : null,
        title: String(title),
        excerpt: typeof excerpt === 'string' ? excerpt : '',
        community_name: String(communityName),
        score: typeof item.score === 'number' ? item.score : Number(item.score || 0),
        sentiment_label: item.sentiment_label ?? item.sentimentLabel ?? '',
        author_name: item.author_name ?? item.authorName ?? ''
      }
    })
    .filter(Boolean)
}

const extractText = (payload) => {
  if (typeof payload === 'string') return payload
  if (!payload || typeof payload !== 'object') return ''

  return (
    payload.delta ??
    payload.content ??
    payload.answer ??
    payload.text ??
    payload.message ??
    ''
  )
}

const extractStatus = (payload) => {
  if (typeof payload === 'string') return payload
  if (!payload || typeof payload !== 'object') return ''

  return String(payload.status ?? payload.message ?? payload.text ?? '')
}

const extractSources = (payload) => {
  if (Array.isArray(payload)) return normalizeSources(payload)
  if (!payload || typeof payload !== 'object') return []

  return normalizeSources(payload.sources ?? payload.data ?? [])
}

const extractError = (payload) => {
  if (typeof payload === 'string') return payload
  if (!payload || typeof payload !== 'object') return '社区智能体暂时不可用'

  return String(payload.error ?? payload.message ?? payload.msg ?? '社区智能体暂时不可用')
}

const dispatchStreamEvent = async (eventType, payload, hooks) => {
  const type = (eventType || '').trim().toLowerCase()

  if (type === 'error') {
    const errorMessage = extractError(payload)
    await hooks.onError?.(errorMessage)
    throw new Error(errorMessage)
  }

  if (type === 'status') {
    await hooks.onStatus?.(extractStatus(payload))
    return
  }

  if (type === 'sources') {
    await hooks.onSources?.(extractSources(payload))
    return
  }

  if (type === 'done') {
    await hooks.onDone?.(payload)
    return
  }

  if (type === 'delta' || type === 'message' || type === '') {
    const deltaText = extractText(payload)
    if (deltaText) {
      await hooks.onDelta?.(deltaText)
    }

    const sources = extractSources(payload)
    if (sources.length > 0) {
      await hooks.onSources?.(sources)
    }

    const status = extractStatus(payload)
    if (status && typeof payload === 'object' && payload.delta === undefined) {
      await hooks.onStatus?.(status)
    }
  }
}

const handleJSONResponse = async (response, hooks) => {
  const data = await response.json()

  if (!response.ok) {
    throw new Error(data?.msg || data?.message || '社区智能体请求失败')
  }

  const sources = extractSources(data)
  if (sources.length > 0) {
    hooks.onSources?.(sources)
  }

  const answer = extractText(data)
  if (answer) {
    hooks.onDelta?.(answer)
  }

  hooks.onDone?.(data)
}

const handleTextStream = async (response, hooks) => {
  const reader = response.body?.getReader()
  if (!reader) throw new Error('浏览器不支持流式响应')

  const decoder = new TextDecoder('utf-8')

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    const chunk = decoder.decode(value, { stream: true })
    if (chunk) {
      hooks.onDelta?.(chunk)
    }
  }

  hooks.onDone?.()
}

const handleSSEStream = async (response, hooks) => {
  const reader = response.body?.getReader()
  if (!reader) throw new Error('浏览器不支持流式响应')

  const decoder = new TextDecoder('utf-8')
  let buffer = ''

  const processEventBlock = async (block) => {
    if (!block.trim()) return

    const lines = block.split(/\r?\n/)
    let eventType = ''
    const dataLines = []

    for (const line of lines) {
      if (line.startsWith('event:')) {
        eventType = line.slice(6).trim()
        continue
      }

      if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trimStart())
      }
    }

    const rawData = dataLines.join('\n')
    if (!rawData) return
    if (rawData === '[DONE]') {
      await hooks.onDone?.()
      return
    }

    const payload = parseMaybeJSON(rawData)
    await dispatchStreamEvent(eventType, payload, hooks)
  }

  while (true) {
    const { done, value } = await reader.read()
    buffer += decoder.decode(value || new Uint8Array(), { stream: !done })

    let separatorIndex = buffer.search(/\r?\n\r?\n/)
    while (separatorIndex !== -1) {
      const block = buffer.slice(0, separatorIndex)
      await processEventBlock(block)
      buffer = buffer.slice(separatorIndex + (buffer[separatorIndex] === '\r' ? 4 : 2))
      separatorIndex = buffer.search(/\r?\n\r?\n/)
    }

    if (done) break
  }

  if (buffer.trim()) {
    await processEventBlock(buffer)
  }
}

export const streamCommunityAgentReply = async ({
  question,
  conversationId,
  signal,
  onStatus,
  onSources,
  onDelta,
  onDone,
  onError
}) => {
  const routeConversationId = conversationId ? String(conversationId) : 'new'
  const payload = {
    query: question
  }

  if (conversationId) {
    payload.conversation_id = String(conversationId)
  }

  const response = await fetch(buildStreamEndpoint(routeConversationId), {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
    signal
  })

  const contentType = response.headers.get('content-type') || ''

  if (contentType.includes('application/json')) {
    await handleJSONResponse(response, { onStatus, onSources, onDelta, onDone, onError })
    return
  }

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || '社区智能体请求失败')
  }

  if (contentType.includes('text/event-stream')) {
    await handleSSEStream(response, { onStatus, onSources, onDelta, onDone, onError })
    return
  }

  await handleTextStream(response, { onStatus, onSources, onDelta, onDone, onError })
}
