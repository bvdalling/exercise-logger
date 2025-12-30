import { watch, nextTick } from 'vue'

/**
 * Composable to save and restore scroll position when opening/closing modals
 * @param {import('vue').Ref<boolean>} isOpen - Ref that tracks if the modal is open
 */
export function useScrollRestore(isOpen) {
  let savedScrollPosition = 0

  watch(isOpen, (newValue) => {
    if (newValue) {
      // Modal is opening - save current scroll position
      // Always use the main element since we've fixed the layout
      const mainElement = document.querySelector('main')
      if (mainElement) {
        savedScrollPosition = mainElement.scrollTop
      }
    } else {
      // Modal is closing - restore scroll position
      nextTick(() => {
        // Use requestAnimationFrame to ensure DOM has updated
        requestAnimationFrame(() => {
          const mainElement = document.querySelector('main')
          if (mainElement) {
            mainElement.scrollTo({
              top: savedScrollPosition,
              behavior: 'instant'
            })
          }
        })
      })
    }
  })
}

