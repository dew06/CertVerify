<template>
  <div class="upload-card">
    <div class="upload-card__header">
      <div class="upload-card__icon-wrap">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
          <polyline points="12 18 12 12"/>
          <polyline points="9 15 12 12 15 15"/>
        </svg>
      </div>
      <div>
        <h3 class="upload-card__title">Upload Your Certificate</h3>
        <p class="upload-card__subtitle">
          Received a certificate by email? Upload the PDF to link it to your profile.
        </p>
      </div>
    </div>

    <!-- Drop zone -->
    <label
      class="dropzone"
      :class="{
        'dropzone--dragging': isDragging,
        'dropzone--disabled': state.uploading,
        'dropzone--done': state.result?.valid
      }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
    >
      <div class="dropzone__inner">
        <!-- Idle -->
        <template v-if="!state.uploading && !state.result && !state.error">
          <svg class="dropzone__icon" width="28" height="28" viewBox="0 0 24 24"
               fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="17 8 12 3 7 8"/>
            <line x1="12" y1="3" x2="12" y2="15"/>
          </svg>
          <span class="dropzone__label">
            <strong>Click to upload</strong> or drag and drop
          </span>
          <span class="dropzone__hint">PDF only · max 10 MB</span>
        </template>

        <!-- Uploading -->
        <template v-else-if="state.uploading">
          <svg class="dropzone__spinner" width="28" height="28" viewBox="0 0 24 24"
               fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          <span class="dropzone__label">Verifying certificate…</span>
          <div class="dropzone__progress">
            <div class="dropzone__progress-bar"></div>
          </div>
        </template>

        <!-- Success -->
        <template v-else-if="state.result?.valid">
          <svg class="dropzone__check" width="28" height="28" viewBox="0 0 24 24"
               fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 6 9 17l-5-5"/>
          </svg>
          <span class="dropzone__label text-success">Certificate verified!</span>
          <span class="dropzone__hint">Click to upload another</span>
        </template>

        <!-- Error / mismatch -->
        <template v-else-if="state.error || (state.result && !state.result.valid)">
          <svg class="dropzone__x" width="28" height="28" viewBox="0 0 24 24"
               fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
          <span class="dropzone__label text-error">Upload failed</span>
          <span class="dropzone__hint">Click to try again</span>
        </template>
      </div>

      <input
        ref="fileInput"
        type="file"
        accept="application/pdf"
        class="dropzone__input"
        :disabled="state.uploading"
        @change="handleFileChange"
      />
    </label>

    <!-- Result card — success -->
    <transition name="slide-up">
      <div v-if="state.result?.valid" class="result result--success">
        <div class="result__top">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2.5">
            <path d="M20 6 9 17l-5-5"/>
          </svg>
          <span class="result__status">Verified &amp; linked to your account</span>
        </div>

        <div class="result__grid">
          <div class="result__field">
            <span class="result__label">Student</span>
            <span class="result__value">{{ state.result.student_name }}</span>
          </div>
          <div class="result__field">
            <span class="result__label">Degree</span>
            <span class="result__value">{{ state.result.degree }}</span>
          </div>
          <div v-if="state.result.gpa" class="result__field">
            <span class="result__label">GPA</span>
            <span class="result__value">{{ state.result.gpa }}</span>
          </div>
          <div class="result__field">
            <span class="result__label">University</span>
            <span class="result__value">{{ state.result.university }}</span>
          </div>
          <div class="result__field">
            <span class="result__label">Issue Date</span>
            <span class="result__value">{{ state.result.issue_date }}</span>
          </div>
          <div class="result__field">
            <span class="result__label">Cert ID</span>
            <span class="result__value result__value--mono">{{ state.result.cert_id }}</span>
          </div>
        </div>

        <!-- Blockchain badge -->
        <div class="result__blockchain" :class="state.result.blockchain?.anchored ? 'result__blockchain--anchored' : 'result__blockchain--pending'">
          <template v-if="state.result.blockchain?.anchored">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2.5">
              <path d="M20 6 9 17l-5-5"/>
            </svg>
            Anchored on Cardano ·
            <a :href="state.result.blockchain.explorer_url" target="_blank" rel="noopener noreferrer"
               class="result__link">
              View on explorer ↗
            </a>
          </template>
          <template v-else>
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
                 stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            Blockchain anchoring pending
          </template>
        </div>

        <!-- IPFS link -->
        <a v-if="state.result.ipfs?.pdf_hash"
           :href="state.result.ipfs.gateway_url" target="_blank" rel="noopener noreferrer"
           class="result__ipfs-link">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2">
            <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
            <polyline points="15 3 21 3 21 9"/>
            <line x1="10" y1="14" x2="21" y2="3"/>
          </svg>
          View original PDF on IPFS
        </a>
      </div>
    </transition>

    <!-- Error banner -->
    <transition name="slide-up">
      <div v-if="state.error || (state.result && !state.result.valid)" class="result result--error">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="15" y1="9" x2="9" y2="15"/>
          <line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
        <span>{{ state.error || state.result?.error || state.result?.message }}</span>
        <button class="result__dismiss" @click="reset" aria-label="Dismiss">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
               stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import api from '../services/api'

const emit = defineEmits(['uploaded'])

const fileInput = ref(null)
const isDragging = ref(false)

const state = reactive({
  uploading: false,
  result: null,
  error: null
})

function reset() {
  state.uploading = false
  state.result = null
  state.error = null
  if (fileInput.value) fileInput.value.value = ''
}

function handleDrop(e) {
  isDragging.value = false
  if (state.uploading) return
  const file = e.dataTransfer.files[0]
  if (file) processFile(file)
}

