"use client"

import { useState } from "react"
import { AddCourseDialog } from "@/components/courses/add-course-dialog"
import { CourseDetailsModal } from "@/components/courses/course-details-modal"
import { Loader2, Calendar, List } from "lucide-react"
import { useCourses, useCreateCourse, useDeleteCourse, useUpdateCourse } from "@/hooks/use-courses"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useRouter, useSearchParams } from "next/navigation"
import CoursesSchedule from "@/components/courses/courses-schedule"
import CoursesTable from "@/components/courses/courses-table"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { course } from "@/wailsjs/go/models"
import { CourseDeleteDialog } from "./course-delete-dialog"

export default function CoursesPage() {
  const { data: courses = [], isLoading, error } = useCourses()
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedDeleteCourseId, setSelectedDeleteCourseId] = useState<number | null>(null)

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
  const deleteMutation = useDeleteCourse()
  const createMutation = useCreateCourse()


  const handleEditCourse = async (courseData: course.LocalCourse, column: string, value: string) => {
    const message = "course " + courseData.Code + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      course: courseData,
      column,
      value
    })
  }

  const handleDeleteCourse = async (course: course.LocalCourse) => {
    const message = "course " + course.Code + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteMutation.mutate(course)
  }

  const handleAddCourse = async (course: course.LocalCourse) => {
    const message = "course " + course.Code + " added"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    createMutation.mutate(course)
  }

  // Handle course selection
  const handleCourseClick = (course: course.LocalCourse) => {
    setSelectedCourseId(course.ID)
  }

  const handleDeleteCourseClick = (course: course.LocalCourse) => {
    setSelectedDeleteCourseId(course.ID)
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
          <Loader2 className="mr-2 w-8 h-8 animate-spin" />
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
      <div className="absolute left-10 top-20 w-72 h-72 rounded-full blur-3xl bg-blue-500/10 animate-float"></div>
      <div className="absolute right-10 bottom-20 w-96 h-96 rounded-full blur-3xl bg-purple-500/10 animate-float-delayed"></div>

      <div className="relative z-10">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
              My Courses
            </h1>
            <p className="mt-2 text-gray-400">
              Manage your enrolled courses ({courses.length} total)
            </p>
          </div>
          <AddCourseDialog onAdd={handleAddCourse} />
        </div>

        <Tabs value={activeView} onValueChange={handleTabChange} className="w-full">
          <TabsList className="grid grid-cols-6 mb-8 w-full border-0 glass">
            <TabsTrigger value="calendar" className="flex items-center space-x-2">
              <Calendar className="w-4 h-4" />
              <span>Calendar</span>
            </TabsTrigger>
            <TabsTrigger value="list" className="flex items-center space-x-2">
              <List className="w-4 h-4" />
              <span>All ({courses.length || 0})</span>
            </TabsTrigger>
          </TabsList>


          <TabsContent value="calendar">
            <CoursesSchedule
              courses={courses || []}
              onCourseClick={handleCourseClick}
            />
          </TabsContent>

          <TabsContent value="list">
            <CoursesTable
              courses={courses || []}
              filter={{ semester: semester || "all", instructor: instructor || "all" }}
              onCourseClick={handleCourseClick}
              onEdit={handleEditCourse}
              onDelete={handleDeleteCourseClick}
            />
          </TabsContent>
        </Tabs>


        <CourseDetailsModal
          isOpen={!!selectedCourseId}
          courseId={selectedCourseId}
          courses={courses || []}
          onClose={() => setSelectedCourseId(null)}
          onEdit={handleEditCourse}
          onDelete={handleDeleteCourseClick}
        />  

        <CourseDeleteDialog
          isOpen={!!selectedDeleteCourseId}
          onClose={() => setSelectedDeleteCourseId(null)}
          courseId={selectedDeleteCourseId}
          courses={courses || []}
          onDelete={handleDeleteCourse}
        />
        
      </div>
    </div>
  )
}
