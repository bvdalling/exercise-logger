<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-5 py-10 sm:px-6 sm:py-12 max-w-2xl">
      <Card>
        <CardHeader class="p-6 sm:p-8">
          <CardTitle class="text-xl sm:text-2xl font-bold">{{ isEditMode ? 'Edit Workout' : 'Record Your Workout' }}</CardTitle>
          <CardDescription class="text-sm sm:text-base font-medium mt-2">{{ isEditMode ? 'Update your exercise performance' : 'Log your exercise performance' }}</CardDescription>
        </CardHeader>
        <CardContent class="p-6 sm:p-8">
          <form @submit.prevent="handleSubmit" class="space-y-6">
            <div class="space-y-3">
              <Label for="exercise_id" class="text-base font-semibold">Exercise *</Label>
              <Combobox
                id="exercise_id"
                v-model="form.exercise_id"
                :options="exercises"
                placeholder="Search for an exercise..."
                :disabled="loading || exercisesLoading"
                @change="onExerciseChange"
              />
            </div>
            <div class="space-y-3">
              <Label for="date" class="text-base font-semibold">Date *</Label>
              <Input
                id="date"
                v-model="form.date"
                type="date"
                required
                :disabled="loading"
              />
            </div>

            <!-- Last workout values display -->
            <div v-if="lastValues" class="p-5 sm:p-6 bg-muted/50 rounded-2xl glass-surface dark:glass-surface">
              <p class="text-base font-semibold mb-3">Last workout ({{ formatDate(lastValues.date) }}):</p>
              <div class="text-sm space-y-1">
                <div v-if="selectedExerciseType === 'strength'">
                  <div v-if="lastValues.weight">Weight: {{ lastValues.weight }}lbs</div>
                  <div v-if="lastValues.weight_per_set && lastValues.weight_per_set.length">
                    <div class="mt-2">Sets:</div>
                    <div v-for="(set, idx) in lastValues.weight_per_set" :key="idx" class="ml-4 text-sm">
                      Set {{ idx + 1 }}: 
                      <span v-if="set.reps">{{ set.reps }} reps</span>
                      <span v-if="set.weight"> @ {{ set.weight }}lbs</span>
                      <span v-if="set.perceived_effort"> (Effort: {{ set.perceived_effort }}/10)</span>
                      <span v-else-if="typeof set === 'number'">{{ set }}lbs</span>
                    </div>
                  </div>
                  <div v-if="lastValues.sets && !lastValues.weight_per_set">Sets: {{ lastValues.sets }}</div>
                  <div v-if="lastValues.reps && !lastValues.weight_per_set">Reps: {{ lastValues.reps }}</div>
                  <div v-if="lastValues.rest_time">Rest time: {{ formatRestTime(lastValues.rest_time) }}</div>
                </div>
                <div v-if="selectedExerciseType === 'cardio'">
                  <div v-if="lastValues.distance">Distance: {{ lastValues.distance }}mi</div>
                  <div v-if="lastValues.duration">Duration: {{ formatDuration(lastValues.duration) }}</div>
                  <div v-if="lastValues.pace">Pace: {{ lastValues.pace }} min/mi</div>
                  <div v-if="lastValues.lap_times && lastValues.lap_times.length">
                    <div class="mt-2">Lap times:</div>
                    <div v-for="(lap, idx) in lastValues.lap_times" :key="idx" class="ml-4">
                      {{ lap.label }} @ {{ formatLapTime(lap.time) }}
                    </div>
                    <div v-if="lastValuesAverageLapTime" class="ml-4 font-medium">
                      Average @ {{ lastValuesAverageLapTime }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Strength training fields -->
            <template v-if="selectedExerciseType === 'strength'">
              <!-- Weight per set -->
              <div class="space-y-4">
                <Label class="text-base font-semibold">Sets with Reps, Weight, and Perceived Effort</Label>
                <div v-if="form.weight_per_set && form.weight_per_set.length > 0" class="space-y-4">
                  <div v-for="(set, index) in form.weight_per_set" :key="index" class="p-4 sm:p-5 rounded-2xl space-y-3 glass-card dark:glass-card-dark">
                    <div class="flex items-center justify-between mb-3">
                      <Label class="font-semibold text-base">Set {{ index + 1 }}</Label>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        @click="removeWeightSet(index)"
                        :disabled="loading"
                      >
                        Remove
                      </Button>
                    </div>
                    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4">
                      <div class="space-y-2">
                        <Label for="reps" class="text-sm font-medium">Reps</Label>
                        <Input
                          :id="`reps-${index}`"
                          v-model.number="set.reps"
                          type="number"
                          min="0"
                          placeholder="Reps"
                          :disabled="loading"
                        />
                      </div>
                      <div class="space-y-2">
                        <Label for="weight" class="text-sm font-medium">Weight (lbs)</Label>
                        <Input
                          :id="`weight-${index}`"
                          v-model.number="set.weight"
                          type="number"
                          min="0"
                          step="0.5"
                          placeholder="Weight"
                          :disabled="loading"
                        />
                      </div>
                      <div class="space-y-2">
                        <Label for="effort" class="text-sm font-medium">Perceived Effort (1-10)</Label>
                        <Input
                          :id="`effort-${index}`"
                          v-model.number="set.perceived_effort"
                          type="number"
                          min="1"
                          max="10"
                          placeholder="Effort"
                          :disabled="loading"
                        />
                      </div>
                    </div>
                  </div>
                </div>
                <div class="flex flex-wrap gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    @click="addWeightSet"
                    :disabled="loading"
                    size="default"
                  >
                    Add Set
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    @click="clearWeightSets"
                    :disabled="loading || !form.weight_per_set || form.weight_per_set.length === 0"
                    size="default"
                  >
                    Clear All
                  </Button>
                </div>
                <p class="text-sm text-muted-foreground font-medium">
                  Perceived effort: 1 = very easy, 10 = maximum effort
                </p>
              </div>

              <div class="space-y-3">
                <Label for="rest_time" class="text-base font-semibold">Rest Time (seconds)</Label>
                <Input
                  id="rest_time"
                  v-model.number="form.rest_time"
                  type="number"
                  min="0"
                  placeholder="e.g., 60"
                  :disabled="loading"
                />
              </div>
            </template>

            <!-- Cardio fields -->
            <template v-if="selectedExerciseType === 'cardio'">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
                <div class="space-y-3">
                  <Label for="distance" class="text-base font-semibold">Distance (miles)</Label>
                  <Input
                    id="distance"
                    v-model.number="form.distance"
                    type="number"
                    min="0"
                    step="0.1"
                    placeholder="e.g., 3.5"
                    :disabled="loading"
                  />
                </div>
                <div class="space-y-3">
                  <Label for="duration" class="text-base font-semibold">Duration (minutes)</Label>
                  <Input
                    id="duration"
                    v-model.number="form.duration"
                    type="number"
                    min="0"
                    placeholder="e.g., 30"
                    :disabled="loading"
                  />
                </div>
              </div>
              <div class="space-y-3">
                <Label for="pace" class="text-base font-semibold">Pace (min/mile)</Label>
                <Input
                  id="pace"
                  v-model.number="form.pace"
                  type="number"
                  min="0"
                  step="0.1"
                  placeholder="e.g., 8.5"
                  :disabled="loading"
                />
              </div>

              <!-- Lap times -->
              <div class="space-y-4">
                <Label class="text-base font-semibold">Lap Times</Label>
                <div v-if="form.lap_times && form.lap_times.length > 0" class="space-y-3">
                  <div v-for="(lap, index) in form.lap_times" :key="index" class="flex gap-2 items-center">
                    <Input
                      v-model="lap.label"
                      type="text"
                      placeholder="e.g., Mile 1"
                      :disabled="loading"
                      class="w-32"
                    />
                    <Input
                      v-model="lap.time"
                      type="text"
                      placeholder="MM:SS or M:SS"
                      :disabled="loading"
                      class="flex-1"
                      pattern="[0-9]{1,2}:[0-5][0-9]"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      @click="removeLapTime(index)"
                      :disabled="loading"
                    >
                      Remove
                    </Button>
                  </div>
                  <div v-if="averageLapTime" class="p-2 bg-muted rounded-md text-sm">
                    <span class="font-medium">Average: {{ averageLapTime }}</span>
                  </div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  @click="addLapTime"
                  :disabled="loading"
                >
                  Add Lap Time
                </Button>
                <p class="text-xs text-muted-foreground">
                  Format: MM:SS (e.g., 10:31 for 10 minutes 31 seconds)
                </p>
              </div>
            </template>

            <div class="space-y-3">
              <Label for="notes" class="text-base font-semibold">Notes</Label>
              <textarea
                id="notes"
                v-model="form.notes"
                class="flex min-h-[100px] w-full rounded-2xl border-2 border-input/50 glass-input dark:glass-input-dark px-5 py-4 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 focus-visible:shadow-lg focus-visible:shadow-primary/20 focus-visible:scale-[1.01]"
                placeholder="Additional notes about your workout"
                :disabled="loading"
              />
            </div>

            <div v-if="error" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
              {{ error }}
            </div>

            <div class="flex flex-col sm:flex-row gap-4 pt-4">
              <Button type="submit" variant="default" :disabled="loading" size="lg" class="flex-1">
                {{ loading ? 'Saving...' : (isEditMode ? 'Update Workout' : 'Log Workout') }}
              </Button>
              <Button type="button" variant="outline" @click="resetForm" :disabled="loading" size="lg" class="flex-1 sm:flex-initial">
                Reset
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getExercises, getLastWorkoutValues, createWorkoutLog, getWorkoutLog, updateWorkoutLog } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Combobox } from '@/components/ui/combobox'
import { localDateToUTC, utcDateToLocal, getTodayLocal, formatDateLocal } from '@/utils/dateUtils'

