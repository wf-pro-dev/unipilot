"use client"

import { useState, useEffect } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AssignmentsCalendar } from "@/components/assignments/assignments-calendar"
import { AssignmentsTable } from "@/components/assignments/assignments-table"
import { AddAssignmentDialog } from "@/components/assignments/add-assignment-dialog"
import { AssignmentDetailsModal } from "@/components/assignments/assignment-details-modal"
import { Calendar, List, Clock, CheckCircle2, AlertTriangle, CalendarDays, Loader2 } from "lucide-react"
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
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format, isSameDay } from "date-fns"
import { DayAssignmentsModal } from "@/components/assignments/day-assignments-modal"
import { toast } from "sonner"
import { AssignmentEditDialog } from "@/components/assignments/assignment-edit-dialog"
import { AssignmentView } from "@/components/assignments/assignment-view"

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
  const currentView = searchParams.get("view") || "today"

  // Valid view values
  const validViews = ["today", "week", "overdue", "exam", "calendar", "list"]

  // Ensure the current view is valid, otherwise default to "today"
  const activeView = validViews.includes(currentView) ? currentView : "today"

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
   * Handles assignment card click to open details modal.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment that was clicked
   */
  const handleAssignmentClick = (assignment: models.LocalAssignment) => {
    setSelectedAssignmentID(assignment.ID)
  }

  /**
   * Handles assignment field updates with optimistic UI updates.
   * 
   * Updates a specific field of an assignment and provides immediate UI feedback
   * through optimistic updates. Logs the change and shows success/error toast notifications.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment to update
   * @param {string} column - The field name to update (e.g., "status_name", "deadline")
   * @param {string} value - The new value for the field
   * @returns {Promise<void>}
   */
  const handleEditAssignment = async (assignment: models.LocalAssignment, column: string, value: string) => {
    const message = "[Frontend] assignment " + assignment.ID + " remote_id " + assignment.RemoteID + " " + column + " changed to " + value
    LogInfo(format(new Date(), "yyyy/MM/dd HH:mm:ssxxx") + " " + message)

    // Use the optimistic update mutation
    updateMutation.mutate({
      assignment,
      column,
      value
    }, {
      onSuccess: () => {
        toast.success("Assignment updated successfully")
      },
      onError: () => {
        toast.error("Assignment update failed")
      }
    })
  }

  /**
   * Toggles assignment completion status between "Done" and "Not started".
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment to toggle
   * @returns {Promise<void>}
   */
  const handleToggleComplete = async (assignment: models.LocalAssignment) => {
    const newStatus = assignment.Status === "Done" ? "Not started" : "Done"
    handleEditAssignment(assignment, "status", newStatus)
  }

  /**
   * Handles assignment deletion with optimistic UI updates.
   * 
   * Deletes an assignment and provides immediate UI feedback. Logs the deletion
   * and shows success/error toast notifications.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment to delete
   * @returns {Promise<void>}
   */
  const handleDeleteAssignment = async (assignment: models.LocalAssignment) => {
    const message = "[Frontend] assignment " + assignment.Title + " deleted"
    LogInfo(format(new Date(), "yyyy/MM/dd HH:mm:ssxxx") + " " + message)
    deleteMutation.mutate(assignment, {
      onSuccess: () => {
        toast.success("Assignment deleted successfully")
      },
      onError: () => {
        toast.error("Assignment deletion failed")
      }
    })
  }


  /**
   * Handles assignment creation with optimistic UI updates.
   * 
   * Creates a new assignment and provides immediate UI feedback. Logs the creation
   * and shows success/error toast notifications.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment to create
   * @returns {Promise<void>}
   */
  const handleAddAssignment = async (assignment: models.LocalAssignment) => {
    const message = "[Frontend] assignment " + assignment.Title + " added"
    LogInfo(format(new Date(), "yyyy/MM/dd HH:mm:ssxxx") + " " + message)
    createMutation.mutate(assignment, {
      onSuccess: () => {
        toast.success("Assignment added successfully")
      },
      onError: () => {
        toast.error("Assignment addition failed")
      }
    })
  }

  /**
   * Handles moving an assignment to a new deadline date.
   * 
   * Updates the assignment's deadline if the new date differs from the current one.
   * Used primarily in the calendar view for drag-and-drop operations.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment to move
   * @param {Date} date - The new deadline date
   * @returns {Promise<void>}
   */
  const handleMoveAssignment = async (assignment: models.LocalAssignment, date: Date) => {
    const newDeadline = format(date, "yyyy-MM-dd HH:mm:ssxxx")
    if (!isSameDay(assignment.Deadline, date)) {
      handleEditAssignment(assignment, "deadline", newDeadline)
    }
  }

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
            <h1 className="text-h1 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              Assignments
            </h1>
            <p className="text-body-small text-gray-400 mt-3">Track and manage your coursework deadlines</p>
          </div>
          <AddAssignmentDialog isOpen={addAssignmentOpen} setOpen={setAddAssignmentOpen} onAdd={handleAddAssignment} />
        </div>

        <Tabs value={activeView} onValueChange={handleTabChange} className="flex flex-col flex-1 w-full">
          
          <TabsList className="flex flex-row bg-white/5 p-1 rounded-xl w-full mb-6 border border-white/5">
            <TabsTrigger 
              value="today" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Clock className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Today ({todayAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger 
              value="week" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <CalendarDays className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">This Week ({weekAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger 
              value="overdue" 
              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <AlertTriangle className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Overdue ({overdueAssignments?.length || 0})</span>
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

          <TabsContent value="today" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentView
              title="Today's Assignments"
              assignments={todayAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment} 
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              onEmptyClick={() => setAddAssignmentOpen(true)}
              isLoading={isLoading}
            />
          </TabsContent>

            <TabsContent value="week" className="flex flex-col data-[state=active]:flex-1 m-0">
              <AssignmentView
              title="Due This Week"
              assignments={weekAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              onEmptyClick={() => setAddAssignmentOpen(true)}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="overdue" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentView
              title="Overdue Assignments"
              assignments={overdueAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              onEmptyClick={() => setAddAssignmentOpen(true)}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="exam" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentView
              title="Exam Assignments"
              assignments={examAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              onEmptyClick={() => setAddAssignmentOpen(true)}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="calendar" className="flex flex-col data-[state=active]:flex-1 m-0">
            <AssignmentsCalendar
              assignments={assignments || []}
              onAddAssignment={() => { }}
              onEdit={handleEditAssignment}
              onMoveAssignment={handleMoveAssignment}
              onAssignmentClick={handleAssignmentClick}
              onDateClick={setSelectedDate}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="list" className="flex flex-col  data-[state=active]:flex-1 m-0">
            <AssignmentsTable
              assignments={assignments || []}
              onToggleComplete={handleToggleComplete}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onAssignmentClick={handleAssignmentClick}
              onOpenEdit={setSelectedAssignmentEdit}
              filter={{ course: courseFilter || "all", status: statusFilter || "all", priority: priorityFilter || "all" }}
              isLoading={isLoading}
            />
          </TabsContent>
        </Tabs>

        <AssignmentDetailsModal
          isOpen={!!selectedAssignmentID}
          onClose={() => setSelectedAssignmentID(null)}
          assignment_id={selectedAssignmentID!}  
          onOpenEdit={setSelectedAssignmentEdit}
          onEdit={handleEditAssignment}
          onDelete={handleDeleteAssignment}
          isLoading={updateMutation.isPending}
        />
        <DayAssignmentsModal
          isOpen={!!selectedDate}
          onClose={() => setSelectedDate(null)}
          date={selectedDate}
          assignments={assignments || []}
          onToggleComplete={handleToggleComplete}
          onAddAssignment={() => { }}
          onEdit={handleEditAssignment}
          onDelete={handleDeleteAssignment}
          onAssignmentClick={handleAssignmentClick}
          onOpenEdit={setSelectedAssignmentEdit}
          isLoading={updateMutation.isPending}
        />

        <AssignmentEditDialog
          open={!!selectedAssignmentEdit}
          setOpen={() => setSelectedAssignmentEdit(null)}
          assignment={selectedAssignmentEdit!}
          onEdit={handleEditAssignment}
        />
      </div>
    </div>
  )
}
