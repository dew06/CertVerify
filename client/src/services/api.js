import axios from 'axios'
import { jwtDecode } from 'jwt-decode'

const API_URL = import.meta.env.VITE_API_URL

const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' }
})

// Request interceptor
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('role')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default {

  get(url, config)        { return api.get(url, config) },
  post(url, data, config) { return api.post(url, data, config) },
  put(url, data, config)  { return api.put(url, data, config) },
  delete(url, config)     { return api.delete(url, config) },

  // ─── Auth ────────────────────────────────────────────────────────────────

  login(credentials)  { return api.post('/auth/login', credentials) },

  // Role-scoped logout — each hits its own protected endpoint so the
  // server can increment token_version and invalidate the session
  logout(role) {
    const endpoints = {
      university: '/auth/logout',
      company:    '/company/logout',
      student:    '/student/logout',
    }
    return api.post(endpoints[role] ?? '/auth/logout')
  },

  getCurrentUser()    { return api.get('/auth/me') },

  // Token refresh — keeps the session alive with a fresh token
  refreshToken()        { return api.post('/auth/refresh') },
  refreshCompanyToken() { return api.post('/company/refresh-token') },
  refreshStudentToken() { return api.post('/student/refresh-token') },

  // ─── University ──────────────────────────────────────────────────────────

  register(data)              { return api.post('/university/register', data) },
  getDomainProof(id)          { return api.get(`/university/${id}/domain-proof`) },
  verifyDomain(id)            { return api.post(`/university/${id}/verify-domain`) },

  // ─── Certificates ────────────────────────────────────────────────────────

  issueCertificate(data)      { return api.post('/certificates/issue', data) },
  getCertificate(certID)      { return api.get(`/certificates/${certID}`) },
  downloadCertificate(certID) {
    return api.get(`/certificates/${certID}/download`, { responseType: 'blob' })
  },
  uploadBulkCSV(formData) {
    return api.post('/certificates/batch-csv', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  batchAnchor(data)           { return api.post('/certificates/batch-anchor', data) },

  // ─── Verify ──────────────────────────────────────────────────────────────

  verifyCertificate(data)     { return api.post('/verify', data) },
  verifyByID(certID)          { return api.get(`/verify/certificate/${certID}`) },
  verifyPDF(formData) {
    return api.post('/verify/pdf', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  getBlockchainInfo(certID)   { return api.get(`/verify/blockchain/${certID}`) },
  getQRData(certID)           { return api.get(`/verify/qr/${certID}`) },

  // ─── Student ─────────────────────────────────────────────────────────────

  registerStudent(data)       { return api.post('/student/register', data) },

  // Profile
  getStudentMe()              { return api.get('/student/me') },
  updateStudentProfile(data)  { return api.put('/student/profile', data) },

  // Skills
  addSkill(data)              { return api.post('/student/skills', data) },
  deleteSkill(skillID)        { return api.delete(`/student/skills/${skillID}`) },

  // Education
  addEducation(data)          { return api.post('/student/education', data) },
  deleteEducation(eduID)      { return api.delete(`/student/education/${eduID}`) },

  // Profile access requests (student side)
  getStudentRequests()        { return api.get('/student/profile-requests') },
  respondToRequest(id, data)  { return api.post(`/student/profile-requests/${id}/respond`, data) },

  // Certificate
  uploadCertificate(formData) {
    return api.post('/student/upload-certificate', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // ─── Company ─────────────────────────────────────────────────────────────

  registerCompany(data)       { return api.post('/company/register', data) },

  // Profile
  getCompanyMe()              { return api.get('/company/me') },
  updateCompanyProfile(data)  { return api.put('/company/profile', data) },

  // Search
  searchStudents(filters)     { return api.post('/company/search', filters) },

  // Profile requests (company side)
  requestProfile(data)        { return api.post('/company/request-profile', data) },
  getMyRequests()             { return api.get('/company/my-requests') },
  getAcceptedProfiles()       { return api.get('/company/accepted-profiles') },
}
