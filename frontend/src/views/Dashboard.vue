<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-5 py-10 sm:px-6 sm:py-12">
      <!-- Week Calendar Section -->
      <Card class="mb-10">
          <CardHeader class="p-6 sm:p-8">
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <CardTitle class="text-xl sm:text-2xl font-bold">Week View</CardTitle>
            <div class="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                class="flex-1 sm:flex-initial text-xs sm:text-sm"
                @click="previousWeek"
              >
                <span class="hidden sm:inline">← Previous</span>
                <span class="sm:hidden">←</span>
              </Button>
              <Button
                variant="outline"
                size="sm"
                class="flex-1 sm:flex-initial text-xs sm:text-sm"
                @click="goToCurrentWeek"
              >
                Today
              </Button>
              <Button
                variant="outline"
                size="sm"
                class="flex-1 sm:flex-initial text-xs sm:text-sm"
                @click="nextWeek"
              >
                <span class="hidden sm:inline">Next →</span>
                <span class="sm:hidden">→</span>
              </Button>
            </div>
          </div>
          <CardDescription class="text-sm sm:text-base font-medium">{{ weekRangeText }}</CardDescription>
          </CardHeader>
        <CardContent class="p-4 sm:p-8">
          <div class="grid grid-cols-7 gap-2 sm:gap-3">
            <div
              v-for="day in weekDays"
              :key="day.dateString"
              :class="[
                'p-3 sm:p-4 rounded-2xl text-center transition-all duration-300 ease-out cursor-pointer min-h-[70px] sm:min-h-[90px] flex flex-col items-center justify-center glass-surface dark:glass-surface',
                day.isToday ? 'day-today' : '',
                day.hasWorkout ? 'day-workout' : '',
                selectedDate === day.dateString ? 'day-selected scale-110 shadow-2xl' : '',
                'hover:scale-105 active:scale-95 hover:shadow-xl',
                'touch-manipulation'
              ]"
              @click="selectDay(day.dateString)"
            >
              <div class="text-xs sm:text-sm font-semibold text-muted-foreground mb-1 sm:mb-2">
                {{ day.dayName }}
              </div>
              <div
                :class="[
                  'text-lg sm:text-xl font-bold',
                  day.isToday ? 'text-primary' : '',
                  day.hasWorkout ? 'text-success' : ''
                ]"
              >
                {{ day.dayNumber }}
              </div>
              <div
                v-if="day.hasWorkout"
                class="mt-1 sm:mt-2 text-xs sm:text-sm badge-success rounded-full px-2 py-1 font-bold"
              >
                ✓
              </div>
            </div>
          </div>
        </CardContent>
        </Card>

      <!-- Exercise Feed Section -->
        <Card class="mt-10">
          <CardHeader class="p-6 sm:p-8">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <div class="flex-1 min-w-0">
                <CardTitle class="text-xl sm:text-2xl font-bold break-words">
                  <span v-if="selectedDate">
                    Workouts on {{ formatDayHeader(selectedDate) }}
                  </span>
                  <span v-else>
                    Workouts This Week
                  </span>
                </CardTitle>
                <CardDescription class="text-sm sm:text-base font-medium break-words mt-1">
                  <span v-if="selectedDate">
                    Exercises for {{ formatDayHeader(selectedDate) }}
                  </span>
                  <span v-else>
                    Exercises logged during this week
                  </span>
                </CardDescription>
              </div>
              <Button
                v-if="selectedDate"
                variant="outline"
                size="sm"
                class="w-full sm:w-auto text-sm sm:text-base"
                @click="clearDaySelection"
              >
                Show All Days
              </Button>
            </div>
          </CardHeader>
          <CardContent class="p-6 sm:p-8">
          <div v-if="groupedLogs.length === 0" class="text-center py-12 text-muted-foreground text-base sm:text-lg">
            No workouts logged this week
            </div>
          <div v-else class="space-y-6 sm:space-y-8">
            <div
              v-for="dayGroup in groupedLogs"
              :key="dayGroup.date"
              class="pb-6 sm:pb-8 last:pb-0"
            >
              <h3 class="text-lg sm:text-xl font-bold mb-4 sm:mb-5">
                {{ formatDayHeader(dayGroup.date) }}
              </h3>
              <div class="space-y-4 sm:space-y-5">
                <div
                  v-for="log in dayGroup.logs"
                :key="log.id"
                  :class="[
                    'p-5 sm:p-6 rounded-3xl cursor-pointer transition-all duration-300 ease-out hover:scale-[1.02] active:scale-[0.98] touch-manipulation glass-card dark:glass-card-dark',
                    log.exercise_type === 'strength' ? 'card-strength' : 'card-cardio'
                  ]"
                  @click="editWorkout(log.id)"
                >
                  <div class="flex items-start justify-between mb-4 gap-3">
                    <div class="font-bold text-base sm:text-lg flex-1 min-w-0">{{ log.exercise_name }}</div>
                    <div 
                      :class="[
                        'text-xs sm:text-sm px-3 py-1.5 rounded-full font-semibold flex-shrink-0 text-white',
                        log.exercise_type === 'strength' ? 'badge-strength' : 'badge-cardio'
                      ]"
                    >
                      {{ log.exercise_type === 'strength' ? 'Strength' : 'Cardio' }}
                    </div>
                  </div>
                  
                  <!-- Strength Exercise Details -->
                  <div v-if="log.exercise_type === 'strength'" class="space-y-2 text-sm sm:text-base text-muted-foreground">
                    <div v-if="log.sets && log.reps" class="flex flex-wrap gap-3 sm:gap-4">
                      <span class="font-medium">Sets: {{ log.sets }}</span>
                      <span class="font-medium">Reps: {{ log.reps }}</span>
                    </div>
                    <div v-if="log.weight" class="flex gap-3 sm:gap-4">
                      <span class="font-medium">Weight: {{ log.weight }}lbs</span>
                    </div>
                    <div v-else-if="log.weight_per_set && log.weight_per_set.length" class="mt-3">
                      <div class="font-semibold mb-2 text-sm sm:text-base">Sets:</div>
                      <div class="space-y-2">
                        <div
                          v-for="(set, idx) in log.weight_per_set"
                          :key="idx"
                          class="text-xs sm:text-sm"
                        >
                          <span v-if="typeof set === 'object' && set !== null">
                            Set {{ idx + 1 }}:
                            <span v-if="set.reps">{{ set.reps }} reps</span>
                            <span v-if="set.weight"> - {{ set.weight }}lbs</span>
                            <span v-if="set.perceived_effort"> (effort: {{ set.perceived_effort }}/10)</span>
                    </span>
                          <span v-else>
                            Set {{ idx + 1 }}: {{ set }}lbs
                  </span>
                        </div>
                      </div>
                    </div>
                    <div v-if="log.rest_time" class="flex gap-3 sm:gap-4">
                      <span class="font-medium">Rest: {{ log.rest_time }}s</span>
                    </div>
                  </div>

                  <!-- Cardio Exercise Details -->
                  <div v-if="log.exercise_type === 'cardio'" class="space-y-2 text-sm sm:text-base text-muted-foreground">
                    <div v-if="log.distance" class="flex flex-wrap gap-3 sm:gap-4">
                      <span class="font-medium">Distance: {{ log.distance }}mi</span>
                    </div>
                    <div v-if="log.duration" class="flex gap-3 sm:gap-4">
                      <span class="font-medium">Duration: {{ formatDuration(log.duration) }}</span>
                    </div>
                    <div v-if="log.pace" class="flex gap-3 sm:gap-4">
                      <span class="font-medium">Pace: {{ log.pace }} min/mi</span>
                    </div>
                    <div v-if="log.lap_times && log.lap_times.length" class="mt-3">
                      <div class="font-semibold mb-2 text-sm sm:text-base">Lap Times:</div>
                      <div class="text-xs sm:text-sm">
                        {{ log.lap_times.map((lap, idx) => `Lap ${idx + 1}: ${formatDuration(lap)}`).join(', ') }}
                      </div>
                    </div>
                  </div>

                  <!-- Notes -->
                  <div v-if="log.notes" class="mt-4 pt-4 border-t border-border/50 text-sm sm:text-base text-muted-foreground">
                    <div class="font-semibold mb-2">Notes:</div>
                    <div class="break-words italic font-medium">{{ log.notes }}</div>
                  </div>
                </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onActivated } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getWorkoutLogs } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { utcDateToLocal, localDateToUTC, isSameLocalDay, formatDateLocal } from '@/utils/dateUtils'

