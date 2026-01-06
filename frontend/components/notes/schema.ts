import { z } from "zod";

export const noteSchema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters").max(200, "Title must be at most 200 characters"),
  subject: z.string().min(3, "Subject must be at least 3 characters").max(200, "Subject must be at most 200 characters"),
  course_code: z.string().min(2, "Course code must be at least 2 characters").max(20, "Course code must be at most 20 characters"),
  course_id: z.number().min(1, "Please select a course"),
  remote_course_id: z.number().min(1, "Remote course ID is required"),
})

export type NoteValues = z.infer<typeof noteSchema>