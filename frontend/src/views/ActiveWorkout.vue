<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-4 py-6 max-w-4xl">
      <!-- Exercise Selection View -->
      <div v-if="!selectedExercise" class="space-y-6">
        <div class="text-center mb-6">
          <h1 class="text-2xl font-bold mb-2">Active Workout</h1>
          <p class="text-muted-foreground">Select an exercise to log</p>
        </div>

        <div v-if="exercisesLoading" class="text-center py-12 text-lg">Loading exercises...</div>
        
        <div v-else-if="exercises.length === 0" class="text-center py-12">
          <p class="text-muted-foreground mb-6">No exercises available</p>
        </div>

        <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <button
            v-for="exercise in exercises"
            :key="exercise.id"
            @click="selectExercise(exercise)"
            :class="[
              'p-6 rounded-2xl text-left transition-all duration-200 ease-out',
              'hover:scale-105 active:scale-95 touch-manipulation',
              'glass-card dark:glass-card-dark',
              exercise.isPublic ? 'border-2 border-primary/50' : ''
            ]"
          >
            <div class="flex items-start justify-between mb-2">
              <h3 class="text-lg font-bold flex-1">{{ exercise.name }}</h3>
              <span
                v-if="exercise.isPublic"
                class="text-xs px-2 py-1 rounded-full bg-primary/20 text-primary font-semibold ml-2"
              >
                Platform
              </span>
            </div>
            <div class="text-sm text-muted-foreground">
              <span v-if="exercise.muscle_group">{{ exercise.muscle_group }}</span>
              <span v-if="exercise.equipment"> • {{ exercise.equipment }}</span>
            </div>
            <div class="mt-2 text-xs text-muted-foreground capitalize">
              {{ exercise.exercise_type }}
            </div>
          </button>
        </div>
      </div>

      <!-- Quick Log View -->
      <div v-else class="space-y-6">
        <!-- Exercise Header -->
        <div class="flex items-center justify-between mb-6">
          <button
            @click="selectedExercise = null"
            class="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            Back
          </button>
          <h2 class="text-xl font-bold">{{ selectedExercise.name }}</h2>
          <div class="w-20"></div>
        </div>

        <!-- Last Workout Values -->
        <Card v-if="lastValues" class="bg-muted/50">
          <CardHeader class="pb-3">
            <CardTitle class="text-base">Last: {{ formatDate(lastValues.date) }}</CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="selectedExercise.exercise_type === 'strength'" class="space-y-2 text-sm">
              <div v-if="lastValues.weight_per_set && lastValues.weight_per_set.length">
                <div class="font-semibold mb-2">Sets:</div>
                <div v-for="(set, idx) in lastValues.weight_per_set" :key="idx" class="ml-4">
                  Set {{ idx + 1 }}:
                  <span v-if="set.reps">{{ set.reps }} reps</span>
                  <span v-if="set.weight"> @ {{ set.weight }}lbs</span>
                  <span v-if="set.perceived_effort"> ({{ set.perceived_effort }}/10)</span>
                </div>
              </div>
              <div v-else>
                <div v-if="lastValues.sets">Sets: {{ lastValues.sets }}</div>
                <div v-if="lastValues.reps">Reps: {{ lastValues.reps }}</div>
                <div v-if="lastValues.weight">Weight: {{ lastValues.weight }}lbs</div>
              </div>
            </div>
            <div v-else class="space-y-1 text-sm">
              <div v-if="lastValues.distance">Distance: {{ lastValues.distance }}mi</div>
              <div v-if="lastValues.duration">Duration: {{ formatDuration(lastValues.duration) }}</div>
              <div v-if="lastValues.pace">Pace: {{ lastValues.pace }} min/mi</div>
            </div>
          </CardContent>
        </Card>

        <!-- Quick Log Form -->
        <Card>
          <CardHeader>
            <CardTitle>Quick Log</CardTitle>
            <CardDescription>Add your workout data</CardDescription>
          </CardHeader>
          <CardContent>
            <form @submit.prevent="handleQuickLog" class="space-y-4">
              <!-- Strength Exercise Fields -->
              <template v-if="selectedExercise.exercise_type === 'strength'">
                <div class="space-y-3">
                  <Label>Reps</Label>
                  <Input
                    v-model.number="quickForm.reps"
                    type="number"
                    min="0"
                    placeholder="Reps"
                    class="text-lg"
                  />
                </div>
                <div class="space-y-3">
                  <Label>Weight (lbs)</Label>
                  <Input
                    v-model.number="quickForm.weight"
                    type="number"
                    min="0"
                    step="0.5"
                    placeholder="Weight"
                    class="text-lg"
                  />
                </div>
                <div class="space-y-3">
                  <Label>Perceived Effort (1-10)</Label>
                  <Input
                    v-model.number="quickForm.perceived_effort"
                    type="number"
                    min="1"
                    max="10"
                    placeholder="Effort"
                    class="text-lg"
                  />
                </div>
              </template>

              <!-- Cardio Exercise Fields -->
              <template v-else>
                <div class="space-y-3">
                  <Label>Distance (miles)</Label>
                  <Input
                    v-model.number="quickForm.distance"
                    type="number"
                    min="0"
                    step="0.1"
                    placeholder="Distance"
                    class="text-lg"
                  />
                </div>
                <div class="space-y-3">
                  <Label>Duration (minutes)</Label>
                  <Input
                    v-model.number="quickForm.duration"
                    type="number"
                    min="0"
                    placeholder="Duration"
                    class="text-lg"
                  />
                </div>
                <div class="space-y-3">
                  <Label>Pace (min/mile)</Label>
                  <Input
                    v-model.number="quickForm.pace"
                    type="number"
                    min="0"
                    step="0.1"
                    placeholder="Pace"
                    class="text-lg"
                  />
                </div>
              </template>

              <div v-if="error" class="text-sm text-red-600 dark:text-red-400 p-3 rounded-lg bg-red-50 dark:bg-red-900/20">
                {{ error }}
              </div>

              <div class="flex gap-3 pt-4">
                <Button
                  type="submit"
                  :disabled="loading || !canSubmit"
                  size="lg"
                  class="flex-1"
                >
                  {{ loading ? 'Saving...' : 'Log Workout' }}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  @click="handleQuickLogAndAddAnother"
                  :disabled="loading || !canSubmit"
                  size="lg"
                >
                  Log & Add Another
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getExercises, getPublicExercises, getLastWorkoutValues, createWorkoutLog } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { localDateToUTC, utcDateToLocal, formatDateLocal } from '@/utils/dateUtils'

