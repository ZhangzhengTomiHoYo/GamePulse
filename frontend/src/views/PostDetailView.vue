<template>
  <div class="detail-container">
    <el-button @click="router.back()" style="margin-bottom: 20px">返回首页</el-button>

    <el-card v-if="post" class="post-detail-card">
      <template #header>
        <div class="detail-header">
          <div class="detail-header-top">
            <h1 class="title">{{ post.title }}</h1>

            <el-button
              v-if="canDeletePost"
              type="danger"
              plain
              :loading="deleting"
              @click="confirmDelete"
            >
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </div>

          <div class="meta">
            <el-tag>作者: {{ post.author_name || post.author_id }}</el-tag>
            <span class="time">{{ new Date(post.create_time).toLocaleString() }}</span>
          </div>
        </div>
      </template>

      <div class="content">
        {{ post.content }}
      </div>
      <div class="detail-images" v-if="post.image_urls && post.image_urls.length > 0">
        <el-image
          v-for="(url, index) in post.image_urls"
          :key="index"
          :src="url"
          class="detail-image-item"
          fit="scale-down"
          :preview-src-list="post.image_urls"
          :initial-index="index"
          preview-teleported
        />
    </div>
      <div class="detail-footer">
        <div class="vote-actions">
          <el-button
            :type="post.vote_status === 1 ? 'warning' : ''"
            circle
            @click="handleVote(1)"
          >
            <el-icon><CaretTop /></el-icon>
          </el-button>

          <span class="vote-count">{{ post.votes || 0 }} 热度</span>

          <el-button
            :type="post.vote_status === -1 ? 'primary' : ''"
            circle
            @click="handleVote(-1)"
          >
            <el-icon><CaretBottom /></el-icon>
          </el-button>
        </div>
      </div>
    </el-card>

    <el-skeleton v-else :rows="5" animated />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CaretBottom, CaretTop, Delete } from '@element-plus/icons-vue'
import { deletePost } from '../api/post'

const route = useRoute()
const router = useRouter()
const post = ref(null)
const deleting = ref(false)
const currentUserId = localStorage.getItem('userid')

const canDeletePost = computed(() => {
  if (!post.value || !currentUserId) return false
  return String(post.value.author_id) === String(currentUserId)
})

const normalizeImageUrls = (postData) => {
  if (!postData) return

  const rawValue = postData.image_urls ?? postData.image_url ?? postData.imageURLs

  if (Array.isArray(rawValue)) {
    postData.image_urls = rawValue.filter(Boolean)
    return
  }

  if (typeof rawValue === 'string' && rawValue.trim()) {
    try {
      const parsed = JSON.parse(rawValue)
      postData.image_urls = Array.isArray(parsed) ? parsed.filter(Boolean) : []
      return
    } catch (error) {
      console.warn('Failed to parse post image urls:', rawValue, error)
    }
  }

  postData.image_urls = []
}

const handleVote = async (direction) => {
  const token = localStorage.getItem('token')
  const currentPost = post.value

  let dirToSend = direction
  if (currentPost.vote_status === direction) {
    dirToSend = 0
  }

  try {
    const res = await axios.post(
      '/api/v1/vote',
      {
        post_id: String(currentPost.id),
        direction: String(dirToSend)
      },
      {
        headers: { Authorization: `Bearer ${token}` }
      }
    )

    if (res.data.code === 1000) {
      ElMessage.success(dirToSend === 0 ? '已取消' : '操作成功')

      if (currentPost.vote_status === 1) currentPost.votes -= 1
      currentPost.vote_status = dirToSend
      if (dirToSend === 1) currentPost.votes += 1
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('投票失败')
  }
}

const confirmDelete = async () => {
  if (!post.value || deleting.value) return

  try {
    await ElMessageBox.confirm(
      '确定要删除这篇帖子吗？该操作不可恢复。',
      '删除确认',
      {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    return
  }

  deleting.value = true
  try {
    const res = await deletePost(post.value.id)

    if (res.data.code === 1000) {
      ElMessage.success('帖子删除成功')
      await router.replace('/')
      return
    }

    ElMessage.error(res.data.msg || '删除失败')
  } catch (error) {
    console.error(error)
    ElMessage.error(error?.response?.data?.msg || '删除失败，请稍后重试')
  } finally {
    deleting.value = false
  }
}

onMounted(async () => {
  const postId = route.params.id
  try {
    const token = localStorage.getItem('token')
    const res = await axios.get(`/api/v1/post/${postId}`, {
      headers: { Authorization: `Bearer ${token}` }
    })

    if (res.data.code === 1000) {
      post.value = res.data.data
      if (post.value.vote_status === undefined) post.value.vote_status = 0
      post.value.votes = Number(post.value.votes)
      normalizeImageUrls(post.value)
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('获取帖子详情失败')
  }
})
</script>

<style scoped>
.detail-container {
  max-width: 800px;
  margin: 40px auto;
  padding: 0 20px;
}

.detail-header {
  margin-bottom: 20px;
}

.detail-header-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.title {
  margin: 0;
  font-size: 24px;
  color: #303133;
  line-height: 1.4;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 15px;
  color: #909399;
  font-size: 13px;
}

.content {
  min-height: 200px;
  font-size: 16px;
  line-height: 1.8;
  color: #333;
  white-space: pre-wrap;
}

.detail-footer {
  margin-top: 40px;
  border-top: 1px solid #eee;
  padding-top: 20px;
  display: flex;
  justify-content: center;
}

.vote-actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.vote-count {
  font-size: 18px;
  font-weight: bold;
  color: #606266;
}

@media (max-width: 640px) {
  .detail-header-top {
    flex-direction: column;
    align-items: stretch;
  }

  .meta {
    justify-content: flex-start;
  }
}
/* 新增：详情页图片样式 */
.detail-images {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-image-item {
  width: 100%;
  max-height: 600px; /* 防止竖图太长霸屏 */
  border-radius: 8px;
  background-color: #f8f9fa; /* 给透明图片加个极浅的底色 */
}
</style>
