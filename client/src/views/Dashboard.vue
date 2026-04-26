<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-7xl mx-auto">
      <div class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-2xl font-bold text-gray-800">University Dashboard</h1>
          <p class="text-gray-600" v-if="auth.user">{{ auth.user.name }} ({{ auth.user.domain }})</p>
        </div>
        <div class="flex gap-4">
          <span v-if="stats.is_verified" class="bg-green-100 text-green-800 px-3 py-1 rounded-full text-sm font-medium flex items-center">
            ✅ Verified Domain
          </span>
          <button @click="auth.logout" class="text-red-600 hover:text-red-800 font-medium">Logout</button>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div class="bg-white p-6 rounded-xl shadow-sm border-l-4 border-blue-500">
          <p class="text-sm text-gray-500 uppercase font-bold">Total Issued</p>
          <p class="text-3xl font-bold">{{ stats.total_certificates || 0 }}</p>
        </div>
        <div class="bg-white p-6 rounded-xl shadow-sm border-l-4 border-green-500">
          <p class="text-sm text-gray-500 uppercase font-bold">Anchored on Chain</p>
          <p class="text-3xl font-bold">{{ stats.anchored_certificates || 0 }}</p>
        </div>
        <div class="bg-white p-6 rounded-xl shadow-sm border-l-4 border-yellow-500">
          <p class="text-sm text-gray-500 uppercase font-bold">Pending Sync</p>
          <p class="text-3xl font-bold">{{ stats.pending_certificates || 0 }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div class="bg-white p-8 rounded-xl shadow-sm">
          <h2 class="text-lg font-bold mb-4">Quick Actions</h2>
          <div class="space-y-4">
            <router-link to="/issue" class="block w-full p-4 text-center border-2 border-blue-100 rounded-lg hover:bg-blue-50 transition font-medium">
              ➕ Issue New Certificate
            </router-link>
            <router-link to="/bulk-upload" class="block w-full p-4 text-center border-2 border-indigo-100 rounded-lg hover:bg-indigo-50 transition font-medium">
              📄 Bulk CSV Upload
            </router-link>
            <router-link to="/batch-anchor" class="block w-full p-4 text-center border-2 border-green-100 rounded-lg hover:bg-green-50 transition font-medium">
              ⛓️ Batch Anchor to Cardano
            </router-link>
          </div>
        </div>

        <!-- <div class="bg-white p-8 rounded-xl shadow-sm">
          <h2 class="text-lg font-bold mb-4">University Wallet</h2>
          <div class="bg-gray-50 p-4 rounded-lg break-all">
            <p class="text-xs text-gray-400 mb-1">Cardano Preview Address</p>
            <p class="text-sm font-mono text-gray-700">{{ auth.university?.cardano_address || 'Not Available' }}</p>
          </div>
          <div class="mt-4 p-4 bg-yellow-50 rounded-lg border border-yellow-100">
            <p class="text-xs text-yellow-800 font-bold">Notice:</p>
            <p class="text-xs text-yellow-700">Ensure your wallet has test ADA to cover transaction fees for anchoring certificates.</p>
          </div>
        </div> -->
        <div v-if="!stats.is_verified" class="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-8 flex justify-between items-center">
            <div class="flex items-center">
                <span class="text-yellow-700 font-bold mr-2">⚠️ Attention:</span>
                <p class="text-yellow-700 text-sm">Your domain is not verified. You cannot anchor certificates to the blockchain yet.</p>
            </div>
            <router-link to="/verify-domain" class="bg-yellow-700 text-white px-4 py-1 rounded text-xs font-bold hover:bg-yellow-800">
                Verify Now
            </router-link>
            </div>
      </div>
    </div>
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