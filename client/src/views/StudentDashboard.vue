<template>
  <div class="min-h-screen bg-gray-50">

    <!-- Top Bar -->
    <header class="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-full bg-teal-600 flex items-center justify-center text-white font-bold text-sm">
          {{ initials }}
        </div>
        <div>
          <p class="font-semibold text-gray-900 text-sm leading-none">{{ student.name || 'Student' }}</p>
          <p class="text-xs text-gray-500 mt-0.5">{{ student.email }}</p>
        </div>
      </div>
      <button @click="logout" class="text-sm text-gray-500 hover:text-red-500 transition-colors">
        Sign out
      </button>
    </header>

    <div class="max-w-5xl mx-auto px-4 py-8">

      <!-- Page title -->
      <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Student Dashboard</h1>
        <p class="text-sm text-gray-500 mt-1">Manage your profile, skills, and privacy settings</p>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 mb-6 bg-gray-100 p-1 rounded-lg w-fit">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="[
            'px-4 py-2 rounded-md text-sm font-medium transition-colors',
            activeTab === tab.id
              ? 'bg-white text-teal-700 shadow-sm'
              : 'text-gray-600 hover:text-gray-900'
          ]"
        >
          {{ tab.label }}
          <span
            v-if="tab.id === 'requests' && pendingRequestCount > 0"
            class="ml-1.5 bg-red-500 text-white text-xs rounded-full px-1.5 py-0.5"
          >
            {{ pendingRequestCount }}
          </span>
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-24">
        <div class="w-8 h-8 border-2 border-teal-600 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <template v-else>

        <!-- ── Overview ──────────────────────────────────────────────────── -->
        <section v-if="activeTab === 'overview'" class="space-y-5">
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">Profile Information</h2>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class="text-xs text-gray-500 uppercase tracking-wide">Full Name</label>
                <p class="text-gray-900 mt-1">{{ student.name || '—' }}</p>
              </div>
              <div>
                <label class="text-xs text-gray-500 uppercase tracking-wide">Email</label>
                <p class="text-gray-900 mt-1">{{ student.email || '—' }}</p>
              </div>
              <div>
                <label class="text-xs text-gray-500 uppercase tracking-wide">Phone</label>
                <p class="text-gray-900 mt-1">{{ student.phone || '—' }}</p>
              </div>
              <div>
                <label class="text-xs text-gray-500 uppercase tracking-wide">Age</label>
                <p class="text-gray-900 mt-1">{{ student.age || '—' }}</p>
              </div>
              <div>
                <label class="text-xs text-gray-500 uppercase tracking-wide">Gender</label>
                <p class="text-gray-900 mt-1 capitalize">{{ student.gender || '—' }}</p>
              </div>
              <div>
                <label class="text-xs text-gray-500 uppercase tracking-wide">Nationality</label>
                <p class="text-gray-900 mt-1">{{ student.nationality || '—' }}</p>
              </div>
              <div class="sm:col-span-2">
                <label class="text-xs text-gray-500 uppercase tracking-wide">LinkedIn</label>
                <p class="text-gray-900 mt-1">
                  <a v-if="student.linkedin_url" :href="student.linkedin_url" target="_blank" rel="noopener noreferrer" class="text-teal-600 hover:underline">
                    {{ student.linkedin_url }}
                  </a>
                  <span v-else>—</span>
                </p>
              </div>
            </div>
          </div>

          <!-- Stats row -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div class="bg-white rounded-xl border border-gray-200 p-4 text-center">
              <p class="text-2xl font-bold text-teal-600">{{ skills.length }}</p>
              <p class="text-xs text-gray-500 mt-1">Skills</p>
            </div>
            <div class="bg-white rounded-xl border border-gray-200 p-4 text-center">
              <p class="text-2xl font-bold text-teal-600">{{ education.length }}</p>
              <p class="text-xs text-gray-500 mt-1">Education entries</p>
            </div>
            <div class="bg-white rounded-xl border border-gray-200 p-4 text-center">
              <p class="text-2xl font-bold text-teal-600">{{ certificates.length }}</p>
              <p class="text-xs text-gray-500 mt-1">Certificates</p>
            </div>
            <div class="bg-white rounded-xl border border-gray-200 p-4 text-center">
              <p class="text-2xl font-bold text-red-500">{{ pendingRequestCount }}</p>
              <p class="text-xs text-gray-500 mt-1">Pending requests</p>
            </div>
          </div>

          <!-- Certificate upload -->
          <CertificateUpload @uploaded="onCertificateUploaded" />

          <!-- Certificates list -->
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">My Certificates</h2>
            <div v-if="certificates.length === 0" class="text-center py-8 text-gray-400 text-sm">
              No certificates issued yet
            </div>
            <ul v-else class="divide-y divide-gray-100">
              <li v-for="cert in certificates" :key="cert.id" class="py-3 flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ cert.degree }}</p>
                  <p class="text-xs text-gray-500">{{ cert.cert_id }} · Issued {{ formatDate(cert.issue_date) }}</p>
                </div>
                <span :class="[
                  'text-xs px-2 py-1 rounded-full font-medium',
                  cert.blockchain_status === 'anchored'
                    ? 'bg-green-50 text-green-700'
                    : 'bg-yellow-50 text-yellow-700'
                ]">
                  {{ cert.blockchain_status || 'pending' }}
                </span>
              </li>
            </ul>
          </div>
        </section>

        <!-- ── Skills ────────────────────────────────────────────────────── -->
        <section v-if="activeTab === 'skills'" class="space-y-5">
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">Add Skill</h2>
            <form @submit.prevent="addSkill" class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <input
                v-model="newSkill.skill_name"
                placeholder="Skill name *"
                required
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              />
              <select
                v-model="newSkill.proficiency_level"
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              >
                <option value="">Proficiency level</option>
                <option value="beginner">Beginner</option>
                <option value="intermediate">Intermediate</option>
                <option value="advanced">Advanced</option>
                <option value="expert">Expert</option>
              </select>
              <div class="flex gap-2">
                <input
                  v-model.number="newSkill.years_of_experience"
                  type="number"
                  min="0"
                  placeholder="Years exp."
                  class="border border-gray-300 rounded-lg px-3 py-2 text-sm flex-1 focus:outline-none focus:ring-2 focus:ring-teal-500"
                />
                <button
                  type="submit"
                  :disabled="skillSaving"
                  class="bg-teal-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-teal-700 disabled:opacity-50 transition-colors"
                >
                  {{ skillSaving ? '...' : 'Add' }}
                </button>
              </div>
            </form>
          </div>

          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">Skills ({{ skills.length }})</h2>
            <div v-if="skills.length === 0" class="text-center py-8 text-gray-400 text-sm">
              No skills added yet
            </div>
            <ul v-else class="divide-y divide-gray-100">
              <li v-for="skill in skills" :key="skill.id" class="py-3 flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ skill.skill_name }}</p>
                  <p class="text-xs text-gray-500 capitalize">
                    {{ skill.proficiency_level || 'No level' }}
                    <span v-if="skill.years_of_experience"> · {{ skill.years_of_experience }}y exp</span>
                  </p>
                </div>
                <button
                  @click="deleteSkill(skill.id)"
                  class="text-xs text-red-500 hover:text-red-700 transition-colors px-2 py-1"
                >
                  Remove
                </button>
              </li>
            </ul>
          </div>
        </section>

        <!-- ── Education ─────────────────────────────────────────────────── -->
        <section v-if="activeTab === 'education'" class="space-y-5">
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">Add Education</h2>
            <form @submit.prevent="addEducation" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input
                v-model="newEdu.degree"
                placeholder="Degree *"
                required
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              />
              <input
                v-model="newEdu.field_of_study"
                placeholder="Field of study"
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              />
              <input
                v-model.number="newEdu.gpa"
                type="number"
                step="0.01"
                min="0"
                max="4"
                placeholder="GPA (e.g. 3.8)"
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              />
              <div class="flex items-center gap-2">
                <input
                  v-model="newEdu.is_current"
                  type="checkbox"
                  id="isCurrent"
                  class="rounded"
                />
                <label for="isCurrent" class="text-sm text-gray-700">Currently studying here</label>
              </div>
              <input
                v-model="newEdu.start_date"
                type="date"
                placeholder="Start date"
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              />
              <input
                v-if="!newEdu.is_current"
                v-model="newEdu.end_date"
                type="date"
                placeholder="End date"
                class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
              />
              <div class="sm:col-span-2">
                <button
                  type="submit"
                  :disabled="eduSaving"
                  class="bg-teal-600 text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-teal-700 disabled:opacity-50 transition-colors"
                >
                  {{ eduSaving ? 'Saving...' : 'Add Education' }}
                </button>
              </div>
            </form>
          </div>

          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">Education History ({{ education.length }})</h2>
            <div v-if="education.length === 0" class="text-center py-8 text-gray-400 text-sm">
              No education entries yet
            </div>
            <ul v-else class="divide-y divide-gray-100">
              <li v-for="edu in education" :key="edu.id" class="py-4 flex items-start justify-between">
                <div>
                  <div class="flex items-center gap-2">
                    <p class="text-sm font-medium text-gray-900">{{ edu.degree }}</p>
                    <!-- Badge shown for auto-created entries from cert upload -->
                    <span v-if="edu.source === 'certificate'"
                      class="text-[10px] px-1.5 py-0.5 rounded bg-teal-50 text-teal-700 font-medium border border-teal-100">
                      from certificate
                    </span>
                  </div>
                  <p class="text-xs text-gray-500">{{ edu.field_of_study }}</p>
                  <p class="text-xs text-gray-400 mt-0.5">
                    <span v-if="edu.gpa">GPA {{ edu.gpa }} ·</span>
                    {{ formatDate(edu.start_date) }} – {{ edu.is_current ? 'Present' : formatDate(edu.end_date) }}
                  </p>
                  <p v-if="edu.university?.name" class="text-xs text-teal-600 mt-0.5">{{ edu.university.name }}</p>
                </div>
                <button
                  @click="deleteEducation(edu.id)"
                  class="text-xs text-red-500 hover:text-red-700 transition-colors px-2 py-1 shrink-0"
                >
                  Remove
                </button>
              </li>
            </ul>
          </div>
        </section>

        <!-- ── Privacy ───────────────────────────────────────────────────── -->
        <section v-if="activeTab === 'privacy'">
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-1">Privacy Settings</h2>
            <p class="text-sm text-gray-500 mb-6">
              Choose what companies can see when they search for you.
              Toggle fields you want to make visible.
            </p>

            <div class="space-y-4">
              <div class="flex items-center justify-between py-3 border-b border-gray-100">
                <div>
                  <p class="text-sm font-medium text-gray-900">Searchable</p>
                  <p class="text-xs text-gray-500">Allow companies to find you in search results</p>
                </div>
                <button @click="privacy.is_searchable = !privacy.is_searchable" :class="toggleClass(privacy.is_searchable)">
                  <span :class="thumbClass(privacy.is_searchable)"></span>
                </button>
              </div>

              <div v-for="field in privacyFields" :key="field.key" class="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ field.label }}</p>
                  <p class="text-xs text-gray-500">{{ field.description }}</p>
                </div>
                <button @click="privacy[field.key] = !privacy[field.key]" :class="toggleClass(privacy[field.key])">
                  <span :class="thumbClass(privacy[field.key])"></span>
                </button>
              </div>
            </div>

            <div class="mt-6 flex items-center gap-3">
              <button
                @click="savePrivacy"
                :disabled="privacySaving"
                class="bg-teal-600 text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-teal-700 disabled:opacity-50 transition-colors"
              >
                {{ privacySaving ? 'Saving...' : 'Save settings' }}
              </button>
              <span v-if="privacySaved" class="text-sm text-green-600">✓ Saved</span>
            </div>
          </div>
        </section>

        <!-- ── Requests ──────────────────────────────────────────────────── -->
        <section v-if="activeTab === 'requests'">
          <div class="bg-white rounded-xl border border-gray-200 p-6">
            <h2 class="font-semibold text-gray-900 mb-4">Profile Access Requests</h2>
            <div v-if="requests.length === 0" class="text-center py-12 text-gray-400 text-sm">
              No companies have requested access to your profile yet
            </div>
            <ul v-else class="divide-y divide-gray-100">
              <li v-for="req in requests" :key="req.id" class="py-4">
                <div class="flex items-start justify-between gap-4">
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-900">{{ req.company?.name || 'Unknown company' }}</p>
                    <p v-if="req.message" class="text-xs text-gray-500 mt-1 line-clamp-2">{{ req.message }}</p>
                    <div class="flex items-center gap-3 mt-2">
                      <span :class="statusBadgeClass(req.status)">{{ req.status }}</span>
                      <span class="text-xs text-gray-400">{{ formatDate(req.requested_at) }}</span>
                      <span v-if="req.status === 'pending'" class="text-xs text-orange-500">
                        Expires {{ formatDate(req.expires_at) }}
                      </span>
                    </div>
                  </div>
                  <div v-if="req.status === 'pending'" class="flex gap-2 shrink-0">
                    <button
                      @click="respond(req.id, 'accepted')"
                      class="bg-teal-600 text-white px-3 py-1.5 rounded-lg text-xs font-medium hover:bg-teal-700 transition-colors"
                    >
                      Accept
                    </button>
                    <button
                      @click="respond(req.id, 'rejected')"
                      class="border border-red-300 text-red-600 px-3 py-1.5 rounded-lg text-xs font-medium hover:bg-red-50 transition-colors"
                    >
                      Decline
                    </button>
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </section>

      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../services/api'
