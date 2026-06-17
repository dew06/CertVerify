<template>
  <div class="min-h-screen bg-[#f7f6f2]">

    <!-- ── Top bar ──────────────────────────────────────────────────────── -->
    <header class="sticky top-0 z-40 h-16 bg-[#f7f6f2]/90 backdrop-blur border-b border-[#dcd9d5] flex items-center px-6">
      <div class="flex-1 flex items-center gap-3">
        <!-- Logo -->
        <router-link to="/" class="flex items-center gap-2 text-sm font-semibold text-[#28251d]">
          <svg width="26" height="26" viewBox="0 0 28 28" fill="none" aria-label="CertChain logo">
            <rect width="28" height="28" rx="6" fill="#01696f"/>
            <path d="M8 14 L12 18 L20 10" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <circle cx="14" cy="14" r="11" stroke="white" stroke-width="1.5" fill="none" opacity="0.4"/>
          </svg>
          CertChain
        </router-link>

        <span class="text-[#dcd9d5]">/</span>

        <div v-if="auth.user" class="text-sm text-[#7a7974]">
          <span class="font-medium text-[#28251d]">{{ auth.user.name }}</span>
          <span class="font-mono text-xs ml-1.5 text-[#7a7974]">{{ auth.user.domain }}</span>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <!-- Verified badge -->
        <span v-if="stats.is_verified"
              class="hidden sm:inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full
                     bg-[#d4dfcc] border border-[#437a22]/20 text-xs font-medium text-[#437a22]">
          <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M20 6 9 17l-5-5"/>
          </svg>
          Verified Domain
        </span>

        <!-- Logout -->
        <button
          @click="auth.logout"
          class="flex items-center gap-1.5 text-sm text-[#7a7974] hover:text-[#a12c7b] transition-colors"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
          Logout
        </button>
      </div>
    </header>

    <!-- ── Domain verification banner ───────────────────────────────────── -->
    <div v-if="!stats.is_verified"
         class="bg-[#fff8f4] border-b border-[#ddcfc6] px-6 py-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div class="flex items-center gap-2.5">
        <svg class="shrink-0 text-[#964219]" width="16" height="16" viewBox="0 0 24 24"
             fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
          <line x1="12" y1="9" x2="12" y2="13"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
        <p class="text-sm text-[#964219]">
          <span class="font-semibold">Domain not verified.</span>
          You cannot anchor certificates to the blockchain until your domain is verified.
        </p>
      </div>
      <router-link to="/verify-domain"
        class="shrink-0 inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold
               bg-[#964219] text-white rounded-lg hover:bg-[#713417] transition-colors">
        Verify now
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M5 12h14M12 5l7 7-7 7"/>
        </svg>
      </router-link>
    </div>

    <!-- ── Main content ──────────────────────────────────────────────────── -->
    <main class="max-w-6xl mx-auto px-6 py-10">

      <!-- Page heading -->
      <div class="mb-8">
        <h1 class="text-xl font-bold text-[#28251d] tracking-tight">Dashboard</h1>
        <p class="text-sm text-[#7a7974] mt-0.5">Overview of your institution's certificate activity.</p>
      </div>

      <!-- ── KPI cards ─────────────────────────────────────────────────── -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">

        <!-- Total Issued -->
        <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5">
          <div class="flex items-start justify-between mb-4">
            <p class="text-xs font-semibold uppercase tracking-widest text-[#7a7974]">Total Issued</p>
            <div class="w-8 h-8 rounded-lg bg-[#cedcd8] flex items-center justify-center shrink-0">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
            </div>
          </div>
          <p class="text-3xl font-bold tabular-nums text-[#28251d]">{{ stats.total_certificates || 0 }}</p>
        </div>

        <!-- Anchored on Chain -->
        <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5">
          <div class="flex items-start justify-between mb-4">
            <p class="text-xs font-semibold uppercase tracking-widest text-[#7a7974]">Anchored on Chain</p>
            <div class="w-8 h-8 rounded-lg bg-[#d4dfcc] flex items-center justify-center shrink-0">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#437a22" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
            </div>
          </div>
          <p class="text-3xl font-bold tabular-nums text-[#28251d]">{{ stats.anchored_certificates || 0 }}</p>
        </div>

        <!-- Pending Sync -->
        <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-5">
          <div class="flex items-start justify-between mb-4">
            <p class="text-xs font-semibold uppercase tracking-widest text-[#7a7974]">Pending Sync</p>
            <div class="w-8 h-8 rounded-lg bg-[#e9e0c6] flex items-center justify-center shrink-0">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#d19900" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/>
                <polyline points="1 20 1 14 7 14"/>
                <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
              </svg>
            </div>
          </div>
          <p class="text-3xl font-bold tabular-nums text-[#28251d]">{{ stats.pending_certificates || 0 }}</p>
        </div>
      </div>

      <!-- ── Quick actions ─────────────────────────────────────────────── -->
      <div class="bg-[#f9f8f5] border border-[#d4d1ca] rounded-xl p-6">
        <h2 class="text-sm font-semibold text-[#28251d] mb-5">Quick actions</h2>

        <div class="grid sm:grid-cols-3 gap-3">
          <!-- Issue Certificate -->
          <router-link to="/issue"
            class="group flex items-center gap-3 p-4 rounded-lg border border-[#d4d1ca]
                   bg-white hover:border-[#01696f] hover:shadow-sm transition-all">
            <div class="w-9 h-9 shrink-0 rounded-lg bg-[#cedcd8] flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#01696f" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="16"/>
                <line x1="8" y1="12" x2="16" y2="12"/>
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-[#28251d] group-hover:text-[#01696f] transition-colors">
                Issue Certificate
              </p>
              <p class="text-xs text-[#7a7974]">Single issuance</p>
            </div>
            <svg class="text-[#bab9b4] group-hover:text-[#01696f] transition-colors shrink-0"
                 width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M5 12h14M12 5l7 7-7 7"/>
            </svg>
          </router-link>

          <!-- Bulk CSV -->
          <router-link to="/bulk-upload"
            class="group flex items-center gap-3 p-4 rounded-lg border border-[#d4d1ca]
                   bg-white hover:border-[#01696f] hover:shadow-sm transition-all">
            <div class="w-9 h-9 shrink-0 rounded-lg bg-[#c6d8e4] flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#006494" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="12" y1="18" x2="12" y2="12"/>
                <line x1="9" y1="15" x2="15" y2="15"/>
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-[#28251d] group-hover:text-[#01696f] transition-colors">
                Bulk CSV Upload
              </p>
              <p class="text-xs text-[#7a7974]">Batch issuance</p>
            </div>
            <svg class="text-[#bab9b4] group-hover:text-[#01696f] transition-colors shrink-0"
                 width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M5 12h14M12 5l7 7-7 7"/>
            </svg>
          </router-link>

          <!-- Batch Anchor -->
          <router-link to="/batch-anchor"
            class="group flex items-center gap-3 p-4 rounded-lg border border-[#d4d1ca]
                   bg-white hover:border-[#01696f] hover:shadow-sm transition-all"
            :class="{ 'opacity-50 pointer-events-none': !stats.is_verified }"
          >
            <div class="w-9 h-9 shrink-0 rounded-lg bg-[#d4dfcc] flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#437a22" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-[#28251d] group-hover:text-[#01696f] transition-colors">
                Batch Anchor
              </p>
              <p class="text-xs text-[#7a7974]">Push to Cardano</p>
            </div>
            <svg class="text-[#bab9b4] group-hover:text-[#01696f] transition-colors shrink-0"
                 width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M5 12h14M12 5l7 7-7 7"/>
            </svg>
          </router-link>
        </div>

        <!-- Disabled hint when domain not verified -->
        <p v-if="!stats.is_verified" class="mt-3 text-xs text-[#7a7974]">
          <span class="text-[#964219] font-medium">Batch Anchor</span> is disabled until your domain is verified.
        </p>
      </div>

    </main>
  </div>
</template>


<script setup>
import { onMounted, ref } from 'vue';
import { useAuthStore } from '../stores/auth';
import api from '../services/api';
import axios from 'axios';


const auth = useAuthStore();
const stats = ref({});


onMounted(async () => {
    try {
        const res = await api.getCurrentUser(); 

        // Ensure you match the Go backend's response structure
        if (res.data && res.data.university) {
        stats.value = {
            // Fallback to 0 if statistics aren't in this specific endpoint
            total_issued: res.data.statistics?.total_certificates || 0,
            pending_anchors: res.data.statistics?.pending_count || 0,
            is_verified: res.data.university.is_verified
        };

        // Also update the store to keep everything in sync
        auth.university = res.data.university;
        }
    } catch (err) {
        console.error("Failed to load dashboard stats:", err.response?.data || err.message);
    }
    });
</script>
