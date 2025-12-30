const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3001/api'

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
    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'Request failed')
    }

    return data
  } catch (error) {
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

export async function register(username, password) {
  return request('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
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

export async function resetPassword(recoveryUuid, recoverySecret, newPassword) {
  return request('/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({ recoveryUuid, recoverySecret, newPassword }),
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

