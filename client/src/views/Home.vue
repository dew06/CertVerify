<template>
  <div class="bg-[#f7f6f2] text-[#28251d]">

    <!-- ======================================================================
         HERO
    ====================================================================== -->
    <section class="relative overflow-hidden">
      <!-- Single quiet tint — used once, not everywhere -->
      <div
        class="absolute inset-0 bg-[radial-gradient(ellipse_70%_60%_at_60%_-5%,rgba(1,105,111,0.07),transparent)]"
        aria-hidden="true"
      />

      <div class="relative max-w-6xl mx-auto px-6 pt-20 pb-24 md:pt-28 md:pb-32">
        <div class="grid md:grid-cols-2 gap-16 items-center">

          <!-- Left: copy -->
          <div>
            <!-- Pill label -->
            <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-[#dcd9d5]
                        bg-[#f9f8f5] text-xs text-[#7a7974] mb-7 select-none">
              <span class="w-1.5 h-1.5 rounded-full bg-[#437a22] inline-block"></span>
              Live on Cardano Preview Network
            </div>

            <h1 class="text-[clamp(2.2rem,4.5vw,4rem)] font-extrabold leading-[1.1] tracking-tight mb-5">
              Academic credentials<br/>
              <span class="text-[#01696f]">the world can verify.</span>
            </h1>

            <p class="text-[clamp(0.95rem,1.3vw,1.1rem)] text-[#7a7974] leading-relaxed mb-8 max-w-md">
              Universities issue tamper-proof certificates anchored to Cardano.
              Employers verify in seconds — no account, no phone call, no waiting.
            </p>

            <div class="flex flex-col sm:flex-row gap-3">
              <router-link
                to="/verify"
                class="inline-flex items-center justify-center gap-2 px-6 py-3 text-sm font-semibold
                       bg-[#01696f] text-white rounded-lg hover:bg-[#0c4e54] transition-colors"
              >
                Verify a Certificate
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              </router-link>
              <router-link
                to="/register"
                class="inline-flex items-center justify-center gap-2 px-6 py-3 text-sm font-semibold
                       border border-[#d4d1ca] text-[#28251d] rounded-lg hover:bg-[#f3f0ec] transition-colors"
              >
                Register University
              </router-link>
            </div>
          </div>

          <!-- Right: certificate card mockup -->
          <div class="hidden md:flex justify-center">
            <div class="relative w-full max-w-sm">
              <!-- Shadow card behind -->
              <div class="absolute inset-0 translate-x-3 translate-y-3 rounded-2xl bg-[#cedcd8] opacity-50"></div>

              <!-- Main card -->
              <div class="relative bg-white border border-[#d4d1ca] rounded-2xl p-6 shadow-lg">
                <!-- Card header -->
                <div class="flex items-center justify-between mb-5">
                  <div class="flex items-center gap-2">
                    <svg width="22" height="22" viewBox="0 0 28 28" fill="none">
                      <rect width="28" height="28" rx="6" fill="#01696f"/>
                      <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    <span class="text-xs font-semibold text-[#28251d]">CertChain</span>
                  </div>
                  <span class="text-xs px-2 py-0.5 rounded-full bg-[#d4dfcc] text-[#437a22] font-medium">Verified</span>
                </div>

                <!-- Cert fields -->
                <div class="space-y-3 mb-5">
                  <div>
                    <p class="text-[10px] uppercase tracking-wider text-[#7a7974] mb-0.5">Recipient</p>
                    <p class="text-sm font-semibold text-[#28251d]">Asel Nurlanovna</p>
                  </div>
                  <div>
                    <p class="text-[10px] uppercase tracking-wider text-[#7a7974] mb-0.5">Programme</p>
                    <p class="text-sm font-medium text-[#28251d]">BSc. Computer Science</p>
                  </div>
                  <div>
                    <p class="text-[10px] uppercase tracking-wider text-[#7a7974] mb-0.5">Issued by</p>
                    <p class="text-sm font-medium text-[#28251d]">University of Moratuwa</p>
                  </div>
                </div>

                <!-- Hash row -->
                <div class="rounded-lg bg-[#f3f0ec] px-3 py-2.5 border border-[#dcd9d5]">
                  <p class="text-[10px] uppercase tracking-wider text-[#7a7974] mb-0.5">Cardano tx hash</p>
                  <p class="text-[11px] font-mono text-[#7a7974] truncate">3a9f2c81dd4e…f7b2d441b8</p>
                </div>

                <!-- Anchored badge -->
                <div class="mt-3 flex items-center gap-1.5 text-xs text-[#7a7974]">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#437a22" stroke-width="2">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                  </svg>
                  Anchored to blockchain · 14 Apr 2025
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Stats row -->
        <div class="mt-16 grid grid-cols-2 md:grid-cols-4 gap-px bg-[#dcd9d5] rounded-xl overflow-hidden border border-[#dcd9d5]">
          <div v-for="stat in stats" :key="stat.label" class="bg-[#f9f8f5] px-6 py-5">
            <div class="text-xl font-bold text-[#28251d] tabular-nums">{{ stat.value }}</div>
            <div class="text-xs text-[#7a7974] mt-1">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- ======================================================================
         HOW IT WORKS
    ====================================================================== -->
    <section class="bg-[#f9f8f5] border-y border-[#dcd9d5] py-24 px-6">
      <div class="max-w-6xl mx-auto">
        <div class="mb-12">
          <p class="text-xs uppercase tracking-widest text-[#7a7974] mb-3">How it works</p>
          <h2 class="text-[clamp(1.4rem,2.8vw,2.25rem)] font-bold leading-tight max-w-lg">
            From issuance to verification in three steps.
          </h2>
        </div>

        <div class="grid md:grid-cols-3 gap-6">
          <div
            v-for="(step, i) in steps"
            :key="step.title"
            class="bg-white border border-[#d4d1ca] rounded-xl p-6 hover:shadow-md transition-shadow"
          >
            <div class="w-8 h-8 rounded-lg bg-[#cedcd8] text-[#01696f] flex items-center justify-center
                        text-sm font-bold mb-5 select-none shrink-0">
              {{ i + 1 }}
            </div>
            <h3 class="font-semibold text-[#28251d] mb-2">{{ step.title }}</h3>
            <p class="text-sm text-[#7a7974] leading-relaxed">{{ step.description }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- ======================================================================
         FEATURES — staggered bento (avoids identical 3-col grid anti-pattern)
    ====================================================================== -->
    <section class="py-24 px-6">
      <div class="max-w-6xl mx-auto">
        <div class="mb-12">
          <p class="text-xs uppercase tracking-widest text-[#7a7974] mb-3">Why CertChain</p>
          <h2 class="text-[clamp(1.4rem,2.8vw,2.25rem)] font-bold leading-tight max-w-xl">
            Built for institutions that take credentials seriously.
          </h2>
        </div>

        <div class="grid md:grid-cols-2 gap-6">
          <!-- Large feature: tamper proof -->
          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-8 flex flex-col justify-between min-h-60 row-span-2">
            <div>
              <div class="w-10 h-10 rounded-lg bg-[#cedcd8] flex items-center justify-center mb-5">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                </svg>
              </div>
              <h3 class="text-lg font-semibold mb-2">Tamper-proof by design</h3>
              <p class="text-sm text-[#7a7974] leading-relaxed max-w-sm">
                Every certificate is SHA-256 hashed and its fingerprint permanently anchored to Cardano.
                Altering a single character produces a completely different hash — forgery is cryptographically impossible.
              </p>
            </div>
            <div class="mt-8 rounded-lg bg-[#f3f0ec] border border-[#dcd9d5] px-4 py-3 font-mono text-xs text-[#7a7974] truncate">
              sha256: e3b0c44298fc1c149afb…
            </div>
          </div>

          <!-- Two stacked smaller cards -->
          <div class="flex flex-col gap-6">
            <!-- Instant verify -->
            <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-6">
              <div class="w-10 h-10 rounded-lg bg-[#cedcd8] flex items-center justify-center mb-4">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <polyline points="12 6 12 12 16 14"/>
                </svg>
              </div>
              <h3 class="font-semibold mb-1">Verified in seconds</h3>
              <p class="text-sm text-[#7a7974] leading-relaxed">
                Paste a Cert ID or scan the QR code on the PDF. Results are instant — no account, no manual lookup.
              </p>
            </div>

            <!-- University owned -->
            <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-6">
              <div class="w-10 h-10 rounded-lg bg-[#cedcd8] flex items-center justify-center mb-4">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                  <ellipse cx="12" cy="5" rx="9" ry="3"/>
                  <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
                  <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
                </svg>
              </div>
              <h3 class="font-semibold mb-1">University-owned records</h3>
              <p class="text-sm text-[#7a7974] leading-relaxed">
                Certificates live on Cardano and IPFS — not a vendor's server. The issuing institution retains full control.
              </p>
            </div>
          </div>
        </div>

        <!-- Second row: privacy + batch -->
        <div class="grid md:grid-cols-2 gap-6 mt-6">
          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-6">
            <div class="w-10 h-10 rounded-lg bg-[#cedcd8] flex items-center justify-center mb-4">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </div>
            <h3 class="font-semibold mb-1">Privacy-first job matching</h3>
            <p class="text-sm text-[#7a7974] leading-relaxed">
              Students control which details companies can see. Employers request profile access; students approve or decline.
            </p>
          </div>

          <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-6">
            <div class="w-10 h-10 rounded-lg bg-[#cedcd8] flex items-center justify-center mb-4">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
                <polyline points="10 9 9 9 8 9"/>
              </svg>
            </div>
            <h3 class="font-semibold mb-1">Batch CSV issuance</h3>
            <p class="text-sm text-[#7a7974] leading-relaxed">
              Upload a CSV and issue hundreds of certificates in one action. Each is individually hashed and anchored.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- ======================================================================
         WHO IT'S FOR — three-audience split, not identical feature cards
    ====================================================================== -->
    <section class="bg-[#f9f8f5] border-y border-[#dcd9d5] py-24 px-6">
      <div class="max-w-6xl mx-auto">
        <div class="mb-12">
          <p class="text-xs uppercase tracking-widest text-[#7a7974] mb-3">Built for everyone</p>
          <h2 class="text-[clamp(1.4rem,2.8vw,2.25rem)] font-bold leading-tight">Who uses CertChain?</h2>
        </div>

        <div class="grid md:grid-cols-3 gap-6">
          <div
            v-for="audience in audiences"
            :key="audience.title"
            class="bg-white border border-[#d4d1ca] rounded-xl p-6"
          >
            <div class="w-10 h-10 rounded-lg mb-5 flex items-center justify-center" :class="audience.iconBg">
              <component :is="audience.icon" />
            </div>
            <h3 class="font-semibold text-[#28251d] mb-2">{{ audience.title }}</h3>
            <p class="text-sm text-[#7a7974] leading-relaxed mb-4">{{ audience.description }}</p>
            <router-link
              :to="audience.route"
              class="inline-flex items-center gap-1 text-xs font-medium text-[#01696f] hover:text-[#0c4e54] transition-colors"
            >
              {{ audience.cta }}
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
            </router-link>
          </div>
        </div>
      </div>
    </section>

    <!-- ======================================================================
         VERIFY CTA — one full-width accent moment, used once
    ====================================================================== -->
    <section class="bg-[#01696f] text-white py-20 px-6">
      <div class="max-w-2xl mx-auto text-center">
        <h2 class="text-[clamp(1.4rem,3vw,2rem)] font-bold mb-4 leading-tight">
          Need to verify a certificate right now?
        </h2>
        <p class="text-[#a8d4d6] text-sm leading-relaxed mb-8 max-w-md mx-auto">
          Enter the Cert ID or upload the original PDF. Verification is instant and requires no account.
        </p>

        <!-- Inline cert-id input -->
        <form
          @submit.prevent="goVerify"
          class="flex flex-col sm:flex-row gap-3 max-w-sm mx-auto"
        >
          <input
            v-model="certId"
            type="text"
            placeholder="Enter Cert ID…"
            class="flex-1 px-4 py-2.5 rounded-lg text-sm text-[#28251d] bg-white placeholder-[#bab9b4]
                   border border-transparent focus:outline-none focus:ring-2 focus:ring-white/40"
          />
          <button
            type="submit"
            class="px-5 py-2.5 bg-white text-[#01696f] font-semibold text-sm rounded-lg
                   hover:bg-[#f9f8f5] transition-colors shrink-0"
          >
            Verify
          </button>
        </form>

        <p class="mt-4 text-xs text-[#a8d4d6]">
          Or
          <router-link to="/verify" class="underline underline-offset-2 hover:text-white transition-colors">
            upload a PDF certificate
          </router-link>
        </p>
      </div>
    </section>

  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const certId = ref('')

const goVerify = () => {
  if (certId.value.trim()) {
    router.push({ path: '/verify', query: { id: certId.value.trim() } })
  } else {
    router.push('/verify')
  }
}

const stats = [
  { value: '10,000+', label: 'Certificates issued' },
  { value: '<2s',     label: 'Avg. verify time' },
  { value: 'IPFS',    label: 'Decentralised storage' },
  { value: 'Cardano', label: 'Blockchain network' },
]

const steps = [
  {
    title: 'University registers',
    description:
      'The institution creates an account. Their domain is verified and a Cardano wallet is set up for signing.',
  },
  {
    title: 'Certificate is issued',
    description:
      'A PDF is generated, SHA-256 hashed, stored on IPFS, and the hash is anchored to Cardano — all in one action.',
  },
  {
    title: 'Anyone verifies',
    description:
      'Enter the Cert ID or scan the QR code. The hash is checked against Cardano on the spot — no account required.',
  },
]

const audiences = [
  {
    title: 'Universities',
    description:
      'Issue, manage, and batch-anchor academic certificates. Domain verification proves institutional authenticity.',
    cta: 'Register your university',
    route: '/register',
    iconBg: 'bg-[#cedcd8]',
    icon: {
      template: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
        <path d="M22 10v6M2 10l10-5 10 5-10 5z"/>
        <path d="M6 12v5c3 3 9 3 12 0v-5"/>
      </svg>`,
    },
  },
  {
    title: 'Employers & Companies',
    description:
      'Instantly verify any certificate without contacting the university. Request student profile access for hiring.',
    cta: 'Start verifying',
    route: '/verify',
    iconBg: 'bg-[#c6d8e4]',
    icon: {
      template: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#006494" stroke-width="2">
        <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
        <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
      </svg>`,
    },
  },
  {
    title: 'Students',
    description:
      'Own your credentials. Share selectively with employers. Approve or decline profile access requests at any time.',
    cta: 'Create student account',
    route: '/student/register',
    iconBg: 'bg-[#dacfde]',
    icon: {
      template: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#7a39bb" stroke-width="2">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
        <circle cx="12" cy="7" r="4"/>
      </svg>`,
    },
  },
]
</script>
