<template>
  <div class="max-w-4xl mx-auto py-10 px-4">
    <div v-if="loading" class="text-center py-20">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-700 mx-auto"></div>
      <p class="mt-4 text-gray-600">Verifying authenticity on-chain...</p>
    </div>

    <div v-else-if="error" class="bg-red-50 border-l-4 border-red-500 p-6 rounded-r-lg">
      <h3 class="text-red-800 font-bold text-lg">Verification Failed</h3>
      <p class="text-red-700 mt-2">{{ error }}</p>
      <router-link to="/verify" class="mt-4 inline-block text-red-600 underline font-medium">Try manual search</router-link>
    </div>

    <div v-else-if="certificate" class="space-y-8">
      <div class="bg-green-600 text-white p-6 rounded-t-2xl flex items-center justify-between shadow-lg">
        <div>
          <h1 class="text-2xl font-bold">Authentic Record</h1>
          <p class="text-green-100 text-sm">Issued by {{ certificate.university }}</p>
        </div>
        <div class="hidden md:block bg-white/20 px-4 py-2 rounded-full border border-white/30 text-xs">
          ID: {{ certificate.cert_id.substring(0, 12) }}...
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div class="lg:col-span-2 space-y-6">
          <div class="bg-white p-8 rounded-2xl shadow-sm border border-gray-100">
            <h2 class="text-gray-400 text-xs uppercase font-bold tracking-widest mb-6">Academic Achievement</h2>
            <div class="space-y-6">
              <div>
                <p class="text-sm text-gray-500">Recipient</p>
                <p class="text-3xl font-bold text-gray-900">{{ certificate.student_name }}</p>
              </div>
              <div>
                <p class="text-sm text-gray-500">Qualification</p>
                <p class="text-xl font-semibold text-gray-800">{{ certificate.degree }}</p>
              </div>
              <div class="flex gap-10">
                <div>
                  <p class="text-sm text-gray-500">GPA</p>
                  <p class="text-lg font-mono font-bold">{{ certificate.gpa }}</p>
                </div>
                <div>
                  <p class="text-sm text-gray-500">Issue Date</p>
                  <p class="text-lg font-medium">{{ formatDate(certificate.issue_date) }}</p>
                </div>
              </div>
            </div>

            <div class="mt-10 pt-6 border-t border-gray-100 flex gap-4">
              <button @click="downloadPDF" class="flex-1 bg-blue-600 text-white py-3 rounded-xl font-bold hover:bg-blue-700 transition">
                Download Original PDF
              </button>
              <a :href="'https://ipfs.io/ipfs/' + certificate.ipfs_pdf_hash" target="_blank" class="px-6 py-3 border border-gray-200 rounded-xl hover:bg-gray-50">
                🌐 IPFS
              </a>
            </div>
          </div>
        </div>

        <div class="space-y-6">
          <div class="bg-gray-900 text-white p-6 rounded-2xl shadow-xl">
            <div class="flex items-center gap-2 mb-4">
              <span class="text-blue-400">⛓️</span>
              <h3 class="font-bold">Cardano Proof</h3>
            </div>
            <div class="space-y-4 text-xs font-mono">
              <div class="p-3 bg-white/5 rounded border border-white/10 overflow-hidden">
                <p class="text-gray-500 mb-1">Transaction ID</p>
                <p class="truncate text-blue-300">{{ certificate.cardano_txid }}</p>
              </div>
              <a :href="'https://preview.cardanoscan.io/transaction/' + certificate.cardano_txid" 
                 target="_blank" 
                 class="block text-center py-2 bg-blue-600 rounded hover:bg-blue-500 transition font-sans text-sm">
                Explore Transaction
              </a>
            </div>
          </div>

          <div class="bg-white p-6 rounded-2xl border border-gray-100 shadow-sm text-center">
            <div v-if="certificate.university_verified" class="inline-flex items-center justify-center w-12 h-12 bg-green-100 text-green-600 rounded-full mb-3">
              🏛️
            </div>
            <h4 class="font-bold text-gray-900">{{ certificate.university }}</h4>
            <p class="text-xs text-gray-500 mt-1">{{ certificate.university_domain }}</p>
            <p v-if="certificate.university_verified" class="mt-3 text-[10px] text-green-600 font-bold uppercase tracking-tighter">
              Verified Institution
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import apiService from '../services/api';

const route = useRoute();
const certificate = ref(null);
const loading = ref(true);
const error = ref(null);

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
};

const downloadPDF = async () => {
  try {
    await apiService.downloadCertificate(certificate.value.cert_id, certificate.value.student_name);
  } catch (err) {
    alert("Download failed. The IPFS gateway might be busy.");
  }
};

onMounted(async () => {
  const certID = route.params.certID;
  try {
    const res = await apiService.getCertificate(certID);
    certificate.value = res.data;
  } catch (err) {
    error.value = "Certificate not found or invalid ID.";
  } finally {
    loading.value = false;
  }
});
</script>