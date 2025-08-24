/**
 * Utility functions for handling timezone-aware date parsing
 */
import { format, differenceInDays, isTomorrow, isToday, isBefore, isAfter, addDays, differenceInMinutes } from "date-fns"
import { course } from '@/wailsjs/go/models'
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
  return difference <= 0 ? difference : difference + 1
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

  if (daysDifference < 0) {
    return `${Math.abs(daysDifference)} days overdue`
  } else if (isToday(deadline)) {
    return "Due today"
  } else if (isTomorrow(deadline)) {
    return "Due tomorrow"
  } else {
    return `${daysDifference} days left`
  }
}

export interface ParsedSchedule {
  days: number[]
  startTime: number // 24-hour format
  startMinute: number
  endTime: number
  endMinute: number
  startTimeString: string
  endTimeString: string
}

/**
 * Parses a schedule string into a ParsedSchedule object
 * @param schedule - The schedule string to parse
 * @returns A ParsedSchedule object if successful, null if parsing fails
 */

export function parseSchedule(schedule: string): ParsedSchedule | null {
  if (!schedule) return null

  try {
    // Day abbreviations mapping
    const DAY_MAPPING: Record<string, number> = {
      'M': 1, 'Mo': 1, 'Mon': 1,
      'T': 2, 'Tu': 2, 'Tue': 2,
      'W': 3, 'We': 3, 'Wed': 3,
      'Th': 4, 'Thu': 4,
      'F': 5, 'Fr': 5, 'Fri': 5,
      'S': 6, 'Sa': 6, 'Sat': 6,
      'Su': 0, 'Sun': 0
    }
    // Split by time separator (looking for patterns like "1:00 PM - 2:00 PM")
    const timeMatch = schedule.match(/(\d{1,2}):(\d{2})\s*(AM|PM)\s*-\s*(\d{1,2}):(\d{2})\s*(AM|PM)/i)
    if (!timeMatch) return null

    const [, startHour, startMin, startPeriod, endHour, endMin, endPeriod] = timeMatch

    // Convert to 24-hour format
    let start24 = parseInt(startHour)
    let end24 = parseInt(endHour)
    let startMinute = parseInt(startMin)
    let endMinute = parseInt(endMin)

    if (startPeriod.toUpperCase() === 'PM' && start24 !== 12) start24 += 12
    if (startPeriod.toUpperCase() === 'AM' && start24 === 12) start24 = 0
    if (endPeriod.toUpperCase() === 'PM' && end24 !== 12) end24 += 12
    if (endPeriod.toUpperCase() === 'AM' && end24 === 12) end24 = 0

    // Parse days (everything before the time)
    const daysPart = schedule.split(/\d{1,2}:\d{2}/)[0].trim()
    const dayTokens = daysPart.split(/[,\s]+/).filter(token => token.length > 0)

    const days: number[] = []
    for (const token of dayTokens) {
      const day = DAY_MAPPING[token]
      if (day !== undefined) {
        days.push(day)
      }
    }

    return {
      days,
      startTime: start24,
      endTime: end24,
      startMinute: startMinute,
      endMinute: endMinute,
      startTimeString: `${startHour}:${startMin} ${startPeriod}`,
      endTimeString: `${endHour}:${endMin} ${endPeriod}`
    }
  } catch (error) {
    console.error('Error parsing schedule:', schedule, error)
    return null
  }
}

/**
 * Gets the next course to be taken
 * @param courses - The courses to search through
 * @returns The next course to be taken, or null if no courses are found
 */

export function getNextCourse(courses: course.LocalCourse[]): { course: course.LocalCourse | null, isOn: boolean, until:number | null } {
  const now = new Date()
  const dayIndex = now.getDay()
  var nextCourse = null
  var isOn = false
  var until = null

  for (var i = 0; i < 7; i++) {
    const day = (dayIndex + i) % 7
    var today_clases = []
    for (const course of courses) {
      const schedule = parseSchedule(course.Schedule)
      if (schedule?.days.includes(day) && ( i != 0 || ( schedule?.endTime >= now.getHours() && schedule?.endMinute >= now.getMinutes() ) )) {
        console.log(schedule?.days.includes(day),i != 0, schedule?.endTime >= now.getHours(), schedule?.endMinute >= now.getMinutes())
        today_clases.push(course)
      }
    }


    if (today_clases.length > 0) {

      nextCourse = today_clases.sort((a, b) => {
        const scheduleA = parseSchedule(a.Schedule)
        const scheduleB = parseSchedule(b.Schedule)
        return (scheduleA?.startTime || 0) - (scheduleB?.startTime || 0)
      })[0]

    }

    if (nextCourse) {
      const schedule = parseSchedule(nextCourse.Schedule)
      if (schedule) {
        if (schedule?.startTime && schedule?.endTime && schedule?.startTime <= now.getHours() && schedule?.endTime >= now.getHours()) {
          isOn = IsOn(schedule) && day == dayIndex
        }
      // Calculate the duration until the next course
        const next_class_date = new Date(now.getFullYear(), now.getMonth(), addDays(now, i).getDate(), schedule?.startTime, schedule?.startMinute)
        until = differenceInMinutes(next_class_date, now)
   
      }

      break
    }

  }

  return { course: nextCourse, isOn: isOn, until: until }
}

export function IsOn(schedule: ParsedSchedule):  boolean {
  const startHour = schedule.startTime
  const startMinute = parseInt(schedule.startTimeString.split(":")[1])

  const endHour = schedule.endTime
  const endMinute = parseInt(schedule.endTimeString.split(":")[1])

  const today = new Date()
  const startDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), startHour, startMinute)
  const endDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), endHour, endMinute)

  return isBefore(startDate, today) && isAfter(endDate, today)
}