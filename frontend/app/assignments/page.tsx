"use client"

import { useState, useEffect, useCallback } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AssignmentsCalendar } from "@/components/assignments/assignments-calendar"
import { AssignmentsTable } from "@/components/assignments/assignments-table"
import { AddAssignmentDialog } from "@/components/assignments/add-assignment-dialog"
import { Calendar, List, CheckCircle2, CalendarDays, Loader2 } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import {
  useAssignments,
  useUpdateAssignment,
  useTodayAssignments,
  useWeekAssignments,
  useOverdueAssignments,
  useExamAssignments,
  useDeleteAssignment,
  useCreateAssignment
} from "@/hooks/use-assignments"
import { DayAssignmentsModal } from "@/components/assignments/day-assignments-modal"

/**
 * Main assignments management page component.
 * 
 * Provides a comprehensive interface for viewing and managing assignments with
 * multiple view modes: Today, Week, Overdue, Exam, Calendar, and List views.
 * Supports full CRUD operations with optimistic updates, URL-based state
 * management, and real-time filtering.
 * 
 * Features:
 * - Tab-based navigation with URL synchronization
 * - Assignment CRUD operations (Create, Read, Update, Delete)
 * - Optimistic UI updates for immediate feedback
 * - Multiple view modes with filtered assignment lists
 * - Calendar view with drag-and-drop date changes
 * - Modal dialogs for assignment details and editing
 * - Deep linking support via URL query parameters
 * 
 * URL Query Parameters:
 * - `view`: Active tab view ("today" | "week" | "overdue" | "exam" | "calendar" | "list")
 * - `course`: Course filter value
 * - `status`: Status filter value
 * - `priority`: Priority filter value
 * - `assignment`: Assignment ID for deep linking to details modal
 * 
 * @returns {JSX.Element} The assignments page with tab navigation and assignment management UI
 */
export default function AssignmentsPage() {
  const searchParams = useSearchParams()
  const router = useRouter()

  // Use the new optimized hooks
  const { data: assignments, isLoading, error } = useAssignments()
  const { data: todayAssignments } = useTodayAssignments()
  const { data: weekAssignments } = useWeekAssignments()
  const { data: overdueAssignments } = useOverdueAssignments()
  const { data: examAssignments } = useExamAssignments()

  // Mutation for updates with optimistic updates
  const updateMutation = useUpdateAssignment()
  const deleteMutation = useDeleteAssignment()
  const createMutation = useCreateAssignment()

  const [selectedAssignmentID, setSelectedAssignmentID] = useState<number | null>(null)
  const [selectedAssignmentEdit, setSelectedAssignmentEdit] = useState<models.LocalAssignment | null>(null)
  const [selectedDate, setSelectedDate] = useState<Date | null>(null)
  const [addAssignmentOpen, setAddAssignmentOpen] = useState(false)

  // Get the current view from URL parameters, default to "today"
  const currentView = searchParams.get("view") || "week"

  // Valid view values
  const validViews = ["week", "exam", "calendar", "list"]

  // Ensure the current view is valid, otherwise default to "today"
  const activeView = validViews.includes(currentView) ? currentView : "week"

  // Get the filter from URL parameters, default to null
  const courseFilter = searchParams.get("course") || null
  const statusFilter = searchParams.get("status") || null
  const priorityFilter = searchParams.get("priority") || null
  const currentAssignment = searchParams.get("assignment") || null


  useEffect(() => {
    if (currentAssignment) {
      const assignment = assignments?.find((assignment) => assignment.ID === parseInt(currentAssignment))
      if (assignment) {
        setSelectedAssignmentID(assignment.ID)
      }
    }
  }, [currentAssignment, assignments])


  /**
   * Handles assignment creation with optimistic UI updates.
   * 
   * Creates a new assignment and provides immediate UI feedback. Logs the creation
   * and shows success/error toast notifications.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment to create
   * @returns {Promise<void>}
   */
  const handleAddAssignment = useCallback((assignment: models.LocalAssignment) => {
    createMutation.mutate(assignment)
  }, [createMutation])


  /**
   * Handles tab change and synchronizes the active view with URL query parameters.
   * 
   * Updates the URL to reflect the selected tab view while preserving other
   * query parameters (filters, assignment ID, etc.).
   * 
   * @param {string} value - The tab value to switch to ("today" | "week" | "overdue" | "exam" | "calendar" | "list")
   */
  const handleTabChange = (value: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("view", value)
    router.push(`/assignments?${params.toString()}`)
  }

  // Show loading state
  if (isLoading) {
    return (
      <div className="">
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin mr-2" />
          <span>Loading assignments...</span>
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
            Error loading assignments: {error.message}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1">
      <div className="flex flex-col flex-1 relative z-10">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-h1 text-white">
              Assignments
            </h1>
            <p className="text-body-small text-gray-400 mt-3">Track and manage your coursework deadlines</p>
          </div>
          <AddAssignmentDialog isOpen={addAssignmentOpen} setOpen={setAddAssignmentOpen} onAdd={handleAddAssignment} />
        </div>

        <Tabs value={activeView} onValueChange={handleTabChange} className="flex flex-col flex-1 w-full">
          
          <TabsList className="flex flex-row bg-white/5 p-1 rounded-xl w-full mb-6 border border-white/5">
           
            <TabsTrigger 
              value="week" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <CalendarDays className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">This Week ({weekAssignments?.length || 0})</span>
            </TabsTrigger>
            
            <TabsTrigger 
              value="exam" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <CheckCircle2 className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Exam ({examAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger 
              value="calendar" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Calendar className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Calendar</span>
            </TabsTrigger>
            <TabsTrigger 
              value="list" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <List className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">All ({assignments?.length || 0})</span>
            </TabsTrigger>
          </TabsList>


            <TabsContent value="week" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentsTable
              assignments={weekAssignments || []}
              filter={{ course: courseFilter || "all", status: statusFilter || "all", priority: priorityFilter || "all" }}
              isLoading={isLoading}
            />
          </TabsContent> 
          <TabsContent value="exam" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentsTable
              assignments={examAssignments || []}
              filter={{ course: courseFilter || "all", status: statusFilter || "all", priority: priorityFilter || "all" }}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="calendar" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentsCalendar
              assignments={assignments || []}
              onDateClick={setSelectedDate}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="list" className="flex flex-col  data-[state=active]:flex-1 m-0">
            <AssignmentsTable
              assignments={assignments || []}
              filter={{ course: courseFilter || "all", status: statusFilter || "all", priority: priorityFilter || "all" }}
              isLoading={isLoading}
            />
          </TabsContent>
        </Tabs>

       
        <DayAssignmentsModal
          isOpen={!!selectedDate}
          onClose={() => setSelectedDate(null)}
          date={selectedDate}
          assignments={assignments || []}
          onAddAssignment={() => { }}
          isLoading={updateMutation.isPending}
        />

      </div>
    </div>
  )
}
