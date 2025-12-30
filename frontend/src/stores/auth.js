import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, register as apiRegister, logout as apiLogout, getCurrentUser, resetPassword as apiResetPassword } from '@/composables/useApi'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!user.value)

  async function login(username, password) {
    loading.value = true
    error.value = null
    try {
      const response = await apiLogin(username, password)
      // Only set user if login is complete (not requiring TOTP)
      if (response.user && !response.requiresTOTP && !response.requiresTOTPSetup) {
        user.value = response.user
      }
      return response
    } catch (err) {
      error.value = err.message || 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(username, email, password) {
    loading.value = true
    error.value = null
    try {
      const response = await apiRegister(username, email, password)
      // Only set user if registration is complete (not requiring TOTP setup)
      if (response.user && !response.requiresTOTPSetup) {
        user.value = response.user
      }
      return response
    } catch (err) {
      error.value = err.message || 'Registration failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function resetPassword(token, newPassword) {
    loading.value = true
    error.value = null
    try {
      const response = await apiResetPassword(token, newPassword)
      return response
    } catch (err) {
      error.value = err.message || 'Password reset failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    loading.value = true
    error.value = null
    try {
      await apiLogout()
      user.value = null
    } catch (err) {
      error.value = err.message || 'Logout failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function checkAuth() {
    try {
      const response = await getCurrentUser()
      user.value = response.user
      return true
    } catch (err) {
      user.value = null
      return false
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    login,
    register,
    logout,
    checkAuth,
    clearError,
    resetPassword
  }
})

