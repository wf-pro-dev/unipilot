
import { BookOpen } from "lucide-react"
import { CourseItem } from "./course-item"
import { models } from "@/wailsjs/go/models"
import { EmptyState } from "../ui/empty-state"

interface CoursesGridProps {
  courses: models.LocalCourse[]
  onCourseClick: (course: models.LocalCourse) => void
  onEdit: (course: models.LocalCourse, column: string, value: string) => void
  onDelete: (course: models.LocalCourse) => void
  disabled?: boolean
}

export function CoursesGrid({ courses, onCourseClick, onEdit, onDelete, disabled = false }: CoursesGridProps) {

  if (courses.length === 0) {
    return (
      <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
        <EmptyState
          icon={BookOpen}
          title="No courses found"
          description="Create a new course to get started"
          className="flex-1 items-center"
        />
      </div>
    )
  }

  return (
    <div className="flex-1 grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
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
