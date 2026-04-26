<template>
  <div class="max-w-4xl mx-auto p-6 bg-white shadow-md rounded-lg mt-10 text-center">
    <h2 class="text-2xl font-bold mb-4">Bulk Certificate Upload</h2>
    <p class="text-gray-600 mb-6">Upload a CSV file containing: student_name, degree, gpa, email</p>
    
    <div class="border-2 border-dashed border-gray-300 p-10 rounded-lg">
      <input type="file" @change="handleFileChange" accept=".csv" class="mb-4" />
      <input v-model="password" type="password" placeholder="University Password" class="block w-full max-w-xs mx-auto p-2 border rounded mb-4" />
      <button @click="uploadCSV" :disabled="!file || loading" class="bg-indigo-600 text-white px-6 py-2 rounded shadow hover:bg-indigo-700">
        {{ loading ? 'Processing Batch...' : 'Upload & Process' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import api from '../services/api';
import { useAuthStore } from '../stores/auth';

const auth = useAuthStore();
const file = ref(null);
const password = ref('');
const loading = ref(false);

const handleFileChange = (e) => { file.value = e.target.files[0]; };

const uploadCSV = async () => {
  // 1. Guard against missing session
  if (!auth.university?.id) {
    alert("University session not found. Please log in again.");
    return;
  }

  // 2. Validate password presence
  if (!password.value) {
    alert("Please enter your password to authorize batch processing.");
    return;
  }

  const formData = new FormData();
  formData.append('csv_file', file.value); // Ensure key matches your Go backend ('file' or 'csv_file')
  formData.append('university_id', auth.university.id);
  formData.append('password', password.value);

  loading.value = true; 
  try {
    const res = await api.uploadBulkCSV(formData);
    
    alert(`${res.data.issued || 'Batch'} certificates successfully processed.`);
    
    // Clear the form after success
    file.value = null;
    password.value = '';
  } catch (err) {
    console.error("Upload error:", err.response?.data);
    alert(err.response?.data?.error || "Upload failed. Please check your CSV format.");
  } finally {
    loading.value = false;
  }
};
</script>