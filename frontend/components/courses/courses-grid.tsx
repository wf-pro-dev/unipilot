"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { BookOpen, Users, Calendar, Clock } from "lucide-react"
import { Assignment, Course } from "@/types/models"
import { useAssignments } from "@/hooks/use-assignments"

interface CoursesGridProps {
  courses: Course[]
  onCourseClick: (course: Course) => void
}

export function CoursesGrid({ courses, onCourseClick }: CoursesGridProps) {
  const { data: assignments = [] } = useAssignments()

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
        const course_assignments = (assignments || []).filter((assignment) => assignment.Course?.Code === course.Code)
        const completed_assignments_count = course_assignments.filter((assignment) => assignment.StatusName === "Done").length
        const completionPercentage = course_assignments.length > 0 ? (completed_assignments_count / course_assignments.length) * 100 : 0

        return (
          <Card 
            key={course.ID} 
            className="glass border-0 hover:bg-white/5 transition-colors cursor-pointer"
            onClick={() => onCourseClick(course)}
          >
            <CardContent className="p-6">
              <div className="space-y-4">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <h3 className="font-semibold text-white">{course.Code}</h3>
                    <p className="text-sm text-gray-400">{course.Name}</p>
                  </div>
                  <div className={`w-3 h-3 rounded-full ${course.Color}`} />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-300">Progress</span>
                    <span className="text-gray-400">
                      {completed_assignments_count}/{course_assignments.length}
                    </span>
                  </div>
                  <Progress value={completionPercentage} className="h-2" />
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <Badge variant="secondary" className="text-xs">
                      {course.Semester}
                    </Badge>
                    <Badge variant="outline" className="text-xs">
                      {course.Credits} credits
                    </Badge>
                  </div>
                  <div className="flex items-center space-x-4 text-xs text-gray-400">
                    <div className="flex items-center space-x-1">
                      <Users className="h-3 w-3" />
                      <span>{course.Instructor}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-4 text-xs text-gray-400">
                  <div className="flex items-center space-x-1">
                    <Calendar className="h-3 w-3" />
                    <span>{new Date(course.StartDate).toLocaleDateString()}</span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <Clock className="h-3 w-3" />
                    <span>{course.Schedule}</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
