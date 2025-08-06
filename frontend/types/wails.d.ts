import { User } from './models'
import { assignment, course, note } from '@/wailsjs/go/models'


declare global {
  interface Window {
    go: {
      main: {
        App: {
          Login: (username: string, password: string) => Promise<void>
          Logout: () => Promise<void>
          Greet: (name: string) => Promise<string>
          IsAuthenticated: () => Promise<storage.LocalCredentials>
          GetAssignment: (id: number) => Promise<assignment.LocalAssignment>
          GetCourse: (id: number) => Promise<course.LocalCourse>
          GetUser: (id: number) => Promise<User>
          GetAssignments: () => Promise<assignment.LocalAssignment[]>
          GetCourseAssignments: (course: course.LocalCourse) => Promise<assignment.LocalAssignment[]>
          GetCourses: () => Promise<course.LocalCourse[]>
          GetNotes: () => Promise<note.LocalNote[]>
          CreateAssignment: (assignment: assignment.LocalAssignment) => Promise<void>
          UpdateAssignment: (assignment: assignment.LocalAssignment, column: string, value: string) => Promise<void>
          DeleteAssignment: (assignment: assignment.LocalAssignment) => Promise<void>
          CreateCourse: (course: course.LocalCourse) => Promise<void>
          UpdateCourse: (course: course.LocalCourse, column: string, value: string) => Promise<void>
          DeleteCourse: (course: course.LocalCourse) => Promise<void>
          CreateNote: (note: note.LocalNote) => Promise<void>
          UpdateNote: (note: note.LocalNote, column: string, value: string) => Promise<void>
          DeleteNote: (note: note.LocalNote) => Promise<void>
        }
      }
    }
  }
}

export {} 