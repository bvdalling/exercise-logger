<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-5 py-10 sm:px-6 sm:py-12">
      <div v-if="loading" class="text-center py-12 text-lg">Loading...</div>
      <div v-else-if="!exercise" class="text-center py-12">
        <p class="text-muted-foreground text-base sm:text-lg">Exercise not found</p>
      </div>
      <div v-else>
        <Card class="mb-10">
          <CardHeader class="p-6 sm:p-8">
            <CardTitle class="text-xl sm:text-2xl font-bold">{{ exercise.name }}</CardTitle>
            <CardDescription v-if="exercise.muscle_group || exercise.equipment" class="text-sm sm:text-base font-medium mt-2">
              <span v-if="exercise.muscle_group">{{ exercise.muscle_group }}</span>
              <span v-if="exercise.equipment"> • {{ exercise.equipment }}</span>
            </CardDescription>
          </CardHeader>
        </Card>

        <div v-if="progress.length === 0" class="text-center py-12">
          <p class="text-muted-foreground mb-6 text-base sm:text-lg">No workout data yet for this exercise.</p>
          <Button @click="$router.push('/log-workout')" size="lg">Log Your First Workout</Button>
        </div>
        <div v-else>
          <Card class="mb-10">
            <CardHeader class="p-6 sm:p-8">
              <CardTitle class="text-xl sm:text-2xl font-bold">Progress Chart</CardTitle>
            </CardHeader>
            <CardContent class="p-6 sm:p-8">
              <div class="h-64 sm:h-80">
                <Line :data="chartData" :options="chartOptions" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader class="p-6 sm:p-8">
              <CardTitle class="text-xl sm:text-2xl font-bold">Workout History</CardTitle>
            </CardHeader>
            <CardContent class="p-6 sm:p-8">
              <div class="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Date</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'strength'">Avg Weight (lbs)</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'strength'">Total Volume (lbs)</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'strength'">Sets</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'strength'">Total Reps</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'strength'">Avg Effort</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'strength'">Avg Rest Time</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'cardio'">Distance (mi)</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'cardio'">Duration (min)</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'cardio'">Pace (min/mi)</TableHead>
                      <TableHead v-if="exercise.exercise_type === 'cardio'">Avg Lap Time</TableHead>
                      <TableHead>Notes</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <TableRow v-for="log in aggregatedProgress" :key="log.id || log.date">
                      <TableCell>{{ formatDate(log.date) }}</TableCell>
                      <TableCell v-if="exercise.exercise_type === 'strength'">
                        {{ log.avgWeight !== null && log.avgWeight !== undefined ? log.avgWeight.toFixed(1) : '-' }}
                      </TableCell>
                      <TableCell v-if="exercise.exercise_type === 'strength'">
                        {{ log.totalVolume !== null && log.totalVolume !== undefined ? log.totalVolume.toFixed(1) : '-' }}
                      </TableCell>
                      <TableCell v-if="exercise.exercise_type === 'strength'">{{ log.sets || '-' }}</TableCell>
                      <TableCell v-if="exercise.exercise_type === 'strength'">{{ log.totalReps || '-' }}</TableCell>
                      <TableCell v-if="exercise.exercise_type === 'strength'">
                        {{ log.avgEffort !== null && log.avgEffort !== undefined ? log.avgEffort.toFixed(1) : '-' }}
                      </TableCell>
                      <TableCell v-if="exercise.exercise_type === 'strength'">
                        {{ log.restTime ? formatRestTime(log.restTime) : '-' }}
                      </TableCell>
                      <TableCell v-if="exercise.exercise_type === 'cardio'">{{ log.distance || '-' }}</TableCell>
                      <TableCell v-if="exercise.exercise_type === 'cardio'">{{ log.duration || '-' }}</TableCell>
                      <TableCell v-if="exercise.exercise_type === 'cardio'">{{ log.pace || '-' }}</TableCell>
                      <TableCell v-if="exercise.exercise_type === 'cardio'">
                        {{ log.avgLapTime || '-' }}
                      </TableCell>
                      <TableCell class="max-w-xs truncate">{{ log.notes || '-' }}</TableCell>
                      <TableCell>
                        <Button
                          variant="outline"
                          size="sm"
                          @click="openDetailDialog(log)"
                        >
                          View Details
                        </Button>
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>

    <!-- Detail View Dialog -->
    <Dialog v-model:open="detailDialogOpen">
      <DialogContent class="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Workout Details</DialogTitle>
          <DialogDescription>
            {{ selectedLog ? formatDate(selectedLog.date) : '' }}
          </DialogDescription>
        </DialogHeader>
        <div v-if="selectedLog" class="space-y-4">
          <div v-if="exercise.exercise_type === 'strength'">
            <div v-if="selectedLog.weight_per_set && Array.isArray(selectedLog.weight_per_set) && selectedLog.weight_per_set.length" class="space-y-3">
              <h4 class="font-semibold">Sets</h4>
              <div v-for="(set, idx) in selectedLog.weight_per_set" :key="idx" class="p-3 rounded-xl glass-card dark:glass-card-dark">
                <div class="font-medium mb-2">Set {{ idx + 1 }}</div>
                <div class="grid grid-cols-3 gap-4 text-sm">
                  <div>
                    <span class="text-muted-foreground">Reps:</span>
                    <span class="ml-2">{{ set.reps || '-' }}</span>
                  </div>
                  <div>
                    <span class="text-muted-foreground">Weight:</span>
                    <span class="ml-2">{{ set.weight ? `${set.weight}lbs` : '-' }}</span>
                  </div>
                  <div>
                    <span class="text-muted-foreground">Effort:</span>
                    <span class="ml-2">{{ set.perceived_effort ? `${set.perceived_effort}/10` : '-' }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="text-muted-foreground">No set data available</div>
            <div v-if="selectedLog.rest_time" class="mt-4">
              <span class="text-muted-foreground">Rest Time:</span>
              <span class="ml-2">{{ formatRestTime(selectedLog.rest_time) }}</span>
            </div>
          </div>
          <div v-if="exercise.exercise_type === 'cardio'">
            <div v-if="selectedLog.lap_times && Array.isArray(selectedLog.lap_times) && selectedLog.lap_times.length" class="space-y-2">
              <h4 class="font-semibold">Lap Times</h4>
              <div v-for="(lap, idx) in selectedLog.lap_times" :key="idx" class="p-2 rounded-xl glass-card dark:glass-card-dark">
                <div class="flex justify-between">
                  <span class="font-medium">{{ lap.label || `Lap ${idx + 1}` }}</span>
                  <span>{{ lap.time }}</span>
                </div>
              </div>
              <div v-if="getAverageLapTime(selectedLog.lap_times)" class="p-2 bg-muted rounded-md font-medium">
                Average: {{ getAverageLapTime(selectedLog.lap_times) }}
              </div>
            </div>
            <div v-else class="text-muted-foreground">No lap time data available</div>
          </div>
          <div v-if="selectedLog.notes" class="mt-4">
            <h4 class="font-semibold mb-2">Notes</h4>
            <p class="text-sm text-muted-foreground whitespace-pre-wrap">{{ selectedLog.notes }}</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'
import { getExercise, getExerciseProgress } from '@/composables/useApi'
import { useScrollRestore } from '@/composables/useScrollRestore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

const route = useRoute()
const exercise = ref(null)
const progress = ref([])
const loading = ref(true)
const detailDialogOpen = ref(false)
const selectedLog = ref(null)

// Restore scroll position when modal closes
useScrollRestore(detailDialogOpen)

// Helper functions for calculations
const formatDate = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleDateString()
}

