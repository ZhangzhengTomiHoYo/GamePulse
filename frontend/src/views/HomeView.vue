<template>
  <div class="common-layout">
    
    <el-header class="app-header">
      <div class="header-inner">
        <div class="logo-area" @click="router.push('/')">
          <span class="logo-icon">🎐</span>
          <span class="logo-text">GamePulse</span>
        </div>
        
        <div class="user-area">
          <span class="welcome-text">Hi, {{ username }}</span>
          <el-button type="primary" round class="write-btn" @click="goCreatePost">
            <el-icon><EditPen /></el-icon>
            <span>写文章</span>
          </el-button>
          <el-dropdown trigger="click" @command="handleCommand">
            <el-avatar :size="36" class="user-avatar">{{ username.charAt(0).toUpperCase() }}</el-avatar>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided style="color: #f56c6c;">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>

    <div class="main-container">
      <el-row :gutter="24">
        
        <el-col :xs="0" :sm="5" :md="5" :lg="4" class="hidden-xs-only">
          <div class="sidebar-left">
            <el-menu
              :default-active="String(currentCommunityId)"
              class="community-menu"
            >
              <div class="menu-title">发现</div>
              <el-menu-item index="0" @click="handleCommunityChange(0)">
                <el-icon><Compass /></el-icon>
                <span>综合</span>
              </el-menu-item>
              
              <div class="menu-title" style="margin-top: 10px;">社区</div>
              <el-menu-item 
                v-for="item in communityList" 
                :key="item.id" 
                :index="String(item.id)"
                @click="handleCommunityChange(item.id)"
              >
                <el-icon><Collection /></el-icon>
                <span>{{ item.name }}</span>
              </el-menu-item>
            </el-menu>
          </div>
        </el-col>

        <el-col :xs="24" :sm="19" :md="14" :lg="15">
          <div class="content-middle">
            <div class="feed-tabs">
              <el-tabs v-model="sortBy" @tab-change="handleTabChange">
                <el-tab-pane label="🔥 热门" name="score"></el-tab-pane>
                <el-tab-pane label="🕒 最新" name="time"></el-tab-pane>
              </el-tabs>
            </div>

            <div v-loading="loading" class="post-list">
              <el-card 
                v-for="post in posts" 
                :key="post.id" 
                class="post-item" 
                shadow="hover"
                @click="goDetail(post.id)"
              >
                <div class="post-meta-top">
                  <span class="author-name">{{ post.author_name || '匿名用户' }}</span>
                  <span class="divider">·</span>
                  <span class="post-time">{{ formatTime(post.create_time) }}</span>
                  <span class="divider">·</span>
                  <el-tag size="small" type="info" effect="plain" class="community-tag">
                    {{ post.community?.name || '综合' }}
                  </el-tag>
                </div>
                
                <h2 class="post-title">{{ post.title }}</h2>
                <p class="post-abstract">{{ post.content }}</p>
                
                <div class="post-actions">
                  <div class="action-group vote-group">
                    <div 
                      class="vote-btn up" 
                      :class="{ active: post.vote_status === 1 }"
                      @click.stop="handleVote(post, 1)"
                    >
                      <el-icon><CaretTop /></el-icon>
                      <span v-if="post.votes !== 0">{{ post.votes }}</span>
                      <span v-else>赞</span>
                    </div>
                    <div 
                      class="vote-btn down" 
                      :class="{ active: post.vote_status === -1 }"
                      @click.stop="handleVote(post, -1)"
                    >
                      <el-icon><CaretBottom /></el-icon>
                    </div>
                  </div>

                  <div class="action-group">
                    <el-icon><ChatLineRound /></el-icon>
                    <span>评论</span>
                  </div>
                  <div class="action-group">
                    <el-icon><Share /></el-icon>
                    <span>分享</span>
                  </div>
                </div>
              </el-card>

              <el-empty v-if="posts.length === 0 && !loading" description="暂无内容，来发布第一篇吧！" />
            </div>
          </div>
        </el-col>

        <el-col :xs="0" :sm="0" :md="5" :lg="5" class="hidden-sm-and-down">
          <div class="sidebar-right">
            
            <el-card class="widget-card welcome-widget" shadow="never">
              <div class="widget-header">
                <h3>GamePulse 社区</h3>
              </div>
              <div class="widget-content">
                <p>专注于游戏社区舆情分析</p>
                <div class="stat-row">
                  <div class="stat-item">
                    <div class="count">100+</div>
                    <div class="label">帖子</div>
                  </div>
                  <div class="stat-item">
                    <div class="count">50+</div>
                    <div class="label">用户</div>
                  </div>
                </div>
              </div>
            </el-card>

            <el-card class="widget-card" shadow="never">
              <template #header>
                <div class="card-header">
                  <span>📢 社区公告</span>
                </div>
              </template>
              <ul class="notice-list">
                <li>🎉 GamePulse v1.0 正式上线！</li>
                <li>🔒 关于社区账号安全的说明</li>
                <li>📝 创作你喜欢的游戏文章</li>
              </ul>
            </el-card>
            
            <div class="site-footer">
              © 2026 GamePulse Blog <br>
              <a href="#">关于我们</a> · <a href="#">联系作者</a>
            </div>

          </div>
        </el-col>

      </el-row>
    </div>

    <el-backtop :right="40" :bottom="40" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  EditPen, Compass, Collection, CaretTop, CaretBottom, 
  ChatLineRound, Share 
} from '@element-plus/icons-vue'

