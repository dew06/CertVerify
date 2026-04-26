<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4">
    <div class="max-w-md w-full space-y-8 bg-white p-10 rounded-xl shadow-lg border border-gray-100">
      
      <div v-if="!registrationSuccess">
        <h2 class="text-center text-3xl font-extrabold text-gray-900">Register Institution</h2>
        <p class="text-center text-sm text-gray-500 mt-2">A Cardano wallet will be generated for you.</p>
        
        <form class="mt-8 space-y-4" @submit.prevent="handleRegister">
          <input v-model="form.name" type="text" placeholder="University Name" required class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500">
          <input v-model="form.email" type="email" placeholder="Admin Email" required class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500">
          <input v-model="form.domain" type="text" placeholder="Official Domain (e.g. uom.lk)" required class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500">
          
          <div class="bg-blue-50 p-3 rounded-md text-[10px] text-blue-700">
            🛡️ Your password will be used to encrypt your blockchain private key. Choose a strong one!
          </div>
          <input v-model="form.password" type="password" placeholder="Master Password (min 12 chars)" required class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500">
          
          <button type="submit" :disabled="loading" class="w-full py-3 bg-blue-700 text-white rounded-md font-bold hover:bg-blue-800 disabled:bg-blue-300">
            {{ loading ? 'Generating Wallet...' : 'Register & Create Wallet' }}
          </button>
        </form>
      </div>

      <div v-else class="text-center space-y-6">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-green-100 text-green-600 rounded-full text-3xl">
          ✅
        </div>
        <h2 class="text-2xl font-bold text-gray-900">Wallet Created!</h2>
        
        <div class="bg-red-50 p-4 rounded-lg border border-red-100 text-left">
          <p class="text-xs font-bold text-red-700 uppercase mb-2">⚠️ Recovery Phrase (Save this now!)</p>
          <p class="text-sm font-mono bg-white p-3 rounded border border-red-200 text-gray-800 break-words">
            {{ walletInfo.recovery_phrase }}
          </p>
          <p class="text-[10px] text-red-600 mt-2">This will never be shown again. If you lose this, you lose access to your certificates.</p>
        </div>

        <div class="text-left bg-gray-50 p-3 rounded text-xs font-mono">
          <p class="text-gray-400">Public Address:</p>
          <p class="truncate text-gray-600">{{ walletInfo.wallet_address }}</p>
        </div>

        <router-link to="/login" class="block w-full py-3 bg-gray-900 text-white rounded-md font-bold">
          Proceed to Login
        </router-link>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import api from '../services/api';

const loading = ref(false);
const registrationSuccess = ref(false);
const walletInfo = ref({});
const form = ref({ name: '', email: '', domain: '', password: '' });

const handleRegister = async () => {
  loading.value = true;
  try {
    const res = await api.register(form.value);
    walletInfo.value = res.data;
    registrationSuccess.value = true;
  } catch (err) {
    alert(err.response?.data?.error || "Registration failed.");
  } finally {
    loading.value = false;
  }
};
</script>