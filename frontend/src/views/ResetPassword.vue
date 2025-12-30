<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-login dark:bg-gradient-login-dark px-5 py-10">
    <Card class="w-full max-w-md glass-strong dark:glass-strong">
      <CardHeader class="p-8">
        <CardTitle class="text-3xl sm:text-4xl font-bold text-center">Reset Password</CardTitle>
        <CardDescription class="text-center text-base font-medium mt-3">
          Enter your recovery credentials to reset your password
        </CardDescription>
      </CardHeader>
      <CardContent class="p-8">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <div class="space-y-3">
            <Label for="recoveryUuid" class="text-base font-semibold">Recovery ID (UUID)</Label>
            <Input
              id="recoveryUuid"
              v-model="recoveryUuid"
              type="text"
              placeholder="Enter your recovery UUID"
              required
              :disabled="authStore.loading"
              class="font-mono text-sm"
            />
          </div>
          <div class="space-y-3">
            <Label for="recoverySecret" class="text-base font-semibold">Recovery Secret</Label>
            <Input
              id="recoverySecret"
              v-model="recoverySecret"
              type="text"
              placeholder="Enter your 32-character recovery secret"
              required
              :disabled="authStore.loading"
              class="font-mono"
            />
            <p class="text-sm text-muted-foreground font-medium">
              Enter the 32-character recovery secret (without hyphens)
            </p>
          </div>
          <div class="space-y-3">
            <Label for="newPassword" class="text-base font-semibold">New Password</Label>
            <Input
              id="newPassword"
              v-model="newPassword"
              type="password"
              placeholder="Enter your new password"
              required
              :disabled="authStore.loading"
            />
          </div>
          <div class="space-y-3">
            <Label for="confirmPassword" class="text-base font-semibold">Confirm New Password</Label>
            <Input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              placeholder="Confirm your new password"
              required
              :disabled="authStore.loading"
            />
          </div>
          <div v-if="passwordMismatch" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            Passwords do not match
          </div>
          <div v-if="authStore.error" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            {{ authStore.error }}
          </div>
          <div v-if="successMessage" class="text-base text-green-600 dark:text-green-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            {{ successMessage }}
          </div>
          <Button
            type="submit"
            variant="default"
            class="w-full"
            size="lg"
            :disabled="authStore.loading || passwordMismatch"
          >
            {{ authStore.loading ? 'Resetting...' : 'Reset Password' }}
          </Button>
          <div class="text-center text-base">
            <button
              type="button"
              @click="router.push('/login')"
              class="text-primary hover:underline"
              :disabled="authStore.loading"
            >
              Back to Login
            </button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const router = useRouter()
const authStore = useAuthStore()

const recoveryUuid = ref('')
const recoverySecret = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const successMessage = ref('')

const passwordMismatch = computed(() => {
  return newPassword.value && confirmPassword.value && newPassword.value !== confirmPassword.value
})

const handleSubmit = async () => {
  if (passwordMismatch.value) {
    return
  }

  if (newPassword.value.length < 6) {
    authStore.error = 'Password must be at least 6 characters'
    return
  }

  // Remove any hyphens from recovery secret if user entered them
  const cleanSecret = recoverySecret.value.replace(/-/g, '')

  if (cleanSecret.length !== 32) {
    authStore.error = 'Recovery secret must be exactly 32 characters'
    return
  }

  try {
    await authStore.resetPassword(recoveryUuid.value, cleanSecret, newPassword.value)
    successMessage.value = 'Password reset successfully! Redirecting to login...'
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (error) {
    // Error is handled by the store
  }
}
</script>