import { useAuthStore } from '../stores/auth'
import CertificateUpload from '../components/CertificateUpload.vue'

const router = useRouter()
const auth = useAuthStore()

// ── State ──────────────────────────────────────────────────────────────────

const loading      = ref(true)
const student      = ref({})
const skills       = ref([])
const education    = ref([])
const certificates = ref([])
const requests     = ref([])
const privacy      = ref({
  is_searchable:    true,
  show_name:        false,
  show_email:       false,
  show_phone:       false,
  show_age:         false,
  show_gender:      false,
  show_nationality: false,
  show_linkedin:    false,
})

const activeTab     = ref('overview')
const skillSaving   = ref(false)
const eduSaving     = ref(false)
const privacySaving = ref(false)
const privacySaved  = ref(false)

const newSkill = ref({ skill_name: '', proficiency_level: '', years_of_experience: 0 })
const newEdu   = ref({ degree: '', field_of_study: '', gpa: null, start_date: '', end_date: '', is_current: false })

// ── Computed ───────────────────────────────────────────────────────────────

const initials = computed(() => {
  const name = student.value.name || ''
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2) || 'S'
})

const pendingRequestCount = computed(() =>
  requests.value.filter(r => r.status === 'pending').length
)

const tabs = [
  { id: 'overview',  label: 'Overview' },
  { id: 'skills',    label: 'Skills' },
  { id: 'education', label: 'Education' },
  { id: 'privacy',   label: 'Privacy' },
  { id: 'requests',  label: 'Requests' },
]