const router = useRouter()
const route = useRoute()
const exercises = ref([])
const exercisesLoading = ref(true)
const loading = ref(false)
const error = ref(null)
const lastValues = ref(null)
const selectedExerciseType = ref(null)
const isEditMode = ref(false)
const logId = ref(null)

const form = ref({
  exercise_id: null,
  date: getTodayLocal(), // Use local date for form input
  sets: null,
  reps: null,
  weight: null,
  weight_per_set: [],
  rest_time: null,
  distance: null,
  duration: null,
  pace: null,
  lap_times: [],
  notes: ''
})

const loadExercises = async () => {
  try {
    exercisesLoading.value = true
    const response = await getExercises()
    exercises.value = response.exercises
  } catch (err) {
    console.error('Failed to load exercises:', err)
    error.value = 'Failed to load exercises'
  } finally {
    exercisesLoading.value = false
  }
}

const loadExistingLog = async (id) => {
  try {
    loading.value = true
    const response = await getWorkoutLog(id)
    const log = response.log
    
    // Populate form with existing log data
    form.value.exercise_id = log.exercise_id.toString()
    // Convert UTC date from API to local date for form input
    form.value.date = utcDateToLocal(log.date)
    form.value.rest_time = log.rest_time
    form.value.distance = log.distance
    form.value.duration = log.duration
    form.value.pace = log.pace
    form.value.notes = log.notes || ''
    
    // Set exercise type
    selectedExerciseType.value = log.exercise_type
    
    // Handle weight_per_set
    if (log.weight_per_set && Array.isArray(log.weight_per_set)) {
      form.value.weight_per_set = log.weight_per_set.map(set => {
        if (typeof set === 'number') {
          // Old format: just a number (weight)
          return {
            reps: log.reps || null,
            weight: set,
            perceived_effort: null
          }
        } else {
          // New format: object with reps, weight, perceived_effort
          return {
            reps: set.reps || null,
            weight: set.weight || null,
            perceived_effort: set.perceived_effort || null
          }
        }
      })
    }
    
    // Handle lap_times
    if (log.lap_times && Array.isArray(log.lap_times)) {
      form.value.lap_times = log.lap_times.map(lap => ({ ...lap }))
    }
    
    // Set legacy fields for backward compatibility
    form.value.sets = log.sets
    form.value.reps = log.reps
    form.value.weight = log.weight
  } catch (err) {
    console.error('Failed to load workout log:', err)
    error.value = 'Failed to load workout log'
  } finally {
    loading.value = false
  }
}

