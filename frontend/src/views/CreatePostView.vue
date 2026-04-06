<template>
  <div class="create-post-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>发布新帖子</span>
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

        <el-form-item label="图片">
          <div class="upload-section">
            <div class="upload-meta">
              <span class="upload-tip">最多上传 9 张图片，选择后会立即开始上传。</span>
              <span class="upload-count">{{ fileList.length }}/9</span>
            </div>

            <el-upload
              v-model:file-list="fileList"
              class="upload-wall"
              list-type="picture-card"
              accept="image/*"
              multiple
              :limit="9"
              :http-request="handleImageUpload"
              :before-upload="beforeImageUpload"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              :on-remove="handleRemove"
              :on-exceed="handleExceed"
            >
              <el-icon><Plus /></el-icon>
            </el-upload>
          </div>
        </el-form-item>

        <el-form-item label="选择社区">
          <el-select
            v-model="form.community_id"
            placeholder="请选择发帖社区"
            style="width: 100%"
          >
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
          <el-button
            type="primary"
            :loading="submitting"
            :disabled="isUploading"
            @click="handleSubmit"
          >
            立即发布
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { createPost, uploadImage } from '../api/post'

const router = useRouter()
const communityList = ref([])
const fileList = ref([])
const uploadedImageUrls = ref([])
const submitting = ref(false)

const form = reactive({
  title: '',
  content: '',
  community_id: null
})

const isUploading = computed(() =>
  fileList.value.some((file) => file.status === 'uploading')
)

const getCommunityList = async () => {
  try {
    const token = localStorage.getItem('token')
    const res = await axios.get('/api/v1/community', {
      headers: { Authorization: `Bearer ${token}` }
    })

    if (res.data.code === 1000) {
      communityList.value = res.data.data
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('获取社区列表失败')
  }
}

const beforeImageUpload = (rawFile) => {
  if (!rawFile.type.startsWith('image/')) {
    ElMessage.error('只能上传图片格式文件')
    return false
  }

  return true
}

const handleImageUpload = async (options) => {
  try {
    const res = await uploadImage(options.file)

    if (res.data.code !== 1000 || !res.data.data) {
      throw new Error(res.data.msg || '图片上传失败')
    }

    options.onSuccess({ url: res.data.data }, options.file)
  } catch (error) {
    options.onError(error)
  }
}

const handleUploadSuccess = (response, uploadFile) => {
  const imageUrl = response?.url
  if (!imageUrl) {
    ElMessage.error('图片上传失败')
    return
  }

  uploadFile.url = imageUrl
  uploadFile.imageUrl = imageUrl

  if (!uploadedImageUrls.value.includes(imageUrl)) {
    uploadedImageUrls.value.push(imageUrl)
  }

  ElMessage.success('图片上传成功')
}

const handleUploadError = (error) => {
  ElMessage.error(error?.message || '图片上传失败')
}

const handleRemove = (uploadFile) => {
  const imageUrl = uploadFile.imageUrl || uploadFile.url || uploadFile.response?.url
  if (!imageUrl) return

  uploadedImageUrls.value = uploadedImageUrls.value.filter((url) => url !== imageUrl)
}

const handleExceed = () => {
  ElMessage.warning('最多只能上传 9 张图片')
}

const handleSubmit = async () => {
  if (!form.title || !form.content || !form.community_id) {
    ElMessage.warning('请填写完整信息')
    return
  }

  if (isUploading.value) {
    ElMessage.warning('图片仍在上传中，请稍候再发布')
    return
  }

  submitting.value = true

  try {
    const payload = {
      ...form,
      image_urls: [...uploadedImageUrls.value]
    }

    const res = await createPost(payload)

    if (res.data.code === 1000) {
      ElMessage.success('发布成功')
      router.push('/')
    } else {
      ElMessage.error(`发布失败：${res.data.msg}`)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('网络请求失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  getCommunityList()
})
</script>

<style scoped>
.create-post-container {
  max-width: 800px;
  margin: 40px auto;
  padding: 0 20px;
}

.card-header {
  font-weight: bold;
}

.upload-section {
  width: 100%;
}

.upload-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.upload-tip {
  font-size: 13px;
  color: #909399;
}

.upload-count {
  flex-shrink: 0;
  padding: 2px 10px;
  border-radius: 999px;
  background: #f4f4f5;
  color: #606266;
  font-size: 12px;
  line-height: 20px;
}

.upload-wall {
  width: 100%;
}

.btn-group {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

@media (max-width: 640px) {
  .upload-meta {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
