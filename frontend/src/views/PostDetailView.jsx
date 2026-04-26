import { useEffect, useMemo, useState } from 'react'
import { ArrowLeft, ChevronDown, ChevronUp, Loader2, Trash2 } from 'lucide-react'
import axios from 'axios'
import { useNavigate, useParams } from 'react-router-dom'
import { deletePost } from '../api/post.js'
import ImagePreview from '../components/ImagePreview.jsx'
import { useToast } from '../components/ToastProvider.jsx'
import { getAuthHeaders, normalizeImageUrls } from '../utils/posts.js'

export default function PostDetailView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const toast = useToast()
  const [post, setPost] = useState(null)
  const [deleting, setDeleting] = useState(false)
  const [preview, setPreview] = useState(null)
  const currentUserId = localStorage.getItem('userid')

  const canDeletePost = useMemo(() => {
    if (!post || !currentUserId) return false
    return String(post.author_id) === String(currentUserId)
  }, [currentUserId, post])

  const fetchPost = async () => {
    try {
      const res = await axios.get(`/api/v1/post/${id}`, {
        headers: getAuthHeaders()
      })

      if (res.data.code === 1000) {
        const nextPost = res.data.data
        if (nextPost.vote_status === undefined) nextPost.vote_status = 0
        nextPost.votes = Number(nextPost.votes)
        normalizeImageUrls(nextPost)
        setPost(nextPost)
      } else {
        toast.error(res.data.msg)
      }
    } catch (error) {
      console.error(error)
      toast.error('获取帖子详情失败')
    }
  }

  const handleVote = async (direction) => {
    if (!post) return

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

      setPost((current) => {
        let nextVotes = Number(current.votes || 0)
        if (current.vote_status === 1) nextVotes -= 1
        if (dirToSend === 1) nextVotes += 1

        return {
          ...current,
          vote_status: dirToSend,
          votes: nextVotes
        }
      })
      toast.success(dirToSend === 0 ? '已取消' : '操作成功')
    } catch (error) {
      console.error(error)
      toast.error('投票失败')
    }
  }

  const confirmDelete = async () => {
    if (!post || deleting) return
    if (!window.confirm('确定要删除这篇帖子吗？该操作不可恢复。')) return

    setDeleting(true)
    try {
      const res = await deletePost(post.id)

      if (res.data.code === 1000) {
        toast.success('帖子删除成功')
        navigate('/', { replace: true })
        return
      }

      toast.error(res.data.msg || '删除失败')
    } catch (error) {
      console.error(error)
      toast.error(error?.response?.data?.msg || '删除失败，请稍后重试')
    } finally {
      setDeleting(false)
    }
  }

  useEffect(() => {
    fetchPost()
  }, [id])

  return (
    <main className="detail-container">
      <button className="ghost-action detail-back" type="button" onClick={() => navigate(-1)}>
        <ArrowLeft size={18} />
        <span>返回首页</span>
      </button>

      {post ? (
        <article className="post-detail-card">
          <header className="detail-header">
            <div className="detail-header-top">
              <h1 className="title">{post.title}</h1>

              {canDeletePost && (
                <button className="danger-action" type="button" disabled={deleting} onClick={confirmDelete}>
                  {deleting ? <Loader2 className="spin" size={18} /> : <Trash2 size={18} />}
                  <span>{deleting ? '删除中...' : '删除'}</span>
                </button>
              )}
            </div>

            <div className="meta">
              <span className="meta-tag">作者：{post.author_name || post.author_id}</span>
              <span className="time">{new Date(post.create_time).toLocaleString()}</span>
            </div>
          </header>

          <div className="content">{post.content}</div>

          {post.image_urls?.length > 0 && (
            <div className="detail-images">
              {post.image_urls.map((url, index) => (
                <button
                  className="detail-image-item"
                  type="button"
                  key={url}
                  onClick={() => setPreview({ images: post.image_urls, index })}
                >
                  <img src={url} alt={`${post.title} 图片 ${index + 1}`} />
                </button>
              ))}
            </div>
          )}

          <footer className="detail-footer">
            <div className="vote-actions">
              <button
                className={`round-action ${post.vote_status === 1 ? 'is-active' : ''}`}
                type="button"
                aria-label="赞"
                onClick={() => handleVote(1)}
              >
                <ChevronUp size={20} />
              </button>

              <span className="vote-count">{post.votes || 0} 热度</span>

              <button
                className={`round-action ${post.vote_status === -1 ? 'is-active' : ''}`}
                type="button"
                aria-label="踩"
                onClick={() => handleVote(-1)}
              >
                <ChevronDown size={20} />
              </button>
            </div>
          </footer>
        </article>
      ) : (
        <div className="detail-skeleton">
          <Loader2 className="spin" size={24} />
          <span>正在加载帖子详情...</span>
        </div>
      )}

      {preview && (
        <ImagePreview
          images={preview.images}
          index={preview.index}
          onIndexChange={(index) => setPreview((current) => ({ ...current, index }))}
          onClose={() => setPreview(null)}
        />
      )}
    </main>
  )
}
