<template>
  <div class="create-post-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>📝 发布新帖子</span>
        </div>
      </template>

      <el-form label-position="top" size="large">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="请输入标题" />
        </el-form-item>

        <el-form-item label="内容">
          <el-input 
            v-model="form.content" 
            type="textarea" 
            :rows="15" 
            placeholder="请输入帖子内容" 
          />
        </el-form-item>

        <el-form-item label="选择社区">
          <el-select v-model="form.community_id" placeholder="请选择发帖社区" style="width: 100%">
            <el-option
              v-for="item in communityList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>

        <div class="btn-group">
          <el-button @click="router.back()">取消</el-button>
          <el-button type="primary" @click="handleSubmit">立即发布</el-button>
        </div>

      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const communityList = ref([])

// 表单数据
const form = reactive({
  title: '',
  content: '',
  community_id: null // 注意初始化为 null
})

// 获取社区列表 (保持不变)
const getCommunityList = async () => {
  try {
    const token = localStorage.getItem('token')
    const res = await axios.get('/api/v1/community', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (res.data.code === 1000) {
      communityList.value = res.data.data
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (error) {
    console.error(error)
  }
}

onMounted(() => {
  getCommunityList()
})

// --- 核心修改：提交给后端 ---
const handleSubmit = async () => {
  if (!form.title || !form.content || !form.community_id) {
    ElMessage.warning("请填写完整信息")
    return
  }

  try {
    const token = localStorage.getItem('token')
    // 发送 POST 请求
    const res = await axios.post('/api/v1/post', form, {
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (res.data.code === 1000) {
      ElMessage.success("发布成功！")
      // 发布成功后跳回首页
      router.push('/')
    } else {
      ElMessage.error("发布失败：" + res.data.msg)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error("网络请求失败")
  }
}
</script>

<style scoped>
.create-post-container {
  max-width: 800px;
  margin: 40px auto;
  padding: 0 20px;
}
.card-header { font-weight: bold; }
.btn-group { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
</style>