const calculateSetsCount = (log) => {
  if (log.weight_per_set && Array.isArray(log.weight_per_set)) {
    return log.weight_per_set.length
  }
  return log.sets || null
}

const calculateTotalVolume = (log) => {
  if (!log.weight_per_set || !Array.isArray(log.weight_per_set)) {
    return null
  }
  
  let totalVolume = 0
  let hasData = false
  
  for (const set of log.weight_per_set) {
    if (typeof set === 'object' && set !== null) {
      const weight = set.weight || 0
      const reps = set.reps || 0
      if (weight > 0 && reps > 0) {
        totalVolume += weight * reps
        hasData = true
      }
    } else if (typeof set === 'number' && log.reps) {
      // Legacy format: weight_per_set is array of numbers, reps is total
      totalVolume += set * (log.reps / log.weight_per_set.length)
      hasData = true
    }
  }
  
  return hasData ? totalVolume : null
}

const calculateTotalReps = (log) => {
  if (!log.weight_per_set || !Array.isArray(log.weight_per_set)) {
    return log.reps || null
  }
  
  let totalReps = 0
  let hasData = false
  
  for (const set of log.weight_per_set) {
    if (typeof set === 'object' && set !== null && set.reps) {
      totalReps += set.reps
      hasData = true
    }
  }
  
  // Fallback to log.reps if no set-level reps data
  if (!hasData && log.reps) {
    return log.reps
  }
  
  return hasData ? totalReps : null
}

