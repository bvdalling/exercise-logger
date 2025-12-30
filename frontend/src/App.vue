<template>
  <div id="app" class="h-screen bg-gradient-animated dark:bg-gradient-animated-dark overflow-hidden">
    <AppLayout v-if="showLayout">
      <router-view />
    </AppLayout>
    <router-view v-else />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'

const route = useRoute()

const showLayout = computed(() => {
  return route.meta.requiresAuth === true
})
</script>
