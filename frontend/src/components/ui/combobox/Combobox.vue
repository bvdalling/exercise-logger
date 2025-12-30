<template>
  <div class="relative" ref="containerRef">
    <div class="relative">
      <Input
        :id="inputId"
        :model-value="searchQuery"
        @update:model-value="handleInput"
        @focus="isOpen = true"
        @blur="handleBlur"
        @keydown="handleKeydown"
        :placeholder="placeholder"
        :disabled="disabled"
        :class="inputClass"
        autocomplete="off"
      />
      <button
        type="button"
        @click="toggleDropdown"
        class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
        :disabled="disabled"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4 transition-transform"
          :class="{ 'rotate-180': isOpen }"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
    </div>
    
    <div
      v-if="isOpen && filteredOptions.length > 0"
      class="absolute z-50 w-full mt-2 glass-strong dark:glass-strong rounded-2xl shadow-2xl max-h-60 overflow-auto border border-white/20"
    >
      <div
        v-for="(option, index) in filteredOptions"
        :key="getOptionValue(option)"
        @mousedown.prevent="selectOption(option)"
        :class="[
          'px-4 py-3 cursor-pointer text-base transition-all duration-200 ease-out',
          index === highlightedIndex ? 'bg-primary/20 text-primary font-semibold' : 'hover:bg-accent/30 hover:text-accent-foreground',
          getOptionValue(option) === modelValue ? 'bg-primary/15 text-primary font-semibold' : '',
          'first:rounded-t-2xl last:rounded-b-2xl'
        ]"
      >
        {{ getOptionLabel(option) }}
      </div>
    </div>
    
    <div
      v-if="isOpen && filteredOptions.length === 0 && searchQuery"
      class="absolute z-50 w-full mt-2 glass-strong dark:glass-strong rounded-2xl shadow-2xl p-4 text-base text-muted-foreground border border-white/20"
    >
      No exercises found
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Input } from '@/components/ui/input'

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: null
  },
  options: {
    type: Array,
    required: true
  },
  getOptionLabel: {
    type: Function,
    default: (option) => option?.name || option?.label || String(option)
  },
  getOptionValue: {
    type: Function,
    default: (option) => option?.id || option?.value || option
  },
  placeholder: {
    type: String,
    default: 'Search...'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  inputId: {
    type: String,
    default: ''
  },
  inputClass: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const containerRef = ref(null)
const isOpen = ref(false)
const searchQuery = ref('')
const highlightedIndex = ref(-1)

const selectedOption = computed(() => {
  if (!props.modelValue) return null
  return props.options.find(opt => props.getOptionValue(opt) === props.modelValue)
})

const displayValue = computed(() => {
  if (selectedOption.value) {
    return props.getOptionLabel(selectedOption.value)
  }
  return searchQuery.value
})

const filteredOptions = computed(() => {
  if (!searchQuery.value.trim()) {
    return props.options
  }
  
  const query = searchQuery.value.toLowerCase().trim()
  return props.options.filter(option => {
    const label = props.getOptionLabel(option).toLowerCase()
    return label.includes(query)
  })
})

const handleInput = (value) => {
  searchQuery.value = value
  isOpen.value = true
  highlightedIndex.value = -1
  
  // If user clears the input, clear selection
  if (!value.trim() && props.modelValue) {
    emit('update:modelValue', null)
    emit('change', null)
  }
}

const selectOption = (option) => {
  const value = props.getOptionValue(option)
  emit('update:modelValue', value)
  emit('change', value)
  searchQuery.value = props.getOptionLabel(option)
  isOpen.value = false
  highlightedIndex.value = -1
}

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value && !searchQuery.value && selectedOption.value) {
    searchQuery.value = props.getOptionLabel(selectedOption.value)
  }
}

const handleBlur = (e) => {
  // Delay to allow click events to fire
  setTimeout(() => {
    if (!containerRef.value?.contains(document.activeElement)) {
      isOpen.value = false
      // Reset search query to selected option label
      if (selectedOption.value) {
        searchQuery.value = props.getOptionLabel(selectedOption.value)
      } else {
        searchQuery.value = ''
      }
      highlightedIndex.value = -1
    }
  }, 200)
}

const handleKeydown = (e) => {
  if (!isOpen.value && (e.key === 'ArrowDown' || e.key === 'Enter')) {
    isOpen.value = true
    e.preventDefault()
    return
  }
  
  if (!isOpen.value) return
  
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      highlightedIndex.value = Math.min(highlightedIndex.value + 1, filteredOptions.value.length - 1)
      break
    case 'ArrowUp':
      e.preventDefault()
      highlightedIndex.value = Math.max(highlightedIndex.value - 1, -1)
      break
    case 'Enter':
      e.preventDefault()
      if (highlightedIndex.value >= 0 && filteredOptions.value[highlightedIndex.value]) {
        selectOption(filteredOptions.value[highlightedIndex.value])
      }
      break
    case 'Escape':
      e.preventDefault()
      isOpen.value = false
      highlightedIndex.value = -1
      if (selectedOption.value) {
        searchQuery.value = props.getOptionLabel(selectedOption.value)
      } else {
        searchQuery.value = ''
      }
      break
  }
}

// Update search query when modelValue changes externally
watch(() => props.modelValue, (newValue) => {
  if (newValue && selectedOption.value) {
    searchQuery.value = props.getOptionLabel(selectedOption.value)
  } else if (!newValue) {
    searchQuery.value = ''
  }
}, { immediate: true })

// Click outside to close
const handleClickOutside = (event) => {
  if (containerRef.value && !containerRef.value.contains(event.target)) {
    isOpen.value = false
    if (selectedOption.value) {
      searchQuery.value = props.getOptionLabel(selectedOption.value)
    } else {
      searchQuery.value = ''
    }
    highlightedIndex.value = -1
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (props.modelValue && selectedOption.value) {
    searchQuery.value = props.getOptionLabel(selectedOption.value)
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