const router = useRouter()
const route = useRoute()
const currentWeekStart = ref(null)
const workoutLogs = ref([])
const weekKey = ref('') // Used to force watch to trigger
const selectedDate = ref(null) // Selected day for filtering

// Initialize current week start (Sunday)
const initializeWeek = () => {
  const today = new Date()
  const dayOfWeek = today.getDay() // 0 = Sunday, 6 = Saturday
  const weekStart = new Date(today)
  weekStart.setDate(today.getDate() - dayOfWeek)
  weekStart.setHours(0, 0, 0, 0)
  currentWeekStart.value = weekStart
  // Update week key to force watch to trigger (use local date string for key)
  const year = weekStart.getFullYear()
  const month = String(weekStart.getMonth() + 1).padStart(2, '0')
  const day = String(weekStart.getDate()).padStart(2, '0')
  weekKey.value = `${year}-${month}-${day}`
}

// Get week start date (Sunday)
const getWeekStartDate = (date) => {
  const d = new Date(date)
  const dayOfWeek = d.getDay()
  const weekStart = new Date(d)
  weekStart.setDate(d.getDate() - dayOfWeek)
  weekStart.setHours(0, 0, 0, 0)
  return weekStart
}

// Get week end date (Saturday)
const getWeekEndDate = (date) => {
  const weekStart = getWeekStartDate(date)
  const weekEnd = new Date(weekStart)
  weekEnd.setDate(weekStart.getDate() + 6)
  weekEnd.setHours(23, 59, 59, 999)
  return weekEnd
}

