/**
 * Utility functions for handling timezone-aware date parsing
 */
import { format, differenceInDays } from "date-fns"

/**
 * Parses a deadline value with timezone awareness
 * Handles RFC3339 format timestamps (e.g., "2024-01-15T00:00:00Z" or "2024-01-15T00:00:00-05:00")
 * @param deadline - The deadline value (Date object, string, or any other type)
 * @returns A valid Date object, or current date as fallback
 */
export function parseDeadline(deadline: any): Date {
  // If it's already a Date object, return it
  if (deadline instanceof Date) {
    return deadline
  }

  // If it's a string, handle timezone-aware parsing
  if (typeof deadline === 'string') {
    try {
      // Handle RFC3339 format (e.g., "2024-01-15T00:00:00Z" or "2024-01-15T00:00:00-05:00")
      const date = new Date(deadline)
      
      // Check if the date is valid
      if (isNaN(date.getTime())) {
        console.warn("Invalid deadline format:", deadline)
        return new Date() // Fallback to current date
      }
      
      return date
    } catch (error) {
      console.warn("Error parsing deadline:", deadline, error)
      return new Date() // Fallback to current date
    }
  }

  // For any other type, try to create a Date object
  try {
    const date = new Date(deadline)
    if (isNaN(date.getTime())) {
      console.warn("Invalid deadline value:", deadline)
      return new Date() // Fallback to current date
    }
    return date
  } catch (error) {
    console.warn("Error creating Date from deadline:", deadline, error)
    return new Date() // Fallback to current date
  }
}

/**
 * Formats a date for display with timezone awareness
 * @param date - The date to format
 * @param formatString - The format string (default: "MMM d, yyyy")
 * @returns Formatted date string
 */
export function formatDeadline(date: Date, formatString: string = "MMM d, yyyy"): string {
  try {
    return format(date, formatString)
  } catch (error) {
    console.warn("Error formatting date:", date, error)
    return date.toLocaleDateString() // Fallback to browser's locale format
  }
}

/**
 * Calculates the difference in days between a deadline and current date
 * @param deadline - The deadline date
 * @returns Number of days difference (negative if overdue, positive if upcoming)
 */
export function calculateDaysDifference(deadline: Date): number {
  var difference = differenceInDays(deadline, new Date())
  return difference < 0 ? difference : difference + 1
}

/**
 * Checks if a deadline is overdue
 * @param deadline - The deadline date
 * @param status - The assignment status
 * @returns True if the assignment is overdue
 */
export function isOverdue(deadline: Date, status: string): boolean {
  const daysDifference = calculateDaysDifference(deadline)
  return daysDifference < 0 && status !== "Done"
}

/**
 * Gets a human-readable description of when an assignment is due
 * @param deadline - The deadline date
 * @param status - The assignment status
 * @returns Human-readable string describing when the assignment is due
 */
export function getDueDescription(deadline: Date, status: string): string {
  if (status === "Done") {
    return "Completed"
  }

  const daysDifference = calculateDaysDifference(deadline)
  
  if (daysDifference <= 0) {
    return `${Math.abs(daysDifference)} days overdue`
  } else if (daysDifference === 1) {
    return "Due today"
  } else if (daysDifference === 2) {
    return "Due tomorrow"
  } else {
    return `${daysDifference} days left`
  }
} 