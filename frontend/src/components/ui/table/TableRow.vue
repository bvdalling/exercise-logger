<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <tr :class="cn('border-b transition-all duration-200 hover:bg-muted/30 data-[state=selected]:bg-muted/50 backdrop-blur-sm', props.class)">
    <slot />
  </tr>
</template>
