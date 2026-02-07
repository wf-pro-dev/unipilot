import { z } from "zod";

export const noteSchema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters").max(200, "Title must be at most 200 characters"),
  subject: z.string().min(3, "Subject must be at least 3 characters").max(200, "Subject must be at most 200 characters"),
  course_id: z.string().min(1, "Please select a course"),
})

export type NoteValues = z.infer<typeof noteSchema>