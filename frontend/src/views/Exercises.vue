<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-5 py-10 sm:px-6 sm:py-12">
      <div v-if="loading" class="text-center py-12 text-lg">Loading...</div>
      <div v-else-if="exercises.length === 0" class="text-center py-12">
        <p class="text-muted-foreground mb-6 text-base sm:text-lg">No exercises yet. Create your first exercise!</p>
        <Button @click="$router.push('/exercises/new')" size="lg">Add Exercise</Button>
      </div>
      <div v-else class="grid gap-6 sm:gap-8 md:grid-cols-2 lg:grid-cols-3">
        <Card
          v-for="exercise in exercises"
          :key="exercise.id"
          class="cursor-pointer hover:shadow-2xl hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 ease-out"
          @click="openExerciseDetail(exercise.id)"
        >
          <CardHeader class="p-6 sm:p-8">
            <CardTitle class="text-lg sm:text-xl font-bold">{{ exercise.name }}</CardTitle>
            <CardDescription class="text-sm sm:text-base font-medium mt-2">
              <span v-if="exercise.muscle_group">{{ exercise.muscle_group }}</span>
              <span v-if="exercise.equipment"> • {{ exercise.equipment }}</span>
            </CardDescription>
          </CardHeader>
          <CardContent class="p-6 sm:p-8 pt-0">
            <p v-if="exercise.description" class="text-sm sm:text-base text-muted-foreground mb-6 line-clamp-2">
              {{ exercise.description }}
            </p>
            <div class="flex flex-wrap gap-3">
              <Button
                variant="outline"
                size="sm"
                @click.stop="$router.push(`/exercises/${exercise.id}/progress`)"
              >
                Progress
              </Button>
              <Button
                variant="outline"
                size="sm"
                @click.stop="$router.push(`/exercises/${exercise.id}/edit`)"
              >
                Edit
              </Button>
              <Button
                variant="destructive"
                size="sm"
                @click.stop="handleDelete(exercise.id)"
              >
                Delete
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

    <!-- Exercise Detail Dialog -->
    <Dialog v-model:open="dialogOpen">
      <DialogContent class="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ selectedExercise?.name || 'Exercise Details' }}</DialogTitle>
          <DialogDescription v-if="selectedExercise">
            <span v-if="selectedExercise.muscle_group">{{ selectedExercise.muscle_group }}</span>
            <span v-if="selectedExercise.equipment"> • {{ selectedExercise.equipment }}</span>
            <span v-if="selectedExercise.exercise_type"> • {{ selectedExercise.exercise_type }}</span>
          </DialogDescription>
        </DialogHeader>

        <div v-if="loadingExercise" class="text-center py-8">
          Loading exercise details...
        </div>

        <div v-else-if="selectedExercise" class="space-y-4">
          <!-- Description -->
          <div v-if="selectedExercise.description">
            <h3 class="text-sm font-semibold mb-2">Description</h3>
            <p class="text-sm text-muted-foreground whitespace-pre-wrap">{{ selectedExercise.description }}</p>
          </div>

          <!-- Instructions -->
          <div v-if="selectedExercise.instructions">
            <h3 class="text-sm font-semibold mb-2">Instructions</h3>
            <p class="text-sm text-muted-foreground whitespace-pre-wrap">{{ selectedExercise.instructions }}</p>
          </div>

          <!-- Image -->
          <div v-if="selectedExercise.image_link">
            <h3 class="text-sm font-semibold mb-2">Image</h3>
            <img
              :src="selectedExercise.image_link"
              :alt="selectedExercise.name"
              class="w-full rounded-lg object-cover max-h-64"
              @error="handleImageError"
            />
          </div>

          <!-- Video -->
          <div v-if="selectedExercise.video_link">
            <h3 class="text-sm font-semibold mb-2">Video</h3>
            <div class="space-y-2">
              <a
                :href="selectedExercise.video_link"
                target="_blank"
                rel="noopener noreferrer"
                class="text-sm text-primary hover:underline inline-flex items-center gap-1"
              >
                Watch Video
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
              </a>
              <!-- Try to embed YouTube videos -->
              <div v-if="isYouTubeLink(selectedExercise.video_link)" class="mt-2">
                <iframe
                  :src="getYouTubeEmbedUrl(selectedExercise.video_link)"
                  class="w-full aspect-video rounded-lg"
                  frameborder="0"
                  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                  allowfullscreen
                ></iframe>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            @click="dialogOpen = false"
          >
            Close
          </Button>
          <Button
            v-if="selectedExercise"
            @click="handleRecordExercise"
          >
            Record Exercise
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getExercises, getExercise, deleteExercise } from '@/composables/useApi'
import { useScrollRestore } from '@/composables/useScrollRestore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

const router = useRouter()
const exercises = ref([])
const loading = ref(true)
const dialogOpen = ref(false)
const selectedExercise = ref(null)
const loadingExercise = ref(false)

// Restore scroll position when modal closes
useScrollRestore(dialogOpen)

const loadExercises = async () => {
  try {
    loading.value = true
    const response = await getExercises()
    exercises.value = response.exercises
  } catch (error) {
    console.error('Failed to load exercises:', error)
  } finally {
    loading.value = false
  }
}

const openExerciseDetail = async (exerciseId) => {
  try {
    dialogOpen.value = true
    loadingExercise.value = true
    selectedExercise.value = null
    
    const response = await getExercise(exerciseId)
    selectedExercise.value = response.exercise
  } catch (error) {
    console.error('Failed to load exercise details:', error)
    alert('Failed to load exercise details')
    dialogOpen.value = false
  } finally {
    loadingExercise.value = false
  }
}

const handleRecordExercise = () => {
  if (selectedExercise.value) {
    router.push(`/log-workout?exerciseId=${selectedExercise.value.id}`)
  }
}

const handleDelete = async (id) => {
  if (!confirm('Are you sure you want to delete this exercise?')) {
    return
  }
  try {
    await deleteExercise(id)
    await loadExercises()
    // Close dialog if the deleted exercise was being viewed
    if (selectedExercise.value?.id === id) {
      dialogOpen.value = false
      selectedExercise.value = null
    }
  } catch (error) {
    console.error('Failed to delete exercise:', error)
    alert('Failed to delete exercise')
  }
}

const isYouTubeLink = (url) => {
  if (!url) return false
  return /(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})/.test(url)
}

const getYouTubeEmbedUrl = (url) => {
  if (!url) return ''
  const regex = /(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})/
  const match = url.match(regex)
  if (match && match[1]) {
    return `https://www.youtube.com/embed/${match[1]}`
  }
  return url
}

const handleImageError = (event) => {
  event.target.style.display = 'none'
}

onMounted(() => {
  loadExercises()
})
</script>

