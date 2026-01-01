import * as z from "zod"

// Validation helper functions matching backend logic
const isValidUsername = (username: string): boolean => {
  const pattern = /^[a-zA-Z0-9_\s-]+$/
  return pattern.test(username)
}

const isValidPassword = (password: string): { valid: boolean; error?: string } => {
  if (password.length > 100) {
    return { valid: false, error: "Password must be less than 100 characters" }
  }

  const uppercase = /[A-Z]/
  const lowercase = /[a-z]/
  const number = /[0-9]/
  const special = /[^a-zA-Z0-9]/

  if (!uppercase.test(password)) {
    return { valid: false, error: "Password must contain at least one uppercase letter" }
  }
  if (!lowercase.test(password)) {
    return { valid: false, error: "Password must contain at least one lowercase letter" }
  }
  if (!number.test(password)) {
    return { valid: false, error: "Password must contain at least one number" }
  }
  if (!special.test(password)) {
    return { valid: false, error: "Password must contain at least one special character" }
  }

  return { valid: true }
}

// Valid language codes matching backend validation
const validLanguages = ["en", "fr", "es", "de", "it", "pt", "nl", "ru", "tr", "ja", "zh", "ko", "ar", "he"] as const

export const loginSchema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
})

export type LoginValues = z.infer<typeof loginSchema>

export const registerStep1Schema = z.object({
  username: z
    .string()
    .min(3, "Username must be at least 3 characters")
    .max(30, "Username must be at most 30 characters")
    .refine(
      (val) => isValidUsername(val),
      {
        message: "Username can only contain letters, numbers, underscores, hyphens, and spaces",
      }
    ),
  email: z
    .string()
    .email("Please enter a valid email address"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(100, "Password must be less than 100 characters")
    .refine(
      (val) => {
        const result = isValidPassword(val)
        return result.valid
      },
      (val) => {
        const result = isValidPassword(val)
        return {
          message: result.error || "Password does not meet requirements",
        }
      }
    ),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords do not match",
  path: ["confirmPassword"],
})

export const registerStep2Schema = z.object({
  university: z
    .string()
    .min(3, "University name must be at least 3 characters")
    .max(100, "University name must be at most 100 characters"),
  semester: z
    .string()
    .max(20, "Semester must be at most 20 characters"),
  year: z
    .string()
    .length(4, "Year must be exactly 4 characters"),
  language: z
    .string()
    .min(1, "Please select a language")
    .refine(
      (val) => validLanguages.includes(val as typeof validLanguages[number]),
      {
        message: "Please select a valid language",
      }
    ),
})

export const registerSchema = z.intersection(registerStep1Schema, registerStep2Schema)

export type RegisterStep1Values = z.infer<typeof registerStep1Schema>
export type RegisterStep2Values = z.infer<typeof registerStep2Schema>
export type RegisterValues = z.infer<typeof registerSchema>

// Export validation functions for use in forms if needed
export { isValidUsername, isValidPassword, validLanguages }

