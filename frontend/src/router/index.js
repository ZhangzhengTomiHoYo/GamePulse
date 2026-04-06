import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import CreatePostView from '../views/CreatePostView.vue'
// 1. 引入详情页
import PostDetailView from '../views/PostDetailView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/post/create', name: 'create-post', component: CreatePostView },
    
    // 2. 添加详情页路由
    // :id 是动态参数，比如 /post/1, /post/100
    { 
      path: '/post/:id', 
      name: 'post-detail', 
      component: PostDetailView 
    }
  ]
})

export default router