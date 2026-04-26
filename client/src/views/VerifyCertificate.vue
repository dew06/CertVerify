<template>
  <div class="max-w-3xl mx-auto py-12 px-4">
    <div class="text-center mb-10">
      <h1 class="text-3xl font-bold text-gray-900">Certificate Verification Portal</h1>
      <p class="text-gray-600 mt-2">Enter details.</p>
    </div>

    <div class="flex justify-center mb-8 bg-gray-100 p-1 rounded-xl w-fit mx-auto">
      <button @click="mode = 'form'" :class="mode === 'form' ? 'bg-white shadow-sm' : ''" class="px-6 py-2 rounded-lg font-medium transition">Manual Details</button>
    </div>

    <div v-if="result" :class="result.valid ? 'border-green-500 bg-green-50' : 'border-red-500 bg-red-50'" class="mb-8 p-6 border-2 rounded-2xl animate-pulse-once">
      <div class="flex items-center gap-4 mb-4">
        <div :class="result.valid ? 'bg-green-500' : 'bg-red-500'" class="w-12 h-12 rounded-full flex items-center justify-center text-white text-2xl">
          {{ result.valid ? '✓' : '✕' }}
        </div>
        <div>
          <h2 class="text-xl font-bold" :class="result.valid ? 'text-green-800' : 'text-red-800'">
            {{ result.valid ? 'Authentic Certificate' : 'Verification Failed' }}
          </h2>
          <p class="text-sm" :class="result.valid ? 'text-green-600' : 'text-red-600'">{{ result.message }}</p>
        </div>
      </div>

      <div v-if="result.valid" class="grid grid-cols-2 gap-4 text-sm mt-6 border-t pt-4 border-green-200">
        <div><span class="text-gray-500">Student:</span> <p class="font-bold">{{ result.student_name }}</p></div>
        <div><span class="text-gray-500">University:</span> <p class="font-bold">{{ result.university }}</p></div>
        <div><span class="text-gray-500">Degree:</span> <p class="font-bold">{{ result.degree }}</p></div>
        <div><span class="text-gray-500">Issue Date:</span> <p class="font-bold">{{ result.issue_date }}</p></div>
      </div>
      
      <div v-if="result.valid && result.blockchain" class="mt-6 bg-white p-4 rounded-lg border border-green-200">
        <p class="text-xs font-bold text-gray-400 uppercase mb-2">Blockchain Proof</p>
        <p class="text-xs break-all font-mono">TX: {{ result.blockchain.cardano_txid }}</p>
        <a :href="result.blockchain.explorer_url" target="_blank" class="text-indigo-600 text-xs mt-2 inline-block font-bold">View on CardanoScan ↗</a>
      </div>
    </div>

    <div v-if="mode === 'form'" class="bg-white p-8 rounded-2xl shadow-sm border border-gray-100">
      <div class="space-y-4">
        <input v-model="formData.cert_id" placeholder="Certificate ID (e.g. CERT-123)" class="w-full p-3 border rounded-xl" />
        <input v-model="formData.student_name" placeholder="Student Full Name" class="w-full p-3 border rounded-xl" />
        <input v-model="formData.degree" placeholder="Degree Name" class="w-full p-3 border rounded-xl" />
        <button @click="verifyByDetails" :disabled="loading" class="w-full bg-indigo-600 text-white py-4 rounded-xl font-bold">
          {{ loading ? 'Verifying...' : 'Verify Authenticity' }}
        </button>
      </div>
    </div>

    
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import api from '../services/api';

const route = useRoute();
const mode = ref('form');
const loading = ref(false);
const result = ref(null);
const formData = ref({ cert_id: '', student_name: '', degree: '' });

// 1. Auto-verify if ID is in URL (for QR codes)
onMounted(async () => {
  // 1. Check for the certID in the URL
  if (route.params.certID) {
    formData.value.cert_id = route.params.certID;
    
    // 2. Automatically trigger the verification for QR scans
    await verifyAuto(route.params.certID);
  }
});

// New function specifically for QR redirects
const verifyAuto = async (id) => {
  loading.value = true;
  result.value = null;
  try {
    const res = await api.get(`/verify/certificate/${id}`);
    result.value = res.data;
  } catch (err) {
    console.error("QR Verification failed", err);
    result.value = { 
      valid: false, 
      message: "The scanned certificate could not be found or verified on-chain." 
    };
  } finally {
    loading.value = false;
  }
};

const verifyByDetails = async () => {
  loading.value = true;
  result.value = null;
  try {
    const res = await api.post('/verify', formData.value);
    result.value = res.data;
  } catch (err) {
    result.value = { valid: false, message: "Certificate not found in our records." };
  } finally {
    loading.value = false;
  }
};

const handlePdfUpload = async (e) => {
  const file = e.target.files[0];
  if (!file) return;

  const data = new FormData();
  data.append('certificate', file);

  loading.value = true;
  result.value = null;
  try {
    const res = await api.post('/verify/pdf', data);
    result.value = res.data;
  } catch (err) {
    result.value = { valid: false, message: "Could not verify PDF file." };
  } finally {
    loading.value = false;
  }
};
</script>