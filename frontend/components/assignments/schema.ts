import * as z from "zod"


// Alternative simpler approach - just check length and basic safety
const isValidAssignmentTitle = (title: string): boolean => {
  // Basic safety check - prevent injection attacks
  const dangerousPatterns = [
    /<script/i,
    /javascript:/i,
    /on\w+\s*=/i,
    /data:/i,
    /vbscript:/i,
  ]

  for (const pattern of dangerousPatterns) {
    if (pattern.test(title)) {
      return false
    }
  }

  return true
}

const isValidLink = (link: string): boolean => {
  if (!link) return true // Link is optional
  try {
    new URL(link)
    return true
  } catch {
    return false
  }
}

const isValidCourseCode = (code: string): boolean => {
  // Course code validation: alphanumeric with spaces, hyphens allowed
  const pattern = /^[a-zA-Z0-9\s\-]+$/
  return pattern.test(code)
}

// Valid types matching frontend options
const validTypes = ["HW", "Group Project", "Exam", "Quiz", "Lab"] as const

// Valid priorities matching frontend options
const validPriorities = ["low", "medium", "high"] as const

// Valid statuses matching frontend options
const validStatuses = ["Not started", "In progress", "Done"] as const

// Step 1: Basic assignment information
export const assignmentStep1Schema = z.object({
  title: z
    .string()
    .min(3, "Title must be at least 3 characters")
    .max(200, "Title must be at most 200 characters")
    .refine(
      (val) => isValidAssignmentTitle(val),
      {
        message: "Title contains unsafe characters or patterns",
      }
    ),
  course_code: z
    .string()
    .min(2, "Course code must be at least 2 characters")
    .max(20, "Course code must be at most 20 characters")
    .refine(
      (val) => isValidCourseCode(val),
      {
        message: "Course code can only contain letters, numbers, spaces, and hyphens",
      }
    ),
  course_id: z
    .number()
    .min(1, "Course ID is required"),
  remote_course_id: z
    .number().optional(),
  type: z
    .string()
    .min(1, "Please select an assignment type")
    .refine(
      (val) => validTypes.includes(val as typeof validTypes[number]),
      {
        message: "Please select a valid assignment type",
      }
    ),
  deadline: z.date({
    required_error: "Deadline is required",
  }),
})

// Step 2: Additional details
export const assignmentStep2Schema = z.object({
  priority: z
    .string()
    .min(1, "Please select a priority")
    .refine(
      (val) => validPriorities.includes(val as typeof validPriorities[number]),
      {
        message: "Please select a valid priority",
      }
    ),
  status: z
    .string()
    .min(1, "Please select a status")
    .refine(
      (val) => validStatuses.includes(val as typeof validStatuses[number]),
      {
        message: "Please select a valid status",
      }
    ),
  todo: z
    .string()
    .max(1000, "Notes must be at most 1000 characters")
    .optional(),
  link: z
    .string()
    .optional()
    .refine(
      (val) => !val || isValidLink(val),
      {
        message: "Please enter a valid URL or leave blank",
      }
    ),
})

// Combined schema
export const assignmentSchema = z.intersection(assignmentStep1Schema, assignmentStep2Schema)

export type AssignmentStep1Values = z.infer<typeof assignmentStep1Schema>
export type AssignmentStep2Values = z.infer<typeof assignmentStep2Schema>
export type AssignmentValues = z.infer<typeof assignmentSchema>

// Export validation functions
export { isValidAssignmentTitle, isValidLink, isValidCourseCode, validTypes, validPriorities, validStatuses }