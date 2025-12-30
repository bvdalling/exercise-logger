// Use relative URL to leverage Vite proxy, or absolute URL if VITE_API_URL is set
const API_BASE_URL = import.meta.env.VITE_API_URL || '/api'

async function request(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`
  const config = {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include', // Important for session cookies
  }

  try {
    const response = await fetch(url, config)
    
    // Check if response has content
    const contentType = response.headers.get('content-type')
    let data
    
    if (contentType && contentType.includes('application/json')) {
      const text = await response.text()
      if (text) {
        try {
          data = JSON.parse(text)
        } catch (parseError) {
          console.error('Failed to parse JSON response:', text)
          throw new Error('Invalid response from server')
        }
      } else {
        console.error('Empty response body')
        throw new Error('Empty response from server')
      }
    } else {
      // Non-JSON response
      const text = await response.text()
      console.error('Non-JSON response:', text)
      throw new Error(text || 'Unexpected response format')
    }

    if (!response.ok) {
      throw new Error(data?.error || data?.message || `Request failed with status ${response.status}`)
    }

    return data
  } catch (error) {
    console.error('API request error:', error)
    if (error.message) {
      throw error
    }
    throw new Error('Network error')
  }
}

export async function login(username, password) {
  return request('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export async function register(username, email, password) {
  return request('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, email, password }),
  })
}

export async function logout() {
  return request('/auth/logout', {
    method: 'POST',
  })
}

export async function getCurrentUser() {
  return request('/auth/me')
}

export async function requestPasswordReset(email) {
  return request('/auth/request-password-reset', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export async function resetPassword(token, newPassword) {
  return request('/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({ token, newPassword }),
  })
}

// Exercise APIs
export async function getExercises() {
  return request('/exercises')
}

export async function getExercise(id) {
  return request(`/exercises/${id}`)
}

export async function createExercise(exerciseData) {
  return request('/exercises', {
    method: 'POST',
    body: JSON.stringify(exerciseData),
  })
}

export async function updateExercise(id, exerciseData) {
  return request(`/exercises/${id}`, {
    method: 'PUT',
    body: JSON.stringify(exerciseData),
  })
}

export async function deleteExercise(id) {
  return request(`/exercises/${id}`, {
    method: 'DELETE',
  })
}

export async function getExerciseProgress(id) {
  return request(`/exercises/${id}/progress`)
}

// Public Exercise APIs
export async function getPublicExercises() {
  return request('/public-exercises')
}

export async function getPublicExercise(id) {
  return request(`/public-exercises/${id}`)
}

// Reports APIs
export async function sendWeeklyReport() {
  return request('/reports/weekly', {
    method: 'POST',
  })
}

// TOTP APIs
export async function setupTOTP(code) {
  return request('/auth/setup-totp', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
}

export async function verifyTOTP(username, code) {
  return request('/auth/verify-totp', {
    method: 'POST',
    body: JSON.stringify({ username, code }),
  })
}

// Workout Log APIs
export async function getWorkoutLogs(params = {}) {
  const queryString = new URLSearchParams(params).toString()
  const endpoint = queryString ? `/workout-logs?${queryString}` : '/workout-logs'
  return request(endpoint)
}

export async function getWorkoutLog(id) {
  return request(`/workout-logs/${id}`)
}

export async function createWorkoutLog(logData) {
  return request('/workout-logs', {
    method: 'POST',
    body: JSON.stringify(logData),
  })
}

export async function updateWorkoutLog(id, logData) {
  return request(`/workout-logs/${id}`, {
    method: 'PUT',
    body: JSON.stringify(logData),
  })
}

export async function deleteWorkoutLog(id) {
  return request(`/workout-logs/${id}`, {
    method: 'DELETE',
  })
}

export async function getLastWorkoutValues(exerciseId) {
  return request(`/workout-logs/exercise/${exerciseId}/last`)
}

