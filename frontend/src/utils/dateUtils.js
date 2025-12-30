/**
 * Date utility functions for handling UTC storage and local timezone display
 */

/**
 * Converts a local date string (YYYY-MM-DD) to UTC date string
 * This ensures the date represents the same calendar day in UTC
 * @param {string} localDateString - Date string in local timezone (YYYY-MM-DD)
 * @returns {string} UTC date string (YYYY-MM-DD)
 */
export function localDateToUTC(localDateString) {
  if (!localDateString) return null
  
  // Parse as local date at midnight
  const localDate = new Date(localDateString + 'T00:00:00')
  
  // Get UTC date components
  const year = localDate.getUTCFullYear()
  const month = String(localDate.getUTCMonth() + 1).padStart(2, '0')
  const day = String(localDate.getUTCDate()).padStart(2, '0')
  
  return `${year}-${month}-${day}`
}

/**
 * Converts a UTC date string (YYYY-MM-DD) to local date string
 * This ensures the date is displayed in the user's local timezone
 * @param {string} utcDateString - Date string in UTC (YYYY-MM-DD)
 * @returns {string} Local date string (YYYY-MM-DD)
 */
export function utcDateToLocal(utcDateString) {
  if (!utcDateString) return null
  
  // Parse as UTC date at midnight
  const utcDate = new Date(utcDateString + 'T00:00:00Z')
  
  // Get local date components
  const year = utcDate.getFullYear()
  const month = String(utcDate.getMonth() + 1).padStart(2, '0')
  const day = String(utcDate.getDate()).padStart(2, '0')
  
  return `${year}-${month}-${day}`
}

/**
 * Converts a UTC datetime string to local Date object
 * @param {string} utcDateTimeString - ISO datetime string in UTC
 * @returns {Date} Local Date object
 */
export function utcDateTimeToLocal(utcDateTimeString) {
  if (!utcDateTimeString) return null
  return new Date(utcDateTimeString)
}

/**
 * Converts a local Date object to UTC datetime string
 * @param {Date} localDate - Local Date object
 * @returns {string} ISO datetime string in UTC
 */
export function localDateTimeToUTC(localDate) {
  if (!localDate) return null
  return localDate.toISOString()
}

/**
 * Gets today's date in local timezone as YYYY-MM-DD
 * @returns {string} Today's date in local timezone
 */
export function getTodayLocal() {
  const today = new Date()
  const year = today.getFullYear()
  const month = String(today.getMonth() + 1).padStart(2, '0')
  const day = String(today.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Formats a date string for display in local timezone
 * @param {string} dateString - Date string (YYYY-MM-DD) in UTC
 * @param {object} options - Intl.DateTimeFormat options
 * @returns {string} Formatted date string
 */
export function formatDateLocal(dateString, options = {}) {
  if (!dateString) return ''
  
  // Parse as UTC date at midnight
  const utcDate = new Date(dateString + 'T00:00:00Z')
  
  const defaultOptions = {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    ...options
  }
  
  return utcDate.toLocaleDateString(undefined, defaultOptions)
}

/**
 * Compares two date strings (UTC) and returns if they represent the same local day
 * @param {string} date1 - First date string (YYYY-MM-DD) in UTC
 * @param {string} date2 - Second date string (YYYY-MM-DD) in UTC
 * @returns {boolean} True if dates represent the same local day
 */
export function isSameLocalDay(date1, date2) {
  if (!date1 || !date2) return false
  
  const local1 = utcDateToLocal(date1)
  const local2 = utcDateToLocal(date2)
  
  return local1 === local2
}