const exercises = ref([])
const exercisesLoading = ref(true)
const selectedExercise = ref(null)
const lastValues = ref(null)
const loading = ref(false)
const error = ref(null)

const quickForm = ref({
  reps: null,
  weight: null,
  perceived_effort: null,
  distance: null,
  duration: null,
  pace: null
})

const canSubmit = computed(() => {
  if (!selectedExercise.value) return false
  
  if (selectedExercise.value.exercise_type === 'strength') {
    return quickForm.value.reps || quickForm.value.weight || quickForm.value.perceived_effort
  } else {
    return quickForm.value.distance || quickForm.value.duration || quickForm.value.pace
  }
})

const loadExercises = async () => {
  try {
    exercisesLoading.value = true
    const [userExercisesResponse, publicExercisesResponse] = await Promise.all([
      getExercises(),
      getPublicExercises()
    ])
    
    const userExercises = userExercisesResponse.exercises.map(ex => ({ ...ex, isPublic: false }))
    const publicExercises = publicExercisesResponse.exercises.map(ex => ({ ...ex, isPublic: true }))
    
    exercises.value = [...userExercises, ...publicExercises].sort((a, b) => 
      a.name.localeCompare(b.name)
    )
  } catch (err) {
    console.error('Failed to load exercises:', err)
    error.value = 'Failed to load exercises'
  } finally {
    exercisesLoading.value = false
  }
}

const selectExercise = async (exercise) => {
  selectedExercise.value = exercise
  error.value = null
  resetQuickForm()
  
  // Load last workout values
  try {
    const response = await getLastWorkoutValues(exercise.id)
    lastValues.value = response.lastLog
    
    // Pre-populate with last values
    if (lastValues.value) {
      if (exercise.exercise_type === 'strength' && lastValues.value.weight_per_set && lastValues.value.weight_per_set.length) {
        const lastSet = lastValues.value.weight_per_set[lastValues.value.weight_per_set.length - 1]
        if (typeof lastSet === 'object' && lastSet !== null) {
          quickForm.value.reps = lastSet.reps || null
          quickForm.value.weight = lastSet.weight || null
          quickForm.value.perceived_effort = lastSet.perceived_effort || null
        }
      } else if (exercise.exercise_type === 'cardio') {
        quickForm.value.distance = lastValues.value.distance || null
        quickForm.value.duration = lastValues.value.duration || null
        quickForm.value.pace = lastValues.value.pace || null
      }
    }
  } catch (err) {
    console.error('Failed to load last values:', err)
    lastValues.value = null
  }
}

const resetQuickForm = () => {
  quickForm.value = {
    reps: null,
    weight: null,
    perceived_effort: null,
    distance: null,
    duration: null,
    pace: null
  }
}

const handleQuickLog = async (addAnother = false) => {
  if (!selectedExercise.value || !canSubmit.value) return

  try {
    loading.value = true
    error.value = null

    const logData = {
      exercise_id: selectedExercise.value.id,
      date: localDateToUTC(new Date().toISOString().split('T')[0]),
    }

    if (selectedExercise.value.exercise_type === 'strength') {
      logData.weight_per_set = [{
        reps: quickForm.value.reps || null,
        weight: quickForm.value.weight || null,
        perceived_effort: quickForm.value.perceived_effort || null
      }]
      logData.sets = 1
    } else {
      logData.distance = quickForm.value.distance || null
      logData.duration = quickForm.value.duration || null
      logData.pace = quickForm.value.pace || null
    }

    await createWorkoutLog(logData)
    
    if (addAnother) {
      resetQuickForm()
      // Reload last values
      await selectExercise(selectedExercise.value)
    } else {
      selectedExercise.value = null
      resetQuickForm()
      lastValues.value = null
    }
  } catch (err) {
    error.value = err.message || 'Failed to log workout'
  } finally {
    loading.value = false
  }
}

const handleQuickLogAndAddAnother = () => {
  handleQuickLog(true)
}

const formatDate = (dateString) => {
  return formatDateLocal(dateString)
}

const formatDuration = (minutes) => {
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  if (hours > 0) {
    return `${hours}h ${mins}m`
  }
  return `${mins}m`
}

onMounted(() => {
  loadExercises()
})
</script>