const calculateAverageWeight = (log) => {
  if (!log.weight_per_set || !Array.isArray(log.weight_per_set)) {
    return log.weight || null
  }
  
  let totalWeight = 0
  let count = 0
  
  for (const set of log.weight_per_set) {
    if (typeof set === 'object' && set !== null && set.weight) {
      totalWeight += set.weight
      count++
    } else if (typeof set === 'number') {
      totalWeight += set
      count++
    }
  }
  
  if (count === 0) {
    return log.weight || null
  }
  
  return totalWeight / count
}

const calculateAverageEffort = (log) => {
  if (!log.weight_per_set || !Array.isArray(log.weight_per_set)) {
    return null
  }
  
  let totalEffort = 0
  let count = 0
  
  for (const set of log.weight_per_set) {
    if (typeof set === 'object' && set !== null && set.perceived_effort !== null && set.perceived_effort !== undefined) {
      totalEffort += set.perceived_effort
      count++
    }
  }
  
  return count > 0 ? totalEffort / count : null
}

const getAverageLapTime = (lapTimes) => {
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

const chartData = computed(() => {
  const labels = progress.value.map(log => formatDate(log.date))
  
  if (!exercise.value) {
    return { labels: [], datasets: [] }
  }
  
  let datasets = []
  
  if (exercise.value.exercise_type === 'strength') {
    // # Sets per workout
    datasets.push({
      label: '# Sets',
      data: progress.value.map(log => calculateSetsCount(log)),
      borderColor: 'rgb(59, 130, 246)',
      backgroundColor: 'rgba(59, 130, 246, 0.1)',
      tension: 0.1,
      yAxisID: 'y'
    })
    
    // Total Volume
    datasets.push({
      label: 'Total Volume (lbs)',
      data: progress.value.map(log => calculateTotalVolume(log)),
      borderColor: 'rgb(34, 197, 94)',
      backgroundColor: 'rgba(34, 197, 94, 0.1)',
      tension: 0.1,
      yAxisID: 'y1'
    })
    
    // Total Reps
    datasets.push({
      label: 'Total Reps',
      data: progress.value.map(log => calculateTotalReps(log)),
      borderColor: 'rgb(251, 146, 60)',
      backgroundColor: 'rgba(251, 146, 60, 0.1)',
      tension: 0.1,
      yAxisID: 'y'
    })
    
    // Average Perceived Effort
    datasets.push({
      label: 'Avg Effort (1-10)',
      data: progress.value.map(log => calculateAverageEffort(log)),
      borderColor: 'rgb(168, 85, 247)',
      backgroundColor: 'rgba(168, 85, 247, 0.1)',
      tension: 0.1,
      yAxisID: 'y2'
    })
  } else if (exercise.value.exercise_type === 'cardio') {
    // Distance
    datasets.push({
      label: 'Distance (miles)',
      data: progress.value.map(log => log.distance || null),
      borderColor: 'rgb(59, 130, 246)',
      backgroundColor: 'rgba(59, 130, 246, 0.1)',
      tension: 0.1,
      yAxisID: 'y'
    })
    
    // Duration
    datasets.push({
      label: 'Duration (min)',
      data: progress.value.map(log => log.duration || null),
      borderColor: 'rgb(34, 197, 94)',
      backgroundColor: 'rgba(34, 197, 94, 0.1)',
      tension: 0.1,
      yAxisID: 'y1'
    })
    
    // Pace
    datasets.push({
      label: 'Pace (min/mile)',
      data: progress.value.map(log => log.pace || null),
      borderColor: 'rgb(251, 146, 60)',
      backgroundColor: 'rgba(251, 146, 60, 0.1)',
      tension: 0.1,
      yAxisID: 'y2'
    })
  }
  
  return {
    labels,
    datasets
  }
})

const chartOptions = computed(() => {
  if (!exercise.value) {
    return {
      responsive: true,
      maintainAspectRatio: false
    }
  }
  
  if (exercise.value.exercise_type === 'strength') {
    return {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          position: 'left',
          title: {
            display: true,
            text: 'Sets / Reps'
          }
        },
        y1: {
          type: 'linear',
          display: true,
          position: 'right',
          beginAtZero: true,
          title: {
            display: true,
            text: 'Total Volume (lbs)'
          },
          grid: {
            drawOnChartArea: false
          }
        },
        y2: {
          type: 'linear',
          display: false,
          beginAtZero: true,
          min: 0,
          max: 10
        }
      },
      plugins: {
        legend: {
          position: 'top'
        },
        title: {
          display: true,
          text: 'Exercise Progress Over Time'
        }
      }
    }
  } else if (exercise.value.exercise_type === 'cardio') {
    return {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          position: 'left',
          title: {
            display: true,
            text: 'Distance (miles)'
          }
        },
        y1: {
          type: 'linear',
          display: true,
          position: 'right',
          beginAtZero: true,
          title: {
            display: true,
            text: 'Duration (min)'
          },
          grid: {
            drawOnChartArea: false
          }
        },
        y2: {
          type: 'linear',
          display: false,
          beginAtZero: true,
          title: {
            display: true,
            text: 'Pace (min/mile)'
          }
        }
      },
      plugins: {
        legend: {
          position: 'top'
        },
        title: {
          display: true,
          text: 'Exercise Progress Over Time'
        }
      }
    }
  }
  
  return {
    responsive: true,
    maintainAspectRatio: false
  }
})

// Aggregated progress for table display
const aggregatedProgress = computed(() => {
  return progress.value.map(log => {
    if (exercise.value?.exercise_type === 'strength') {
      return {
        ...log,
        sets: calculateSetsCount(log),
        totalVolume: calculateTotalVolume(log),
        totalReps: calculateTotalReps(log),
        avgWeight: calculateAverageWeight(log),
        avgEffort: calculateAverageEffort(log),
        restTime: log.rest_time
      }
    } else if (exercise.value?.exercise_type === 'cardio') {
      return {
        ...log,
        avgLapTime: getAverageLapTime(log.lap_times)
      }
    }
    return log
  })
})

const openDetailDialog = (log) => {
  selectedLog.value = log
  detailDialogOpen.value = true
}

const loadData = async () => {
  try {
    loading.value = true
    const exerciseId = route.params.id
    
    const [exerciseResponse, progressResponse] = await Promise.all([
      getExercise(exerciseId),
      getExerciseProgress(exerciseId)
    ])
    
    exercise.value = exerciseResponse.exercise
    progress.value = progressResponse.progress
  } catch (error) {
    console.error('Failed to load exercise data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
