<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-login dark:bg-gradient-login-dark px-5 py-10">
    <Card class="w-full max-w-md glass-strong dark:glass-strong">
      <CardHeader class="p-8">
        <CardTitle class="text-3xl sm:text-4xl font-bold text-center">Set Up Two-Factor Authentication</CardTitle>
        <CardDescription class="text-center text-base font-medium mt-3">
          Scan the QR code with your authenticator app
        </CardDescription>
      </CardHeader>
      <CardContent class="p-8">
        <div v-if="loading" class="text-center py-12 text-lg">Setting up...</div>
        
        <div v-else-if="!qrCode" class="space-y-6">
          <Button
            @click="initiateSetup"
            :disabled="loading"
            class="w-full"
            size="lg"
          >
            Start TOTP Setup
          </Button>
        </div>

        <div v-else-if="!verified" class="space-y-6">
          <!-- QR Code -->
          <div class="flex justify-center">
            <img :src="qrCode" alt="TOTP QR Code" class="w-64 h-64 border-2 border-border rounded-lg" />
          </div>

          <div class="space-y-3">
            <Label for="totpCode" class="text-base font-semibold">Enter 6-digit code from your app</Label>
            <Input
              id="totpCode"
              v-model="totpCode"
              type="text"
              placeholder="000000"
              maxlength="6"
              pattern="[0-9]{6}"
              class="text-center text-2xl font-mono tracking-widest"
              @input="handleCodeInput"
            />
          </div>

          <div v-if="error" class="text-base text-red-600 dark:text-red-400 font-medium p-4 rounded-2xl glass-card dark:glass-card-dark">
            {{ error }}
          </div>

          <Button
            @click="verifyCode"
            :disabled="!totpCode || totpCode.length !== 6 || verifying"
            class="w-full"
            size="lg"
          >
            {{ verifying ? 'Verifying...' : 'Verify & Enable' }}
          </Button>
        </div>

        <!-- Backup Codes Display -->
        <div v-else class="space-y-6">
          <div class="bg-yellow-50 dark:bg-yellow-900/20 rounded-xl p-4">
            <p class="text-sm text-yellow-800 dark:text-yellow-200 font-semibold mb-3">
              ⚠️ Save these backup codes in a secure location!
            </p>
            <p class="text-sm text-yellow-800 dark:text-yellow-200 mb-4">
              You can use these codes to access your account if you lose your authenticator device.
            </p>
            <div class="space-y-2">
              <div
                v-for="(code, index) in backupCodes"
                :key="index"
                class="font-mono text-sm bg-white dark:bg-gray-800 p-2 rounded border"
              >
                {{ code }}
              </div>
            </div>
          </div>

          <Button
            @click="completeSetup"
            class="w-full"
            size="lg"
          >
            I've Saved My Backup Codes
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { setupTOTP, verifyTOTP } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const qrCode = ref('')
const totpCode = ref('')
const verifying = ref(false)
const verified = ref(false)
const backupCodes = ref([])
const error = ref(null)

const handleCodeInput = (e) => {
  // Only allow numbers
  totpCode.value = e.target.value.replace(/\D/g, '').slice(0, 6)
}

const initiateSetup = async () => {
  try {
    loading.value = true
    error.value = null
    
    const response = await setupTOTP('')
    qrCode.value = response.qrCode
  } catch (err) {
    error.value = err.message || 'Failed to initiate TOTP setup'
  } finally {
    loading.value = false
  }
}

const verifyCode = async () => {
  if (!totpCode.value || totpCode.value.length !== 6) {
    error.value = 'Please enter a 6-digit code'
    return
  }

  try {
    verifying.value = true
    error.value = null
    
    const response = await setupTOTP(totpCode.value)
    if (response.backupCodes && response.backupCodes.length > 0) {
      backupCodes.value = response.backupCodes
      verified.value = true
    } else {
      error.value = 'Verification failed'
    }
  } catch (err) {
    error.value = err.message || 'Invalid code. Please try again.'
  } finally {
    verifying.value = false
  }
}

const completeSetup = () => {
  router.push('/')
}

onMounted(() => {
  // Auto-initiate setup if coming from registration
  if (route.query.from === 'register') {
    initiateSetup()
  }
})
</script>

