// Assignment type matching the Go struct
interface Assignment {
  ID: number
  RemoteID: number | null
  NotionID: string
  Title: string
  Todo: string
  Deadline: Date 
  Link: string
  TypeName: string
  StatusName: "Not started" | "In progress" | "Done"
  Priority: "low" | "medium" | "high"
  Completed: boolean
  CreatedAt: Date 
  UpdatedAt: Date
  DeletedAt?: Date 
  User?: User
  Course?: Course
  Type?: AssignmentType
  Status?: AssignmentStatus
}


// Course type matching the Go struct
interface Course {
  ID: number
  UserID: number
  NotionID: string
  Code: string
  Name: string
  Duration: string
  RoomNumber: string
  Color: string
  StartDate: Date
  EndDate: Date
  Schedule: string
  Semester: string
  Instructor: string
  InstructorEmail: string
  Credits: number
  CreatedAt: Date 
  UpdatedAt: Date
  DeletedAt?: Date 
  User?: User
}

// User type matching the Go struct
interface User {
  ID: number
  Username: string
  Email: string
  PasswordHash: string
  NotionAPIKey: string
  AssignmentsDbId: string
  NotionID: string
  CoursesDbId: string
  LastSync?: Date 
  CreatedAt: Date 
  UpdatedAt: Date 
  DeletedAt?: Date 
}

// AssignmentType type matching the Go struct
interface AssignmentType {
  ID: number
  Name: string
  Color: string
  NotionID: string
}

// AssignmentStatus type matching the Go struct
interface AssignmentStatus {
  ID: number
  Name: string
  Color: string
  NotionID: string
}

// Legacy interface for backward compatibility
interface LegacyAssignment {
  id: number
  title: string
  course: string
  courseColor: string
  type: string
  priority: "low" | "medium" | "high"
  status: "pending" | "in-progress" | "completed" | "overdue"
  dueDate: Date
  description?: string
  completed: boolean
}

export type { 
  Assignment, 
  Course, 
  User, 
  AssignmentType, 
  AssignmentStatus,
  LegacyAssignment 
}