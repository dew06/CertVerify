import axios from 'axios'
import { jwtDecode } from 'jwt-decode'

const API_URL = import.meta.env.VITE_API_URL

const api = axios.create({
  baseURL: API_URL, 
  headers: {
    'Content-Type': 'application/json'
  }
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
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default {

  get(url, config) {
    return api.get(url, config)
  },
  post(url, data, config) {
    return api.post(url, data, config)
  },

  // Auth
  login(credentials) {
    return api.post('/auth/login', credentials)
  },
  
  register(data) {
    return api.post('/university/register', data)
  },
  
  getCurrentUser() {
    return api.get('/auth/me')
  },
  
  // Certificates
  issueCertificate(data) {
    return api.post('/certificates/issue', data)
  },
  
  uploadBulkCSV(formData) {
    return api.post('/certificates/batch-csv', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  
  batchAnchor(data) {
    return api.post('/certificates/batch-anchor', data)
  },
  
  getCertificate(certID) {
    return api.get(`/certificates/${certID}`)
  },
  
  getDomainProof(id) {
    return api.get(`/university/${id}/domain-proof`)
  },

  verifyDomain(id) {
    return api.post(`/university/${id}/verify-domain`)
  },

  verifyCertificate(data) {
    return api.post('/verify', data)
  },
  verifyByID(certID) {
    return api.get(`/verify/certificate/${certID}`)
  },
  
  verifyPDF(formData) {
    return api.post('/verify/pdf', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  
  getBlockchainInfo(certID) {
    return api.get(`/verify/blockchain/${certID}`)
  },
  
  getQRData(certID) {
    return api.get(`/verify/qr/${certID}`)
  },
  
  downloadCertificate(certID) {
    return api.get(`/certificates/${certID}/download`, {
      responseType: 'blob'
    })
  }
}