<template>
  <div class="min-h-screen bg-[#f7f6f2] flex flex-col">

    <!-- ── Navbar ──────────────────────────────────────────────────────── -->
    <nav class="h-16 border-b border-[#dcd9d5] bg-[#f7f6f2]/90 backdrop-blur flex items-center px-6 sticky top-0 z-50">
      <router-link to="/" class="flex items-center gap-2 text-sm font-semibold text-[#28251d]">
        <svg width="26" height="26" viewBox="0 0 28 28" fill="none" aria-label="CertChain logo">
          <rect width="28" height="28" rx="6" fill="#01696f"/>
          <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          ircle cx="14" cy="14" r="11" stroke="white" stroke-width="1.5" fill="none" opacity="0.4"/>
        </svg>
        CertChain
      </router-link>
    </nav>

    <div class="flex-1 flex items-center justify-center px-4 py-12">
      <div class="w-full max-w-lg">

        <!-- ================================================================
             STEP 1 — Role selection
        ================================================================ -->
        <div v-if="step === 'role'">
          <div class="mb-8 text-center">
            <h1 class="text-2xl font-bold tracking-tight text-[#28251d]">Create an account</h1>
            <p class="mt-1.5 text-sm text-[#7a7974]">Choose the type of account that describes you.</p>
          </div>

          <div class="grid gap-4">
            <button
              v-for="role in roles"
              :key="role.value"
              type="button"
              @click="selectRole(role.value)"
              class="group w-full text-left flex items-start gap-4 bg-[#f9f8f5] border border-[#d4d1ca]
                     rounded-xl p-5 hover:border-[#01696f] hover:bg-white hover:shadow-md transition-all"
            >
              <div
                class="w-12 h-12 shrink-0 rounded-xl flex items-center justify-center"
                :class="role.iconBg"
              >
                <svg v-if="role.value === 'university'" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                  <path d="M22 10v6M2 10l10-5 10 5-10 5z"/>
                  <path d="M6 12v5c3 3 9 3 12 0v-5"/>
                </svg>
                <svg v-else-if="role.value === 'company'" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#006494" stroke-width="2">
                  <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
                  <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
                </svg>
                <svg v-else width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#7a39bb" stroke-width="2">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  ircle cx="12" cy="7" r="4"/>
                </svg>
              </div>

              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-0.5">
                  <p class="font-semibold text-[#28251d]">{{ role.label }}</p>
                  <svg class="text-[#bab9b4] group-hover:text-[#01696f] transition-colors shrink-0"
                       width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M5 12h14M12 5l7 7-7 7"/>
                  </svg>
                </div>
                <p class="text-sm text-[#7a7974] leading-relaxed">{{ role.description }}</p>
              </div>
            </button>
          </div>

          <p class="mt-8 text-center text-sm text-[#7a7974]">
            Already have an account?
            <router-link to="/login" class="font-medium text-[#01696f] hover:text-[#0c4e54] transition-colors">
              Sign in
            </router-link>
          </p>
        </div>

        <!-- ================================================================
             STEP 2 — Role-specific form
        ================================================================ -->
        <div v-else-if="step === 'form' && !registrationSuccess">

          <!-- Back + role badge -->
          <div class="flex items-center gap-3 mb-6">
            <button
              type="button"
              @click="step = 'role'"
              class="flex items-center gap-1.5 text-sm text-[#7a7974] hover:text-[#28251d] transition-colors"
            >
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M19 12H5M12 19l-7-7 7-7"/>
              </svg>
              Back
            </button>
            <span class="text-xs font-medium px-2.5 py-1 rounded-full border" :class="activeRole.chipClass">
              {{ activeRole.label }}
            </span>
          </div>

          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-2xl p-8 shadow-sm">
            <!-- Heading -->
            <div class="mb-7">
              <h1 class="text-2xl font-bold tracking-tight text-[#28251d]">{{ activeRole.formTitle }}</h1>
              <p class="mt-1 text-sm text-[#7a7974] leading-relaxed">{{ activeRole.formSubtitle }}</p>
            </div>

            <form @submit.prevent="handleRegister" novalidate class="space-y-4">

              <!-- Name -->
              <div>
                <label for="reg-name" class="block text-sm font-medium text-[#28251d] mb-1.5">
                  {{ currentRole === 'university' ? 'University name' : currentRole === 'company' ? 'Company name' : 'Full name' }}
                </label>
                <input
                  id="reg-name"
                  v-model.trim="form.name"
                  type="text"
                  required
                  :disabled="loading"
                  :placeholder="activeRole.namePlaceholder"
                  class="field"
                />
              </div>

              <!-- Email -->
              <div>
                <label for="reg-email" class="block text-sm font-medium text-[#28251d] mb-1.5">
                  {{ currentRole === 'university' ? 'Admin email' : currentRole === 'company' ? 'Business email' : 'Email address' }}
                </label>
                <input
                  id="reg-email"
                  v-model.trim="form.email"
                  type="email"
                  autocomplete="email"
                  required
                  :disabled="loading"
                  :placeholder="activeRole.emailPlaceholder"
                  class="field"
                />
              </div>

              <!-- Domain (university only) -->
              <div v-if="currentRole === 'university'">
                <label for="reg-domain" class="block text-sm font-medium text-[#28251d] mb-1.5">Official domain</label>
                <input
                  id="reg-domain"
                  v-model.trim="form.domain"
                  type="text"
                  required
                  :disabled="loading"
                  placeholder="uom.lk"
                  class="field"
                />
                <p class="mt-1.5 text-xs text-[#7a7974]">Used for domain verification and institutional trust.</p>
              </div>

              <!-- Company extras -->
              <template v-if="currentRole === 'company'">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label for="reg-industry" class="block text-sm font-medium text-[#28251d] mb-1.5">Industry</label>
                    <input id="reg-industry" v-model.trim="form.industry" type="text"
                           :disabled="loading" placeholder="Technology" class="field"/>
                  </div>
                  <div>
                    <label for="reg-size" class="block text-sm font-medium text-[#28251d] mb-1.5">Company size</label>
                    <select id="reg-size" v-model="form.company_size" :disabled="loading" class="field">
                      <option value="">Select…</option>
                      <option>1–10</option>
                      <option>11–50</option>
                      <option>51–200</option>
                      <option>201–500</option>
                      <option>500+</option>
                    </select>
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label for="reg-location" class="block text-sm font-medium text-[#28251d] mb-1.5">Location</label>
                    <input id="reg-location" v-model.trim="form.location" type="text"
                           :disabled="loading" placeholder="Colombo, Sri Lanka" class="field"/>
                  </div>
                  <div>
                    <label for="reg-website" class="block text-sm font-medium text-[#28251d] mb-1.5">Website</label>
                    <input id="reg-website" v-model.trim="form.website" type="url"
                           :disabled="loading" placeholder="https://company.com" class="field"/>
                  </div>
                </div>
              </template>

              <!-- Student extras -->
              <template v-if="currentRole === 'student'">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label for="reg-phone" class="block text-sm font-medium text-[#28251d] mb-1.5">
                      Phone <span class="text-[#7a7974] font-normal">(optional)</span>
                    </label>
                    <input id="reg-phone" v-model.trim="form.phone" type="tel"
                           :disabled="loading" placeholder="+94 77 000 0000" class="field"/>
                  </div>
                  <div>
                    <label for="reg-age" class="block text-sm font-medium text-[#28251d] mb-1.5">
                      Age <span class="text-[#7a7974] font-normal">(optional)</span>
                    </label>
                    <input id="reg-age" v-model.number="form.age" type="number"
                           min="16" max="100" :disabled="loading" placeholder="22" class="field"/>
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label for="reg-gender" class="block text-sm font-medium text-[#28251d] mb-1.5">
                      Gender <span class="text-[#7a7974] font-normal">(optional)</span>
                    </label>
                    <select id="reg-gender" v-model="form.gender" :disabled="loading" class="field">
                      <option value="">Prefer not to say</option>
                      <option value="male">Male</option>
                      <option value="female">Female</option>
                      <option value="other">Other</option>
                    </select>
                  </div>
                  <div>
                    <label for="reg-nationality" class="block text-sm font-medium text-[#28251d] mb-1.5">
                      Nationality <span class="text-[#7a7974] font-normal">(optional)</span>
                    </label>
                    <input id="reg-nationality" v-model.trim="form.nationality" type="text"
                           :disabled="loading" placeholder="Sri Lankan" class="field"/>
                  </div>
                </div>
                <div>
                  <label for="reg-linkedin" class="block text-sm font-medium text-[#28251d] mb-1.5">
                    LinkedIn URL <span class="text-[#7a7974] font-normal">(optional)</span>
                  </label>
                  <input id="reg-linkedin" v-model.trim="form.linkedin_url" type="url"
                         :disabled="loading" placeholder="https://linkedin.com/in/yourname" class="field"/>
                </div>
              </template>

              <!-- Security notice (university only) -->
              <div v-if="currentRole === 'university'"
                   class="rounded-lg border border-[#d4dfcc] bg-[#f6faf2] px-4 py-3">
                <div class="flex items-start gap-2.5">
                  <svg class="mt-0.5 shrink-0" width="16" height="16" viewBox="0 0 24 24"
                       fill="none" stroke="#437a22" stroke-width="2">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                  </svg>
                  <div>
                    <p class="text-sm font-medium text-[#28251d]">Choose a strong master password</p>
                    <p class="text-xs text-[#7a7974] mt-0.5 leading-relaxed">
                      Your password protects your institution account and wallet setup.
                    </p>
                  </div>
                </div>
              </div>

              <!-- Password -->
              <div>
                <label for="reg-password" class="block text-sm font-medium text-[#28251d] mb-1.5">Password</label>
                <div class="relative">
                  <input
                    id="reg-password"
                    v-model="form.password"
                    :type="showPassword ? 'text' : 'password'"
                    autocomplete="new-password"
                    required
                    :minlength="currentRole === 'university' ? 12 : 8"
                    :disabled="loading"
                    :placeholder="currentRole === 'university' ? 'Minimum 12 characters' : 'Minimum 8 characters'"
                    class="field pr-10"
                  />
                  <button
                    type="button"
                    @click="showPassword = !showPassword"
                    :aria-label="showPassword ? 'Hide password' : 'Show password'"
                    class="absolute right-3 top-1/2 -translate-y-1/2 text-[#7a7974] hover:text-[#28251d] transition-colors"
                  >
                    <svg v-if="!showPassword" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>ircle cx="12" cy="12" r="3"/>
                    </svg>
                    <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                      <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                      e x1="1" y1="1" x2="23" y2="23"/>
                    </svg>
                  </button>
                </div>

                <!-- Strength bar -->
                <div class="mt-2">
                  <div class="h-1.5 rounded-full bg-[#e6e4df] overflow-hidden">
                    <div class="h-full transition-all duration-300"
                         :class="passwordStrength.barClass"
                         :style="{ width: passwordStrength.width }"/>
                  </div>
                  <p class="mt-1 text-xs" :class="passwordStrength.textClass">
                    {{ passwordStrength.label }}
                  </p>
                </div>
              </div>

              <!-- Inline error -->
              <div v-if="errorMessage" role="alert"
                   class="flex items-start gap-2.5 rounded-lg border border-[#e0ced7] bg-[#fdf5fa] px-4 py-3 text-sm text-[#a12c7b]">
                <svg class="mt-0.5 shrink-0" width="15" height="15" viewBox="0 0 24 24"
                     fill="none" stroke="currentColor" stroke-width="2">
                  ircle cx="12" cy="12" r="10"/>
                  e x1="12" y1="8" x2="12" y2="12"/>
                  e x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
                {{ errorMessage }}
              </div>

              <!-- Submit -->
              <button
                type="submit"
                :disabled="loading || !isFormValid"
                class="w-full flex items-center justify-center gap-2 py-2.5 px-4 text-sm font-semibold
                       rounded-lg bg-[#01696f] text-white hover:bg-[#0c4e54] transition-colors mt-2
                       disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg v-if="loading" class="animate-spin" width="16" height="16" viewBox="0 0 24 24"
                     fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56" stroke-linecap="round"/>
                </svg>
                {{ loading ? 'Creating account…' : activeRole.submitLabel }}
              </button>
            </form>

            <div class="mt-6 pt-6 border-t border-[#dcd9d5] text-center text-sm text-[#7a7974]">
              Already registered?
              <router-link to="/login" class="font-medium text-[#01696f] hover:text-[#0c4e54] transition-colors">
                Sign in
              </router-link>
            </div>
          </div>
        </div>

        <!-- ================================================================
             STEP 3 — Success
        ================================================================ -->
        <div v-else-if="registrationSuccess">
          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-2xl p-8 shadow-sm">

            <!-- Check icon -->
            <div class="text-center mb-6">
              <div class="inline-flex items-center justify-center w-14 h-14 rounded-full bg-[#d4dfcc] mb-4">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#437a22" stroke-width="2.5">
                  <path d="M20 6 9 17l-5-5"/>
                </svg>
              </div>
              <h2 class="text-2xl font-bold text-[#28251d]">{{ successConfig.title }}</h2>
              <p class="mt-2 text-sm text-[#7a7974] leading-relaxed max-w-xs mx-auto">
                {{ successConfig.message }}
              </p>
            </div>

            <!-- Recovery phrase — university only -->
            <template v-if="currentRole === 'university' && walletInfo.recovery_phrase">
              <div class="rounded-xl border border-[#ddcfc6] bg-[#fff8f4] p-4 mb-4">
                <p class="text-xs font-semibold uppercase tracking-wider text-[#964219] mb-2">
                  Recovery phrase
                </p>
                <p class="text-sm font-mono leading-relaxed break-words rounded-lg
                          border border-[#eaded6] bg-white p-3 text-[#28251d]">
                  {{ walletInfo.recovery_phrase }}
                </p>
                <p class="mt-2 text-xs text-[#964219]">
                  Store this offline. It will never be shown again.
                </p>
              </div>

              <div class="rounded-xl border border-[#dcd9d5] bg-white p-4 mb-5">
                <p class="text-xs uppercase tracking-wider text-[#7a7974] mb-1">Public address</p>
                <p class="text-xs font-mono text-[#28251d] break-all">
                  {{ walletInfo.wallet_address || 'Not provided' }}
                </p>
              </div>

              <label class="flex items-start gap-3 text-sm text-[#28251d] cursor-pointer mb-4">
                <input v-model="hasSavedPhrase" type="checkbox"
                       class="mt-0.5 rounded border-[#d4d1ca] text-[#01696f] focus:ring-[#01696f]"/>
                <span>I have saved the recovery phrase in a secure place.</span>
              </label>

              <button
                type="button"
                @click="copyRecoveryPhrase"
                class="w-full py-2.5 text-sm font-semibold rounded-lg border border-[#d4d1ca]
                       text-[#28251d] hover:bg-[#f3f0ec] transition-colors mb-3"
              >
                {{ copied ? 'Copied!' : 'Copy recovery phrase' }}
              </button>
            </template>

            <!-- Proceed -->
            <router-link
              to="/login"
              :class="[
                'block w-full text-center py-2.5 px-4 text-sm font-semibold rounded-lg transition-colors',
                canProceed
                  ? 'bg-[#01696f] text-white hover:bg-[#0c4e54]'
                  : 'bg-[#dcd9d5] text-[#7a7974] pointer-events-none'
              ]"
            >
              Proceed to login
            </router-link>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import api from '../services/api'

// ── State ──────────────────────────────────────────────────────────────────
const step        = ref('role')
const currentRole = ref('')
const loading     = ref(false)
const showPassword = ref(false)
const errorMessage = ref('')
const registrationSuccess = ref(false)
const walletInfo  = ref({})
const hasSavedPhrase = ref(false)
const copied = ref(false)

const form = reactive({
  // shared
  name: '', email: '', password: '',
  // university
  domain: '',
  // company
  industry: '', company_size: '', location: '', website: '', description: '',
  // student
  phone: '', age: null, gender: '', nationality: '', linkedin_url: '',
})

// ── Role config ────────────────────────────────────────────────────────────
const roles = [
  {
    value: 'university',
    label: 'University',
    description: 'Issue and manage blockchain-anchored academic certificates for your students.',
    iconBg: 'bg-[#cedcd8]',
    chipClass: 'bg-[#cedcd8] border-[#01696f]/20 text-[#01696f]',
    formTitle: 'Register your institution',
    formSubtitle: 'A Cardano wallet will be generated for your institution during setup.',
    namePlaceholder: 'University of Moratuwa',
    emailPlaceholder: 'admin@university.edu',
    submitLabel: 'Register & create wallet',
  },
  {
    value: 'company',
    label: 'Company',
    description: 'Verify candidate credentials instantly and find talent through privacy-first job matching.',
    iconBg: 'bg-[#c6d8e4]',
    chipClass: 'bg-[#c6d8e4] border-[#006494]/20 text-[#006494]',
    formTitle: 'Register your company',
    formSubtitle: 'Access verified graduate profiles and instant certificate verification.',
    namePlaceholder: 'Acme Technologies Pvt Ltd',
    emailPlaceholder: 'hr@company.com',
    submitLabel: 'Create company account',
  },
  {
    value: 'student',
    label: 'Student',
    description: 'Own your credentials and share them selectively with employers on your own terms.',
    iconBg: 'bg-[#dacfde]',
    chipClass: 'bg-[#dacfde] border-[#7a39bb]/20 text-[#7a39bb]',
    formTitle: 'Create student account',
    formSubtitle: 'Your personal details are private by default. You control what companies can see.',
    namePlaceholder: 'Asel Nurlanovna',
    emailPlaceholder: 'student@university.edu',
    submitLabel: 'Create student account',
  },
]

const activeRole = computed(() => roles.find((r) => r.value === currentRole.value) ?? roles[0])

const successConfig = computed(() => ({
  university: {
    title: 'Wallet created!',
    message: 'Your institution is registered. Save the recovery phrase before continuing.',
  },
  company: {
    title: 'Company registered!',
    message: 'Your company account is ready. Sign in to start verifying credentials.',
  },
  student: {
    title: 'Account created!',
    message: 'Check your email to verify your address, then sign in.',
  },
}[currentRole.value] ?? { title: 'Done!', message: '' }))

// ── Computed ───────────────────────────────────────────────────────────────
const passwordStrength = computed(() => {
  const pwd = form.password
  if (!pwd) return { label: 'Use at least 8 characters.', width: '0%', barClass: 'bg-[#dcd9d5]', textClass: 'text-[#7a7974]' }
  let score = 0
  if (pwd.length >= 8)          score++
  if (pwd.length >= 12)         score++
  if (/[A-Z]/.test(pwd))        score++
  if (/\d/.test(pwd))           score++
  if (/[^A-Za-z0-9]/.test(pwd)) score++
  if (score <= 2) return { label: 'Weak',   width: '33%',  barClass: 'bg-[#a13544]', textClass: 'text-[#a13544]' }
  if (score <= 4) return { label: 'Good',   width: '66%',  barClass: 'bg-[#d19900]', textClass: 'text-[#7a7974]'  }
  return               { label: 'Strong', width: '100%', barClass: 'bg-[#437a22]', textClass: 'text-[#437a22]' }
})

const isFormValid = computed(() => {
  const minPwd = currentRole.value === 'university' ? 12 : 8
  const base = form.name && form.email && form.password.length >= minPwd
  if (currentRole.value === 'university') return !!(base && form.domain)
  return !!base
})

const canProceed = computed(() =>
  currentRole.value !== 'university' || hasSavedPhrase.value
)

// ── Actions ────────────────────────────────────────────────────────────────
const selectRole = (role) => {
  currentRole.value = role
  errorMessage.value = ''
  step.value = 'form'
}

const normalizeDomain = (val) =>
  val.trim()
    .replace(/^https?:\/\//i, '')
    .replace(/^www\./i, '')
    .replace(/\/.*$/, '')

const handleRegister = async () => {
  errorMessage.value = ''
  loading.value = true
  try {
    let res
    if (currentRole.value === 'university') {
      res = await api.post('/university/register', {
        name:     form.name,
        email:    form.email,
        domain:   normalizeDomain(form.domain),
        password: form.password,
      })
      walletInfo.value = res.data
    } else if (currentRole.value === 'company') {
      await api.post('/company/register', {
        name:         form.name,
        email:        form.email,
        password:     form.password,
        industry:     form.industry     || undefined,
        company_size: form.company_size || undefined,
        location:     form.location     || undefined,
        website:      form.website      || undefined,
        description:  form.description  || undefined,
      })
    } else {
      await api.post('/student/register', {
        name:         form.name,
        email:        form.email,
        password:     form.password,
        phone:        form.phone        || undefined,
        age:          form.age          || undefined,
        gender:       form.gender       || undefined,
        nationality:  form.nationality  || undefined,
        linkedin_url: form.linkedin_url || undefined,
      })
    }
    registrationSuccess.value = true
  } catch (err) {
    errorMessage.value =
      err.response?.data?.error ||
      err.response?.data?.message ||
      'Registration failed. Please try again.'
  } finally {
    loading.value = false
  }
}

const copyRecoveryPhrase = async () => {
  const phrase = walletInfo.value?.recovery_phrase
  if (!phrase) return
  try {
    await navigator.clipboard.writeText(phrase)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1800)
  } catch {
    copied.value = false
  }
}
</script>

<style scoped>
.field {
  @apply w-full px-3 py-2.5 text-sm rounded-lg border transition-colors
         bg-white text-[#28251d] placeholder-[#bab9b4]
         border-[#d4d1ca] focus:outline-none focus:ring-2 focus:ring-[#01696f]/30
         focus:border-[#01696f] disabled:opacity-50 disabled:cursor-not-allowed;
}
</style>