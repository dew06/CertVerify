<template>
  <nav class="bg-white shadow-sm border-b border-gray-200">
    <div class="container mx-auto px-4 flex justify-between items-center h-16">
      <div class="flex items-center gap-2">
        <router-link to="/" class="text-xl font-bold text-blue-700 flex items-center gap-2">
          <span class="text-2xl">🎓</span> CertVerify
        </router-link>
      </div>

      <div class="flex items-center gap-6 text-sm font-medium">
        <router-link to="/verify" class="text-gray-600 hover:text-blue-600">Verify</router-link>

        <template v-if="auth.isAuthenticated">
          <router-link to="/dashboard" class="text-gray-600 hover:text-blue-600">Dashboard</router-link>
          <router-link to="/issue" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700">
            Issue Cert
          </router-link>
          <button @click="handleLogout" class="text-red-500 hover:text-red-700">Logout</button>
        </template>

        <template v-else>
          <router-link to="/login" class="text-gray-600 hover:text-blue-600">University Login</router-link>
        </template>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';

const auth = useAuthStore();
const router = useRouter();

const handleLogout = () => {
  auth.logout();
  router.push('/');
};
</script>