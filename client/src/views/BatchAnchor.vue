<template>
  <div class="max-w-4xl mx-auto p-6 bg-white shadow-md rounded-lg mt-10">
    <h2 class="text-2xl font-bold mb-4 text-green-700">Anchor Certificates to Cardano</h2>
    <p class="mb-6 text-gray-600">This will group all pending certificates into a Merkle Tree and record the root on the blockchain.</p>
    
    <div class="space-y-4">
      <div class="flex items-center gap-2">
        <input type="checkbox" v-model="sendEmails" id="email" />
        <label for="email">Send automated emails to students with PDF attachments</label>
      </div>
      <input v-model="password" type="password" placeholder="Confirm Wallet Password" class="w-full p-2 border rounded" />
      <button @click="startAnchoring" :disabled="loading" class="w-full bg-green-600 text-white py-4 font-bold rounded shadow-lg hover:bg-green-700">
        {{ loading ? 'Communicating with Cardano Network...' : 'Confirm Blockchain Anchor' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/auth';

const auth = useAuthStore();
const password = ref('');
const sendEmails = ref(true);
const loading = ref(false);

const startAnchoring = async () => {
    if (!auth.university?.id) {
        alert("University session not found. Please log in again.");
        return;
    }

    if (!password.value) {
        alert("Please enter your wallet password.");
        return;
    }
  loading.value = true;
  try {
    const res = await api.batchAnchor({
      university_id: auth.university.id, 
      password: password.value,
      send_emails: sendEmails.value
    });

    alert(`Success! Root anchored to Cardano.`);
    
    password.value = ''; 
  } catch (err) {
    console.error("Anchoring Error:", err.response?.data);
    alert(err.response?.data?.error || "Anchoring failed. Check if you have enough tADA.");
  } finally {
    loading.value = false;
  }
};

</script>