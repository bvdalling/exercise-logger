<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <p :class="cn('text-sm sm:text-base text-muted-foreground font-medium', props.class)">
    <slot />
  </p>
</template>
