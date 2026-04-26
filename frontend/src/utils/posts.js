export const getAuthHeaders = () => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
})

export const normalizeImageUrls = (post) => {
  if (!post) return []

  const rawValue = post.image_urls ?? post.image_url ?? post.imageURLs

  if (Array.isArray(rawValue)) {
    post.image_urls = rawValue.filter(Boolean)
    return post.image_urls
  }

  if (typeof rawValue === 'string' && rawValue.trim()) {
    try {
      const parsed = JSON.parse(rawValue)
      post.image_urls = Array.isArray(parsed) ? parsed.filter(Boolean) : []
      return post.image_urls
    } catch (error) {
      console.warn('Failed to parse post image urls:', rawValue, error)
    }
  }

  post.image_urls = []
  return post.image_urls
}

export const formatPostTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit'
  })}`
}

const sentimentDisplayMap = {
  positive: { text: '正向', type: 'success' },
  neutral: { text: '中性', type: 'info' },
  negative: { text: '负向', type: 'danger' }
}

export const normalizeSentimentLabel = (label) => {
  if (typeof label !== 'string') return ''
  return label.trim().toLowerCase()
}

export const getSentimentLabelText = (label) =>
  sentimentDisplayMap[normalizeSentimentLabel(label)]?.text || ''

export const getSentimentTagType = (label) =>
  sentimentDisplayMap[normalizeSentimentLabel(label)]?.type || 'info'