// Generate array of 7 day objects
const getWeekDays = (startDate) => {
  const days = []
  const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const todayLocalDate = today.toISOString().split('T')[0]

  for (let i = 0; i < 7; i++) {
    const date = new Date(startDate)
    date.setDate(startDate.getDate() + i)
    
    // Get local date string for this day
    const localDateString = utcDateToLocal(date.toISOString().split('T')[0])
    const dayName = dayNames[date.getDay()]
    const dayNumber = date.getDate()
    
    // Check if this is today in local timezone
    const isToday = localDateString === todayLocalDate
    
    // Check if this day has workouts (compare using UTC dates)
    const hasWorkout = workoutLogs.value.some(log => {
      // log.date is in UTC, convert to local for comparison
      const logLocalDate = utcDateToLocal(log.date)
      return logLocalDate === localDateString
    })

    days.push({
      date,
      dateString: localDateString, // Use local date string for selection
      utcDateString: date.toISOString().split('T')[0], // Keep UTC for API calls
      dayName,
      dayNumber,
      isToday,
      hasWorkout
    })
  }

  return days
}

// Group logs by date (using local dates for grouping)
const groupLogsByDate = (logs) => {
  const grouped = {}
  
  logs.forEach(log => {
    // Convert UTC date to local date for grouping
    const localDate = utcDateToLocal(log.date)
    if (!grouped[localDate]) {
      grouped[localDate] = {
        date: localDate, // Use local date for display
        utcDate: log.date, // Keep UTC date for reference
        logs: []
      }
    }
    grouped[localDate].logs.push(log)
  })

  // Convert to array and sort by date (newest first)
  return Object.values(grouped).sort((a, b) => {
    return new Date(b.utcDate + 'T00:00:00Z') - new Date(a.utcDate + 'T00:00:00Z')
  })
}

// Format week range for display
const formatWeekRange = (startDate, endDate) => {
  const start = new Date(startDate)
  const end = new Date(endDate)
  
  const startMonth = start.toLocaleDateString('en-US', { month: 'short' })
  const startDay = start.getDate()
  const endMonth = end.toLocaleDateString('en-US', { month: 'short' })
  const endDay = end.getDate()
  const year = start.getFullYear()

  if (startMonth === endMonth) {
    return `${startMonth} ${startDay} - ${endDay}, ${year}`
  } else {
    return `${startMonth} ${startDay} - ${endMonth} ${endDay}, ${year}`
  }
}

// Format day header (dateString is local date)
const formatDayHeader = (dateString) => {
  if (!dateString) return ''
  
  // Parse as local date at midnight
  const date = new Date(dateString + 'T00:00:00')
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const dateOnly = new Date(date)
  dateOnly.setHours(0, 0, 0, 0)
  
  const isToday = dateOnly.getTime() === today.getTime()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)
  const isYesterday = dateOnly.getTime() === yesterday.getTime()
  
  if (isToday) {
    return `Today, ${date.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}`
  } else if (isYesterday) {
    return `Yesterday, ${date.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}`
  } else {
    return date.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })
  }
}

// Format duration
const formatDuration = (minutes) => {
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  if (hours > 0) {
    return `${hours}h ${mins}m`
  }
  return `${mins}m`
}

// Computed properties
const weekDays = computed(() => {
  if (!currentWeekStart.value) return []
  return getWeekDays(currentWeekStart.value)
})

const weekRangeText = computed(() => {
  if (!currentWeekStart.value) return ''
  const weekEnd = getWeekEndDate(currentWeekStart.value)
  return formatWeekRange(currentWeekStart.value, weekEnd)
})

const groupedLogs = computed(() => {
  const grouped = groupLogsByDate(workoutLogs.value)
  
  // Filter by selected date if one is selected
  if (selectedDate.value) {
    return grouped.filter(dayGroup => dayGroup.date === selectedDate.value)
  }
  
  return grouped
})