const onExerciseChange = async () => {
  if (!form.value.exercise_id) {
    selectedExerciseType.value = null
    lastValues.value = null
    return
  }

  // Find the selected exercise to get its type
  const selectedExercise = exercises.value.find(e => e.id === parseInt(form.value.exercise_id))
  selectedExerciseType.value = selectedExercise?.exercise_type || null

  // Load last values
  await loadLastValues()
}

const loadLastValues = async () => {
  if (!form.value.exercise_id) {
    lastValues.value = null
    return
  }
  try {
    const response = await getLastWorkoutValues(form.value.exercise_id)
    lastValues.value = response.lastLog
    
    // Pre-populate form with last values if available
    if (lastValues.value) {
      if (lastValues.value.weight_per_set && Array.isArray(lastValues.value.weight_per_set)) {
        // Handle both old format (array of numbers) and new format (array of objects)
        form.value.weight_per_set = lastValues.value.weight_per_set.map(set => {
          if (typeof set === 'number') {
            // Old format: just a number (weight)
            return {
              reps: lastValues.value.reps || null,
              weight: set,
              perceived_effort: null
            }
          } else {
            // New format: object with reps, weight, perceived_effort
            return {
              reps: set.reps || null,
              weight: set.weight || null,
              perceived_effort: set.perceived_effort || null
            }
          }
        })
      }
      if (lastValues.value.lap_times && Array.isArray(lastValues.value.lap_times)) {
        form.value.lap_times = lastValues.value.lap_times.map(lap => ({ ...lap }))
      }
    }
  } catch (err) {
    console.error('Failed to load last values:', err)
    lastValues.value = null
  }
}

