import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/Home.vue')
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../views/Register.vue')
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true }
    },
   
    {
      path: '/bulk-upload',
      name: 'BulkUpload',
      component: () => import('../views/BulkUpload.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/batch-anchor',
      name: 'BatchAnchor',
      component: () => import('../views/BatchAnchor.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/verify',
      name: 'VerifyCertificate',
      component: () => import('../views/VerifyCertificate.vue')
    },
      {
        path: '/verify/:certID',
        name: 'VerifyCertificateWithID',
        component: () => import('../views/VerifyCertificate.vue')
      },
    {
      path: '/verify-domain',
      name: 'VerifyDomain',
      component: () => import('../views/VerifyDomain.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

export default router
