"use client"

import { useEffect, useState } from "react"
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
import { LinkRequestModal } from "@/components/community/link-request-modal"

/**
 * Courses page component for course management and viewing.
 * 
 * Provides a comprehensive interface for viewing and managing courses with
 * multiple view modes: Schedule (calendar view) and List (table view).
 * Supports full CRUD operations with optimistic updates, URL-based state
 * management, and course sharing functionality.
 * 
 * Features:
 * - Tab-based navigation with URL synchronization
 * - Course CRUD operations (Create, Read, Update, Delete)
 * - Schedule and list view modes
 * - Course details modal integration
 * - Course sharing via link requests
 * - Deep linking support via URL query parameters
 * 
 * URL Query Parameters:
 * - `view`: Active tab view ("schedule" | "list")
 * - `course`: Course code for deep linking to course details
 * - `semester`: Semester filter value
 * - `instructor`: Instructor filter value
 * 
 * @returns {JSX.Element} The courses page with tab navigation and course management UI
 */
export default function CoursesPage() {
  // Fetch courses data with default empty array to prevent undefined errors
  const { data: courses = [], isLoading, error } = useCourses()
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedDeleteCourseId, setSelectedDeleteCourseId] = useState<number | null>(null)
  const [isLinkRequestModalOpen, setIsLinkRequestModalOpen] = useState(false)

  const searchParams = useSearchParams()
  const router = useRouter()

  // Extract view from URL for deep linking, default to schedule view
  const currentView = searchParams.get("view") || "schedule"
  // Extract course code from URL for deep linking to course details
  const currentCourse = searchParams.get("course") || null

  // Sync selected course with URL parameter when course code is present
  // Matches course code from URL to course ID for modal display
  useEffect(() => {
    if (currentCourse) {
      const course = courses.find((course) => course.Code === currentCourse)
      if (course) {
        setSelectedCourseId(course.ID)
      }
    }
  }, [currentCourse, courses])

  // Valid view values for tab navigation
  const validViews = ["schedule", "list"]

  // Validate and sanitize view parameter to prevent invalid states
  const activeView = validViews.includes(currentView) ? currentView : "schedule"

  const semester = searchParams.get("semester") || null
  const instructor = searchParams.get("instructor") || null

  const updateMutation = useUpdateCourse()
  const deleteMutation = useDeleteCourse()
  const createMutation = useCreateCourse()

  /**
   * Handles course field updates with optimistic UI updates.
   * 
   * Updates a specific field of a course and provides immediate UI feedback
   * through optimistic updates. Logs the change for audit purposes.
   * 
   * @param {course.LocalCourse} courseData - The course to update
   * @param {string} column - The field name to update
   * @param {string} value - The new value for the field
   * @returns {Promise<void>}
   */
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

  /**
   * Handles course deletion with optimistic UI updates.
   * 
   * Deletes a course and provides immediate UI feedback. Logs the deletion
   * for audit purposes. Note: This triggers the delete confirmation dialog.
   * 
   * @param {course.LocalCourse} course - The course to delete
   * @returns {Promise<void>}
   */
  const handleDeleteCourse = async (course: course.LocalCourse) => {
    const message = "course " + course.Code + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteMutation.mutate(course)
  }

  /**
   * Handles course creation with optimistic UI updates.
   * 
   * Creates a new course and provides immediate UI feedback. Logs the creation
   * for audit purposes.
   * 
   * @param {course.LocalCourse} course - The course to create
   * @returns {Promise<void>}
   */
  const handleAddCourse = async (course: course.LocalCourse) => {
    const message = "course " + course.Code + " added"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    createMutation.mutate(course)
  }

  /**
   * Handles course card click to open details modal.
   * 
   * @param {course.LocalCourse} course - The course that was clicked
   */
  const handleCourseClick = (course: course.LocalCourse) => {
    setSelectedCourseId(course.ID)
  }

  /**
   * Handles delete button click to open delete confirmation dialog.
   * 
   * @param {course.LocalCourse} course - The course to delete
   */
  const handleDeleteCourseClick = (course: course.LocalCourse) => {
    setSelectedDeleteCourseId(course.ID)
  }

  /**
   * Handles tab change and synchronizes the active view with URL query parameters.
   * 
   * Updates the URL to reflect the selected tab view while preserving other
   * query parameters (filters, course code, etc.).
   * 
   * @param {string} value - The tab value to switch to ("schedule" | "list")
   */
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
      {/* Decorative background elements for visual depth */}
      <div className="absolute left-10 top-20 w-72 h-72 rounded-full blur-3xl bg-blue-500/10 animate-float"></div>
      <div className="absolute right-10 bottom-20 w-96 h-96 rounded-full blur-3xl bg-purple-500/10 animate-float-delayed"></div>

      <div className="relative z-10">
        {/* Page header with course count and add course button */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-h1 text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
              Courses
            </h1>
            <p className="mt-3 text-body-small text-gray-400">
              Manage your enrolled courses ({courses.length} total)
            </p>
          </div>
          <AddCourseDialog onAdd={handleAddCourse} />
        </div>

        {/* Tab navigation with URL synchronization */}
        <Tabs value={activeView} onValueChange={handleTabChange} className="w-full">
          <TabsList className="h-full flex w-fit bg-white/5 p-1 rounded-xl mb-6 border border-white/5">
            <TabsTrigger 
              value="schedule" 
              className="flex w-60 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Calendar className="w-4 h-4" />
              <span className="hidden sm:inline text-sm font-medium">Schedule</span>
            </TabsTrigger>
            <TabsTrigger 
              value="list" 
              className="flex w-60 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <List className="w-4 h-4" />
              <span className="hidden sm:inline text-sm font-medium">All ({courses.length || 0})</span>
            </TabsTrigger>
          </TabsList>

          {/* Schedule view: Calendar-based course display */}
          <TabsContent value="schedule">
            <CoursesSchedule
              onEdit={handleEditCourse}
              onDelete={handleDeleteCourseClick}
              courses={courses || []}
              onCourseClick={handleCourseClick}
            />
          </TabsContent>

          {/* List view: Table-based course display with filtering */}
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
          onLinkRequest={() => setIsLinkRequestModalOpen(true)}
        />

        <CourseDeleteDialog
          isOpen={!!selectedDeleteCourseId}
          onClose={() => setSelectedDeleteCourseId(null)}
          courseId={selectedDeleteCourseId}
          courses={courses || []}
          onDelete={handleDeleteCourse}
        />

        <LinkRequestModal
          courseID={selectedCourseId!}
          isOpen={isLinkRequestModalOpen}
          onClose={() => setIsLinkRequestModalOpen(false)}
        />
        
      </div>
    </div>
  )
}
