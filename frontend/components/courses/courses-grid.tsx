
import { BookOpen } from "lucide-react"
import { CourseItem } from "./course-item"
import { course } from "@/wailsjs/go/models"

interface CoursesGridProps {
  courses: course.LocalCourse[]
  onCourseClick: (course: course.LocalCourse) => void
  onEdit: (course: course.LocalCourse, column: string, value: string) => void
  onDelete: (course: course.LocalCourse) => void
  disabled?: boolean
}

export function CoursesGrid({ courses, onCourseClick, onEdit, onDelete, disabled = false }: CoursesGridProps) {

  if (courses.length === 0) {
    return (
      <div className="py-12 text-center">
        <BookOpen className="mx-auto mb-4 w-12 h-12 text-gray-400" />
        <h3 className="mb-2 text-lg font-medium text-gray-300">No courses found</h3>
        <p className="text-gray-400">Try adjusting your search or filters</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {courses.map((course) => {
        

        return (
          <CourseItem 
            course={course} 
            onEdit={onEdit} 
            onDelete={onDelete}
            onCourseClick={onCourseClick} 
            disabled={disabled}
          />
        )
      })}
    </div>
  )
}
