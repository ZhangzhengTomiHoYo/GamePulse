<template>
  <div class="detail-container">
    <el-button @click="router.back()" style="margin-bottom: 20px">← 返回首页</el-button>
    
    <el-card v-if="post" class="post-detail-card">
      <template #header>
        <div class="detail-header">
          <h1 class="title">{{ post.title }}</h1>
          <div class="meta">
            <el-tag>作者: {{ post.author_name || post.author_id }}</el-tag>
            <span class="time">{{ new Date(post.create_time).toLocaleString() }}</span>
          </div>
        </div>
      </template>

      <div class="content">
        {{ post.content }}
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
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { CaretTop, CaretBottom } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const post = ref(null)

const handleVote = async (direction) => {
  const token = localStorage.getItem('token')
  const currentPost = post.value
  
  let dirToSend = direction
  if (currentPost.vote_status === direction) {
    dirToSend = 0
  }

  try {
    const res = await axios.post('/api/v1/vote', {
      post_id: String(currentPost.id),
      direction: String(dirToSend)
    }, {
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (res.data.code === 1000) {
      ElMessage.success(dirToSend === 0 ? "已取消" : "操作成功")
      
      // 乐观更新
      if (currentPost.vote_status === 1) currentPost.votes -= 1
      currentPost.vote_status = dirToSend
      if (dirToSend === 1) currentPost.votes += 1

    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (error) {
    console.error(error)
  }
}

onMounted(async () => {
  const postId = route.params.id
  try {
    const token = localStorage.getItem('token')
    const res = await axios.get(`/api/v1/post/${postId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (res.data.code === 1000) {
      post.value = res.data.data
      // 初始化
      if (post.value.vote_status === undefined) post.value.vote_status = 0
      // 确保是数字
      post.value.votes = Number(post.value.votes)
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (error) {
    console.error(error)
  }
})
</script>

<style scoped>
/* 样式与之前保持一致 */
.detail-container { max-width: 800px; margin: 40px auto; padding: 0 20px; }
.detail-header { text-align: center; margin-bottom: 20px; }
.title { font-size: 24px; margin-bottom: 10px; color: #303133; }
.meta { display: flex; justify-content: center; align-items: center; gap: 15px; color: #909399; font-size: 13px; }
.content { font-size: 16px; line-height: 1.8; color: #333; white-space: pre-wrap; min-height: 200px; }

.detail-footer {
  margin-top: 40px;
  border-top: 1px solid #eee;
  padding-top: 20px;
  display: flex;
  justify-content: center;
}
.vote-actions { display: flex; align-items: center; gap: 20px; }
.vote-count { font-size: 18px; font-weight: bold; color: #606266; }
</style>  