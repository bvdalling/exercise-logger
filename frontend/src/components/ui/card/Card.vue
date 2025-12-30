<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <div
    :class="
      cn(
        'rounded-3xl text-card-foreground',
        'glass-card dark:glass-card-dark',
        'transition-all duration-300 ease-out',
        'hover:scale-[1.01] hover:shadow-2xl',
        'active:scale-[0.99]',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
