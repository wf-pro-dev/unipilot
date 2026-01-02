import * as z from "zod"

// Validation helper functions matching backend logic
const isValidCourseCode = (code: string): boolean => {
  // Course code validation: alphanumeric with spaces allowed (e.g., "CS 101")
  const pattern = /^[a-zA-Z0-9\s-]+$/
  return pattern.test(code)
}

const isValidCourseName = (name: string): boolean => {
  // Course name can contain letters, numbers, spaces, and common punctuation
  const pattern = /^[a-zA-Z0-9\s\-.,:;()&]+$/
  return pattern.test(name)
}

const isValidEmail = (email: string): boolean => {
  const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailPattern.test(email)
}

const isValidSchedule = (schedule: string): { valid: boolean; error?: string } => {
  if (schedule.length === 0) {
    return { valid: false, error: "Schedule is required" }
  }

  // Accept async formats
  if (schedule.toLowerCase() === "async" || schedule.toLowerCase() === "asynchronous") {
    return { valid: true }
  }

  // Validate format: "<day>, <day> <hour>:<minutes> <period> - <hour>:<minutes> <period>"
  const schedulePattern = /^((?:M|T|W|Th|F|Sa|Su)(?:,\s(?:M|T|W|Th|F|Sa|Su))*)\s+(\d{1,2}:[0-5]\d\s(?:AM|PM))\s-\s(\d{1,2}:[0-5]\d\s(?:AM|PM))$/

  const match = schedule.match(schedulePattern)

  if (!match) {
    return { 
      valid: false, 
      error: "Invalid schedule format. Expected: 'M, T, W 9:00 AM - 10:30 AM'" 
    }
  }

  const [, daysStr, startTime, endTime] = match

  // Validate at least one day
  if (!daysStr || daysStr.trim().length === 0) {
    return { valid: false, error: "At least one day must be specified" }
  }

  // Validate individual days
  const days = daysStr.trim().split(', ')
  const validDays = ['M', 'T', 'W', 'Th', 'F', 'Sa', 'Su']

  for (const day of days) {
    if (!validDays.includes(day)) {
      return { 
        valid: false, 
        error: `Invalid day '${day}'. Valid days: M, T, W, Th, F, Sa, Su` 
      }
    }
  }

  // Validate time format
  const timePattern = /^(\d{1,2}):([0-5]\d)\s(AM|PM)$/
  const startMatch = startTime.match(timePattern)
  const endMatch = endTime.match(timePattern)

  if (!startMatch || !endMatch) {
    return { 
      valid: false, 
      error: "Invalid time format. Use format like '9:00 AM' or '12:30 PM'" 
    }
  }

  // Validate hour ranges (1-12 for 12-hour format)
  const startHour = parseInt(startMatch[1])
  const endHour = parseInt(endMatch[1])

  if (startHour < 1 || startHour > 12 || endHour < 1 || endHour > 12) {
    return { valid: false, error: "Hour must be between 1 and 12" }
  }

  // Convert to minutes for comparison
  const convertToMinutes = (hour: number, minute: number, period: string): number => {
    if (period === 'AM') {
      return hour === 12 ? 0 * 60 + minute : hour * 60 + minute
    } else {
      return hour === 12 ? 12 * 60 + minute : (hour + 12) * 60 + minute
    }
  }

  const startMinutes = convertToMinutes(startHour, parseInt(startMatch[2]), startMatch[3])
  const endMinutes = convertToMinutes(endHour, parseInt(endMatch[2]), endMatch[3])

  if (startMinutes >= endMinutes) {
    return { valid: false, error: "Start time must be before end time" }
  }

  return { valid: true }
}

// Valid semesters matching backend
const validSemesters = [
  "FALL 2024", "SPRING 2025", "SUMMER 2025", "FALL 2025",
  "SPRING 2026", "SUMMER 2026", "FALL 2026", "SPRING 2027",
  "SUMMER 2027", "FALL 2027", "SPRING 2028", "SUMMER 2028"
] as const

// Valid colors matching frontend options
const validColors = [
  "bg-blue-500", "bg-green-500", "bg-purple-500", 
  "bg-red-500", "bg-orange-500", "bg-pink-500"
] as const

// Step 1: Basic course information
export const courseStep1Schema = z.object({
  name: z
    .string()
    .min(3, "Course name must be at least 3 characters")
    .max(100, "Course name must be at most 100 characters")
    .refine(
      (val) => isValidCourseName(val),
      {
        message: "Course name can only contain letters, numbers, spaces, and common punctuation",
      }
    ),
  code: z
    .string()
    .min(2, "Course code must be at least 2 characters")
    .max(20, "Course code must be at most 20 characters")
    .refine(
      (val) => isValidCourseCode(val),
      {
        message: "Course code can only contain letters, numbers, spaces, and hyphens",
      }
    ),
  credits: z
    .string()
    .refine(
      (val) => {
        const num = parseInt(val)
        return !isNaN(num) && num >= 1 && num <= 4
      },
      {
        message: "Credits must be between 1 and 4",
      }
    ),
  semester: z
    .string()
    .min(1, "Please select a semester")
    .refine(
      (val) => validSemesters.includes(val as typeof validSemesters[number]),
      {
        message: "Please select a valid semester",
      }
    ),
})

// Step 2: Schedule and location
export const courseStep2Schema = z.object({
  schedule: z
    .string()
    .min(1, "Schedule is required")
    .max(100, "Schedule must be at most 100 characters")
    .refine(
      (val) => {
        const result = isValidSchedule(val)
        return result.valid
      },
      (val) => {
        const result = isValidSchedule(val)
        return {
          message: result.error || "Invalid schedule format",
        }
      }
    ),
  location: z
    .string()
    .min(3, "Location must be at least 3 characters")
    .max(100, "Location must be at most 100 characters"),
  startDate: z.date({
    required_error: "Start date is required",
  }),
  endDate: z.date({
    required_error: "End date is required",
  }),
}).refine(
  (data) => data.startDate <= data.endDate,
  {
    message: "Start date must be before end date",
    path: ["endDate"],
  }
)

// Step 3: Instructor and color
export const courseStep3Schema = z.object({
  instructor: z
    .string()
    .min(3, "Instructor name must be at least 3 characters")
    .max(100, "Instructor name must be at most 100 characters"),
  instructor_email: z
    .string()
    .email("Please enter a valid email address")
    .refine(
      (val) => isValidEmail(val),
      {
        message: "Please enter a valid email address",
      }
    ),
  color: z
    .string()
    .min(1, "Please select a color")
    .refine(
      (val) => validColors.includes(val as typeof validColors[number]),
      {
        message: "Please select a valid color",
      }
    ),
})

// Combined schema
export const courseSchema = z.intersection(
  courseStep1Schema,
  z.intersection(courseStep2Schema, courseStep3Schema)
)

export type CourseStep1Values = z.infer<typeof courseStep1Schema>
export type CourseStep2Values = z.infer<typeof courseStep2Schema>
export type CourseStep3Values = z.infer<typeof courseStep3Schema>
export type CourseValues = z.infer<typeof courseSchema>

// Export validation functions for use in forms if needed
export { isValidCourseCode, isValidCourseName, isValidSchedule, validSemesters, validColors }