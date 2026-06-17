<template>
  <div id="app" class="min-h-screen bg-gray-50">
    <Navbar v-if="!isAuthPage && !isDashboardPage" />

    <!-- Dashboard pages manage their own layout entirely -->
    <router-view v-if="isDashboardPage" />

    <!-- All other pages get the standard container wrapper -->
    <main v-else class="container mx-auto px-4 py-8">
      <router-view />
    </main>

    <Footer v-if="!isAuthPage && !isDashboardPage" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'

const route = useRoute()

const isAuthPage = computed(() =>
  ['Login', 'Register'].includes(route.name)
)

const isDashboardPage = computed(() =>
  ['StudentDashboard', 'CompanyDashboard', 'UniversityDashboard'].includes(route.name)
)
</script>