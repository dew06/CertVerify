<template>
  <div class="min-h-screen bg-white">
    <header class="bg-gradient-to-r from-blue-700 to-indigo-900 text-white py-20 px-6">
      <div class="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between">
        <div class="md:w-1/2 space-y-6 text-center md:text-left">
          <h1 class="text-5xl font-extrabold leading-tight">
            Trust, Verified by <span class="text-blue-300">Blockchain</span>.
          </h1>
          <p class="text-xl text-blue-100">
            Secure, immutable, and instantly verifiable academic credentials powered by Cardano.
          </p>
          <div class="flex flex-col sm:flex-row gap-4 pt-4 justify-center md:justify-start">
            <router-link to="/verify" class="px-8 py-3 bg-white text-blue-700 font-bold rounded-lg shadow-lg hover:bg-blue-50 transition">
              Verify a Certificate
            </router-link>
            <router-link to="/login" class="px-8 py-3 bg-transparent border-2 border-white text-white font-bold rounded-lg hover:bg-white/10 transition">
              University Portal
            </router-link>
          </div>
        </div>
        <div class="md:w-1/2 mt-12 md:mt-0 flex justify-center">
          <div class="w-80 h-80 bg-white/10 rounded-full flex items-center justify-center border-4 border-white/20 animate-pulse">
             <span class="text-6xl">🎓</span>
          </div>
        </div>
      </div>
    </header>

    <section class="py-20 px-6 max-w-6xl mx-auto grid md:grid-cols-3 gap-12 text-center">
      <div class="p-6">
        <div class="text-4xl mb-4">🛡️</div>
        <h3 class="text-xl font-bold mb-2">Tamper Proof</h3>
        <p class="text-gray-600">Certificates are hashed and anchored to the Cardano blockchain, making them impossible to alter.</p>
      </div>
      <div class="p-6">
        <div class="text-4xl mb-4">⚡</div>
        <h3 class="text-xl font-bold mb-2">Instant Verification</h3>
        <p class="text-gray-600">Zero manual checks. Verify credentials in seconds via QR code or Cert ID.</p>
      </div>
      <div class="p-6">
        <div class="text-4xl mb-4">🌐</div>
        <h3 class="text-xl font-bold mb-2">Decentralized</h3>
        <p class="text-gray-600">Owned by the university, verified by the network, accessible globally.</p>
      </div>
    </section>
  </div>
</template>