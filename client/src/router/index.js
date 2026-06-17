import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const roleDashboardName = {
  university: 'UniversityDashboard',
  company: 'CompanyDashboard',
  student: 'StudentDashboard',
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Public
    {
      path: '/',
      name: 'Home',
      component: () => import('../views/Home.vue'),
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../views/Register.vue'),
    },
    {
      path: '/verify',
      name: 'VerifyCertificate',
      component: () => import('../views/VerifyCertificate.vue'),
    },
    {
      path: '/verify/:certID',
      name: 'VerifyCertificateWithID',
      component: () => import('../views/VerifyCertificate.vue'),
    },

    // University
    {
      path: '/university/dashboard',
      name: 'UniversityDashboard',
      component: () => import('../views/UniversityDashboard.vue'),
      meta: { requiresAuth: true, role: 'university' },
    },
    // {
    //   path: '/certificates/issue',
    //   name: 'IssueCertificate',
    //   component: () => import('../views/IssueCertificate.vue'),
    //   meta: { requiresAuth: true, role: 'university' },
    // },
    {
      path: '/bulk-upload',
      name: 'BulkUpload',
      component: () => import('../views/BulkUpload.vue'),
      meta: { requiresAuth: true, role: 'university' },
    },
    {
      path: '/batch-anchor',
      name: 'BatchAnchor',
      component: () => import('../views/BatchAnchor.vue'),
      meta: { requiresAuth: true, role: 'university' },
    },
    {
      path: '/verify-domain',
      name: 'VerifyDomain',
      component: () => import('../views/VerifyDomain.vue'),
      meta: { requiresAuth: true, role: 'university' },
    },

    // Company
    {
      path: '/company/dashboard',
      name: 'CompanyDashboard',
      component: () => import('../views/CompanyDashboard.vue'),
      meta: { requiresAuth: true, role: 'company' },
    },

    // Student
    {
      path: '/student/dashboard',
      name: 'StudentDashboard',
      component: () => import('../views/StudentDashboard.vue'),
      meta: { requiresAuth: true, role: 'student' },
    },

    // Fallback
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()

  // Block unauthenticated access
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next({ name: 'Login' })
  }

  // Block wrong-role access
  if (to.meta.requiresAuth && to.meta.role && to.meta.role !== auth.role) {
    const dashName = roleDashboardName[auth.role]
    return dashName ? next({ name: dashName }) : next({ name: 'Login' })
  }

  // Already logged in → don't allow login/register again
  if (auth.isAuthenticated && (to.name === 'Login' || to.name === 'Register')) {
    const dashName = roleDashboardName[auth.role]
    return dashName ? next({ name: dashName }) : next()
  }

  next()
})

export default router