function handleFileChange(e) {
  const file = e.target.files[0]
  if (file) processFile(file)
  // Reset result/error when user picks a new file
  state.result = null
  state.error = null
}

async function processFile(file) {
  if (file.type !== 'application/pdf') {
    state.error = 'Only PDF files are accepted.'
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    state.error = 'File too large. Maximum allowed size is 10 MB.'
    return
  }

  state.uploading = true
  state.result = null
  state.error = null

  const formData = new FormData()
  formData.append('certificate', file)

  try {
    const res = await api.uploadCertificate(formData)
    state.result = res.data
    if (res.data.valid) {
      emit('uploaded', res.data)
    }
  } catch (err) {
    const msg = err.response?.data?.error
      || err.response?.data?.message
      || 'Upload failed. Please try again.'
    state.error = msg
  } finally {
    state.uploading = false
    if (fileInput.value) fileInput.value.value = ''
  }
}
</script>

<style scoped>
/* ── Card ──────────────────────────────────────────────────────────── */
.upload-card {
  background: #f9f8f5;
  border: 1px solid #d4d1ca;
  border-radius: 0.75rem;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.upload-card__header {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.upload-card__icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 0.5rem;
  background: #cedcd8;
  color: #01696f;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.upload-card__title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #28251d;
  line-height: 1.3;
}

.upload-card__subtitle {
  font-size: 0.75rem;
  color: #7a7974;
  margin-top: 0.125rem;
  line-height: 1.5;
}

/* ── Drop zone ─────────────────────────────────────────────────────── */
.dropzone {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 120px;
  border: 2px dashed #d4d1ca;
  border-radius: 0.625rem;
  background: #ffffff;
  cursor: pointer;
  transition: border-color 180ms ease, background 180ms ease;
  overflow: hidden;
}

.dropzone:hover:not(.dropzone--disabled) {
  border-color: #01696f;
  background: #f3f9f8;
}

.dropzone--dragging {
  border-color: #01696f;
  background: #eaf4f3;
}

.dropzone--done {
  border-color: #437a22;
  background: #f5faf2;
}

.dropzone--disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.dropzone__input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  width: 100%;
  height: 100%;
}

.dropzone__inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.375rem;
  padding: 1.25rem;
  pointer-events: none;
  text-align: center;
}

.dropzone__icon { color: #bab9b4; }
.dropzone__check { color: #437a22; }
.dropzone__x { color: #a12c7b; }

.dropzone__spinner {
  color: #01696f;
  animation: spin 0.9s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.dropzone__label {
  font-size: 0.8125rem;
  color: #7a7974;
}

.dropzone__label strong { color: #28251d; }

.text-success { color: #437a22 !important; font-weight: 600; }
.text-error   { color: #a12c7b !important; font-weight: 600; }

.dropzone__hint {
  font-size: 0.6875rem;
  color: #bab9b4;
}

.dropzone__progress {
  width: 120px;
  height: 3px;
  background: #e6e4df;
  border-radius: 9999px;
  overflow: hidden;
  margin-top: 0.25rem;
}

.dropzone__progress-bar {
  height: 100%;
  width: 40%;
  background: #01696f;
  border-radius: 9999px;
  animation: progress-indeterminate 1.2s ease-in-out infinite;
}

@keyframes progress-indeterminate {
  0%   { transform: translateX(-200%); }
  100% { transform: translateX(400%); }
}

/* ── Result cards ──────────────────────────────────────────────────── */
.result {
  border-radius: 0.625rem;
  padding: 0.875rem 1rem;
  font-size: 0.78125rem;
}

.result--success {
  background: #f0f7ee;
  border: 1px solid rgba(67, 122, 34, 0.18);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.result--error {
  background: #faf1f7;
  border: 1px solid rgba(161, 44, 123, 0.18);
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  color: #a12c7b;
  position: relative;
}

.result--error svg { flex-shrink: 0; margin-top: 1px; }

.result__top {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #437a22;
  font-weight: 600;
}

.result__status { font-size: 0.8125rem; }

.result__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem 1.5rem;
}

.result__field {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.result__label {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #7a7974;
  font-weight: 500;
}

.result__value {
  font-size: 0.8125rem;
  color: #28251d;
  font-weight: 500;
}

.result__value--mono {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.6875rem;
  color: #7a7974;
  word-break: break-all;
}

.result__blockchain {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  padding: 0.4375rem 0.625rem;
  border-radius: 0.375rem;
}

.result__blockchain--anchored {
  background: #cedcd8;
  color: #01696f;
}

.result__blockchain--pending {
  background: #e9e0c6;
  color: #b07a00;
}

.result__link {
  color: inherit;
  font-weight: 600;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.result__ipfs-link {
  display: inline-flex;
  align-items: center;
  gap: 0.3125rem;
  font-size: 0.75rem;
  color: #7a7974;
  text-decoration: underline;
  text-underline-offset: 2px;
  width: fit-content;
}

.result__ipfs-link:hover { color: #01696f; }

.result__dismiss {
  margin-left: auto;
  flex-shrink: 0;
  color: #a12c7b;
  opacity: 0.6;
  padding: 0.125rem;
  cursor: pointer;
  background: none;
  border: none;
  line-height: 1;
}

.result__dismiss:hover { opacity: 1; }

/* ── Transition ────────────────────────────────────────────────────── */
.slide-up-enter-active { transition: all 220ms cubic-bezier(0.16, 1, 0.3, 1); }
.slide-up-leave-active { transition: all 160ms ease-in; }
.slide-up-enter-from  { opacity: 0; transform: translateY(8px); }
.slide-up-leave-to    { opacity: 0; transform: translateY(-4px); }
</style>