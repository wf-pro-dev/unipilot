"use client"

import { useState } from "react"
import { AddCourseDialog } from "@/components/courses/add-course-dialog"
import { CourseDetailsModal } from "@/components/courses/course-details-modal"
import { Loader2, Calendar, List } from "lucide-react"
import { useCourses, useUpdateCourse } from "@/hooks/use-courses"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useRouter, useSearchParams } from "next/navigation"
import CoursesSchedule from "@/components/courses/courses-schedule"
import CoursesTable from "@/components/courses/courses-table"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { course } from "@/wailsjs/go/models"

export default function CoursesPage() {
  const { data: courses = [], isLoading, error } = useCourses()
  const [selectedCourse, setSelectedCourse] = useState<course.LocalCourse | null>(null)

  const searchParams = useSearchParams()
  const router = useRouter()

  // Get the current view from URL parameters, default to "today"
  const currentView = searchParams.get("view") || "calendar"

  // Valid view values
  const validViews = ["calendar", "list"]

  // Ensure the current view is valid, otherwise default to "today"
  const activeView = validViews.includes(currentView) ? currentView : "calendar"

  const semester = searchParams.get("semester") || null
  const instructor = searchParams.get("instructor") || null

  const updateMutation = useUpdateCourse()

  const handleEditCourse = async (courseData: course.LocalCourse, column: string, value: string) => {
    const message = "course " + courseData.Code + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      courseData,
      column,
      value
    })
  }


  // Handle tab change and update URL
  const handleTabChange = (value: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("view", value)
    router.push(`/courses?${params.toString()}`)
  }
  // Show loading state
  if (isLoading) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin mr-2" />
          <span>Loading courses...</span>
        </div>
      </div>
    )
  }

  // Show error state
  if (error) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500">
            Error loading courses: {error.message}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="relative z-10">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              My Courses
            </h1>
            <p className="text-gray-400 mt-2">
              Manage your enrolled courses ({courses.length} total)
            </p>
          </div>
          <AddCourseDialog />
        </div>

        <Tabs value={activeView} onValueChange={handleTabChange} className="w-full">
          <TabsList className="grid w-full grid-cols-6 glass border-0 mb-8">
            <TabsTrigger value="calendar" className="flex items-center space-x-2">
              <Calendar className="h-4 w-4" />
              <span>Calendar</span>
            </TabsTrigger>
            <TabsTrigger value="list" className="flex items-center space-x-2">
              <List className="h-4 w-4" />
              <span>All ({courses.length || 0})</span>
            </TabsTrigger>
          </TabsList>


          <TabsContent value="calendar">
            <CoursesSchedule
              courses={courses || []}
              onCourseClick={setSelectedCourse}
            />
          </TabsContent>

          <TabsContent value="list">
            <CoursesTable
              courses={courses || []}
              filter={{ semester: semester || "all", instructor: instructor || "all" }}
              onCourseClick={setSelectedCourse}
            />
          </TabsContent>
        </Tabs>


        <CourseDetailsModal
          isOpen={!!selectedCourse}
          course={selectedCourse}
          onClose={() => setSelectedCourse(null)}
          onEdit={handleEditCourse}
        />  
        
      </div>
    </div>
  )
}
