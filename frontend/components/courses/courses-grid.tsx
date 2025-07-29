"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { BookOpen, Users, Calendar, Clock } from "lucide-react"
import { Assignment, Course } from "@/types/models"
import { useAssignments } from "@/hooks/use-assignments"
import { CourseItem } from "./course-item"

interface CoursesGridProps {
  courses: Course[]
  onCourseClick: (course: Course) => void
  disabled?: boolean
}

export function CoursesGrid({ courses, onCourseClick, disabled = false }: CoursesGridProps) {

  if (courses.length === 0) {
    return (
      <div className="text-center py-12">
        <BookOpen className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-300 mb-2">No courses found</h3>
        <p className="text-gray-400">Try adjusting your search or filters</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {courses.map((course) => {
        

        return (
          <CourseItem 
            course={course} 
            onEdit={() => {}} 
            onDelete={() => {}} 
            onToggleComplete={() => {}} 
            onCourseClick={onCourseClick} 
            disabled={disabled}
          />
        )
      })}
    </div>
  )
}
