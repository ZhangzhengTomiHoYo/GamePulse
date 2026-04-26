import { useEffect, useMemo, useRef, useState } from 'react'
import { Compass, Copy, Edit3, Plus, RefreshCw, Send, Sparkles, Square, X } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { streamCommunityAgentReply } from '../api/agent.js'
import { useToast } from '../components/ToastProvider.jsx'

const suggestedPrompts = [
  '最近社区里大家对哪个游戏活动讨论最多？',
  '帮我总结一下玩家对当前版本更新的主要意见。',
  '社区里关于角色平衡性最常见的抱怨是什么？',
  '最近有哪些帖子明显偏负面，原因主要集中在哪些点？'
]

const sentimentDisplayMap = {
  positive: '正向',
  neutral: '中性',
  negative: '负向'
}

const createMessageId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`

const createWelcomeMessage = () => ({
  id: createMessageId(),
  role: 'assistant',
  content: '我是游小脉，可以帮你基于社区帖子内容回答问题、梳理观点和定位讨论来源。',
  createdAt: Date.now(),
  excludeFromContext: true,
  pending: false,
  failed: false,
  statusText: '',
  sources: []
})

const createUserMessage = (content) => ({
  id: createMessageId(),
  role: 'user',
  content,
  createdAt: Date.now(),
  excludeFromContext: false,
  pending: false,
  failed: false,
  statusText: '',
  sources: []
})

const createAssistantPlaceholder = () => ({
  id: createMessageId(),
  role: 'assistant',
  content: '',
  createdAt: Date.now(),
  excludeFromContext: false,
  pending: true,
  failed: false,
  statusText: '正在连接社区智能体...',
  sources: []
})

const normalizeSentimentLabel = (label) => {
  if (typeof label !== 'string') return 'neutral'

  const normalized = label.trim().toLowerCase()
  if (normalized === 'positive' || normalized === 'negative') return normalized
  return 'neutral'
}

const getSentimentLabelText = (label) =>
  sentimentDisplayMap[normalizeSentimentLabel(label)] || '中性'

const formatScore = (score) => `相关度 ${(Number(score) || 0).toFixed(2)}`

const formatMessageTime = (time) =>
  new Date(time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

export default function CommunityAgentView() {
  const navigate = useNavigate()
  const toast = useToast()
  const [username] = useState(localStorage.getItem('username') || '用户')
  const [draft, setDraft] = useState('')
  const [messages, setMessages] = useState([])
  const [conversationId, setConversationId] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [activeStatusText, setActiveStatusText] = useState('')
  const abortControllerRef = useRef(null)
  const messageListRef = useRef(null)
  const scrollScheduledRef = useRef(false)

  const hasConversation = useMemo(
    () => messages.some((message) => message.role === 'user'),
    [messages]
  )

  const updateMessage = (messageId, updater) => {
    setMessages((items) =>
      items.map((item) => (item.id === messageId ? updater(item) : item))
    )
  }

  const scrollToBottom = (behavior = 'smooth') => {
    window.requestAnimationFrame(() => {
      const target = messageListRef.current
      if (!target) return

      target.scrollTo({
        top: target.scrollHeight,
        behavior
      })
    })
  }

  const scheduleScrollToBottom = () => {
    if (scrollScheduledRef.current) return

    scrollScheduledRef.current = true
    window.requestAnimationFrame(() => {
      scrollScheduledRef.current = false
      scrollToBottom('auto')
    })
  }

  const ensureAuthenticated = () => {
    if (localStorage.getItem('token')) return true

    navigate('/login', { replace: true })
    return false
  }

  const resetConversation = () => {
    if (streaming) {
      abortControllerRef.current?.abort()
    }

    setMessages([createWelcomeMessage()])
    setConversationId('')
    setDraft('')
    setActiveStatusText('')
    abortControllerRef.current = null
    setStreaming(false)
  }

  const stopStreaming = () => {
    abortControllerRef.current?.abort()
  }

  const copyMessage = async (content) => {
    try {
      await navigator.clipboard.writeText(content)
      toast.success('已复制到剪贴板')
    } catch (error) {
      console.error(error)
      toast.error('复制失败')
    }
  }

  const openSource = (source) => {
    if (!source?.post_id) {
      toast.warning('这条参考帖子暂时无法打开')
      return
    }

    navigate(`/post/${source.post_id}`)
  }

  const isLatestAssistantMessage = (messageId) => {
    const latestAssistant = [...messages]
      .reverse()
      .find((message) => message.role === 'assistant' && !message.excludeFromContext)

    return latestAssistant?.id === messageId
  }

  const sendPromptWithContext = async (question) => {
    const assistantMessage = createAssistantPlaceholder()
    const assistantId = assistantMessage.id
    let receivedContent = ''

    setMessages((items) => [...items, assistantMessage])
    setStreaming(true)
    setActiveStatusText(assistantMessage.statusText)

    const controller = new AbortController()
    abortControllerRef.current = controller
    scrollToBottom('auto')

    try {
      await streamCommunityAgentReply({
        question,
        conversationId,
        signal: controller.signal,
        onStatus: (status) => {
          if (!status) return

          updateMessage(assistantId, (message) => ({
            ...message,
            statusText: status
          }))
          setActiveStatusText(status)
        },
        onSources: (sources) => {
          updateMessage(assistantId, (message) => ({
            ...message,
            sources
          }))
        },
        onDelta: (delta) => {
          if (!delta) return

          receivedContent += delta
          updateMessage(assistantId, (message) => ({
            ...message,
            content: message.content + delta,
            statusText: ''
          }))
          setActiveStatusText('')
          scheduleScrollToBottom()
        },
        onDone: (payload) => {
          if (payload?.conversation_id !== undefined && payload?.conversation_id !== null) {
            setConversationId(String(payload.conversation_id))
          }

          updateMessage(assistantId, (message) => ({
            ...message,
            pending: false,
            statusText: ''
          }))
          setActiveStatusText('')
        }
      })

      if (!receivedContent.trim()) {
        updateMessage(assistantId, (message) => ({
          ...message,
          content: '这次请求没有返回正文内容，你可以稍后再试一次。'
        }))
      }
    } catch (error) {
      if (error?.name === 'AbortError') {
        updateMessage(assistantId, (message) => ({
          ...message,
          content:
            message.content.trim() || '这次回答已经停止，你可以继续追问或重新生成。'
        }))
      } else {
        updateMessage(assistantId, (message) => ({
          ...message,
          failed: true,
          content:
            message.content.trim() ||
            error?.message ||
            '社区智能体暂时不可用，请稍后重试。'
        }))
        toast.error(error?.message || '社区智能体请求失败')
      }
    } finally {
      updateMessage(assistantId, (message) => ({
        ...message,
        pending: false,
        statusText: ''
      }))
      setActiveStatusText('')
      setStreaming(false)
      abortControllerRef.current = null
      scrollToBottom('auto')
    }
  }

  const submitPrompt = async (promptText = draft) => {
    const question = String(promptText || '').trim()
    if (!question || streaming) return
    if (!ensureAuthenticated()) return

    setMessages((items) => [...items, createUserMessage(question)])
    setDraft('')
    await sendPromptWithContext(question)
  }

  const retryAssistantMessage = async (messageId) => {
    if (streaming) return

    const assistantIndex = messages.findIndex((message) => message.id === messageId)
    if (assistantIndex <= 0) return

    let userIndex = assistantIndex - 1
    while (userIndex >= 0 && messages[userIndex].role !== 'user') {
      userIndex -= 1
    }

    if (userIndex < 0) {
      toast.warning('没有找到可重新生成的问题')
      return
    }

    const question = messages[userIndex].content
    setMessages(messages.slice(0, assistantIndex))
    await sendPromptWithContext(question)
  }

  const handleComposerKeydown = (event) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      submitPrompt()
    }
  }

  useEffect(() => {
    if (!ensureAuthenticated()) return undefined

    resetConversation()
    return () => abortControllerRef.current?.abort()
  }, [])

  return (
    <div className="agent-page">
      <header className="agent-header">
        <div className="agent-header-inner">
          <button className="brand-block" type="button" onClick={() => navigate('/')}>
            <div className="brand-mark">GP</div>
            <div className="brand-copy">
              <div className="brand-title">GamePulse</div>
              <div className="brand-subtitle">游小脉社区智能体</div>
            </div>
          </button>

          <div className="header-actions">
            <button className="ghost-action" type="button" onClick={() => navigate('/')}>
              <Compass size={17} />
              <span>返回首页</span>
            </button>
            <button className="ghost-action new-chat-btn" type="button" onClick={resetConversation}>
              <Plus size={17} />
              <span>新对话</span>
            </button>
            <button className="primary-action" type="button" onClick={() => navigate('/post/create')}>
              <Edit3 size={17} />
              <span>写文章</span>
            </button>
            <div className="user-chip">{username.charAt(0).toUpperCase()}</div>
          </div>
        </div>
      </header>

      <main className="agent-main">
        <aside className="left-panel">
          <section className="panel-card nav-card">
            <div className="panel-title">发现</div>
            <button className="nav-item" type="button" onClick={() => navigate('/')}>
              <Compass size={18} />
              <span>综合</span>
            </button>
            <button className="nav-item is-active" type="button">
              <Sparkles size={18} />
              <span>游小脉社区智能体</span>
            </button>
            <button className="nav-item" type="button" onClick={() => navigate('/post/create')}>
              <Plus size={18} />
              <span>发布帖子</span>
            </button>
          </section>

          <section className="panel-card brief-card">
            <div className="panel-title">使用方式</div>
            <ul className="brief-list">
              <li>围绕社区帖子内容提问</li>
              <li>回答会优先引用相关帖子</li>
              <li>适合做舆情归纳与信息定位</li>
            </ul>
          </section>
        </aside>

        <section className="chat-shell">
          <div className="chat-topbar">
            <div>
              <div className="chat-title">游小脉社区智能体</div>
              <div className="chat-subtitle">基于社区帖子内容进行检索增强问答</div>
            </div>
            {streaming && (
              <div className="status-chip">{activeStatusText || '正在生成回答...'}</div>
            )}
          </div>

          <div ref={messageListRef} className="message-list">
            {!hasConversation && (
              <div className="empty-state">
                <div className="empty-badge">COMMUNITY RAG</div>
                <h1>问问社区里已经发生过什么</h1>
                <p>
                  你可以直接提问游戏版本反馈、活动评价、角色讨论、争议话题，智能体会优先结合社区帖子给出答案。
                </p>

                <div className="suggestion-grid">
                  {suggestedPrompts.map((prompt) => (
                    <button
                      key={prompt}
                      type="button"
                      className="suggestion-card"
                      disabled={streaming}
                      onClick={() => submitPrompt(prompt)}
                    >
                      {prompt}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {messages.map((message) => (
              <div key={message.id} className={`message-row role-${message.role}`}>
                <div className="message-avatar">
                  <span>{message.role === 'assistant' ? 'AI' : username.charAt(0).toUpperCase()}</span>
                </div>

                <div className="message-body">
                  <div className="message-meta">
                    <span className="message-name">
                      {message.role === 'assistant' ? '游小脉' : username}
                    </span>
                    <span className="message-time">{formatMessageTime(message.createdAt)}</span>
                  </div>

                  <div
                    className={`message-bubble ${message.pending ? 'is-pending' : ''} ${
                      message.failed ? 'is-failed' : ''
                    }`}
                  >
                    <div className="message-content">{message.content}</div>

                    {message.pending && message.statusText && (
                      <div className="message-status">{message.statusText}</div>
                    )}

                    {message.sources?.length > 0 && (
                      <div className="source-group">
                        <div className="source-group-title">参考帖子</div>
                        {message.sources.map((source) => (
                          <button
                            key={`${message.id}-${source.post_id || source.title}`}
                            type="button"
                            className="source-card"
                            onClick={() => openSource(source)}
                          >
                            <div className="source-card-top">
                              <span className="source-title">{source.title}</span>
                              {source.score > 0 && (
                                <span className="source-score">{formatScore(source.score)}</span>
                              )}
                            </div>
                            <div className="source-meta">
                              <span>{source.community_name}</span>
                              {source.author_name && <span>{source.author_name}</span>}
                              {source.sentiment_label && (
                                <span
                                  className={`sentiment-pill sentiment-${normalizeSentimentLabel(
                                    source.sentiment_label
                                  )}`}
                                >
                                  {getSentimentLabelText(source.sentiment_label)}
                                </span>
                              )}
                            </div>
                            {source.excerpt && <p className="source-excerpt">{source.excerpt}</p>}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>

                  {message.role === 'assistant' && !message.excludeFromContext && !message.pending && (
                    <div className="message-actions">
                      <button className="text-action" type="button" onClick={() => copyMessage(message.content)}>
                        <Copy size={14} />
                        <span>复制</span>
                      </button>
                      {isLatestAssistantMessage(message.id) && (
                        <button
                          className="text-action"
                          type="button"
                          onClick={() => retryAssistantMessage(message.id)}
                        >
                          <RefreshCw size={14} />
                          <span>重新生成</span>
                        </button>
                      )}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>

          <form className="composer-shell" onSubmit={(event) => event.preventDefault()}>
            <div className="composer-box">
              <textarea
                value={draft}
                rows={2}
                placeholder="输入你想了解的社区问题，Enter 发送，Shift + Enter 换行"
                onChange={(event) => setDraft(event.target.value)}
                onKeyDown={handleComposerKeydown}
              />

              <div className="composer-actions">
                <div className="composer-hint">回答仅基于社区内容检索与大模型生成，不代表官方结论。</div>
                <div className="composer-buttons">
                  {streaming ? (
                    <button className="ghost-action" type="button" onClick={stopStreaming}>
                      <Square size={16} />
                      <span>停止回答</span>
                    </button>
                  ) : (
                    <button className="primary-action" type="button" disabled={!draft.trim()} onClick={() => submitPrompt()}>
                      <Send size={16} />
                      <span>发送</span>
                    </button>
                  )}
                </div>
              </div>
            </div>
          </form>
        </section>

        <aside className="right-panel">
          <section className="panel-card">
            <div className="panel-title">推荐提问</div>
            <div className="prompt-list">
              {suggestedPrompts.map((prompt) => (
                <button
                  key={`side-${prompt}`}
                  type="button"
                  className="prompt-chip"
                  disabled={streaming}
                  onClick={() => submitPrompt(prompt)}
                >
                  {prompt}
                </button>
              ))}
            </div>
          </section>

          <section className="panel-card">
            <div className="panel-title">回答特点</div>
            <ul className="brief-list">
              <li>优先召回相关帖子再组织答案</li>
              <li>支持流式输出和中途停止</li>
              <li>展示参考帖子，方便回到原文核对</li>
            </ul>
          </section>
        </aside>
      </main>
    </div>
  )
}
