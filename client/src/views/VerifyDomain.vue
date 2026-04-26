<template>
  <div class="max-w-4xl mx-auto py-10 px-4">
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
      <div class="bg-indigo-600 p-6 text-white">
        <h1 class="text-2xl font-bold text-white">Domain Verification</h1>
        <p class="text-indigo-100 mt-1">Prove ownership of <strong>{{ auth.university?.domain }}</strong></p>
      </div>

      <div class="p-8 space-y-8">
        <div class="flex gap-6">
          <div class="flex-shrink-0 w-10 h-10 bg-indigo-100 text-indigo-600 rounded-full flex items-center justify-center font-bold">1</div>
          <div class="flex-1">
            <h3 class="font-bold text-gray-900">Create Verification File</h3>
            <p class="text-sm text-gray-600 mb-4">Create a file named <code class="bg-gray-100 px-1 rounded text-red-600">cardano-key.json</code> with the following content:</p>
            
            <div class="relative">
              <pre class="bg-gray-900 text-green-400 p-4 rounded-lg text-xs overflow-x-auto">{{ JSON.stringify(verificationData?.file_content, null, 2) }}</pre>
              <button @click="copyToClipboard" class="absolute top-2 right-2 bg-white/10 hover:bg-white/20 text-white text-[10px] px-2 py-1 rounded">
                Copy JSON
              </button>
            </div>
          </div>
        </div>

        <div class="flex gap-6">
          <div class="flex-shrink-0 w-10 h-10 bg-indigo-100 text-indigo-600 rounded-full flex items-center justify-center font-bold">2</div>
          <div class="flex-1">
            <h3 class="font-bold text-gray-900">Upload to Server</h3>
            <p class="text-sm text-gray-600">Upload the file to your university's web server at the following path:</p>
            <p class="mt-2 font-mono text-sm bg-blue-50 p-2 border border-blue-100 rounded text-blue-700">
              https://{{ auth.university?.domain }}/.well-known/cardano-key.json
            </p>
          </div>
        </div>

        <div class="flex gap-6">
          <div class="flex-shrink-0 w-10 h-10 bg-indigo-100 text-indigo-600 rounded-full flex items-center justify-center font-bold">3</div>
          <div class="flex-1">
            <h3 class="font-bold text-gray-900">Confirm Verification</h3>
            <p class="text-sm text-gray-600 mb-4">Once the file is publicly accessible, click the button below.</p>
            
            <button 
              @click="handleVerify" 
              :disabled="loading"
              class="bg-indigo-600 text-white px-8 py-3 rounded-xl font-bold hover:bg-indigo-700 transition disabled:bg-gray-400"
            >
              {{ loading ? 'Checking Server...' : 'Verify Domain Now' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useAuthStore } from '../stores/auth';
import api from '../services/api';
import { useRouter } from 'vue-router';

const auth = useAuthStore();
const router = useRouter();
const verificationData = ref(null);
const loading = ref(false);

// Change these lines in VerifyDomain.vue
const fetchInstructions = async () => {
  // Use university instead of user
  if (!auth.university?.id) {
    console.error("University ID not found in store");
    return;
  }
  
  try {
    const res = await api.getDomainProof(auth.university.id);
    verificationData.value = res.data;
  } catch (err) {
    console.error("Failed to load instructions:", err);
  }
};

const handleVerify = async () => {
  if (!auth.university?.id) {
    alert("University ID missing. Please log in again.");
    return;
  }

  loading.value = true;
  try {
    const res = await api.verifyDomain(auth.university.id);
    alert(res.data.message);
    auth.university.is_verified = true;
    localStorage.setItem('university', JSON.stringify(auth.university));
    router.push('/dashboard');
  } catch (err) {
    alert("Verification failed.");
  } finally {
    loading.value = false;
  }
};

const copyToClipboard = () => {
  navigator.clipboard.writeText(JSON.stringify(verificationData.value.file_content, null, 2));
  alert("Copied to clipboard!");
};

onMounted(fetchInstructions);
</script>