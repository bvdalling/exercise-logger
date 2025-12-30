<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-login dark:bg-gradient-login-dark px-5 py-10">
    <Card class="w-full max-w-md glass-strong dark:glass-strong">
      <CardHeader class="p-8">
        <CardTitle class="text-3xl sm:text-4xl font-bold text-center">Forgot Password</CardTitle>
        <CardDescription class="text-center text-base font-medium mt-3">
          Enter your email address and we'll send you a password reset link
        </CardDescription>
      </CardHeader>
      <CardContent class="p-8">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <div class="space-y-3">
            <Label for="email" class="text-base font-semibold">Email</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              placeholder="Enter your email"
              required
              :disabled="loading"
            />
          </div>
          <div v-if="error" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            {{ error }}
          </div>
          <div v-if="successMessage" class="text-base text-green-600 dark:text-green-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            {{ successMessage }}
          </div>
          <Button
            type="submit"
            variant="default"
            class="w-full"
            size="lg"
            :disabled="loading || successMessage"
          >
            {{ loading ? 'Sending...' : 'Send Reset Link' }}
          </Button>
          <div class="text-center text-base">
            <button
              type="button"
              @click="router.push('/login')"
              class="text-primary hover:underline"
              :disabled="loading"
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { requestPasswordReset } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const router = useRouter()
const email = ref('')
const loading = ref(false)
const error = ref(null)
const successMessage = ref('')

const handleSubmit = async () => {
  if (!email.value) {
    error.value = 'Please enter your email address'
    return
  }

  try {
    loading.value = true
    error.value = null
    successMessage.value = ''
    
    await requestPasswordReset(email.value)
    successMessage.value = 'If the email exists, a password reset link has been sent. Please check your inbox.'
  } catch (err) {
    error.value = err.message || 'Failed to send reset email'
  } finally {
    loading.value = false
  }
}
</script>

