/**
 * Utility functions for handling timezone-aware date parsing
 */
import { format, differenceInDays, isTomorrow, isToday, isBefore, isAfter, addDays, differenceInMinutes, weeksToDays, isWithinInterval } from "date-fns"
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
  startHour: number // 24-hour format
  startMinute: number
  nextClassStart: Date
  endHour: number
  endMinute: number
  nextClassEnd: Date
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

    // Calculate the start and end dates for the next class
    const now = new Date()
    const dayIndex = now.getDay()
    var next_date = now

    var valid_days = days.filter(day => day >= dayIndex)
    if (valid_days.length > 0) {
      console.log(schedule, valid_days.length, Math.max(...valid_days) - dayIndex)
      next_date = addDays(now, Math.min(...valid_days) - dayIndex)
    } else {
      next_date = addDays(now,  (7 - dayIndex) + Math.min(...days))
    }
    
    const nextClassStart = new Date(next_date.getFullYear(), next_date.getMonth(), next_date.getDate(), start24, startMinute, 0, 0)
    const nextClassEnd = new Date(next_date.getFullYear(), next_date.getMonth(), next_date.getDate(), end24, endMinute, 0, 0)
     return {
      days,
      startHour: start24,
      endHour: end24,
      nextClassStart: nextClassStart,
      nextClassEnd: nextClassEnd,
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

export function getNextCourse(courses: course.LocalCourse[]): { course: course.LocalCourse | null, isOn: boolean, until: number | null } {
  const now = new Date()
  var isOn = false
  var until = null


  var next_course = courses
    // Filter out classes that are already over
    .filter(course => {

      const schedule = parseSchedule(course.Schedule)

      if (schedule && isAfter(parseDeadline(course.EndDate), now) ) {
        return differenceInMinutes(schedule.nextClassEnd, now) > 0
      }

      return false
    })
    // Sort by the next class end date
    .sort((a, b) => {
      const scheduleA = parseSchedule(a.Schedule)
      const scheduleB = parseSchedule(b.Schedule)
      return differenceInMinutes(scheduleA?.nextClassEnd || new Date(), scheduleB?.nextClassEnd || new Date())
    })[0]

  if (next_course) {
    const schedule = parseSchedule(next_course.Schedule)
    if (schedule) {
      isOn = isBefore(schedule.nextClassStart, now) && isAfter(schedule.nextClassEnd, now)

      until = isOn ? differenceInMinutes(schedule.nextClassEnd, now) : differenceInMinutes(schedule.nextClassStart, now)
    }
  }


  return { course: next_course, isOn: isOn, until: until }
}

export function IsOn(schedule: ParsedSchedule): boolean {
  const startHour = schedule.startHour
  const startMinute = parseInt(schedule.startTimeString.split(":")[1])

  const endHour = schedule.endHour
  const endMinute = parseInt(schedule.endTimeString.split(":")[1])

  const today = new Date()
  const startDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), startHour, startMinute)
  const endDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), endHour, endMinute)

  return isBefore(startDate, today) && isAfter(endDate, today)
}
