<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { reactiveOmit } from "@vueuse/core"
import { watch, onUnmounted } from "vue"
import { X } from "lucide-vue-next"
import {
  DialogClose,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  useForwardPropsEmits,
} from "reka-ui"
import { cn } from "@/lib/utils"

const props = defineProps<DialogContentProps & { class?: HTMLAttributes["class"] }>()
const emits = defineEmits<DialogContentEmits>()

const delegatedProps = reactiveOmit(props, "class")

const forwarded = useForwardPropsEmits(delegatedProps, emits)

// Prevent body scroll when modal is open
let originalBodyOverflow = ''

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    // Save original overflow and prevent body scroll
    originalBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
  } else {
    // Restore original overflow
    document.body.style.overflow = originalBodyOverflow
  }
})

onUnmounted(() => {
  // Ensure we restore body scroll on unmount
  document.body.style.overflow = originalBodyOverflow || ''
})
</script>

<template>
  <DialogPortal>
    <DialogOverlay
      class="fixed inset-0 z-50 glass-overlay data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
    />
    <DialogContent
      v-bind="forwarded"
      :class="
        cn(
          'fixed left-1/2 top-1/2 z-50 flex flex-col w-full max-w-lg max-h-[90vh] -translate-x-1/2 -translate-y-1/2 rounded-3xl duration-300 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]',
          'glass-dialog dark:glass-dialog-dark',
          'overflow-hidden',
          props.class,
        )"
    >
      <div class="flex-1 overflow-y-auto overflow-x-hidden min-h-0">
        <div class="p-8">
          <slot />
        </div>
      </div>

      <DialogClose
        class="absolute right-6 top-6 rounded-2xl w-10 h-10 flex items-center justify-center opacity-70 ring-offset-background transition-all duration-200 hover:opacity-100 hover:bg-accent/30 hover:scale-110 active:scale-95 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none min-w-[44px] min-h-[44px] z-10"
      >
        <X class="w-5 h-5" />
        <span class="sr-only">Close</span>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
