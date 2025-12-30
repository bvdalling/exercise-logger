<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-login dark:bg-gradient-login-dark px-5 py-10">
    <Card class="w-full max-w-md glass-strong dark:glass-strong">
      <CardHeader class="p-8">
        <CardTitle class="text-3xl sm:text-4xl font-bold text-center">Simple Logger</CardTitle>
        <CardDescription class="text-center text-base font-medium mt-3">
          {{ isLogin ? 'Sign in to your account' : 'Create a new account' }}
        </CardDescription>
      </CardHeader>
      <CardContent class="p-8">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <div class="space-y-3">
            <Label for="username" class="text-base font-semibold">Username</Label>
            <Input
              id="username"
              v-model="username"
              type="text"
              placeholder="Enter your username"
              required
              :disabled="authStore.loading"
            />
          </div>
          <div v-if="!isLogin" class="space-y-3">
            <Label for="email" class="text-base font-semibold">Email</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              placeholder="Enter your email"
              required
              :disabled="authStore.loading"
            />
          </div>
          <div class="space-y-3">
            <Label for="password" class="text-base font-semibold">Password</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="Enter your password"
              required
              :disabled="authStore.loading || requiresTOTP"
            />
          </div>
          <div v-if="requiresTOTP" class="space-y-3">
            <Label for="totpCode" class="text-base font-semibold">Enter 6-digit code from your authenticator app</Label>
            <Input
              id="totpCode"
              v-model="totpCode"
              type="text"
              placeholder="000000"
              maxlength="6"
              pattern="[0-9]{6}"
              class="text-center text-2xl font-mono tracking-widest"
              :disabled="verifyingTOTP"
              @input="handleTOTPInput"
            />
          </div>
          <div v-if="authStore.error" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            {{ authStore.error }}
          </div>
          <Button
            type="submit"
            variant="default"
            class="w-full"
            size="lg"
            :disabled="authStore.loading"
          >
            {{ authStore.loading ? 'Loading...' : (isLogin ? 'Sign In' : 'Sign Up') }}
          </Button>
          <div class="text-center text-base space-y-3">
            <button
              type="button"
              @click="toggleMode"
              class="text-primary hover:underline block"
              :disabled="authStore.loading"
            >
              {{ isLogin ? "Don't have an account? Sign up" : 'Already have an account? Sign in' }}
            </button>
            <button
              v-if="isLogin"
              type="button"
              @click="router.push('/forgot-password')"
              class="text-primary hover:underline block"
              :disabled="authStore.loading"
            >
              Forgot Password?
            </button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- Recovery Credentials Dialog -->
    <Dialog v-model:open="showRecoveryDialog">
      <DialogContent class="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Save Your Recovery Credentials</DialogTitle>
          <DialogDescription>
            <strong class="text-red-600 dark:text-red-400">IMPORTANT:</strong> Please screenshot or save these credentials. 
            You will need them to reset your password if you forget it. This is the only time they will be shown.
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div class="space-y-2">
            <Label>Recovery ID (UUID)</Label>
            <div class="flex gap-2">
              <Input
                :value="recoveryCredentials?.uuid || ''"
                readonly
                class="font-mono text-sm"
              />
              <Button
                type="button"
                variant="outline"
                @click="copyToClipboard(recoveryCredentials?.uuid || '')"
              >
                Copy
              </Button>
            </div>
          </div>
          <div class="space-y-2">
            <Label>Recovery Secret (32 characters)</Label>
            <div class="flex gap-2">
              <Input
                :value="formatRecoverySecret(recoveryCredentials?.secret || '')"
                readonly
                class="font-mono text-sm"
              />
              <Button
                type="button"
                variant="outline"
                @click="copyToClipboard(recoveryCredentials?.secret || '')"
              >
                Copy
              </Button>
            </div>
          </div>
          <div class="bg-yellow-50 dark:bg-yellow-900/20 rounded-xl p-3">
            <p class="text-sm text-yellow-800 dark:text-yellow-200">
              ⚠️ Make sure to save these credentials in a secure location. Without them, you cannot reset your password.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button @click="acknowledgeRecovery" class="w-full">
            I've Saved My Credentials
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useScrollRestore } from '@/composables/useScrollRestore'
import { verifyTOTP } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

const router = useRouter()
const authStore = useAuthStore()

const isLogin = ref(true)
const username = ref('')
const email = ref('')
const password = ref('')
const showRecoveryDialog = ref(false)
const recoveryCredentials = ref(null)
const requiresTOTP = ref(false)
const requiresTOTPSetup = ref(false)
const totpCode = ref('')
const verifyingTOTP = ref(false)

// Restore scroll position when modal closes
useScrollRestore(showRecoveryDialog)

const toggleMode = () => {
  isLogin.value = !isLogin.value
  authStore.clearError()
}

const formatRecoverySecret = (secret) => {
  if (!secret) return ''
  // Format as groups of 4 characters with hyphens for readability
  return secret.match(/.{1,4}/g)?.join('-') || secret
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    // You could add a toast notification here
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}

const acknowledgeRecovery = () => {
  showRecoveryDialog.value = false
  recoveryCredentials.value = null
  router.push('/')
}

const handleTOTPInput = (e) => {
  // Only allow numbers
  totpCode.value = e.target.value.replace(/\D/g, '').slice(0, 6)
  // Auto-submit when 6 digits entered
  if (totpCode.value.length === 6) {
    handleTOTPVerify()
  }
}

const handleTOTPVerify = async () => {
  if (!totpCode.value || totpCode.value.length !== 6) {
    authStore.error = 'Please enter a 6-digit code'
    return
  }

  try {
    verifyingTOTP.value = true
    authStore.error = null
    
    const response = await verifyTOTP(username.value, totpCode.value)
    authStore.user = response.user
    router.push('/')
  } catch (error) {
    authStore.error = error.message || 'Invalid TOTP code'
    totpCode.value = ''
  } finally {
    verifyingTOTP.value = false
  }
}

const handleSubmit = async () => {
  try {
    if (isLogin.value) {
      const response = await authStore.login(username.value, password.value)
      
      // Check if TOTP is required
      if (response.requiresTOTP) {
        requiresTOTP.value = true
        return
      }
      
      if (response.requiresTOTPSetup) {
        router.push('/setup-totp?from=login')
        return
      }
      
      router.push('/')
    } else {
      if (!email.value) {
        authStore.error = 'Email is required'
        return
      }
      const response = await authStore.register(username.value, email.value, password.value)
      
      if (response.requiresTOTPSetup) {
        router.push('/setup-totp?from=register')
        return
      }
      
      if (response.recovery) {
        recoveryCredentials.value = response.recovery
        showRecoveryDialog.value = true
      } else {
        router.push('/')
      }
    }
  } catch (error) {
    // Error is handled by the store
  }
}
</script>

