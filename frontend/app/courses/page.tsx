"use client"

import { useEffect, useMemo, useState } from "react"
import { CourseAddDialog } from "@/components/courses/course-add-dialog"
import { CourseDetailsModal } from "@/components/courses/course-details-modal"
import { Loader2, Calendar, List } from "lucide-react"
import { useCourses, useCreateCourse, useDeleteCourse, useUpdateCourse } from "@/hooks/use-courses"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useRouter, useSearchParams } from "next/navigation"
import CoursesSchedule from "@/components/courses/courses-schedule"
import CoursesTable from "@/components/courses/courses-table"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { models } from "@/wailsjs/go/models"
import { CourseDeleteDialog } from "./course-delete-dialog"
import { LinkRequestModal } from "@/components/community/link-request-modal"
import { Link as LinkIcon } from "lucide-react"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"
import { useAuthContext } from "@/components/provider/auth-provider"

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
  const { user } = useAuthContext()
  const { data: courses = [], isLoading, error } = useCourses()
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedDeleteCourseId, setSelectedDeleteCourseId] = useState<number | null>(null)
  const [isLinkRequestModalOpen, setIsLinkRequestModalOpen] = useState(false)
  const [selectedSemester, setSelectedSemester] = useState<string>(user?.Semester || "")



  // Get unique semesters
  const semesters = useMemo(() => {
    if (!courses) return []
    const uniqueSemesters = [...new Set([...courses.map(course => course.Semester), user?.Semester || ""])]

    // Custom sorting function for semester format: "<season> <year>"
    const sortSemesters = (a: string, b: string) => {
      const parseSemester = (semester: string) => {
        const parts = semester.split(' ')
        if (parts.length !== 2) return { season: '', year: 0, seasonPriority: 0 }

        const season = parts[0].toUpperCase()
        const year = parseInt(parts[1])

        // Season priority: SPRING = 1, SUMMER = 2, FALL = 3
        const seasonPriority: Record<string, number> = {
          'FALL': 1,
          'SUMMER': 2,
          'SPRING': 3
        }

        return {
          season,
          year,
          seasonPriority: seasonPriority[season] || 0
        }
      }

      const semesterA = parseSemester(a)
      const semesterB = parseSemester(b)

      // First sort by year (descending)
      if (semesterA.year !== semesterB.year) {
        return semesterB.year - semesterA.year
      }

      // Then sort by season (SPRING -> SUMMER -> FALL)
      return semesterA.seasonPriority - semesterB.seasonPriority
    }

    return uniqueSemesters.sort(sortSemesters)
  }, [courses])


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
  const handleEditCourse = async (courseData: models.LocalCourse, column: string, value: string) => {
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
  const handleDeleteCourse = async (course: models.LocalCourse) => {
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
  const handleAddCourse = async (course: models.LocalCourse) => {
    const message = "course " + course.Code + " added"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    createMutation.mutate(course)
  }

  /**
   * Handles course card click to open details modal.
   * 
   * @param {course.LocalCourse} course - The course that was clicked
   */
  const handleCourseClick = (course: models.LocalCourse) => {
    setSelectedCourseId(course.ID)
  }

  /**
   * Handles delete button click to open delete confirmation dialog.
   * 
   * @param {course.LocalCourse} course - The course to delete
   */
  const handleDeleteCourseClick = (course: models.LocalCourse) => {
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
      <div className="">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500">
            Error loading courses: {error.message}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1">

      <div className="flex flex-col flex-1">
        {/* Page header with course count and add course button */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-h1 text-white">
              Courses
            </h1>
            <p className="mt-3 text-body-small text-gray-400">
              Manage your enrolled courses ({courses.length} total)
            </p>
          </div>
          <CourseAddDialog onAdd={handleAddCourse} />
        </div>

        {/* Tab navigation with URL synchronization */}
        <Tabs value={activeView} onValueChange={handleTabChange} className="flex flex-col flex-1 w-full">
          <div className="flex justify-between items-center mb-6">
            <TabsList className="h-full flex w-fit bg-white/5 p-1 rounded-xl  border border-white/5">
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
            <div>
              <Select value={selectedSemester} onValueChange={setSelectedSemester}>
                <SelectTrigger className="bg-white/5 border-white/10 text-white hover:bg-white/10 transition-all duration-200 backdrop-blur-sm">
                  <SelectValue placeholder="Filter by semester" />
                </SelectTrigger>
                <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl text-gray-200">
                  {semesters.map(semester => (
                    <SelectItem
                      key={semester}
                      value={semester}
                      className="focus:bg-white/10 focus:text-white cursor-pointer"
                    >
                      {semester}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Schedule view: Calendar-based course display */}
          <TabsContent value="schedule" className="flex flex-col data-[state=active]:flex-1 m-0">
            <CoursesSchedule
              selectedSemester={selectedSemester}
              onEdit={handleEditCourse}
              onDelete={handleDeleteCourseClick}
              courses={courses || []}
              onCourseClick={handleCourseClick}
            />
          </TabsContent>

          {/* List view: Table-based course display with filtering */}
          <TabsContent value="list" className="flex flex-col data-[state=active]:flex-1 m-0">
            <CoursesTable
              courses={courses || []}
              filter={{ semester: semester || null, instructor: instructor || null }}
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
          onViewLinks={() => handleTabChange("linked")}
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
