import { useEffect, useMemo, useRef, useState } from 'react'
import { ArrowLeft, ImagePlus, Plus, Send, Trash2, UploadCloud, X } from 'lucide-react'
import axios from 'axios'
import { useNavigate } from 'react-router-dom'
import { createPost, uploadImage } from '../api/post.js'
import { useToast } from '../components/ToastProvider.jsx'
import { getAuthHeaders } from '../utils/posts.js'

const emptyForm = {
  title: '',
  content: '',
  community_id: ''
}

const createUploadId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`

export default function CreatePostView() {
  const navigate = useNavigate()
  const toast = useToast()
  const fileInputRef = useRef(null)
  const fileListRef = useRef([])
  const [communityList, setCommunityList] = useState([])
  const [fileList, setFileList] = useState([])
  const [uploadedImageUrls, setUploadedImageUrls] = useState([])
  const [form, setForm] = useState(emptyForm)
  const [submitting, setSubmitting] = useState(false)

  const isUploading = useMemo(() => fileList.some((file) => file.status === 'uploading'), [fileList])

  const updateForm = (field, value) => {
    setForm((current) => ({ ...current, [field]: value }))
  }

  const getCommunityList = async () => {
    try {
      const res = await axios.get('/api/v1/community', {
        headers: getAuthHeaders()
      })

      if (res.data.code === 1000) {
        setCommunityList(res.data.data || [])
      } else {
        toast.error(res.data.msg)
      }
    } catch (error) {
      console.error(error)
      toast.error('获取社区列表失败')
    }
  }

  const uploadOneFile = async (entry) => {
    try {
      const res = await uploadImage(entry.file)
      const responseData = res.data !== undefined ? res.data : res

      if (responseData.code !== 1000 || !responseData.data) {
        throw new Error(responseData.msg || '图片上传失败')
      }

      const imageUrl = responseData.data
      setFileList((items) =>
        items.map((item) =>
          item.id === entry.id
            ? {
                ...item,
                url: imageUrl,
                status: 'success'
              }
            : item
        )
      )
      setUploadedImageUrls((items) => (items.includes(imageUrl) ? items : [...items, imageUrl]))
      toast.success('图片上传成功')
    } catch (error) {
      console.error(error)
      setFileList((items) =>
        items.map((item) =>
          item.id === entry.id
            ? {
                ...item,
                status: 'error',
                error: error?.message || '图片上传失败'
              }
            : item
        )
      )
      toast.error(error?.message || '图片上传失败')
    }
  }

  const handleFileChange = (event) => {
    const files = Array.from(event.target.files || [])
    event.target.value = ''

    if (files.length === 0) return

    const availableSlots = Math.max(9 - fileList.length, 0)
    if (availableSlots === 0) {
      toast.warning('最多只能上传 9 张图片')
      return
    }

    const acceptedFiles = files.slice(0, availableSlots).filter((file) => {
      if (file.type.startsWith('image/')) return true
      toast.error(`${file.name} 不是图片格式`)
      return false
    })

    if (files.length > availableSlots) {
      toast.warning('最多只能上传 9 张图片')
    }

    const entries = acceptedFiles.map((file) => ({
      id: createUploadId(),
      name: file.name,
      file,
      previewUrl: URL.createObjectURL(file),
      url: '',
      status: 'uploading',
      error: ''
    }))

    if (entries.length === 0) return

    setFileList((items) => [...items, ...entries])
    entries.forEach(uploadOneFile)
  }

  const handleRemove = (entry) => {
    if (entry.previewUrl) URL.revokeObjectURL(entry.previewUrl)
    setFileList((items) => items.filter((item) => item.id !== entry.id))
    if (entry.url) {
      setUploadedImageUrls((items) => items.filter((url) => url !== entry.url))
    }
  }

  const handleSubmit = async (event) => {
    event.preventDefault()

    if (!form.title || !form.content || !form.community_id) {
      toast.warning('请填写完整信息')
      return
    }

    if (isUploading) {
      toast.warning('图片仍在上传中，请稍候再发布')
      return
    }

    setSubmitting(true)

    try {
      const payload = {
        ...form,
        community_id: String(form.community_id),
        image_urls: [...uploadedImageUrls]
      }

      const res = await createPost(payload)

      if (res.data.code === 1000) {
        toast.success('发布成功')
        navigate('/')
      } else {
        toast.error(`发布失败：${res.data.msg}`)
      }
    } catch (error) {
      console.error(error)
      toast.error('网络请求失败')
    } finally {
      setSubmitting(false)
    }
  }

  useEffect(() => {
    getCommunityList()
  }, [])

  useEffect(() => {
    fileListRef.current = fileList
  }, [fileList])

  useEffect(
    () => () => {
      fileListRef.current.forEach((file) => {
        if (file.previewUrl) URL.revokeObjectURL(file.previewUrl)
      })
    },
    []
  )

  return (
    <main className="create-post-container">
      <div className="bg-glow-top-left" />
      <div className="bg-glow-top-right" />
      <div className="bg-glow-bottom" />
      <section className="editor-card">
        <div className="editor-header">
          <button className="ghost-action" type="button" onClick={() => navigate(-1)}>
            <ArrowLeft size={18} />
            <span>返回</span>
          </button>
          <h1>发布新帖子</h1>
        </div>

        <form className="post-form" onSubmit={handleSubmit}>
          <label className="field-block">
            <span>标题</span>
            <input
              className="glass-input"
              value={form.title}
              placeholder="请输入标题"
              onChange={(event) => updateForm('title', event.target.value)}
            />
          </label>

          <label className="field-block">
            <span>内容</span>
            <textarea
              className="glass-input"
              value={form.content}
              rows={15}
              placeholder="请输入帖子内容"
              onChange={(event) => updateForm('content', event.target.value)}
            />
          </label>

          <div className="field-block">
            <div className="upload-meta">
              <span>图片</span>
              <span className="upload-count">{fileList.length}/9</span>
            </div>
            <p className="upload-tip">最多上传 9 张图片，选择后会立即开始上传。</p>

            <div className="upload-wall">
              {fileList.map((entry) => (
                <div className={`upload-tile status-${entry.status}`} key={entry.id}>
                  <img src={entry.url || entry.previewUrl} alt={entry.name} />
                  <div className="upload-state">
                    {entry.status === 'uploading' && <UploadCloud size={18} />}
                    {entry.status === 'error' && <X size={18} />}
                  </div>
                  <button
                    className="upload-remove icon-only"
                    type="button"
                    aria-label="移除图片"
                    onClick={() => handleRemove(entry)}
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              ))}

              {fileList.length < 9 && (
                <button
                  className="upload-add"
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <ImagePlus size={24} />
                  <span>上传图片</span>
                </button>
              )}
            </div>

            <input
              ref={fileInputRef}
              className="visually-hidden"
              type="file"
              accept="image/*"
              multiple
              onChange={handleFileChange}
            />
          </div>

          <label className="field-block">
            <span>选择社区</span>
            <select
              className="glass-input"
              value={form.community_id}
              onChange={(event) => updateForm('community_id', event.target.value)}
            >
              <option value="">请选择发帖社区</option>
              {communityList.map((item) => (
                <option value={item.id} key={item.id}>
                  {item.name}
                </option>
              ))}
            </select>
          </label>

          <div className="btn-group">
            <button className="ghost-action" type="button" onClick={() => navigate(-1)}>
              取消
            </button>
            <button className="primary-action primary-gradient-button" type="submit" disabled={submitting || isUploading}>
              {submitting ? <UploadCloud size={18} /> : <Send size={18} />}
              <span>{submitting ? '发布中...' : '立即发布'}</span>
            </button>
          </div>
        </form>
      </section>
    </main>
  )
}
