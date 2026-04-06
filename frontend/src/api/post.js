import axios from 'axios'

const getAuthHeaders = () => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
})

export const createPost = (payload) => {
  return axios.post('/api/v1/post', payload, {
    headers: getAuthHeaders()
  })
}

export const uploadImage = (file) => {
  const formData = new FormData()
  formData.append('image', file)

  return axios.post('/api/v1/upload', formData, {
    // 只需要带上 Token 即可，千万别手动写 Content-Type
    headers: getAuthHeaders() 
  })
}

export const deletePost = (id) => {
  return axios.delete(`/api/v1/post/${id}`, {
    headers: getAuthHeaders()
  })
}
