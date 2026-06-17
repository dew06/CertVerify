<template>
  <div class="min-h-screen bg-[#f7f6f2]">

    <!-- ── Top bar ─────────────────────────────────────────────────────── -->
    <header class="sticky top-0 z-40 h-16 bg-[#f7f6f2]/90 backdrop-blur border-b border-[#dcd9d5] flex items-center px-6 gap-4">
      <router-link to="/" class="flex items-center gap-2 text-sm font-semibold text-[#28251d] shrink-0">
        <svg width="26" height="26" viewBox="0 0 28 28" fill="none" aria-label="CertChain logo">
          <rect width="28" height="28" rx="6" fill="#01696f"/>
          <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <circle cx="14" cy="14" r="11" stroke="white" stroke-width="1.5" fill="none" opacity="0.4"/>
        </svg>
        CertChain
      </router-link>

      <span class="text-[#dcd9d5]">/</span>

      <div class="flex-1 text-sm text-[#7a7974]">
        <span class="font-medium text-[#28251d]">{{ company.name || 'Company' }}</span>
        <span v-if="company.industry" class="ml-2 text-xs px-2 py-0.5 rounded-full bg-[#c6d8e4] text-[#006494] font-medium">
          {{ company.industry }}
        </span>
      </div>

      <button
        @click="handleLogout"
        class="flex items-center gap-1.5 text-sm text-[#7a7974] hover:text-[#a12c7b] transition-colors"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
          <polyline points="16 17 21 12 16 7"/>
          <line x1="21" y1="12" x2="9" y2="12"/>
        </svg>
        Sign out
      </button>
    </header>

    <!-- ── Main ───────────────────────────────────────────────────────── -->
    <main class="max-w-6xl mx-auto px-6 py-10">

      <!-- Page heading + tabs -->
      <div class="mb-6">
        <h1 class="text-xl font-bold text-[#28251d] tracking-tight mb-4">Dashboard</h1>

        <div class="flex gap-1 border-b border-[#dcd9d5]">
          <button
            v-for="tab in tabs" :key="tab.id"
            @click="activeTab = tab.id"
            class="relative px-4 py-2.5 text-sm font-medium transition-colors"
            :class="activeTab === tab.id
              ? 'text-[#01696f] after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:bg-[#01696f] after:rounded-t'
              : 'text-[#7a7974] hover:text-[#28251d]'"
          >
            {{ tab.label }}
            <span
              v-if="tab.id === 'requests' && pendingCount > 0"
              class="ml-1.5 inline-flex items-center justify-center w-4 h-4 rounded-full bg-[#a12c7b] text-white text-[10px] font-bold"
            >{{ pendingCount }}</span>
          </button>
        </div>
      </div>

      <!-- ── TAB: Overview ─────────────────────────────────────────────── -->
      <div v-if="activeTab === 'overview'">

        <!-- KPI cards -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5">
            <div class="flex items-start justify-between mb-4">
              <p class="text-xs font-semibold uppercase tracking-widest text-[#7a7974]">Searches Done</p>
              <div class="w-8 h-8 rounded-lg bg-[#c6d8e4] flex items-center justify-center">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#006494" stroke-width="2">
                  <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
                </svg>
              </div>
            </div>
            <p class="text-3xl font-bold tabular-nums text-[#28251d]">{{ stats.searchCount }}</p>
          </div>

          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5">
            <div class="flex items-start justify-between mb-4">
              <p class="text-xs font-semibold uppercase tracking-widest text-[#7a7974]">Requests Sent</p>
              <div class="w-8 h-8 rounded-lg bg-[#e9e0c6] flex items-center justify-center">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#d19900" stroke-width="2">
                  <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07A19.5 19.5 0 0 1 4.69 13"/>
                  <path d="M22 16.92A19.5 19.5 0 0 0 13 4.69 19.79 19.79 0 0 0 2.08 2a2 2 0 0 0-2 2.18v3"/>
                </svg>
              </div>
            </div>
            <p class="text-3xl font-bold tabular-nums text-[#28251d]">{{ stats.totalRequests }}</p>
          </div>

          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5">
            <div class="flex items-start justify-between mb-4">
              <p class="text-xs font-semibold uppercase tracking-widest text-[#7a7974]">Accepted</p>
              <div class="w-8 h-8 rounded-lg bg-[#d4dfcc] flex items-center justify-center">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#437a22" stroke-width="2">
                  <path d="M20 6 9 17l-5-5"/>
                </svg>
              </div>
            </div>
            <p class="text-3xl font-bold tabular-nums text-[#28251d]">{{ stats.acceptedRequests }}</p>
          </div>
        </div>

        <!-- Company profile card -->
        <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-6">
          <h2 class="text-sm font-semibold text-[#28251d] mb-4">Company Profile</h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Name</p>
              <p class="text-sm text-[#28251d] font-medium">{{ company.name || '—' }}</p>
            </div>
            <div>
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Email</p>
              <p class="text-sm text-[#28251d]">{{ company.email || '—' }}</p>
            </div>
            <div>
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Industry</p>
              <p class="text-sm text-[#28251d]">{{ company.industry || '—' }}</p>
            </div>
            <div>
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Company Size</p>
              <p class="text-sm text-[#28251d]">{{ company.company_size || '—' }}</p>
            </div>
            <div>
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Location</p>
              <p class="text-sm text-[#28251d]">{{ company.location || '—' }}</p>
            </div>
            <div>
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Website</p>
              <a v-if="company.website" :href="company.website" target="_blank" rel="noopener noreferrer"
                 class="text-sm text-[#01696f] hover:underline">{{ company.website }}</a>
              <p v-else class="text-sm text-[#28251d]">—</p>
            </div>
            <div class="sm:col-span-2" v-if="company.description">
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-1">Description</p>
              <p class="text-sm text-[#28251d]">{{ company.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- ── TAB: Search Students ───────────────────────────────────────── -->
      <div v-if="activeTab === 'search'">

        <!-- Filters -->
        <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5 mb-5">
          <h2 class="text-sm font-semibold text-[#28251d] mb-4">Search Filters</h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            <div>
              <label class="block text-xs text-[#7a7974] mb-1">Degree</label>
              <input v-model="filters.degree" type="text" placeholder="e.g. Computer Science"
                class="w-full px-3 py-2 text-sm border border-[#d4d1ca] rounded-lg bg-white text-[#28251d] placeholder-[#bab9b4] focus:outline-none focus:border-[#01696f]" />
            </div>
            <div>
              <label class="block text-xs text-[#7a7974] mb-1">Skills (comma separated)</label>
              <input v-model="filters.skills" type="text" placeholder="e.g. Python, React"
                class="w-full px-3 py-2 text-sm border border-[#d4d1ca] rounded-lg bg-white text-[#28251d] placeholder-[#bab9b4] focus:outline-none focus:border-[#01696f]" />
            </div>
           
            <div>
              <label class="block text-xs text-[#7a7974] mb-1">Min GPA</label>
              <input v-model.number="filters.min_gpa" type="number" min="0" max="4" step="0.1" placeholder="0.0"
                class="w-full px-3 py-2 text-sm border border-[#d4d1ca] rounded-lg bg-white text-[#28251d] placeholder-[#bab9b4] focus:outline-none focus:border-[#01696f]" />
            </div>
            <div>
              <label class="block text-xs text-[#7a7974] mb-1">Min Experience (years)</label>
              <input v-model.number="filters.min_experience" type="number" min="0" placeholder="0"
                class="w-full px-3 py-2 text-sm border border-[#d4d1ca] rounded-lg bg-white text-[#28251d] placeholder-[#bab9b4] focus:outline-none focus:border-[#01696f]" />
            </div>
          </div>
          <div class="flex items-center gap-3 mt-4">
            <button @click="runSearch"
              :disabled="searching"
              class="px-4 py-2 text-sm font-semibold bg-[#01696f] text-white rounded-lg hover:bg-[#0c4e54] disabled:opacity-50 transition-colors flex items-center gap-2">
              <svg v-if="searching" class="animate-spin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
              {{ searching ? 'Searching…' : 'Search' }}
            </button>
            <button @click="clearSearch" class="px-4 py-2 text-sm text-[#7a7974] hover:text-[#28251d] transition-colors">
              Clear
            </button>
            <span v-if="searchTotal > 0" class="text-xs text-[#7a7974] ml-auto">
              {{ searchTotal }} result{{ searchTotal !== 1 ? 's' : '' }}
            </span>
          </div>
        </div>

        <!-- Results -->
        <div v-if="searchResults.length > 0" class="space-y-3">
          <div
            v-for="s in searchResults" :key="s.id"
            class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5 flex flex-col sm:flex-row sm:items-start gap-4"
          >
            <!-- Avatar -->
            <div class="w-10 h-10 rounded-full bg-[#dacfde] flex items-center justify-center text-sm font-bold text-[#7a39bb] uppercase shrink-0">
              {{ (s.name || '?').charAt(0) }}
            </div>

            <!-- Info -->
            <div class="flex-1 min-w-0">
              <div class="flex flex-wrap items-center gap-2 mb-1">
                <p class="text-sm font-semibold text-[#28251d]">{{ s.name || 'Anonymous Student' }}</p>
                <span v-if="s.certificates_count"
                  class="text-xs px-1.5 py-0.5 rounded-full bg-[#cedcd8] text-[#01696f] font-medium">
                  {{ s.certificates_count }} certificate{{ s.certificates_count !== 1 ? 's' : '' }}
                </span>
                <span
                  class="text-xs px-1.5 py-0.5 rounded-full font-medium"
                  :class="requestStatusClass(s.request_status)"
                >{{ requestStatusLabel(s.request_status) }}</span>
              </div>

              <div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-[#7a7974] mb-2">
                <span v-if="s.email">{{ s.email }}</span>
                <span v-if="s.nationality">{{ s.nationality }}</span>
                <span v-if="s.gender" class="capitalize">{{ s.gender }}</span>
                <span v-if="s.age">Age {{ s.age }}</span>
              </div>

              <!-- Skills -->
              <div v-if="s.skills && s.skills.length" class="flex flex-wrap gap-1.5 mb-2">
                <span v-for="sk in s.skills" :key="sk.id"
                  class="text-xs px-2 py-0.5 rounded-full bg-[#f3f0ec] border border-[#d4d1ca] text-[#28251d]">
                  {{ sk.skill_name }}
                  <span v-if="sk.proficiency_level" class="text-[#7a7974]">· {{ sk.proficiency_level }}</span>
                </span>
              </div>

              <!-- Education -->
              <div v-if="s.education && s.education.length" class="text-xs text-[#7a7974]">
                <span v-for="(ed, i) in s.education" :key="ed.id">
                  {{ ed.degree }}
                  <span v-if="ed.gpa"> · GPA {{ ed.gpa }}</span>
                  <span v-if="i < s.education.length - 1"> &nbsp;|&nbsp; </span>
                </span>
              </div>
            </div>

            <!-- Action -->
            <div class="shrink-0">
              <button
                v-if="s.can_request"
                @click="openRequestModal(s)"
                class="px-3 py-2 text-xs font-semibold bg-[#01696f] text-white rounded-lg hover:bg-[#0c4e54] transition-colors"
              >
                Request Profile
              </button>
              <span v-else-if="s.request_status === 'pending'"
                class="text-xs text-[#7a7974] italic">Awaiting response</span>
              <span v-else-if="s.request_status === 'accepted'"
                class="text-xs text-[#437a22] font-medium">Profile Unlocked</span>
            </div>
          </div>
        </div>

        <div v-else-if="hasSearched && !searching"
          class="flex flex-col items-center justify-center py-16 text-center">
          <svg class="text-[#bab9b4] mb-3" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <p class="text-sm font-medium text-[#28251d]">No students found</p>
          <p class="text-xs text-[#7a7974] mt-1">Try adjusting your filters</p>
        </div>

        <div v-else-if="!hasSearched"
          class="flex flex-col items-center justify-center py-16 text-center">
          <svg class="text-[#bab9b4] mb-3" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <p class="text-sm font-medium text-[#28251d]">Search for students</p>
          <p class="text-xs text-[#7a7974] mt-1">Use the filters above and click Search</p>
        </div>
      </div>

      <!-- ── TAB: Requests ──────────────────────────────────────────────── -->
      <div v-if="activeTab === 'requests'">
        <div v-if="requests.length > 0" class="space-y-3">
          <div
            v-for="req in requests" :key="req.id"
            class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5 flex flex-col sm:flex-row sm:items-center gap-4"
          >
            <div class="flex-1 min-w-0">
              <div class="flex flex-wrap items-center gap-2 mb-1">
                <p class="text-sm font-semibold text-[#28251d]">{{ req.student_email }}</p>
                <span
                  class="text-xs px-2 py-0.5 rounded-full font-medium"
                  :class="requestStatusClass(req.status)"
                >{{ requestStatusLabel(req.status) }}</span>
              </div>
              <p v-if="req.message" class="text-xs text-[#7a7974] mb-1 italic">"{{ req.message }}"</p>
              <div class="flex gap-4 text-xs text-[#7a7974]">
                <span>Sent {{ formatDate(req.requested_at) }}</span>
                <span v-if="req.expires_at && req.status === 'pending'">
                  Expires {{ formatDate(req.expires_at) }}
                </span>
                <span v-if="req.responded_at && req.status !== 'pending'">
                  Responded {{ formatDate(req.responded_at) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="flex flex-col items-center justify-center py-16 text-center">
          <svg class="text-[#bab9b4] mb-3" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07"/>
            <path d="M2 2l20 20"/>
          </svg>
          <p class="text-sm font-medium text-[#28251d]">No requests sent yet</p>
          <p class="text-xs text-[#7a7974] mt-1">Search for students and request access to their profiles</p>
        </div>
      </div>

      <!-- ── TAB: Accepted Profiles ─────────────────────────────────────── -->
      <div v-if="activeTab === 'accepted'">
        <div v-if="acceptedProfiles.length > 0" class="space-y-4">
          <div
            v-for="p in acceptedProfiles" :key="p.student.id"
            class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5"
          >
            <!-- Header -->
            <div class="flex items-start gap-3 mb-4">
              <div class="w-10 h-10 rounded-full bg-[#dacfde] flex items-center justify-center text-sm font-bold text-[#7a39bb] uppercase shrink-0">
                {{ (p.student.name || '?').charAt(0) }}
              </div>
              <div class="flex-1">
                <p class="text-sm font-semibold text-[#28251d]">{{ p.student.name }}</p>
                <p class="text-xs text-[#7a7974]">{{ p.student.email }}</p>
              </div>
              <span class="text-xs px-2 py-0.5 rounded-full bg-[#d4dfcc] text-[#437a22] font-medium">
                {{ p.certificates_count }} cert{{ p.certificates_count !== 1 ? 's' : '' }}
              </span>
            </div>

            <!-- Details grid -->
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 mb-4">
              <div v-if="p.student.phone">
                <p class="text-xs text-[#7a7974] mb-0.5">Phone</p>
                <p class="text-sm text-[#28251d]">{{ p.student.phone }}</p>
              </div>
              <div v-if="p.student.age">
                <p class="text-xs text-[#7a7974] mb-0.5">Age</p>
                <p class="text-sm text-[#28251d]">{{ p.student.age }}</p>
              </div>
              <div v-if="p.student.gender">
                <p class="text-xs text-[#7a7974] mb-0.5">Gender</p>
                <p class="text-sm text-[#28251d] capitalize">{{ p.student.gender }}</p>
              </div>
              <div v-if="p.student.nationality">
                <p class="text-xs text-[#7a7974] mb-0.5">Nationality</p>
                <p class="text-sm text-[#28251d]">{{ p.student.nationality }}</p>
              </div>
              <div v-if="p.student.linkedin_url" class="sm:col-span-2">
                <p class="text-xs text-[#7a7974] mb-0.5">LinkedIn</p>
                <a :href="p.student.linkedin_url" target="_blank" rel="noopener noreferrer"
                   class="text-sm text-[#01696f] hover:underline">{{ p.student.linkedin_url }}</a>
              </div>
            </div>

            <!-- Skills -->
            <div v-if="p.skills && p.skills.length" class="mb-3">
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-2">Skills</p>
              <div class="flex flex-wrap gap-1.5">
                <span v-for="sk in p.skills" :key="sk.id"
                  class="text-xs px-2 py-0.5 rounded-full bg-[#f3f0ec] border border-[#d4d1ca] text-[#28251d]">
                  {{ sk.skill_name }}
                  <span v-if="sk.proficiency_level" class="text-[#7a7974]">· {{ sk.proficiency_level }}</span>
                  <span v-if="sk.years_of_experience" class="text-[#7a7974]"> {{ sk.years_of_experience }}y</span>
                </span>
              </div>
            </div>

            <!-- Education -->
            <div v-if="p.education && p.education.length">
              <p class="text-xs text-[#7a7974] uppercase tracking-wide mb-2">Education</p>
              <div class="space-y-1.5">
                <div v-for="ed in p.education" :key="ed.id"
                  class="flex items-start gap-2 text-xs text-[#28251d]">
                  <svg class="shrink-0 mt-0.5 text-[#01696f]" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c0 2 2 3 6 3s6-1 6-3v-5"/>
                  </svg>
                  <span>
                    {{ ed.degree }}<span v-if="ed.field_of_study"> in {{ ed.field_of_study }}</span>
                    <span v-if="ed.gpa" class="text-[#7a7974]"> · GPA {{ ed.gpa }}</span>
                    <span v-if="ed.is_current" class="text-[#437a22]"> · Current</span>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="flex flex-col items-center justify-center py-16 text-center">
          <svg class="text-[#bab9b4] mb-3" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
          <p class="text-sm font-medium text-[#28251d]">No accepted profiles yet</p>
          <p class="text-xs text-[#7a7974] mt-1">Students who accept your requests will appear here</p>
        </div>
      </div>
    </main>

    <!-- ── Request Modal ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="requestModal.open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
        @click.self="requestModal.open = false"
      >
        <div class="w-full max-w-md bg-[#f9f8f5] rounded-2xl border border-[#d4d1ca] shadow-lg p-6">
          <h3 class="text-base font-bold text-[#28251d] mb-1">Request Profile Access</h3>
          <p class="text-sm text-[#7a7974] mb-4">
            Requesting access to <span class="font-medium text-[#28251d]">{{ requestModal.student?.name || 'this student' }}</span>'s full profile.
          </p>

          <label class="block text-xs text-[#7a7974] mb-1.5">Message (optional)</label>
          <textarea
            v-model="requestModal.message"
            rows="3"
            placeholder="Briefly describe why you'd like to view this profile…"
            class="w-full px-3 py-2 text-sm border border-[#d4d1ca] rounded-lg bg-white text-[#28251d] placeholder-[#bab9b4] focus:outline-none focus:border-[#01696f] resize-none mb-4"
          />

          <p class="text-xs text-[#7a7974] mb-4">
            This request expires in 30 days. The student will be notified and can accept or decline.
          </p>

          <div class="flex gap-2 justify-end">
            <button @click="requestModal.open = false"
              class="px-4 py-2 text-sm text-[#7a7974] hover:text-[#28251d] border border-[#d4d1ca] rounded-lg transition-colors">
              Cancel
            </button>
            <button @click="submitRequest"
              :disabled="requestModal.submitting"
              class="px-4 py-2 text-sm font-semibold bg-[#01696f] text-white rounded-lg hover:bg-[#0c4e54] disabled:opacity-50 transition-colors">
              {{ requestModal.submitting ? 'Sending…' : 'Send Request' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const router = useRouter()
const auth   = useAuthStore()

// ── State ────────────────────────────────────────────────────────────────
const company        = ref({})
const requests       = ref([])
const acceptedProfiles = ref([])
const activeTab      = ref('overview')

const searching      = ref(false)
const hasSearched    = ref(false)
const searchResults  = ref([])
const searchTotal    = ref(0)

const filters = ref({
  degree: '', skills: '', nationality: '', gender: '',
  min_gpa: null, min_experience: null,
})

const requestModal = ref({
  open: false, student: null, message: '', submitting: false,
})

const stats = ref({ searchCount: 0, totalRequests: 0, acceptedRequests: 0 })

// ── Tabs ─────────────────────────────────────────────────────────────────
const tabs = [
  { id: 'overview', label: 'Overview' },
  { id: 'search',   label: 'Find Talent' },
  { id: 'requests', label: 'Requests' },
  { id: 'accepted', label: 'Accepted Profiles' },
]

const pendingCount = computed(() =>
  requests.value.filter(r => r.status === 'pending').length
)

// ── Load ─────────────────────────────────────────────────────────────────
onMounted(async () => {
  try {
    const [meRes, reqRes, accRes] = await Promise.all([
      api.getCompanyMe(),
      api.getMyRequests(),
      api.getAcceptedProfiles(),
    ])

    console.log('📦 /company/me:', meRes.data)

    const d = meRes.data
    company.value = d.company || d || {}

    requests.value       = reqRes.data?.requests || reqRes.data || []
    acceptedProfiles.value = accRes.data?.profiles || accRes.data || []

    stats.value = {
      searchCount:      d.stats?.search_count      || 0,
      totalRequests:    requests.value.length,
      acceptedRequests: acceptedProfiles.value.length,
    }
  } catch (err) {
    console.error('❌ Failed to load company dashboard:', err.response?.data || err.message)
    company.value = auth.user || {}
  }
})

// ── Search ───────────────────────────────────────────────────────────────
async function runSearch() {
  searching.value  = true
  hasSearched.value = true
  try {
    const payload = {
      degree:         filters.value.degree         || undefined,
      skills:         filters.value.skills         || undefined,
      min_gpa:        filters.value.min_gpa        || undefined,
      min_experience: filters.value.min_experience || undefined,
      limit: 50, offset: 0,
    }
    const res = await api.searchStudents(payload)
    searchResults.value = res.data?.students || []
    searchTotal.value   = res.data?.total    || 0
    stats.value.searchCount++
  } catch (err) {
    console.error('Search failed:', err.response?.data || err.message)
  } finally {
    searching.value = false
  }
}

function clearSearch() {
  filters.value = { degree: '', skills: '', min_gpa: null, min_experience: null }
  searchResults.value = []
  searchTotal.value   = 0
  hasSearched.value   = false
}

// ── Request modal ─────────────────────────────────────────────────────────
function openRequestModal(student) {
  requestModal.value = { open: true, student, message: '', submitting: false }
}

async function submitRequest() {
  requestModal.value.submitting = true
  try {
    await api.requestProfile({
      student_id: requestModal.value.student.id,
      message:    requestModal.value.message,
    })
    // update local request_status
    const s = searchResults.value.find(r => r.id === requestModal.value.student.id)
    if (s) { s.request_status = 'pending'; s.can_request = false }
    stats.value.totalRequests++
    requestModal.value.open = false
    // reload requests tab
    const res = await api.getMyRequests()
    requests.value = res.data?.requests || res.data || []
  } catch (err) {
    console.error('Request failed:', err.response?.data || err.message)
    alert(err.response?.data?.error || 'Failed to send request')
  } finally {
    requestModal.value.submitting = false
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────
function handleLogout() {
  auth.logout()
  router.push('/login')
}

function formatDate(d) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-US', { day: 'numeric', month: 'short', year: 'numeric' })
}

function requestStatusLabel(status) {
  return { pending: 'Pending', accepted: 'Accepted', rejected: 'Declined', expired: 'Expired' }[status] || '—'
}

function requestStatusClass(status) {
  return {
    pending:  'bg-[#e9e0c6] text-[#d19900]',
    accepted: 'bg-[#d4dfcc] text-[#437a22]',
    rejected: 'bg-[#e0ced7] text-[#a12c7b]',
    expired:  'bg-[#f3f0ec] text-[#7a7974]',
  }[status] || 'bg-[#f3f0ec] text-[#7a7974]'
}
</script>
