import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(to, from, savedPosition) {
    // If navigating to a new route (not back/forward), scroll to top
    if (savedPosition) {
      // User is navigating back/forward - restore saved position
      return savedPosition
    } else {
      // New navigation - scroll to top
      return { top: 0, behavior: 'smooth' }
    }
  },
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/forgot-password',
      name: 'ForgotPassword',
      component: () => import('@/views/ForgotPassword.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/reset-password',
      name: 'ResetPassword',
      component: () => import('@/views/ResetPassword.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/setup-totp',
      name: 'TOTPSetup',
      component: () => import('@/views/TOTPSetup.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('@/views/Dashboard.vue'),
      meta: { requiresAuth: true, title: 'Dashboard' }
    },
    {
      path: '/exercises',
      name: 'Exercises',
      component: () => import('@/views/Exercises.vue'),
      meta: { requiresAuth: true, title: 'Exercises' }
    },
    {
      path: '/exercises/new',
      name: 'ExerciseNew',
      component: () => import('@/views/ExerciseForm.vue'),
      meta: { requiresAuth: true, title: 'New Exercise' }
    },
    {
      path: '/exercises/:id/edit',
      name: 'ExerciseEdit',
      component: () => import('@/views/ExerciseForm.vue'),
      meta: { requiresAuth: true, title: 'Edit Exercise' }
    },
    {
      path: '/exercises/:id/progress',
      name: 'ExerciseProgress',
      component: () => import('@/views/ExerciseProgress.vue'),
      meta: { requiresAuth: true, title: 'Exercise Progress' }
    },
    {
      path: '/log-workout',
      name: 'LogWorkout',
      component: () => import('@/views/LogWorkout.vue'),
      meta: { requiresAuth: true, title: 'Log Workout' }
    },
    {
      path: '/log-workout/:id/edit',
      name: 'EditWorkout',
      component: () => import('@/views/LogWorkout.vue'),
      meta: { requiresAuth: true, title: 'Edit Workout' }
    },
    {
      path: '/active-workout',
      name: 'ActiveWorkout',
      component: () => import('@/views/ActiveWorkout.vue'),
      meta: { requiresAuth: true, title: 'Active Workout' }
    },
    {
      path: '/email-settings',
      name: 'EmailSettings',
      component: () => import('@/views/EmailSettings.vue'),
      meta: { requiresAuth: true, title: 'Email Settings' }
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // Check authentication status if not already checked
  if (!authStore.user) {
    await authStore.checkAuth()
  }

  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const requiresGuest = to.matched.some(record => record.meta.requiresGuest)

  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (requiresGuest && authStore.isAuthenticated) {
    next('/')
  } else if (requiresAuth && authStore.isAuthenticated) {
    // Check if TOTP is required but not set up
    // Allow access to TOTP setup page
    if (to.path !== '/setup-totp' && authStore.user && authStore.user.totp_enabled === false) {
      next('/setup-totp')
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router

