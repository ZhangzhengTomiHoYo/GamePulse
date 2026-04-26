import { useEffect, useState } from 'react'
import {
  ChevronDown,
  ChevronUp,
  Compass,
  Edit3,
  LogOut,
  MessageCircle,
  Send,
  Share2,
  Sparkles,
  User
} from 'lucide-react'
import axios from 'axios'
import { useNavigate } from 'react-router-dom'
import ImagePreview from '../components/ImagePreview.jsx'
import { useToast } from '../components/ToastProvider.jsx'
import logoImg from '../assets/gamepulse.png'
import {
  formatPostTime,
  getAuthHeaders,
  getSentimentLabelText,
  getSentimentTagType,
  normalizeImageUrls
} from '../utils/posts.js'

export default function HomeView() {
  const navigate = useNavigate()
  const toast = useToast()
  const [username] = useState(localStorage.getItem('username') || '用户')
  const [posts, setPosts] = useState([])
  const [communityList, setCommunityList] = useState([])
  const [loading, setLoading] = useState(false)
  const [sortBy, setSortBy] = useState('score')
  const [currentCommunityId, setCurrentCommunityId] = useState(0)
  const [menuOpen, setMenuOpen] = useState(false)
  const [preview, setPreview] = useState(null)

  const fetchCommunityList = async () => {
    try {
      const res = await axios.get('/api/v1/community', {
        headers: getAuthHeaders()
      })

      if (res.data.code === 1000) {
        setCommunityList(res.data.data || [])
      }
    } catch (error) {
      console.error(error)
    }
  }

  const fetchPosts = async (nextCommunityId = currentCommunityId, nextSortBy = sortBy) => {
    if (!localStorage.getItem('token')) return

    setLoading(true)

    try {
      const params = { page: 1, size: 20, order: nextSortBy }
      if (nextCommunityId !== 0) params.community_id = nextCommunityId

      const res = await axios.get('/api/v1/posts2', {
        params,
        headers: getAuthHeaders()
      })

      if (res.data.code === 1000) {
        const list = res.data.data || []
        list.forEach((item) => {
          if (item.vote_status === undefined) item.vote_status = 0
          item.votes = Number(item.votes)
          normalizeImageUrls(item)
        })
        setPosts(list)
      }
    } catch (error) {
      console.error(error)
      toast.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCommunityChange = (id) => {
    setCurrentCommunityId(id)
    fetchPosts(id, sortBy)
  }

  const handleSortChange = (value) => {
    setSortBy(value)
    fetchPosts(currentCommunityId, value)
  }

  const handleVote = async (post, direction) => {
    const dirToSend = post.vote_status === direction ? 0 : direction

    try {
      const res = await axios.post(
        '/api/v1/vote',
        {
          post_id: String(post.id),
          direction: String(dirToSend)
        },
        { headers: getAuthHeaders() }
      )

      if (res.data.code !== 1000) {
        toast.error(res.data.msg)
        return
      }

      setPosts((items) =>
        items.map((item) => {
          if (item.id !== post.id) return item

          let nextVotes = Number(item.votes || 0)
          if (item.vote_status === 1) nextVotes -= 1
          if (dirToSend === 1) nextVotes += 1

          return {
            ...item,
            vote_status: dirToSend,
            votes: nextVotes
          }
        })
      )
    } catch (error) {
      console.error(error)
      toast.error('投票失败')
    }
  }

  const handleCommand = (command) => {
    setMenuOpen(false)
    if (command === 'logout') {
      localStorage.clear()
      toast.success('已安全退出')
      navigate('/login')
    }
  }

  useEffect(() => {
    if (!localStorage.getItem('token')) {
      navigate('/login')
      return
    }

    fetchCommunityList()
    fetchPosts()
  }, [])

  return (
    <div className="common-layout">
      <div className="bg-glow-top-left" />
      <div className="bg-glow-top-right" />
      <div className="bg-glow-bottom" />
      <header className="app-header">
        <div className="header-inner">
          <button className="logo-area" type="button" onClick={() => navigate('/')}>
            <img src={logoImg} alt="GamePulse" className="logo-image" />
            <span className="logo-text">GamePulse</span>
          </button>

          <div className="user-area">
            <span className="welcome-text">Hi, {username}</span>
            <button className="primary-action write-btn" type="button" onClick={() => navigate('/post/create')}>
              <Edit3 size={17} />
              <span>写文章</span>
            </button>

            <div className="profile-menu">
              <button
                className="user-avatar"
                type="button"
                aria-label="打开用户菜单"
                onClick={() => setMenuOpen((value) => !value)}
              >
                {username.charAt(0).toUpperCase()}
              </button>

              {menuOpen && (
                <div className="profile-dropdown">
                  <button type="button" onClick={() => handleCommand('profile')}>
                    <User size={16} />
                    <span>个人中心</span>
                  </button>
                  <button className="danger-item" type="button" onClick={() => handleCommand('logout')}>
                    <LogOut size={16} />
                    <span>退出登录</span>
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="main-container">
        <aside className="sidebar-left">
          <nav className="community-menu">
            <div className="menu-title">发现</div>
            <button
              className={`menu-item ${currentCommunityId === 0 ? 'is-active' : ''}`}
              type="button"
              onClick={() => handleCommunityChange(0)}
            >
              <Compass size={18} />
              <span>综合</span>
            </button>
            <button className="menu-item" type="button" onClick={() => navigate('/agent/community')}>
              <Sparkles size={18} />
              <span>游小脉社区智能体</span>
            </button>

            <div className="menu-title menu-title-spaced">社区</div>
            {communityList.map((item) => (
              <button
                className={`menu-item ${currentCommunityId === item.id ? 'is-active' : ''}`}
                type="button"
                key={item.id}
                onClick={() => handleCommunityChange(item.id)}
              >
                <MessageCircle size={18} />
                <span>{item.name}</span>
              </button>
            ))}
          </nav>
        </aside>

        <section className="content-middle">
          <div className="feed-tabs">
            <button
              className={`feed-tab ${sortBy === 'score' ? 'active is-active' : ''}`}
              type="button"
              onClick={() => handleSortChange('score')}
            >
              热门
            </button>
            <button
              className={`feed-tab ${sortBy === 'time' ? 'active is-active' : ''}`}
              type="button"
              onClick={() => handleSortChange('time')}
            >
              最新
            </button>
          </div>

          <div className={`post-list ${loading ? 'is-loading' : ''}`}>
            {loading && (
              <div className="loading-block">
                <div className="loading-spinner" />
                <span>正在加载帖子...</span>
              </div>
            )}

            {!loading &&
              posts.map((post) => (
                <article className="post-item" key={post.id} onClick={() => navigate(`/post/${post.id}`)}>
                  <div className="post-meta-top">
                    <span className="author-name">{post.author_name || '匿名用户'}</span>
                    <span className="divider">·</span>
                    <span className="post-time">{formatPostTime(post.create_time)}</span>
                    <span className="divider">·</span>
                    <span className="community-tag">{post.community?.name || '综合'}</span>
                    {post.sentiment_label && (
                      <span className={`sentiment-tag sentiment-${getSentimentTagType(post.sentiment_label)}`}>
                        {getSentimentLabelText(post.sentiment_label)}
                      </span>
                    )}
                  </div>

                  <h2 className="post-title">{post.title}</h2>
                  <p className="post-abstract">{post.content}</p>

                  {post.image_urls?.length > 0 && (
                    <div className="post-images" onClick={(event) => event.stopPropagation()}>
                      {post.image_urls.map((url, index) => (
                        <button
                          className="image-thumb"
                          type="button"
                          key={`${post.id}-${url}`}
                          onClick={() => setPreview({ images: post.image_urls, index })}
                        >
                          <img src={url} alt={`${post.title} 图片 ${index + 1}`} />
                        </button>
                      ))}
                    </div>
                  )}

                  <div className="post-actions">
                    <div className="vote-group" onClick={(event) => event.stopPropagation()}>
                      <button
                        className={`vote-btn up ${post.vote_status === 1 ? 'active' : ''}`}
                        type="button"
                        aria-label="赞"
                        onClick={() => handleVote(post, 1)}
                      >
                        <ChevronUp size={17} />
                        <span>{post.votes !== 0 ? post.votes : '赞'}</span>
                      </button>
                      <button
                        className={`vote-btn down ${post.vote_status === -1 ? 'active' : ''}`}
                        type="button"
                        aria-label="踩"
                        onClick={() => handleVote(post, -1)}
                      >
                        <ChevronDown size={17} />
                      </button>
                    </div>

                    <button className="action-group" type="button" onClick={(event) => event.stopPropagation()}>
                      <MessageCircle size={17} />
                      <span>评论</span>
                    </button>
                    <button className="action-group" type="button" onClick={(event) => event.stopPropagation()}>
                      <Share2 size={17} />
                      <span>分享</span>
                    </button>
                  </div>
                </article>
              ))}

            {!loading && posts.length === 0 && (
              <div className="empty-glass-card">
                <Sparkles size={48} />
                <h3>尚未产生动态</h3>
                <p>还没有社区动态，发布第一篇帖子，开始收集玩家情绪。</p>
              </div>
            )}
          </div>
        </section>

        <aside className="sidebar-right">
          <section className="widget-card welcome-widget">
            <div className="widget-header">
              <h3>GamePulse 社区</h3>
            </div>
            <div className="widget-content">
              <p>专注于游戏社区舆情分析</p>
              <div className="stat-row">
                <div className="stat-item">
                  <div className="count">100+</div>
                  <div className="label">帖子</div>
                </div>
                <div className="stat-item">
                  <div className="count">50+</div>
                  <div className="label">用户</div>
                </div>
              </div>
            </div>
          </section>

          <section className="widget-card">
            <div className="card-header">
              <span>社区公告</span>
            </div>
            <ul className="notice-list">
              <li>GamePulse v1.0 正式上线</li>
              <li>关于社区账号安全的说明</li>
              <li>创作你喜欢的游戏文章</li>
            </ul>
          </section>

          <footer className="site-footer">
            © 2026 GamePulse Blog
            <br />
            <a href="#">关于我们</a> · <a href="#">联系作者</a>
          </footer>
        </aside>
      </main>

      <button className="backtop icon-only" type="button" aria-label="回到顶部" onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}>
        <ChevronUp size={20} />
      </button>

      {preview && (
        <ImagePreview
          images={preview.images}
          index={preview.index}
          onIndexChange={(index) => setPreview((current) => ({ ...current, index }))}
          onClose={() => setPreview(null)}
        />
      )}
    </div>
  )
}
