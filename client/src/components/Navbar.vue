<template>
  <nav class="sticky top-0 z-50 bg-[#f7f6f2]/90 backdrop-blur border-b border-[#dcd9d5]">
    <div class="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">

      <!-- Logo -->
      <router-link to="/" class="flex items-center gap-2 text-sm font-semibold text-[#28251d] shrink-0">
        <svg width="26" height="26" viewBox="0 0 28 28" fill="none" aria-label="CertChain logo">
          <rect width="28" height="28" rx="6" fill="#01696f"/>
          <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="14" cy="14" r="11" stroke="white" stroke-width="1.5" fill="none" opacity="0.4"/>
        </svg>
        CertChain
      </router-link>

      <!-- Desktop nav -->
      <div class="hidden md:flex items-center gap-1 text-sm">

        <!-- Always visible -->
        <router-link
          to="/verify"
          class="px-3 py-1.5 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
        >
          Verify
        </router-link>

        <!-- Authenticated: university -->
        <template v-if="auth.isAuthenticated && auth.role === 'university'">
          <router-link :to="{ name: 'UniversityDashboard' }">Dashboard</router-link>
          <router-link
            to="/certificates/issue"
            class="px-3 py-1.5 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
          >
            Issue Certificate
          </router-link>
        </template>

        <!-- Authenticated: company -->
        <template v-else-if="auth.isAuthenticated && auth.role === 'company'">
          <router-link
            to="/dashboard/company"
            class="px-3 py-1.5 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
          >
            Dashboard
          </router-link>
          <router-link
            to="/company/search"
            class="px-3 py-1.5 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
          >
            Find Talent
          </router-link>
        </template>

        <!-- Authenticated: student -->
        <template v-else-if="auth.isAuthenticated && auth.role === 'student'">
          <router-link
            to="/student/dashboard"
            class="px-3 py-1.5 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
          >
            My Profile
          </router-link>
          <router-link
            to="/student/certificates"
            class="px-3 py-1.5 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
          >
            My Certificates
          </router-link>
        </template>
      </div>

      <!-- Right-side actions -->
      <div class="hidden md:flex items-center gap-3">
        <!-- Not logged in -->
        <template v-if="!auth.isAuthenticated">
          <router-link
            to="/login"
            class="px-4 py-2 text-sm font-medium rounded-lg border border-[#d4d1ca]
                   text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
          >
            Sign in
          </router-link>
          <router-link
            to="/register"
            class="px-4 py-2 text-sm font-semibold rounded-lg bg-[#01696f] text-white
                   hover:bg-[#0c4e54] transition-colors"
          >
            Register
          </router-link>
        </template>

        <!-- Logged in: user chip + logout -->
        <template v-else>
          <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-[#f3f0ec] border border-[#d4d1ca]">
            <span class="w-6 h-6 rounded-full bg-[#cedcd8] flex items-center justify-center text-xs font-bold text-[#01696f] uppercase select-none">
              {{ userInitial }}
            </span>
            <span class="text-sm font-medium text-[#28251d] max-w-[120px] truncate">{{ auth.user?.name }}</span>
            <span class="text-xs px-1.5 py-0.5 rounded-full font-medium" :class="roleChipClass">
              {{ auth.role }}
            </span>
          </div>

          <button
            @click="handleLogout"
            class="px-3 py-2 text-sm font-medium rounded-lg text-[#7a7974]
                   hover:text-[#a12c7b] hover:bg-[#fdf5fa] transition-colors"
            aria-label="Logout"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
              <polyline points="16 17 21 12 16 7"/>
              <line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
          </button>
        </template>
      </div>

      <!-- Mobile hamburger -->
      <button
        class="md:hidden p-2 rounded-lg text-[#7a7974] hover:text-[#28251d] hover:bg-[#f3f0ec] transition-colors"
        @click="mobileOpen = !mobileOpen"
        :aria-expanded="mobileOpen"
        aria-label="Toggle navigation"
      >
        <svg v-if="!mobileOpen" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="3" y1="6"  x2="21" y2="6"/>
          <line x1="3" y1="12" x2="21" y2="12"/>
          <line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
        <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/>
          <line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>

    <!-- Mobile drawer -->
    <div
      v-if="mobileOpen"
      class="md:hidden border-t border-[#dcd9d5] bg-[#f7f6f2] px-6 py-4 space-y-1"
    >
      <router-link
        to="/verify"
        @click="mobileOpen = false"
        class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d] transition-colors"
      >
        Verify certificate
      </router-link>

      <template v-if="auth.isAuthenticated && auth.role === 'university'">
        <router-link to="/dashboard/university"   @click="mobileOpen = false" class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d]">Dashboard</router-link>
        <router-link to="/certificates/issue" @click="mobileOpen = false" class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d]">Issue Certificate</router-link>
      </template>

      <template v-else-if="auth.isAuthenticated && auth.role === 'company'">
        <router-link to="/dashboard/company" @click="mobileOpen = false" class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d]">Dashboard</router-link>
        <router-link to="/company/search"    @click="mobileOpen = false" class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d]">Find Talent</router-link>
      </template>

      <template v-else-if="auth.isAuthenticated && auth.role === 'student'">
        <router-link to="/student/dashboard"      @click="mobileOpen = false" class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d]">My Profile</router-link>
        <router-link to="/student/certificates"   @click="mobileOpen = false" class="flex items-center py-2.5 text-sm text-[#7a7974] hover:text-[#28251d]">My Certificates</router-link>
      </template>

      <div class="pt-3 border-t border-[#dcd9d5] space-y-2">
        <template v-if="!auth.isAuthenticated">
          <router-link to="/login"    @click="mobileOpen = false" class="block w-full text-center py-2.5 text-sm font-medium rounded-lg border border-[#d4d1ca] text-[#28251d] hover:bg-[#f3f0ec]">Sign in</router-link>
          <router-link to="/register" @click="mobileOpen = false" class="block w-full text-center py-2.5 text-sm font-semibold rounded-lg bg-[#01696f] text-white hover:bg-[#0c4e54]">Register</router-link>
        </template>

        <template v-else>
          <div class="flex items-center gap-2 py-2">
            <span class="w-7 h-7 rounded-full bg-[#cedcd8] flex items-center justify-center text-xs font-bold text-[#01696f] uppercase">
              {{ userInitial }}
            </span>
            <div>
              <p class="text-sm font-medium text-[#28251d]">{{ auth.user?.name }}</p>
              <p class="text-xs text-[#7a7974] capitalize">{{ auth.role }}</p>
            </div>
          </div>
          <button
            @click="handleLogout"
            class="w-full py-2.5 text-sm font-medium rounded-lg text-[#a12c7b]
                   border border-[#e0ced7] hover:bg-[#fdf5fa] transition-colors"
          >
            Sign out
          </button>
        </template>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth   = useAuthStore()
const router = useRouter()

const mobileOpen = ref(false)

const userInitial = computed(() => {
  return auth.user?.name?.charAt(0)?.toUpperCase() || '?'
})

const roleChipClass = computed(() => ({
  'university': 'bg-[#cedcd8] text-[#01696f]',
  'company':    'bg-[#c6d8e4] text-[#006494]',
  'student':    'bg-[#dacfde] text-[#7a39bb]',
}[auth.role] ?? 'bg-[#f3f0ec] text-[#7a7974]'))

const handleLogout = () => {
  auth.logout()
  mobileOpen.value = false
  router.push('/')
}
</script>