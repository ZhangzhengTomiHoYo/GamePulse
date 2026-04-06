<template>
  <div class="login-container">
    <div class="card">
      <div class="header">
        <h2>GamePulse社区</h2>
        <p>{{ isLogin ? '账号登录' : '注册新账号' }}</p>
      </div>

      <el-form label-position="top" size="large">
        
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>

        <el-form-item label="密码">
          <el-input 
            v-model="form.password" 
            type="password" 
            placeholder="请输入密码" 
            show-password 
          />
        </el-form-item>

        <el-form-item label="确认密码" v-if="!isLogin">
          <el-input 
            v-model="form.re_password" 
            type="password" 
            placeholder="请再次输入密码" 
            show-password 
          />
        </el-form-item>

        <el-button type="primary" class="submit-btn" @click="handleSubmit">
          {{ isLogin ? '立 即 登 录' : '注 册' }}
        </el-button>

        <div class="footer-links">
          <el-link type="primary" @click="toggleType">
            {{ isLogin ? '没有账号？去注册' : '已有账号？去登录' }}
          </el-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const isLogin = ref(true) // true=登录模式, false=注册模式

// 表单数据 (对应 Go 后端的 json tag)
const form = reactive({
  username: '',
  password: '',
  re_password: ''
})

// 切换登录/注册
const toggleType = () => {
  isLogin.value = !isLogin.value
  // 切换时清空表单，防止数据残留
  form.username = ''
  form.password = ''
  form.re_password = ''
}

// 核心提交逻辑
const handleSubmit = async () => {
  // 1. 前端基础非空校验
  if(!form.username || !form.password) {
    ElMessage.warning("用户名和密码不能为空")
    return
  }
  
  // 2. 注册时的密码一致性校验
  if (!isLogin.value && form.password !== form.re_password) {
    ElMessage.error("两次输入的密码不一致")
    return
  }

  try {
    // 3. 确定请求地址
    // /api 开头会触发 vite.config.js 的代理转发，连到 localhost:8080
    // 修改后 (加上 /v1)
    const url = isLogin.value ? '/api/v1/login' : '/api/v1/signup'
    
    // 4. 发送请求给 Go 后端
    const res = await axios.post(url, form)

    // 5. 打印结果方便调试
    console.log("后端返回:", res.data)

    // --- 重点修复：根据后端 Code 判断成败 ---
    // 你的后端定义：CodeSuccess = 1000
    if (res.data.code === 1000) {
      
      // A. 成功的情况
      if (isLogin.value) {
        ElMessage.success("登录成功！")
        
        // 1. 提取后端返回的数据对象
        // 后端返回的是: { user_id: 1, user_name: "xx", token: "..." }
        const loginData = res.data.data 

        // 2. 分别保存到 localStorage
        localStorage.setItem('token', loginData.token)
        localStorage.setItem('username', loginData.user_name) // 存名字，首页要显示
        localStorage.setItem('userid', loginData.user_id)     // 存ID，以备后用
        
        // 跳转到首页
        router.push('/')
      } else {
        ElMessage.success("注册成功，请登录")
        // 注册成功后，自动切回登录界面
        isLogin.value = true 
      }

    } else {
      // B. 失败的情况 (Code 不是 1000)
      // 比如：1004(密码错误), 1003(用户不存在)
      // 直接显示后端返回的 msg 错误提示
      ElMessage.error("失败：" + res.data.msg)
      
      // 注意：这里没有 router.push，所以用户会停留在登录页重试
    }

  } catch (error) {
    // C. 网络层面彻底挂了 (如 404, 500, 后端没启动)
    console.error(error)
    ElMessage.error("网络连接失败，请检查后端服务")
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f7fa;
}
.card {
  width: 380px;
  padding: 40px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.header {
  text-align: center;
  margin-bottom: 30px;
}
.header h2 { margin: 0; color: #409EFF; }
.header p { margin: 10px 0 0; color: #999; font-size: 14px; }
.submit-btn { width: 100%; margin-top: 10px; font-weight: bold; }
.footer-links { text-align: center; margin-top: 20px; }
</style>