import axios from 'axios'
import router from '../router'

const api = axios.create({
  baseURL: '/api'
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use((response) => response, (error) => {
  if (error?.response?.status === 401) {
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')

    const current = router.currentRoute.value
    if (current?.name !== 'Login') {
      router.replace({ name: 'Login', query: { redirect: current?.fullPath || '/' } })
    }
  }
  return Promise.reject(error)
})

export { api }
