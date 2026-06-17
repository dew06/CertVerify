<template>
  <footer class="bg-[#1c1b19] text-[#cdccca] border-t border-[#262523]">
    <div class="max-w-6xl mx-auto px-6 py-12">

      <!-- Top row: brand + nav columns -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-8 pb-10 border-b border-[#262523]">

        <!-- Brand -->
        <div class="col-span-2 md:col-span-1">
          <router-link to="/" class="inline-flex items-center gap-2 text-sm font-semibold text-[#cdccca] mb-3">
            <svg width="26" height="26" viewBox="0 0 28 28" fill="none" aria-label="CertChain logo">
              <rect width="28" height="28" rx="6" fill="#01696f"/>
              <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
              <circle cx="14" cy="14" r="11" stroke="white" stroke-width="1.5" fill="none" opacity="0.4"/>
            </svg>
            CertChain
          </router-link>
          <p class="text-xs text-[#5a5957] leading-relaxed max-w-[180px]">
            Tamper-proof academic credentials anchored to the Cardano blockchain.
          </p>

          <!-- Cardano badge -->
          <div class="mt-4 inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-[#393836] text-xs text-[#5a5957]">
            <span class="w-1.5 h-1.5 rounded-full bg-[#4f98a3] inline-block"></span>
            Cardano Preview Network
          </div>
        </div>

        <!-- Platform -->
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-[#797876] mb-4">Platform</p>
          <ul class="space-y-2.5 text-sm">
            <li>
              <router-link to="/verify" class="text-[#5a5957] hover:text-[#cdccca] transition-colors">
                Verify Certificate
              </router-link>
            </li>
            <li>
              <router-link to="/register" class="text-[#5a5957] hover:text-[#cdccca] transition-colors">
                Register University
              </router-link>
            </li>
            <li>
              <router-link to="/login" class="text-[#5a5957] hover:text-[#cdccca] transition-colors">
                University Portal
              </router-link>
            </li>
          </ul>
        </div>

        <!-- Resources -->
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-[#797876] mb-4">Resources</p>
          <ul class="space-y-2.5 text-sm">
            <li>
              <a
                href="/swagger/index.html"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[#5a5957] hover:text-[#cdccca] transition-colors"
              >
                API Documentation
              </a>
            </li>
            <li>
              <a
                href="https://cardano.org"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[#5a5957] hover:text-[#cdccca] transition-colors inline-flex items-center gap-1"
              >
                About Cardano
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
                  <polyline points="15 3 21 3 21 9"/>
                  <line x1="10" y1="14" x2="21" y2="3"/>
                </svg>
              </a>
            </li>
          </ul>
        </div>

        <!-- Legal -->
        <div>
          <p class="text-xs font-semibold uppercase tracking-widest text-[#797876] mb-4">Legal</p>
          <ul class="space-y-2.5 text-sm">
            <li>
              <router-link to="/privacy" class="text-[#5a5957] hover:text-[#cdccca] transition-colors">
                Privacy Policy
              </router-link>
            </li>
            <li>
              <router-link to="/terms" class="text-[#5a5957] hover:text-[#cdccca] transition-colors">
                Terms of Use
              </router-link>
            </li>
          </ul>
        </div>
      </div>

      <!-- Bottom row: copyright -->
      <div class="pt-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-[#5a5957]">
        <span>&copy; 2026 UOM AI Specialization Project. All rights reserved.</span>
        <div class="flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full bg-[#437a22] inline-block"></span>
          <span>Network: Cardano Preview</span>
        </div>
      </div>

    </div>
  </footer>
</template>