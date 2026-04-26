<template>
  <div class="agent-page">
    <header class="agent-header">
      <div class="agent-header-inner">
        <div class="brand-block" @click="router.push('/')">
          <div class="brand-mark">GP</div>
          <div class="brand-copy">
            <div class="brand-title">GamePulse</div>
            <div class="brand-subtitle">游小脉社区智能体</div>
          </div>
        </div>

        <div class="header-actions">
          <el-button plain @click="router.push('/')">返回首页</el-button>
          <el-button class="new-chat-btn" @click="resetConversation">新对话</el-button>
          <el-button type="primary" round @click="router.push('/post/create')">
            <el-icon><EditPen /></el-icon>
            <span>写文章</span>
          </el-button>
          <div class="user-chip">{{ username.charAt(0).toUpperCase() }}</div>
        </div>
      </div>
    </header>

    <main class="agent-main">
      <aside class="left-panel">
        <section class="panel-card nav-card">
          <div class="panel-title">发现</div>
          <button class="nav-item" type="button" @click="router.push('/')">
            <el-icon><Compass /></el-icon>
            <span>综合</span>
          </button>
          <button class="nav-item is-active" type="button">
            <el-icon><ChatLineRound /></el-icon>
            <span>游小脉社区智能体</span>
          </button>
          <button class="nav-item" type="button" @click="router.push('/post/create')">
            <el-icon><Plus /></el-icon>
            <span>发布帖子</span>
          </button>
        </section>

        <section class="panel-card brief-card">
          <div class="panel-title">使用方式</div>
          <ul class="brief-list">
            <li>围绕社区帖子内容提问</li>
            <li>回答会优先引用相关帖子</li>
            <li>适合做舆情归纳与信息定位</li>
          </ul>
        </section>
      </aside>

      <section class="chat-shell">
        <div class="chat-topbar">
          <div>
            <div class="chat-title">游小脉社区智能体</div>
            <div class="chat-subtitle">基于社区帖子内容进行检索增强问答</div>
          </div>
          <div class="status-chip" v-if="streaming">
            {{ activeStatusText || '正在生成回答...' }}
          </div>
        </div>

        <div ref="messageListRef" class="message-list">
          <div v-if="!hasConversation" class="empty-state">
            <div class="empty-badge">COMMUNITY RAG</div>
            <h1>问问社区里已经发生过什么</h1>
            <p>
              你可以直接提问游戏版本反馈、活动评价、角色讨论、争议话题，智能体会优先结合社区帖子给出答案。
            </p>

            <div class="suggestion-grid">
              <button
                v-for="prompt in suggestedPrompts"
                :key="prompt"
                type="button"
                class="suggestion-card"
                :disabled="streaming"
                @click="submitPrompt(prompt)"
              >
                {{ prompt }}
              </button>
            </div>
          </div>

          <div
            v-for="message in messages"
            :key="message.id"
            class="message-row"
            :class="`role-${message.role}`"
          >
            <div class="message-avatar">
              <span v-if="message.role === 'assistant'">AI</span>
              <span v-else>{{ username.charAt(0).toUpperCase() }}</span>
            </div>

            <div class="message-body">
              <div class="message-meta">
                <span class="message-name">
                  {{ message.role === 'assistant' ? '游小脉' : username }}
                </span>
                <span class="message-time">{{ formatMessageTime(message.createdAt) }}</span>
              </div>

              <div
                class="message-bubble"
                :class="{
                  'is-pending': message.pending,
                  'is-failed': message.failed
                }"
              >
                <div class="message-content">{{ message.content }}</div>

                <div v-if="message.pending && message.statusText" class="message-status">
                  {{ message.statusText }}
                </div>

                <div v-if="message.sources?.length" class="source-group">
                  <div class="source-group-title">参考帖子</div>
                  <button
                    v-for="source in message.sources"
                    :key="`${message.id}-${source.post_id || source.title}`"
                    type="button"
                    class="source-card"
                    @click="openSource(source)"
                  >
                    <div class="source-card-top">
                      <span class="source-title">{{ source.title }}</span>
                      <span class="source-score" v-if="source.score > 0">
                        {{ formatScore(source.score) }}
                      </span>
                    </div>
                    <div class="source-meta">
                      <span>{{ source.community_name }}</span>
                      <span v-if="source.author_name">{{ source.author_name }}</span>
                      <span
                        v-if="source.sentiment_label"
                        class="sentiment-pill"
                        :class="`sentiment-${normalizeSentimentLabel(source.sentiment_label)}`"
                      >
                        {{ getSentimentLabelText(source.sentiment_label) }}
                      </span>
                    </div>
                    <p v-if="source.excerpt" class="source-excerpt">{{ source.excerpt }}</p>
                  </button>
                </div>
              </div>

              <div
                v-if="message.role === 'assistant' && !message.excludeFromContext && !message.pending"
                class="message-actions"
              >
                <button class="text-action" type="button" @click="copyMessage(message.content)">
                  复制
                </button>
                <button
                  v-if="isLatestAssistantMessage(message.id)"
                  class="text-action"
                  type="button"
                  @click="retryAssistantMessage(message.id)"
                >
                  重新生成
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="composer-shell">
          <div class="composer-box">
            <el-input
              v-model="draft"
              type="textarea"
              resize="none"
              :autosize="{ minRows: 2, maxRows: 6 }"
              placeholder="输入你想了解的社区问题，Enter 发送，Shift + Enter 换行"
              @keydown="handleComposerKeydown"
            />

            <div class="composer-actions">
              <div class="composer-hint">回答仅基于社区内容检索与大模型生成，不代表官方结论。</div>
              <div class="composer-buttons">
                <el-button v-if="streaming" plain @click="stopStreaming">停止回答</el-button>
                <el-button
                  v-else
                  type="primary"
                  :disabled="!draft.trim()"
                  @click="submitPrompt()"
                >
                  发送
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <aside class="right-panel">
        <section class="panel-card">
          <div class="panel-title">推荐提问</div>
          <div class="prompt-list">
            <button
              v-for="prompt in suggestedPrompts"
              :key="`side-${prompt}`"
              type="button"
              class="prompt-chip"
              :disabled="streaming"
              @click="submitPrompt(prompt)"
            >
              {{ prompt }}
            </button>
          </div>
        </section>

        <section class="panel-card">
          <div class="panel-title">回答特点</div>
          <ul class="brief-list">
            <li>优先召回相关帖子再组织答案</li>
            <li>支持流式输出和中途停止</li>
            <li>展示参考帖子，方便回到原文核对</li>
          </ul>
        </section>
      </aside>
    </main>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChatLineRound, Compass, EditPen, Plus } from '@element-plus/icons-vue'