const privacyFields = [
  { key: 'show_name',        label: 'Name',        description: 'Show your full name to companies' },
  { key: 'show_email',       label: 'Email',       description: 'Show your email address' },
  { key: 'show_phone',       label: 'Phone',       description: 'Show your phone number' },
  { key: 'show_age',         label: 'Age',         description: 'Show your age' },
  { key: 'show_gender',      label: 'Gender',      description: 'Show your gender' },
  { key: 'show_nationality', label: 'Nationality', description: 'Show your nationality' },
  { key: 'show_linkedin',    label: 'LinkedIn',    description: 'Show your LinkedIn profile URL' },
]

// ── Data loading ───────────────────────────────────────────────────────────

async function loadAll() {
  loading.value = true
  try {
    const [meRes, reqRes] = await Promise.all([
      api.getStudentMe(),
      api.getStudentRequests().catch(() => ({ data: { requests: [] } })),
    ])

    const data = meRes.data
    student.value      = data.student      || {}
    skills.value       = data.skills       || []
    education.value    = data.education    || []
    certificates.value = data.certificates || []

    const s = data.student || {}
    privacy.value = {
      is_searchable:    s.is_searchable    ?? true,
      show_name:        s.show_name        ?? false,
      show_email:       s.show_email       ?? false,
      show_phone:       s.show_phone       ?? false,
      show_age:         s.show_age         ?? false,
      show_gender:      s.show_gender      ?? false,
      show_nationality: s.show_nationality ?? false,
      show_linkedin:    s.show_linkedin    ?? false,
    }

    requests.value = reqRes.data?.requests || []
  } catch (err) {
    console.error('Failed to load dashboard:', err)
  } finally {
    loading.value = false
  }
}

