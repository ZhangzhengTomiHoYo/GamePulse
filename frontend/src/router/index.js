import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import CreatePostView from '../views/CreatePostView.vue'
import PostDetailView from '../views/PostDetailView.vue'
import CommunityAgentView from '../views/CommunityAgentView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/post/create', name: 'create-post', component: CreatePostView },
    { path: '/agent/community', name: 'community-agent', component: CommunityAgentView },
    { 
      path: '/post/:id', 
      name: 'post-detail', 
      component: PostDetailView 
    }
  ]
})

export default router