import { streamCommunityAgentReply } from '../api/agent'

const router = useRouter()
const username = ref(localStorage.getItem('username') || '用户')
const draft = ref('')
const messages = ref([])
const conversationId = ref('')
const streaming = ref(false)
const abortController = ref(null)
const messageListRef = ref(null)
const activeStatusText = ref('')

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

const hasConversation = computed(() =>
  messages.value.some((message) => message.role === 'user')
)

const createMessageId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`

const createWelcomeMessage = () => ({
  id: createMessageId(),
  role: 'assistant',
  content:
    '我是游小脉，可以帮你基于社区帖子内容回答问题、梳理观点和定位讨论来源。',
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

const scrollToBottom = async (behavior = 'smooth') => {
  await nextTick()

  const target = messageListRef.value
  if (!target) return

  target.scrollTo({
    top: target.scrollHeight,
    behavior
  })
}

const ensureAuthenticated = () => {
  if (localStorage.getItem('token')) return true

  router.replace('/login')
  return false
}

const resetConversation = () => {
  if (streaming.value) {
    abortController.value?.abort()
  }

  messages.value = [createWelcomeMessage()]
  conversationId.value = ''
  draft.value = ''
  activeStatusText.value = ''
  abortController.value = null
  streaming.value = false
}

const stopStreaming = () => {
  abortController.value?.abort()
}

const copyMessage = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    console.error(error)
    ElMessage.error('复制失败')
  }
}

const openSource = (source) => {
  if (!source?.post_id) {
    ElMessage.warning('这条参考帖子暂时无法打开')
    return
  }

  router.push(`/post/${source.post_id}`)
}

const isLatestAssistantMessage = (messageId) => {
  const latestAssistant = [...messages.value]
    .reverse()
    .find((message) => message.role === 'assistant' && !message.excludeFromContext)

  return latestAssistant?.id === messageId
}

const sendPromptWithContext = async (question) => {
  const assistantMessage = reactive(createAssistantPlaceholder())
  messages.value.push(assistantMessage)
  streaming.value = true
  activeStatusText.value = assistantMessage.statusText

  const controller = new AbortController()
  abortController.value = controller
  let scrollScheduled = false

  const scheduleScrollToBottom = () => {
    if (scrollScheduled) return

    scrollScheduled = true

    const scheduler =
      typeof window !== 'undefined' && typeof window.requestAnimationFrame === 'function'
        ? window.requestAnimationFrame.bind(window)
        : (callback) => window.setTimeout(callback, 16)

    scheduler(async () => {
      scrollScheduled = false
      await scrollToBottom('auto')
    })
  }

  await scrollToBottom('auto')

  try {
    await streamCommunityAgentReply({
      question,
      conversationId: conversationId.value,
      signal: controller.signal,
      onStatus: (status) => {
        if (!status) return
        assistantMessage.statusText = status
        activeStatusText.value = status
      },
      onSources: (sources) => {
        assistantMessage.sources = sources
      },
      onDelta: (delta) => {
        if (!delta) return
        assistantMessage.content += delta
        assistantMessage.statusText = ''
        activeStatusText.value = ''
        scheduleScrollToBottom()
      },
      onDone: (payload) => {
        if (payload?.conversation_id !== undefined && payload?.conversation_id !== null) {
          conversationId.value = String(payload.conversation_id)
        }
        assistantMessage.pending = false
        assistantMessage.statusText = ''
        activeStatusText.value = ''
      }
    })

    if (!assistantMessage.content.trim()) {
      assistantMessage.content = '这次请求没有返回正文内容，你可以稍后再试一次。'
    }
  } catch (error) {
    if (error?.name === 'AbortError') {
      assistantMessage.content =
        assistantMessage.content.trim() || '这次回答已经停止，你可以继续追问或重新生成。'
    } else {
      assistantMessage.failed = true
      assistantMessage.content =
        assistantMessage.content.trim() ||
        error?.message ||
        '社区智能体暂时不可用，请稍后重试。'
      ElMessage.error(error?.message || '社区智能体请求失败')
    }
  } finally {
    assistantMessage.pending = false
    assistantMessage.statusText = ''
    activeStatusText.value = ''
    streaming.value = false
    abortController.value = null
    await scrollToBottom('auto')
  }
}

const submitPrompt = async (promptText = draft.value) => {
  const question = promptText.trim()
  if (!question || streaming.value) return
  if (!ensureAuthenticated()) return

  messages.value.push(createUserMessage(question))
  draft.value = ''

  await sendPromptWithContext(question)
}

const retryAssistantMessage = async (messageId) => {
  if (streaming.value) return

  const assistantIndex = messages.value.findIndex((message) => message.id === messageId)
  if (assistantIndex <= 0) return

  let userIndex = assistantIndex - 1
  while (userIndex >= 0 && messages.value[userIndex].role !== 'user') {
    userIndex -= 1
  }

  if (userIndex < 0) {
    ElMessage.warning('没有找到可重新生成的问题')
    return
  }

  const question = messages.value[userIndex].content
  messages.value = messages.value.slice(0, assistantIndex)

  await sendPromptWithContext(question)
}

const handleComposerKeydown = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    submitPrompt()
  }
}

onMounted(() => {
  if (!ensureAuthenticated()) return

  resetConversation()
})
</script>

<style scoped>
.agent-page {
  --agent-bg: linear-gradient(180deg, #f7f8fb 0%, #edf2ff 100%);
  --panel-bg: rgba(255, 255, 255, 0.84);
  --panel-border: rgba(92, 123, 250, 0.12);
  --panel-shadow: 0 20px 40px rgba(59, 76, 145, 0.08);
  --brand-main: #2457f5;
  --brand-soft: #eef3ff;
  --accent: #1fb7a6;
  --text-main: #172033;
  --text-subtle: #5c6680;
  --danger: #e45757;
  min-height: 100vh;
  background: var(--agent-bg);
  color: var(--text-main);
}

.agent-header {
  position: sticky;
  top: 0;
  z-index: 20;
  backdrop-filter: blur(18px);
  background: rgba(247, 248, 251, 0.72);
  border-bottom: 1px solid rgba(36, 87, 245, 0.08);
}

.agent-header-inner {
  max-width: 1440px;
  margin: 0 auto;
  padding: 16px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
}

.brand-mark {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: linear-gradient(135deg, #2457f5 0%, #1fb7a6 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.08em;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-title {
  font-size: 18px;
  font-weight: 700;
}

.brand-subtitle {
  font-size: 13px;
  color: var(--text-subtle);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.new-chat-btn {
  border-color: rgba(36, 87, 245, 0.2);
}

.user-chip {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: #fff;
  border: 1px solid rgba(36, 87, 245, 0.14);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: var(--brand-main);
  box-shadow: 0 8px 24px rgba(59, 76, 145, 0.08);
}

.agent-main {
  max-width: 1440px;
  margin: 0 auto;
  padding: 24px;
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr) 280px;
  gap: 20px;
}

.left-panel,
.right-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-card,
.chat-shell {
  background: var(--panel-bg);
  border: 1px solid var(--panel-border);
  box-shadow: var(--panel-shadow);
  backdrop-filter: blur(16px);
}

.panel-card {
  border-radius: 24px;
  padding: 18px;
}

.panel-title {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-subtle);
  text-transform: uppercase;
  margin-bottom: 14px;
}

.nav-card {
  padding: 14px;
}

.nav-item {
  width: 100%;
  border: none;
  background: transparent;
  border-radius: 18px;
  padding: 14px 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-main);
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.nav-item:hover {
  background: rgba(36, 87, 245, 0.06);
  transform: translateY(-1px);
}

.nav-item.is-active {
  background: linear-gradient(135deg, rgba(36, 87, 245, 0.14) 0%, rgba(31, 183, 166, 0.12) 100%);
  color: var(--brand-main);
  font-weight: 700;
}

.brief-list {
  margin: 0;
  padding-left: 18px;
  color: var(--text-subtle);
  font-size: 14px;
  line-height: 1.8;
}

.chat-shell {
  min-height: calc(100vh - 130px);
  border-radius: 30px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-topbar {
  padding: 22px 24px 18px;
  border-bottom: 1px solid rgba(36, 87, 245, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background:
    radial-gradient(circle at top left, rgba(36, 87, 245, 0.12), transparent 42%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(255, 255, 255, 0.74));
}

.chat-title {
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.chat-subtitle {
  margin-top: 4px;
  color: var(--text-subtle);
  font-size: 13px;
}

.status-chip {
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(36, 87, 245, 0.08);
  color: var(--brand-main);
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 18px 24px 28px;
}

.empty-state {
  padding: 36px 8px 22px;
}

.empty-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(31, 183, 166, 0.1);
  color: var(--accent);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.empty-state h1 {
  margin: 18px 0 10px;
  font-size: 38px;
  line-height: 1.1;
  letter-spacing: -0.04em;
  max-width: 560px;
}

.empty-state p {
  margin: 0;
  max-width: 700px;
  font-size: 15px;
  line-height: 1.8;
  color: var(--text-subtle);
}

.suggestion-grid {
  margin-top: 28px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.suggestion-card {
  border: 1px solid rgba(36, 87, 245, 0.1);
  border-radius: 22px;
  padding: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(238, 243, 255, 0.72));
  text-align: left;
  color: var(--text-main);
  font-size: 15px;
  line-height: 1.7;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.suggestion-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 30px rgba(59, 76, 145, 0.08);
}

.message-row {
  display: flex;
  gap: 14px;
  margin-bottom: 22px;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 16px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
}

.role-assistant .message-avatar {
  background: linear-gradient(135deg, #2457f5 0%, #1fb7a6 100%);
  color: #fff;
}

.role-user {
  flex-direction: row-reverse;
}

.role-user .message-avatar {
  background: #fff;
  color: var(--brand-main);
  border: 1px solid rgba(36, 87, 245, 0.14);
}

.message-body {
  flex: 1;
  min-width: 0;
}

.role-user .message-body {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  color: var(--text-subtle);
  font-size: 12px;
}

.message-name {
  font-weight: 700;
  color: var(--text-main);
}

.message-bubble {
  max-width: min(100%, 840px);
  padding: 16px 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(36, 87, 245, 0.08);
}

.role-user .message-bubble {
  background: linear-gradient(135deg, rgba(36, 87, 245, 0.95), rgba(54, 103, 255, 0.88));
  color: #fff;
  border-color: transparent;
}

.message-bubble.is-pending {
  box-shadow: inset 0 0 0 1px rgba(31, 183, 166, 0.1);
}

.message-bubble.is-failed {
  border-color: rgba(228, 87, 87, 0.24);
}

.message-content {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 15px;
  line-height: 1.8;
}

.message-status {
  margin-top: 12px;
  font-size: 13px;
  color: var(--accent);
}

.message-actions {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 10px;
}

.text-action {
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-subtle);
  font-size: 13px;
  cursor: pointer;
}

.text-action:hover {
  color: var(--brand-main);
}

.source-group {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid rgba(36, 87, 245, 0.08);
  display: grid;
  gap: 10px;
}

.source-group-title {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-subtle);
  text-transform: uppercase;
}

.source-card {
  width: 100%;
  text-align: left;
  border: 1px solid rgba(36, 87, 245, 0.08);
  background: rgba(247, 248, 251, 0.92);
  border-radius: 18px;
  padding: 14px;
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease;
}

.source-card:hover {
  transform: translateY(-1px);
  border-color: rgba(36, 87, 245, 0.22);
}

.source-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.source-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-main);
}

.source-score {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--brand-main);
}

.source-meta {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: var(--text-subtle);
}

.sentiment-pill {
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 600;
}

.sentiment-positive {
  background: rgba(47, 179, 93, 0.12);
  color: #22874a;
}

.sentiment-neutral {
  background: rgba(92, 102, 128, 0.12);
  color: #5c6680;
}

.sentiment-negative {
  background: rgba(228, 87, 87, 0.12);
  color: #c34848;
}

.source-excerpt {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-subtle);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  overflow: hidden;
}

.composer-shell {
  padding: 18px 24px 24px;
  border-top: 1px solid rgba(36, 87, 245, 0.08);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.3), rgba(255, 255, 255, 0.9));
}

.composer-box {
  border: 1px solid rgba(36, 87, 245, 0.1);
  background: #fff;
  border-radius: 24px;
  padding: 12px;
}

:deep(.composer-box .el-textarea__inner) {
  box-shadow: none;
  border: none;
  background: transparent;
  padding: 8px 10px 10px;
  font-size: 15px;
  line-height: 1.8;
}

.composer-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 0 6px 4px;
}

.composer-hint {
  font-size: 12px;
  color: var(--text-subtle);
  line-height: 1.6;
}

.composer-buttons {
  flex-shrink: 0;
}

.prompt-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.prompt-chip {
  width: 100%;
  border: 1px solid rgba(36, 87, 245, 0.1);
  background: #fff;
  border-radius: 16px;
  padding: 12px 14px;
  text-align: left;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-main);
  cursor: pointer;
}

.prompt-chip:hover {
  border-color: rgba(36, 87, 245, 0.24);
  background: rgba(238, 243, 255, 0.8);
}

@media (max-width: 1200px) {
  .agent-main {
    grid-template-columns: 220px minmax(0, 1fr);
  }

  .right-panel {
    display: none;
  }
}

@media (max-width: 900px) {
  .agent-header-inner,
  .agent-main {
    padding-left: 16px;
    padding-right: 16px;
  }

  .agent-main {
    grid-template-columns: minmax(0, 1fr);
  }

  .left-panel {
    display: none;
  }

  .chat-shell {
    min-height: calc(100vh - 104px);
  }

  .suggestion-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .agent-header-inner {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .chat-topbar,
  .message-list,
  .composer-shell {
    padding-left: 16px;
    padding-right: 16px;
  }

  .empty-state h1 {
    font-size: 30px;
  }

  .composer-actions {
    align-items: flex-start;
    flex-direction: column;
  }

  .composer-buttons {
    width: 100%;
  }

  .composer-buttons :deep(.el-button) {
    width: 100%;
  }
}
</style>
