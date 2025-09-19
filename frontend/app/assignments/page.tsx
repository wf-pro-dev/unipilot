"use client"

import { useState, useEffect } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AssignmentsCalendar } from "@/components/assignments/assignments-calendar"
import { AssignmentsTable } from "@/components/assignments/assignments-table"
import { AddAssignmentDialog } from "@/components/assignments/add-assignment-dialog"
import { AssignmentDetailsModal } from "@/components/assignments/assignment-details-modal"
import { Calendar, List, Clock, CheckCircle2, AlertTriangle, CalendarDays, Loader2 } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
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
  const [selectedAssignmentEdit, setSelectedAssignmentEdit] = useState<assignment.LocalAssignment | null>(null)
  const [selectedDate, setSelectedDate] = useState<Date | null>(null)

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

  const handleAssignmentClick = (assignment: assignment.LocalAssignment) => {
    setSelectedAssignmentID(assignment.ID)
  }

  console.log("filter (page)", { courseFilter, statusFilter, priorityFilter })

  const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
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

  const handleToggleComplete = async (assignment: assignment.LocalAssignment) => {
    const newStatus = assignment.StatusName === "Done" ? "Not started" : "Done"
    handleEditAssignment(assignment, "status_name", newStatus)
  }

  const handleDeleteAssignment = async (assignment: assignment.LocalAssignment) => {
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


  const handleAddAssignment = async (assignment: assignment.LocalAssignment) => {
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

  const handleMoveAssignment = async (assignment: assignment.LocalAssignment, date: Date) => {
    const newDeadline = format(date, "yyyy-MM-dd HH:mm:ssxxx")
    if (!isSameDay(assignment.Deadline, date)) {
      handleEditAssignment(assignment, "deadline", newDeadline)
    }
  }

  // Handle tab change and update URL
  const handleTabChange = (value: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("view", value)
    router.push(`/assignments?${params.toString()}`)
  }

  // Show loading state
  if (isLoading) {
    return (
      <div className="page">
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
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500">
            Error loading assignments: {error.message}
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
            <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              Assignments
            </h1>
            <p className="text-gray-400 mt-2">Track and manage your coursework deadlines</p>
          </div>
          <AddAssignmentDialog onAdd={handleAddAssignment} />
        </div>

        <Tabs value={activeView} onValueChange={handleTabChange} className="w-full">
          <TabsList className="grid w-full grid-cols-6 glass border-0 mb-4">
            <TabsTrigger value="today" className="flex items-center space-x-2 ">
              <Clock className="h-4 w-4" />
              <span>Today ({todayAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="week" className="flex items-center space-x-2 ">
              <CalendarDays className="h-4 w-4" />
              <span>This Week ({weekAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="overdue" className="flex items-center space-x-2 ">
              <AlertTriangle className="h-4 w-4" />
              <span>Overdue ({overdueAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="exam" className="flex items-center space-x-2 ">
              <CheckCircle2 className="h-4 w-4" />
              <span>Exam ({examAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="calendar" className="flex items-center space-x-2 ">
              <Calendar className="h-4 w-4" />
              <span>Calendar</span>
            </TabsTrigger>
            <TabsTrigger value="list" className="flex items-center space-x-2 ">
              <List className="h-4 w-4" />
              <span>All ({assignments?.length || 0})</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="today">
            <AssignmentView
              title="Today's Assignments"
              assignments={todayAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment} 
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="week">
            <AssignmentView
              title="Due This Week"
              assignments={weekAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="overdue">
            <AssignmentView
              title="Overdue Assignments"
              assignments={overdueAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="exam">
            <AssignmentView
              title="Exam Assignments"
              assignments={examAssignments}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={handleAssignmentClick}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onOpenEdit={setSelectedAssignmentEdit}
              isLoading={isLoading}
            />
          </TabsContent>

          <TabsContent value="calendar">
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

          <TabsContent value="list">
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