// ── Certificate upload callback ────────────────────────────────────────────

// Called by CertificateUpload when a cert is successfully verified.
// Refreshes both certificates AND education so both lists update instantly.
async function onCertificateUploaded(data) {
  try {
    const res = await api.getStudentMe()
    certificates.value = res.data.certificates || []
    education.value    = res.data.education    || []

    // If a new education entry was created, switch to the Education tab
    // so the student can see it immediately
    if (data.education_created) {
      activeTab.value = 'education'
    }
  } catch (err) {
    console.error('Failed to refresh after upload:', err)
  }
}

// ── Skills ─────────────────────────────────────────────────────────────────

async function addSkill() {
  if (!newSkill.value.skill_name.trim()) return
  skillSaving.value = true
  try {
    const res = await api.addSkill(newSkill.value)
    skills.value.push(res.data.skill || res.data)
    newSkill.value = { skill_name: '', proficiency_level: '', years_of_experience: 0 }
  } catch (err) {
    console.error('Failed to add skill:', err)
  } finally {
    skillSaving.value = false
  }
}

async function deleteSkill(id) {
  try {
    await api.deleteSkill(id)
    skills.value = skills.value.filter(s => s.id !== id)
  } catch (err) {
    console.error('Failed to delete skill:', err)
  }
}

