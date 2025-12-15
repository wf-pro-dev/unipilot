
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
      <div className="py-16 text-center">
        <div className="w-20 h-20 rounded-full bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center mx-auto mb-6">
          <BookOpen className="w-10 h-10 text-white/40" />
        </div>
        <h3 className="mb-2 text-h4 text-gray-300">No courses found</h3>
        <p className="text-body-small text-gray-400">Try adjusting your search or filters</p>
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
