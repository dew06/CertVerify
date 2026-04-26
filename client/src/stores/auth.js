import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token'))
  const university = ref(JSON.parse(localStorage.getItem('university') || 'null'))
  
  const isAuthenticated = computed(() => !!token.value)
  
  async function login(credentials) {
    const { data } = await api.login(credentials)
    token.value = data.token
    university.value = data.university
    localStorage.setItem('token', data.token)
    localStorage.setItem('university', JSON.stringify(data.university))
    return data
  }
  
  async function register(formData) {
    const { data } = await api.register(formData)
    return data
  }
  
  function logout() {
    token.value = null
    university.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('university')
  }
  
  return {
    token,
    university,
    isAuthenticated,
    login,
    register,
    logout
  }
})
