import { Assignment, Course, User } from './models'

declare global {
  interface Window {
    go: {
      main: {
        App: {
          Login: (username: string, password: string) => Promise<void>
          Logout: () => Promise<void>
          Greet: (name: string) => Promise<string>
          IsAuthenticated: () => Promise<storage.LocalCredentials>
          GetAssignment: (id: number) => Promise<Assignment>
          GetCourse: (id: number) => Promise<Course>
          GetUser: (id: number) => Promise<User>
          GetAssignments: () => Promise<Assignment[]>
          GetCourses: () => Promise<Course[]>
          CreateAssignment: (assignment: Assignment) => Promise<void>
          UpdateAssignment: (assignment: Assignment, column: string, value: string) => Promise<void>
          DeleteAssignment: (assignment: Assignment) => Promise<void>
          CreateCourse: (course: Course) => Promise<void>
          UpdateCourse: (course: Course) => Promise<void>
          DeleteCourse: (course: Course) => Promise<void>
        }
      }
    }
  }
}

export {} 