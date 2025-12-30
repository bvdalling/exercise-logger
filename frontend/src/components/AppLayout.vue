<template>
  <div class="flex flex-col h-screen bg-gradient-colorful dark:bg-gradient-colorful-dark overflow-hidden">
    <!-- Top Bar -->
    <header ref="headerRef" class="fixed top-0 left-0 right-0 z-50 glass-nav dark:glass-nav-dark flex-shrink-0">
      <div class="flex items-center justify-between px-5 h-16">
        <!-- Back Button -->
        <button
          v-if="route.path !== '/'"
          @click="handleBack"
          class="flex items-center justify-center w-12 h-12 rounded-2xl hover:bg-muted/50 active:scale-95 transition-all duration-200 min-w-[44px] min-h-[44px]"
          aria-label="Go back"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-6 w-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 19l-7-7 7-7"
            />
          </svg>
        </button>
        <div v-else class="w-12"></div>

        <!-- Page Title -->
        <h1 class="text-xl font-bold text-center flex-1">
          {{ pageTitle }}
        </h1>

        <!-- Create Button -->
        <div class="w-10 flex items-center justify-end">
          <Button
            v-if="createButtonConfig && route.path !== '/'"
            @click="handleCreate"
            size="sm"
            variant="ghost"
            class="h-8 w-8 p-0 flex items-center justify-center text-lg font-bold"
            aria-label="Add"
          >
            +
          </Button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main 
      ref="mainRef"
      class="flex-1 overflow-y-auto overflow-x-hidden"
      :style="{ 
        paddingTop: `${headerHeight}px`,
        paddingBottom: `${footerHeight}px`,
        height: `calc(100vh - ${headerHeight + footerHeight}px)`
      }"
    >
      <slot />
    </main>

    <!-- Bottom Tab Bar -->
    <nav ref="footerRef" class="fixed bottom-0 left-0 right-0 z-50 glass-nav-footer border-t border-white/10 flex-shrink-0">
      <div class="flex items-center justify-around h-20 px-2">
        <!-- Exercises Tab -->
        <button
          @click="$router.push('/exercises')"
          :class="[
            'flex flex-col items-center justify-center flex-1 h-full transition-all duration-300 ease-out rounded-2xl min-h-[44px]',
            isActiveTab('/exercises') ? 'text-primary scale-110 font-semibold' : 'text-muted-foreground hover:text-primary/80 hover:scale-105 active:scale-95'
          ]"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-7 w-7 mb-1"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
          <span class="text-xs font-medium">Exercises</span>
        </button>

        <!-- Dashboard Tab -->
        <button
          @click="$router.push('/')"
          :class="[
            'flex flex-col items-center justify-center flex-1 h-full transition-all duration-300 ease-out rounded-2xl min-h-[44px]',
            isActiveTab('/') ? 'text-primary scale-110 font-semibold' : 'text-muted-foreground hover:text-primary/80 hover:scale-105 active:scale-95'
          ]"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-7 w-7 mb-1"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
            />
          </svg>
          <span class="text-xs font-medium">Dashboard</span>
        </button>

        <!-- Log Tab -->
        <button
          @click="$router.push('/log-workout')"
          :class="[
            'flex flex-col items-center justify-center flex-1 h-full transition-all duration-300 ease-out rounded-2xl min-h-[44px]',
            isActiveTab('/log-workout') ? 'text-primary scale-110 font-semibold' : 'text-muted-foreground hover:text-primary/80 hover:scale-105 active:scale-95'
          ]"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-7 w-7 mb-1"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 4v16m8-8H4"
            />
          </svg>
          <span class="text-xs font-medium">Log</span>
        </button>
      </div>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Button } from '@/components/ui/button'

const router = useRouter()
const route = useRoute()

const headerRef = ref(null)
const footerRef = ref(null)
const mainRef = ref(null)
const headerHeight = ref(64) // h-16 = 64px
const footerHeight = ref(80) // h-20 = 80px

const updateHeights = () => {
  if (headerRef.value) {
    headerHeight.value = headerRef.value.offsetHeight
  }
  if (footerRef.value) {
    footerHeight.value = footerRef.value.offsetHeight
  }
}

onMounted(() => {
  updateHeights()
  // Update on resize
  window.addEventListener('resize', updateHeights)
  // Use ResizeObserver for more accurate height tracking
  if (headerRef.value && footerRef.value) {
    const headerObserver = new ResizeObserver(() => updateHeights())
    const footerObserver = new ResizeObserver(() => updateHeights())
    headerObserver.observe(headerRef.value)
    footerObserver.observe(footerRef.value)
    
    onUnmounted(() => {
      headerObserver.disconnect()
      footerObserver.disconnect()
    })
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', updateHeights)
})

const pageTitle = computed(() => {
  return route.meta.title || route.name || 'Simple Logger'
})

const isActiveTab = (path) => {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}

const handleBack = () => {
  router.back()
}

const createButtonConfig = computed(() => {
  const path = route.path
  
  // Don't show create button on form pages
  if (path === '/exercises/new' || path.includes('/edit') || path.includes('/progress')) {
    return null
  }
  
  if (path === '/exercises') {
    return {
      label: 'Add Exercise',
      action: () => router.push('/exercises/new')
    }
  }
  
  if (path === '/log-workout') {
    return {
      label: 'Add Log',
      action: () => {
        // Navigate to a fresh log entry (will reset form if coming from edit page)
        router.push('/log-workout')
      }
    }
  }
  
  return null
})

const handleCreate = () => {
  if (createButtonConfig.value) {
    createButtonConfig.value.action()
  }
}
</script>