// Navigation functions
const previousWeek = () => {
  if (!currentWeekStart.value) return
  const newWeekStart = new Date(currentWeekStart.value)
  newWeekStart.setDate(currentWeekStart.value.getDate() - 7)
  currentWeekStart.value = newWeekStart
  const year = newWeekStart.getFullYear()
  const month = String(newWeekStart.getMonth() + 1).padStart(2, '0')
  const day = String(newWeekStart.getDate()).padStart(2, '0')
  weekKey.value = `${year}-${month}-${day}`
  // Clear day selection when changing weeks
  selectedDate.value = null
  // Always fetch to ensure data is loaded
  fetchWorkoutLogs()
}

const nextWeek = () => {
  if (!currentWeekStart.value) return
  const newWeekStart = new Date(currentWeekStart.value)
  newWeekStart.setDate(currentWeekStart.value.getDate() + 7)
  currentWeekStart.value = newWeekStart
  const year = newWeekStart.getFullYear()
  const month = String(newWeekStart.getMonth() + 1).padStart(2, '0')
  const day = String(newWeekStart.getDate()).padStart(2, '0')
  weekKey.value = `${year}-${month}-${day}`
  // Clear day selection when changing weeks
  selectedDate.value = null
  // Always fetch to ensure data is loaded
  fetchWorkoutLogs()
}

const goToCurrentWeek = () => {
  const today = new Date()
  const dayOfWeek = today.getDay()
  const weekStart = new Date(today)
  weekStart.setDate(today.getDate() - dayOfWeek)
  weekStart.setHours(0, 0, 0, 0)
  
  // Always update and fetch, even if it's the same week
  currentWeekStart.value = weekStart
  const year = weekStart.getFullYear()
  const month = String(weekStart.getMonth() + 1).padStart(2, '0')
  const day = String(weekStart.getDate()).padStart(2, '0')
  weekKey.value = `${year}-${month}-${day}`
  // Clear day selection when going to current week
  selectedDate.value = null
  // Force fetch even if weekKey is the same
  fetchWorkoutLogs()
}

// Navigate to edit workout
const editWorkout = (logId) => {
  router.push(`/log-workout/${logId}/edit`)
}

// Select a day to filter exercises (dateString is local date)
const selectDay = (dateString) => {
  // Toggle selection: if clicking the same day, deselect it
  if (selectedDate.value === dateString) {
    selectedDate.value = null
  } else {
    selectedDate.value = dateString
  }
}

// Clear day selection
const clearDaySelection = () => {
  selectedDate.value = null
}

// Fetch workout logs for current week
const fetchWorkoutLogs = async () => {
  if (!currentWeekStart.value) return
  
  try {
    const weekEnd = getWeekEndDate(currentWeekStart.value)
    // Convert local week dates to UTC for API query
    // currentWeekStart is a local Date object, so we need to get the local date string first
    const startYear = currentWeekStart.value.getFullYear()
    const startMonth = String(currentWeekStart.value.getMonth() + 1).padStart(2, '0')
    const startDay = String(currentWeekStart.value.getDate()).padStart(2, '0')
    const startLocalDate = `${startYear}-${startMonth}-${startDay}`
    
    const endYear = weekEnd.getFullYear()
    const endMonth = String(weekEnd.getMonth() + 1).padStart(2, '0')
    const endDay = String(weekEnd.getDate()).padStart(2, '0')
    const endLocalDate = `${endYear}-${endMonth}-${endDay}`
    
    const startDateString = localDateToUTC(startLocalDate)
    const endDateString = localDateToUTC(endLocalDate)
    
    const response = await getWorkoutLogs({
      start_date: startDateString,
      end_date: endDateString
    })
    // Logs come back with UTC dates, which is what we want
    workoutLogs.value = response.logs
  } catch (error) {
    console.error('Failed to load workout logs:', error)
    workoutLogs.value = []
  }
}

// Watch for week changes using weekKey to ensure it triggers
watch(weekKey, () => {
  fetchWorkoutLogs()
}, { immediate: false })

// Also watch currentWeekStart as backup
watch(currentWeekStart, (newVal, oldVal) => {
  if (newVal && oldVal && newVal.getTime() !== oldVal.getTime()) {
    fetchWorkoutLogs()
  }
}, { immediate: false })

// Watch route to refresh when returning to dashboard (e.g., after editing)
watch(() => route.path, (newPath, oldPath) => {
  if (newPath === '/' && oldPath && oldPath !== '/') {
    // User navigated back to dashboard, refresh data
    fetchWorkoutLogs()
  }
})

// Refresh data when component is activated (e.g., returning from edit page)
onActivated(() => {
  fetchWorkoutLogs()
})

onMounted(() => {
  initializeWeek()
  fetchWorkoutLogs()
})
</script>
