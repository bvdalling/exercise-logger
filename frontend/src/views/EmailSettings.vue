<template>
  <div class="min-h-screen">
    <div class="container mx-auto px-5 py-10 sm:px-6 sm:py-12 max-w-2xl">
      <Card>
        <CardHeader class="p-6 sm:p-8">
          <CardTitle class="text-xl sm:text-2xl font-bold">Email Settings</CardTitle>
          <CardDescription class="text-sm sm:text-base font-medium mt-2">
            Manage your email preferences and weekly reports
          </CardDescription>
        </CardHeader>
        <CardContent class="p-6 sm:p-8">
          <div v-if="loading" class="text-center py-12 text-lg">Loading...</div>
          
          <div v-else class="space-y-6">
            <!-- Current Email -->
            <div class="space-y-3">
              <Label class="text-base font-semibold">Current Email</Label>
              <Input
                :value="user?.email || 'Not set'"
                disabled
                class="bg-muted"
              />
            </div>

            <!-- Weekly Report Toggle -->
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <div>
                  <Label class="text-base font-semibold">Weekly Report Emails</Label>
                  <p class="text-sm text-muted-foreground mt-1">
                    Receive a weekly summary of your workouts via email
                  </p>
                </div>
                <Button
                  :variant="weeklyReportEnabled ? 'default' : 'outline'"
                  @click="toggleWeeklyReports"
                  :disabled="saving"
                  size="sm"
                >
                  {{ weeklyReportEnabled ? 'Enabled' : 'Disabled' }}
                </Button>
              </div>
            </div>

            <!-- Send Weekly Report Button -->
            <div class="space-y-3">
              <Label class="text-base font-semibold">Send Weekly Report</Label>
              <p class="text-sm text-muted-foreground mb-3">
                Request an email with your weekly workout summary
              </p>
              <Button
                @click="sendWeeklyReport"
                :disabled="sendingReport || !user?.email"
                variant="outline"
                class="w-full"
              >
                {{ sendingReport ? 'Sending...' : 'Send Weekly Report Now' }}
              </Button>
            </div>

            <div v-if="error" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
              {{ error }}
            </div>

            <div v-if="successMessage" class="text-base text-green-600 dark:text-green-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
              {{ successMessage }}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getCurrentUser, sendWeeklyReport as apiSendWeeklyReport } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const authStore = useAuthStore()
const user = ref(null)
const loading = ref(true)
const saving = ref(false)
const sendingReport = ref(false)
const error = ref(null)
const successMessage = ref(null)
const weeklyReportEnabled = ref(true)

const loadUser = async () => {
  try {
    loading.value = true
    const response = await getCurrentUser()
    user.value = response.user
    weeklyReportEnabled.value = response.user.weekly_report_enabled !== false
  } catch (err) {
    console.error('Failed to load user:', err)
    error.value = 'Failed to load user settings'
  } finally {
    loading.value = false
  }
}

const toggleWeeklyReports = async () => {
  // Note: This would require a backend endpoint to update the preference
  // For now, we'll just show a message
  weeklyReportEnabled.value = !weeklyReportEnabled.value
  successMessage.value = 'Weekly report preference updated (backend endpoint needed to persist)'
  setTimeout(() => {
    successMessage.value = null
  }, 3000)
}

const sendWeeklyReport = async () => {
  if (!user.value?.email) {
    error.value = 'Email not set for your account'
    return
  }

  try {
    sendingReport.value = true
    error.value = null
    successMessage.value = null

    await apiSendWeeklyReport()
    successMessage.value = 'Weekly report sent successfully! Check your email.'
    setTimeout(() => {
      successMessage.value = null
    }, 5000)
  } catch (err) {
    error.value = err.message || 'Failed to send weekly report'
  } finally {
    sendingReport.value = false
  }
}

onMounted(() => {
  loadUser()
})
</script>

