import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { safeParse } from '../utils/safeParse'
import api from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const token      = ref('')
  const role       = ref('')
  const user       = ref(null)
  const university = ref(null)
  const isAuthenticated = computed(() => !!token.value)

  async function login(credentials) {
    const { data } = await api.login(credentials)

    token.value = data.token
    role.value  = data.role

    localStorage.setItem('token', data.token)
    localStorage.setItem('role',  data.role)

    if (data.user) {
      user.value = data.user
      localStorage.setItem('user', JSON.stringify(data.user))
    }

    if (data.role === 'university' && data.user) {
      university.value = data.user
      localStorage.setItem('university', JSON.stringify(data.user))
    }

    return data
  }

  async function register(formData) {
    const { data } = await api.register(formData)
    return data
  }

  function clearLocalState() {
    token.value      = ''
    role.value       = ''
    user.value       = null
    university.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    localStorage.removeItem('user')
    localStorage.removeItem('university')
  }

  async function logout() {
    try {
      await api.logout(role.value) // role.value is 'university' | 'company' | 'student'
    } catch {
      // If the request fails (expired token, network error), still clear state
    } finally {
      clearLocalState()
    }
  }

  return { token, role, user, university, isAuthenticated, login, register, logout }
})
