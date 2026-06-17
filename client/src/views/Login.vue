<template>
  <div class="min-h-screen bg-[#f7f6f2] flex flex-col">

    <!-- Minimal header -->
    <nav class="h-16 border-b border-[#dcd9d5] bg-[#f7f6f2]/90 backdrop-blur flex items-center px-6">
      <router-link to="/" class="flex items-center gap-2 text-sm font-semibold text-[#28251d]">
        <svg width="26" height="26" viewBox="0 0 28 28" fill="none" aria-label="CertChain logo">
          <rect width="28" height="28" rx="6" fill="#01696f"/>
          <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="14" cy="14" r="11" stroke="white" stroke-width="1.5" fill="none" opacity="0.4"/>
        </svg>
        CertChain
      </router-link>
    </nav>

    <!-- Login card — centered -->
    <div class="flex-1 flex items-center justify-center px-4 py-16">
      <div class="w-full max-w-sm">

        <!-- Heading -->
        <div class="mb-8">
          <h1 class="text-2xl font-bold text-[#28251d] tracking-tight">Sign in</h1>
          <p class="mt-1 text-sm text-[#7a7974]">
            Access your university portal or student dashboard.
          </p>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleLogin" novalidate class="space-y-4">

          <!-- Email -->
          <div>
            <label for="email" class="block text-sm font-medium text-[#28251d] mb-1.5">
              Email address
            </label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              :disabled="loading"
              class="w-full px-3 py-2.5 text-sm rounded-lg border transition-colors
                     bg-white text-[#28251d] placeholder-[#bab9b4]
                     border-[#d4d1ca] focus:outline-none focus:ring-2 focus:ring-[#01696f]/30
                     focus:border-[#01696f] disabled:opacity-50 disabled:cursor-not-allowed"
              placeholder="you@university.edu"
            />
          </div>

          <!-- Password -->
          <div>
            <label for="password" class="block text-sm font-medium text-[#28251d] mb-1.5">
              Password
            </label>
            <div class="relative">
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                required
                :disabled="loading"
                class="w-full px-3 py-2.5 pr-10 text-sm rounded-lg border transition-colors
                       bg-white text-[#28251d] placeholder-[#bab9b4]
                       border-[#d4d1ca] focus:outline-none focus:ring-2 focus:ring-[#01696f]/30
                       focus:border-[#01696f] disabled:opacity-50 disabled:cursor-not-allowed"
                placeholder="••••••••"
              />
              <!-- Show / hide toggle -->
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-[#7a7974]
                       hover:text-[#28251d] transition-colors"
                :aria-label="showPassword ? 'Hide password' : 'Show password'"
              >
                <!-- Eye open -->
                <svg v-if="!showPassword" width="16" height="16" viewBox="0 0 24 24"
                     fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                  <circle cx="12" cy="12" r="3"/>
                </svg>
                <!-- Eye off -->
                <svg v-else width="16" height="16" viewBox="0 0 24 24"
                     fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                  <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                  <line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Inline error — replaces alert() -->
          <div
            v-if="errorMessage"
            role="alert"
            class="flex items-start gap-2.5 rounded-lg border border-[#e0ced7] bg-[#fdf5fa]
                   px-4 py-3 text-sm text-[#a12c7b]"
          >
            <svg class="mt-0.5 shrink-0" width="15" height="15" viewBox="0 0 24 24"
                 fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ errorMessage }}
          </div>

          <!-- Submit -->
          <button
            type="submit"
            :disabled="loading || !form.email || !form.password"
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 text-sm
                   font-semibold rounded-lg transition-colors mt-2
                   bg-[#01696f] text-white hover:bg-[#0c4e54]
                   disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg
              v-if="loading"
              class="animate-spin"
              width="16" height="16" viewBox="0 0 24 24"
              fill="none" stroke="currentColor" stroke-width="2"
            >
              <path d="M21 12a9 9 0 1 1-6.219-8.56" stroke-linecap="round"/>
            </svg>
            {{ loading ? 'Signing in…' : 'Sign in' }}
          </button>
        </form>

        <!-- Divider + register link -->
        <div class="mt-6 pt-6 border-t border-[#dcd9d5] text-center">
          <p class="text-sm text-[#7a7974]">
            New university?
            <router-link to="/register" class="font-medium text-[#01696f] hover:text-[#0c4e54] transition-colors">
              Register your institution
            </router-link>
          </p>
        </div>

        <!-- Verify shortcut -->
        <p class="mt-3 text-center text-xs text-[#7a7974]">
          Just need to verify a certificate?
          <router-link to="/verify" class="text-[#01696f] hover:underline">Go to verifier →</router-link>
        </p>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { roleDashboard } from '../utils/roles'


const auth         = useAuthStore()
const router       = useRouter()
const form         = ref({ email: '', password: '' })
const loading      = ref(false)
const showPassword = ref(false)
const errorMessage = ref('')

const handleLogin = async () => {
  errorMessage.value = ''
  loading.value = true

  try {
    const result = await auth.login(form.value)

    // ✅ Use roleDashboard — ignore result.redirect (backend path doesn't match frontend routes)
    router.push(roleDashboard(result?.role))

  } catch (err) {
    const status = err.response?.status
    const msg    = err.response?.data?.error || err.response?.data?.message

    if (status === 401) {
      errorMessage.value = 'Invalid email or password. Please try again.'
    } else if (status === 403) {
      errorMessage.value = 'Your email is not verified. Please check your inbox.'
    } else if (status === 429) {
      errorMessage.value = 'Too many attempts. Please wait a moment and try again.'
    } else if (msg) {
      errorMessage.value = msg
    } else {
      errorMessage.value = 'Something went wrong. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>
