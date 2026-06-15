import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2026-06-15',
  srcDir: '.',
  css: ['~/assets/css/tailwind.css'],
  runtimeConfig: {
    apiInternalBase: '',
    public: {
      apiBase: 'http://localhost:8080/api/v1',
      siteName: 'Support Us'
    }
  },
  typescript: {
    strict: true
  },
  vite: {
    plugins: [tailwindcss()]
  },
  nitro: {
    preset: 'node-server'
  }
})