// ── Education ──────────────────────────────────────────────────────────────

async function addEducation() {
  if (!newEdu.value.degree.trim()) return
  eduSaving.value = true
  try {
    const payload = { ...newEdu.value }
    if (payload.is_current) delete payload.end_date
    const res = await api.addEducation(payload)
    education.value.push(res.data.education || res.data)
    newEdu.value = { degree: '', field_of_study: '', gpa: null, start_date: '', end_date: '', is_current: false }
  } catch (err) {
    console.error('Failed to add education:', err)
  } finally {
    eduSaving.value = false
  }
}

async function deleteEducation(id) {
  try {
    await api.deleteEducation(id)
    education.value = education.value.filter(e => e.id !== id)
  } catch (err) {
    console.error('Failed to delete education:', err)
  }
}

// ── Privacy ────────────────────────────────────────────────────────────────

async function savePrivacy() {
  privacySaving.value = true
  privacySaved.value  = false
  try {
    await api.updateStudentProfile(privacy.value)
    privacySaved.value = true
    setTimeout(() => { privacySaved.value = false }, 3000)
  } catch (err) {
    console.error('Failed to save privacy:', err)
  } finally {
    privacySaving.value = false
  }
}

// ── Requests ───────────────────────────────────────────────────────────────

async function respond(requestId, status) {
  try {
    await api.respondToRequest(requestId, { status })
    const req = requests.value.find(r => r.id === requestId)
    if (req) req.status = status
  } catch (err) {
    console.error('Failed to respond to request:', err)
  }
}

// ── Auth ───────────────────────────────────────────────────────────────────

function logout() {
  auth.logout()
  router.push('/login')
}

// ── Helpers ────────────────────────────────────────────────────────────────

function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

function toggleClass(val) {
  return [
    'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
    val ? 'bg-teal-600' : 'bg-gray-300'
  ]
}

function thumbClass(val) {
  return [
    'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
    val ? 'translate-x-6' : 'translate-x-1'
  ]
}

function statusBadgeClass(status) {
  const map = {
    pending:  'text-xs px-2 py-0.5 rounded-full bg-yellow-50 text-yellow-700 font-medium capitalize',
    accepted: 'text-xs px-2 py-0.5 rounded-full bg-green-50 text-green-700 font-medium capitalize',
    rejected: 'text-xs px-2 py-0.5 rounded-full bg-red-50 text-red-700 font-medium capitalize',
    expired:  'text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-500 font-medium capitalize',
  }
  return map[status] || map.expired
}

// ── Init ───────────────────────────────────────────────────────────────────

onMounted(loadAll)
</script>
