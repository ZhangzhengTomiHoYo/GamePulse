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
    headers: {
      ...getAuthHeaders(),
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const deletePost = (id) => {
  return axios.delete(`/api/v1/post/${id}`, {
    headers: getAuthHeaders()
  })
}
