<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-5 py-10 sm:px-6 sm:py-12 max-w-2xl">
      <Card>
        <CardHeader class="p-6 sm:p-8">
          <CardTitle class="text-xl sm:text-2xl font-bold">{{ isEdit ? 'Edit Exercise' : 'Create New Exercise' }}</CardTitle>
          <CardDescription class="text-sm sm:text-base font-medium mt-2">Add details about your exercise</CardDescription>
        </CardHeader>
        <CardContent class="p-6 sm:p-8">
          <form @submit.prevent="handleSubmit" class="space-y-6">
            <div class="space-y-3">
              <Label for="name">Exercise Name *</Label>
              <Input
                id="name"
                v-model="form.name"
                type="text"
                placeholder="e.g., Bench Press"
                required
                :disabled="loading"
              />
            </div>
            <div class="space-y-2">
              <Label for="exercise_type">Exercise Type *</Label>
              <select
                id="exercise_type"
                v-model="form.exercise_type"
                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                required
                :disabled="loading"
              >
                <option value="strength">Strength</option>
                <option value="cardio">Cardio</option>
              </select>
            </div>
            <div class="space-y-2">
              <Label for="muscle_group">Muscle Group</Label>
              <Input
                id="muscle_group"
                v-model="form.muscle_group"
                type="text"
                placeholder="e.g., Chest, Legs, Back"
                :disabled="loading"
              />
            </div>
            <div class="space-y-3">
              <Label for="equipment">Equipment</Label>
              <Input
                id="equipment"
                v-model="form.equipment"
                type="text"
                placeholder="e.g., Barbell, Dumbbell, Bodyweight"
                :disabled="loading"
              />
            </div>
            <div class="space-y-3">
              <Label for="description">Description</Label>
              <textarea
                id="description"
                v-model="form.description"
                class="flex min-h-[100px] w-full rounded-2xl border-2 border-input/50 glass-input dark:glass-input-dark px-5 py-4 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 focus-visible:shadow-lg focus-visible:shadow-primary/20 focus-visible:scale-[1.01]"
                placeholder="Brief description of the exercise"
                :disabled="loading"
              />
            </div>
            <div class="space-y-3">
              <Label for="instructions">Instructions</Label>
              <textarea
                id="instructions"
                v-model="form.instructions"
                class="flex min-h-[140px] w-full rounded-2xl border-2 border-input/50 glass-input dark:glass-input-dark px-5 py-4 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 focus-visible:shadow-lg focus-visible:shadow-primary/20 focus-visible:scale-[1.01]"
                placeholder="Step-by-step instructions"
                :disabled="loading"
              />
            </div>
            <div class="space-y-3">
              <Label for="video_link">Video Link</Label>
              <Input
                id="video_link"
                v-model="form.video_link"
                type="url"
                placeholder="https://..."
                :disabled="loading"
              />
            </div>
            <div class="space-y-3">
              <Label for="image_link">Image Link</Label>
              <Input
                id="image_link"
                v-model="form.image_link"
                type="url"
                placeholder="https://..."
                :disabled="loading"
              />
            </div>
            <div v-if="error" class="text-sm text-red-600 dark:text-red-400">
              {{ error }}
            </div>
            <div class="flex gap-4">
              <Button type="submit" variant="default" :disabled="loading">
                {{ loading ? 'Saving...' : (isEdit ? 'Update' : 'Create') }}
              </Button>
              <Button type="button" variant="outline" @click="$router.push('/exercises')" :disabled="loading">
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getExercise, createExercise, updateExercise } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const router = useRouter()
const route = useRoute()
const isEdit = ref(false)
const loading = ref(false)
const error = ref(null)

const form = ref({
  name: '',
  exercise_type: 'strength',
  muscle_group: '',
  equipment: '',
  description: '',
  instructions: '',
  video_link: '',
  image_link: ''
})

const loadExercise = async () => {
  if (route.params.id) {
    isEdit.value = true
    try {
      loading.value = true
      const response = await getExercise(route.params.id)
      form.value = {
        name: response.exercise.name || '',
        exercise_type: response.exercise.exercise_type || 'strength',
        muscle_group: response.exercise.muscle_group || '',
        equipment: response.exercise.equipment || '',
        description: response.exercise.description || '',
        instructions: response.exercise.instructions || '',
        video_link: response.exercise.video_link || '',
        image_link: response.exercise.image_link || ''
      }
    } catch (err) {
      error.value = 'Failed to load exercise'
      console.error(err)
    } finally {
      loading.value = false
    }
  }
}

const handleSubmit = async () => {
  try {
    loading.value = true
    error.value = null
    
    if (isEdit.value) {
      await updateExercise(route.params.id, form.value)
    } else {
      await createExercise(form.value)
    }
    
    router.push('/exercises')
  } catch (err) {
    error.value = err.message || 'Failed to save exercise'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadExercise()
})
</script>