const formatDate = (dateString) => {
  // dateString is UTC, convert to local for display
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

const formatRestTime = (seconds) => {
  if (seconds < 60) {
    return `${seconds}s`
  }
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  if (secs === 0) {
    return `${mins}m`
  }
  return `${mins}m ${secs}s`
}

const formatLapTime = (timeString) => {
  // timeString is in MM:SS format
  return timeString || 'N/A'
}

// Helper function to calculate average lap time from lap times array
const calculateAverageLapTime = (lapTimes) => {
  if (!lapTimes || lapTimes.length === 0) {
    return null
  }

  const validLaps = lapTimes.filter(lap => lap.time && /^\d{1,2}:[0-5]\d$/.test(lap.time))
  if (validLaps.length === 0) {
    return null
  }

  // Convert MM:SS to total seconds
  const totalSeconds = validLaps.reduce((sum, lap) => {
    const [minutes, seconds] = lap.time.split(':').map(Number)
    return sum + (minutes * 60 + seconds)
  }, 0)

  // Calculate average
  const avgSeconds = Math.round(totalSeconds / validLaps.length)
  const avgMinutes = Math.floor(avgSeconds / 60)
  const avgSecs = avgSeconds % 60

  // Format as MM:SS
  return `${avgMinutes}:${avgSecs.toString().padStart(2, '0')}`
}

// Calculate average lap time for form
const averageLapTime = computed(() => {
  return calculateAverageLapTime(form.value.lap_times)
})

// Calculate average lap time for last values
const lastValuesAverageLapTime = computed(() => {
  if (!lastValues.value || !lastValues.value.lap_times) {
    return null
  }
  return calculateAverageLapTime(lastValues.value.lap_times)
})

const addWeightSet = () => {
  if (!form.value.weight_per_set) {
    form.value.weight_per_set = []
  }
  form.value.weight_per_set.push({
    reps: null,
    weight: null,
    perceived_effort: null
  })
}

const removeWeightSet = (index) => {
  if (form.value.weight_per_set) {
    form.value.weight_per_set.splice(index, 1)
  }
}

const clearWeightSets = () => {
  form.value.weight_per_set = []
}

const addLapTime = () => {
  if (!form.value.lap_times) {
    form.value.lap_times = []
  }
  form.value.lap_times.push({ label: '', time: '' })
}

const removeLapTime = (index) => {
  if (form.value.lap_times) {
    form.value.lap_times.splice(index, 1)
  }
}

const resetForm = () => {
  form.value = {
    exercise_id: null,
    date: getTodayLocal(), // Use local date
    sets: null,
    reps: null,
    weight: null,
    weight_per_set: [],
    rest_time: null,
    distance: null,
    duration: null,
    pace: null,
    lap_times: [],
    notes: ''
  }
  lastValues.value = null
  selectedExerciseType.value = null
  error.value = null
}

const handleSubmit = async () => {
  try {
    loading.value = true
    error.value = null

    // Clean up data
    // Validate required fields
    if (!form.value.exercise_id) {
      error.value = 'Please select an exercise'
      loading.value = false
      return
    }
    
    // Convert local date to UTC before sending to API
    const logData = {
      exercise_id: parseInt(form.value.exercise_id),
      date: localDateToUTC(form.value.date),
      sets: form.value.sets || null,
      reps: form.value.reps || null,
      weight: form.value.weight || null,
      rest_time: form.value.rest_time || null,
      distance: form.value.distance || null,
      duration: form.value.duration || null,
      pace: form.value.pace || null,
      notes: form.value.notes || null
    }

    // Add weight_per_set if it has values
    if (form.value.weight_per_set && form.value.weight_per_set.length > 0) {
      const validSets = form.value.weight_per_set.filter(set => {
        // Keep sets that have at least one field filled
        return set && (set.reps !== null && set.reps !== undefined && set.reps !== '' ||
                       set.weight !== null && set.weight !== undefined && set.weight !== '' ||
                       set.perceived_effort !== null && set.perceived_effort !== undefined && set.perceived_effort !== '')
      })
      if (validSets.length > 0) {
        // Clean up the sets - remove null/undefined values
        logData.weight_per_set = validSets.map(set => ({
          reps: set.reps || null,
          weight: set.weight || null,
          perceived_effort: set.perceived_effort || null
        }))
      }
    }

    // Add lap_times if it has values
    if (form.value.lap_times && form.value.lap_times.length > 0) {
      const validLaps = form.value.lap_times.filter(lap => lap.label && lap.time)
      if (validLaps.length > 0) {
        logData.lap_times = validLaps
      }
    }

    // Calculate sets from weight_per_set length for backward compatibility
    if (logData.weight_per_set && logData.weight_per_set.length > 0) {
      logData.sets = logData.weight_per_set.length
    }

    if (isEditMode.value && logId.value) {
      await updateWorkoutLog(logId.value, logData)
    } else {
      await createWorkoutLog(logData)
    }
    router.push('/')
  } catch (err) {
    error.value = err.message || 'Failed to log workout'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // Check if we're in edit mode
  if (route.params.id) {
    isEditMode.value = true
    logId.value = parseInt(route.params.id)
  }
  
  await loadExercises()
  
  // Load existing log data if in edit mode
  if (isEditMode.value && logId.value) {
    await loadExistingLog(logId.value)
  } else {
    // Check for pre-selected exercise from query parameter
    const exerciseId = route.query.exerciseId
    if (exerciseId) {
      const exerciseIdNum = parseInt(exerciseId)
      // Verify the exercise exists in the loaded exercises
      const exercise = exercises.value.find(e => e.id === exerciseIdNum)
      if (exercise) {
        form.value.exercise_id = exerciseIdNum.toString()
        // Trigger exercise change to load last values
        await onExerciseChange()
      }
    }
  }
})
</script>
