"use client"

import { useState } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AssignmentsCalendar } from "@/components/assignments/assignments-calendar"
import { AssignmentsTable } from "@/components/assignments/assignments-table"
import { TodayAssignments } from "@/components/assignments/today-assignments"
import { WeekAssignments } from "@/components/assignments/week-assignments"
import { OverdueAssignments } from "@/components/assignments/overdue-assignments"
import { ExamAssignments } from "@/components/assignments/exam-assignments"
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

  const [selectedAssignment, setSelectedAssignment] = useState<assignment.LocalAssignment | null>(null)
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

  const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
    console.log("Editing assignment:", assignment)
    const message = "assignment " + assignment.ID + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      assignment,
      column,
      value
    })
  }

  const handleToggleComplete = async (assignment: assignment.LocalAssignment) => {
    const newStatus = assignment.StatusName === "Done" ? "Not started" : "Done"
    handleEditAssignment(assignment, "status_name", newStatus)
  }

  const handleDeleteAssignment = async (assignment: assignment.LocalAssignment) => {
    const message = "assignment " + assignment.Title + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteMutation.mutate(assignment)
  }


  const handleAddAssignment = async (assignment: assignment.LocalAssignment) => {
    const message = "assignment " + assignment.Title + " added"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    createMutation.mutate(assignment)
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
            <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              Assignments
            </h1>
            <p className="text-gray-400 mt-2">Track and manage your coursework deadlines</p>
          </div>
          <AddAssignmentDialog onAdd={handleAddAssignment} />
        </div>

        <Tabs value={activeView} onValueChange={handleTabChange} className="w-full">
          <TabsList className="grid w-full grid-cols-6 glass border-0 mb-8">
            <TabsTrigger value="today" className="flex items-center space-x-2">
              <Clock className="h-4 w-4" />
              <span>Today ({todayAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="week" className="flex items-center space-x-2">
              <CalendarDays className="h-4 w-4" />
              <span>This Week ({weekAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="overdue" className="flex items-center space-x-2">
              <AlertTriangle className="h-4 w-4" />
              <span>Overdue ({overdueAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="exam" className="flex items-center space-x-2">
              <CheckCircle2 className="h-4 w-4" />
              <span>Exam ({examAssignments?.length || 0})</span>
            </TabsTrigger>
            <TabsTrigger value="calendar" className="flex items-center space-x-2">
              <Calendar className="h-4 w-4" />
              <span>Calendar</span>
            </TabsTrigger>
            <TabsTrigger value="list" className="flex items-center space-x-2">
              <List className="h-4 w-4" />
              <span>All ({assignments?.length || 0})</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="today">
            <TodayAssignments
              assignments={todayAssignments || []}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={setSelectedAssignment}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              isLoading={updateMutation.isPending}
            />
          </TabsContent>

          <TabsContent value="week">
            <WeekAssignments
              assignments={weekAssignments || []}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={setSelectedAssignment}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              isLoading={updateMutation.isPending}
            />
          </TabsContent>

          <TabsContent value="overdue">
            <OverdueAssignments
              assignments={overdueAssignments || []}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={setSelectedAssignment}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              isLoading={updateMutation.isPending}
            />
          </TabsContent>

          <TabsContent value="exam">
            <ExamAssignments
              assignments={examAssignments || []}
              onToggleComplete={handleToggleComplete}
              onAssignmentClick={setSelectedAssignment}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              isLoading={updateMutation.isPending}
            />
          </TabsContent>

          <TabsContent value="calendar">
            <AssignmentsCalendar
              assignments={assignments || []}
              onAddAssignment={() => { }}
              onEdit={handleEditAssignment}
              onMoveAssignment={handleMoveAssignment}
              onAssignmentClick={setSelectedAssignment}
              onDateClick={setSelectedDate}
              isLoading={updateMutation.isPending}
            />
          </TabsContent>

          <TabsContent value="list">
            <AssignmentsTable
              assignments={assignments || []}
              onToggleComplete={handleToggleComplete}
              onEdit={handleEditAssignment}
              onDelete={handleDeleteAssignment}
              onAssignmentClick={setSelectedAssignment}
              filter={{ course: courseFilter || "all", status: statusFilter || "all", priority: priorityFilter || "all" }}
              isLoading={updateMutation.isPending}
            />
          </TabsContent>
        </Tabs>

        <AssignmentDetailsModal
          isOpen={!!selectedAssignment}
          onClose={() => setSelectedAssignment(null)}
          assignment={selectedAssignment}
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
          onAddAssignment={() => {}}
          onEdit={handleEditAssignment}
          onDelete={handleDeleteAssignment}
          onAssignmentClick={setSelectedAssignment}
          isLoading={updateMutation.isPending}
        />
      </div>
    </div>
  )
}
