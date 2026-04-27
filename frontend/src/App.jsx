import { Route, Routes } from 'react-router-dom'
import HomeView from './views/HomeView.jsx'
import LoginView from './views/LoginView.jsx'
import CreatePostView from './views/CreatePostView.jsx'
import PostDetailView from './views/PostDetailView.jsx'
import CommunityAgentView from './views/CommunityAgentView.jsx'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomeView />} />
      <Route path="/login" element={<LoginView />} />
      <Route path="/signup" element={<LoginView initialMode="signup" />} />
      <Route path="/post/create" element={<CreatePostView />} />
      <Route path="/agent/community" element={<CommunityAgentView />} />
      <Route path="/post/:id" element={<PostDetailView />} />
    </Routes>
  )
}