const router = useRouter()
const username = ref(localStorage.getItem('username') || '用户')
const posts = ref([])
const communityList = ref([])
const loading = ref(false)
const sortBy = ref('score')
const currentCommunityId = ref(0)

const goCreatePost = () => { router.push('/post/create') }
const goDetail = (id) => { router.push(`/post/${id}`) }

// 处理下拉菜单命令
const handleCommand = (command) => {
  if (command === 'logout') {
    localStorage.clear()
    router.push('/login')
    ElMessage.success('已安全退出')
  }
}

// 简单的时间格式化
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const getCommunityList = async () => {
  try {
    const token = localStorage.getItem('token')
    const res = await axios.get('/api/v1/community', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (res.data.code === 1000) communityList.value = res.data.data
  } catch (error) { console.error(error) }
}

const fetchPosts = async () => {
  const token = localStorage.getItem('token')
  if (!token) return
  loading.value = true
  try {
    const params = { page: 1, size: 20, order: sortBy.value }
    if (currentCommunityId.value !== 0) params.community_id = currentCommunityId.value
    
    const res = await axios.get('/api/v1/posts2', {
      params, headers: { 'Authorization': `Bearer ${token}` }
    })

    if (res.data.code === 1000) {
      const list = res.data.data || []
      list.forEach(item => {
        if (item.vote_status === undefined) item.vote_status = 0
        item.votes = Number(item.votes)
      })
      posts.value = list
    }
  } catch (error) { ElMessage.error("获取数据失败") } 
  finally { loading.value = false }
}

const handleCommunityChange = (id) => {
  currentCommunityId.value = id
  fetchPosts()
}

const handleTabChange = () => { fetchPosts() }

const handleVote = async (post, direction) => {
  const token = localStorage.getItem('token')
  let dirToSend = direction
  if (post.vote_status === direction) dirToSend = 0

  try {
    const res = await axios.post('/api/v1/vote', {
      post_id: String(post.id),
      direction: String(dirToSend)
    }, { headers: { 'Authorization': `Bearer ${token}` } })

    if (res.data.code === 1000) {
      if (post.vote_status === 1) post.votes -= 1
      post.vote_status = dirToSend
      if (dirToSend === 1) post.votes += 1
    } else { ElMessage.error(res.data.msg) }
  } catch (error) { ElMessage.error("投票失败") }
}

onMounted(() => {
  if (!localStorage.getItem('token')) {
    router.push('/login')
    return
  }
  getCommunityList()
  fetchPosts()
})
</script>

<style scoped>
/* 全局背景色，模拟主流社区的浅灰底色 */
.common-layout {
  background-color: #f2f3f5;
  min-height: 100vh;
}

/* --- 1. Header 样式重构 --- */
.app-header {
  background-color: #fff;
  border-bottom: 1px solid #e4e6eb;
  position: fixed;
  width: 100%;
  top: 0;
  z-index: 1000;
  padding: 0;
  height: 60px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.header-inner {
  max-width: 1240px; /* 限制最大宽度，与内容区对齐 */
  margin: 0 auto;
  height: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.logo-area {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}
.logo-icon { font-size: 24px; }
.logo-text {
  font-size: 22px;
  font-weight: 700;
  color: #1e80ff; /* 品牌色 */
  letter-spacing: 1px;
}

.user-area { display: flex; align-items: center; gap: 16px; }
.welcome-text { color: #515767; font-size: 14px; display: none; } /* 移动端隐藏 */
@media (min-width: 768px) { .welcome-text { display: block; } }

.user-avatar { cursor: pointer; border: 2px solid #fff; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.user-avatar:hover { transform: scale(1.05); transition: transform 0.2s; }

/* --- 2. 主体布局 --- */
.main-container {
  margin-top: 60px; /* 给 Header 让位 */
  max-width: 1240px; /* 核心：限制宽度并居中 */
  margin-left: auto;
  margin-right: auto;
  padding: 20px;
}

/* --- 3. 左侧侧边栏 (Sticky) --- */
.sidebar-left {
  position: sticky;
  top: 80px; /* 吸顶距离 */
  background: #fff;
  border-radius: 4px;
  padding: 8px 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.community-menu { border-right: none; }
.menu-title {
  padding: 10px 20px;
  font-size: 12px;
  color: #86909c;
}

/* --- 4. 中间内容区 --- */
.content-middle {
  min-height: 500px;
}

.feed-tabs {
  background: #fff;
  padding: 10px 20px 0;
  border-radius: 4px 4px 0 0;
  border-bottom: 1px solid #e4e6eb;
}

/* 帖子卡片优化 */
.post-item {
  margin-bottom: 10px;
  border: none;
  border-radius: 0;
  cursor: pointer;
  transition: all 0.2s;
}
.post-item:first-child { border-radius: 0 0 4px 4px; } /* 第一个圆角处理 */
.post-item:hover { background-color: #fafafa; }

.post-meta-top {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #86909c;
  margin-bottom: 8px;
}
.author-name { color: #515767; font-weight: 500; }
.divider { margin: 0 8px; color: #e5e6eb; }

.post-title {
  font-size: 18px;
  font-weight: 700;
  color: #1d2129;
  margin: 0 0 8px;
  line-height: 24px;
}
.post-abstract {
  color: #86909c;
  font-size: 14px;
  line-height: 22px;
  margin-bottom: 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2; /* 限制2行 */
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 底部操作区 */
.post-actions { display: flex; align-items: center; gap: 24px; }
.action-group {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #86909c;
  font-size: 13px;
  cursor: pointer;
  transition: color 0.2s;
}
.action-group:hover { color: #1e80ff; }

/* 投票按钮特化 */
.vote-group {
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  padding: 2px 0;
}
.vote-btn {
  padding: 0 8px;
  display: flex;
  align-items: center;
  gap: 2px;
}
.vote-btn:hover { color: #1e80ff; }
.vote-btn.up { border-right: 1px solid #e5e6eb; }
.vote-btn.up.active { color: #1e80ff; }
.vote-btn.down.active { color: #1e80ff; }

/* --- 5. 右侧侧边栏 (新增) --- */
.sidebar-right {
  position: sticky;
  top: 80px;
}

.widget-card {
  margin-bottom: 16px;
  border: none;
  border-radius: 4px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

/* 欢迎卡片 */
.welcome-widget { background: linear-gradient(135deg, #e0f2fe 0%, #fff 100%); }
.stat-row { display: flex; margin-top: 15px; border-top: 1px solid rgba(0,0,0,0.05); padding-top: 15px; }
.stat-item { flex: 1; text-align: center; }
.stat-item .count { font-weight: 700; color: #1d2129; }
.stat-item .label { font-size: 12px; color: #86909c; }

/* 公告列表 */
.notice-list { list-style: none; padding: 0; margin: 0; font-size: 13px; color: #515767; }
.notice-list li { padding: 8px 0; border-bottom: 1px solid #f2f3f5; }
.notice-list li:last-child { border-bottom: none; }

/* 页脚 */
.site-footer { font-size: 12px; color: #c9cdd4; text-align: center; line-height: 20px; }
.site-footer a { color: #c9cdd4; text-decoration: none; }
.site-footer a:hover { color: #86909c; }
</style>