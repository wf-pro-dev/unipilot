"use client"

import { useState, useEffect, useRef } from "react"
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { Button } from "@/components/ui/button"
import { Calendar, ChevronLeft, ChevronRight, Loader2, Plus } from "lucide-react"
import { format } from "date-fns"
import { assignment, models } from "@/wailsjs/go/models"
import { parseDeadline } from "@/lib/date-utils"
import { CalendarContainer } from "./calendar-container"
import { DndProvider } from "react-dnd"
import { HTML5Backend } from "react-dnd-html5-backend"

interface AssignmentsCalendarProps {
  assignments: assignment.LocalAssignment[]
  onAddAssignment: () => void
  onMoveAssignment: (assignment: assignment.LocalAssignment, date: Date) => void
  isLoading?: boolean
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  onDateClick: (date: Date) => void
}

export function AssignmentsCalendar({
  assignments,
  onAddAssignment,
  onMoveAssignment,
  onEdit,
  onAssignmentClick,
  onDateClick,
  isLoading = false,
}: AssignmentsCalendarProps) {
  const [currentDate, setCurrentDate] = useState(new Date())
  // Local state to manage assignments for optimistic updates
  const [localAssignments, setLocalAssignments] = useState<assignment.LocalAssignment[]>(assignments)
  // Track pending optimistic updates: map of assignment ID to new deadline string
  const pendingUpdatesRef = useRef<Map<number, string>>(new Map())
  // Track if we should skip the next sync (to prevent overwriting optimistic updates)
  const skipNextSyncRef = useRef(false)

  // Sync local state with props when assignments change, but preserve optimistic updates
  useEffect(() => {
    // Skip sync if we just made an optimistic update
    if (skipNextSyncRef.current) {
      skipNextSyncRef.current = false
      return
    }

    if (pendingUpdatesRef.current.size === 0) {
      // No pending updates, sync normally
      setLocalAssignments(assignments)
      return
    }

    // We have pending updates - merge props with our optimistic updates
    setLocalAssignments(prev => {
      const updated = assignments.map(a => {
        const pendingDeadline = pendingUpdatesRef.current.get(a.ID)
        if (pendingDeadline) {
          // Check if the prop already reflects our optimistic update
          // Compare dates by day (not time) since we only care about the date
          const propDeadline = a.Deadline ? parseDeadline(a.Deadline) : null
          const pendingDeadlineDate = parseDeadline(pendingDeadline)
          
          if (propDeadline && format(propDeadline, "yyyy-MM-dd") === format(pendingDeadlineDate, "yyyy-MM-dd")) {
            // Prop matches our optimistic update, remove from pending
            pendingUpdatesRef.current.delete(a.ID)
            return a
          }
          
          // Prop doesn't match yet, keep our optimistic update
          return { ...a, Deadline: pendingDeadline } as assignment.LocalAssignment
        }
        return a
      })
      return updated
    })
  }, [assignments])

  // Wrapper for onMoveAssignment that optimistically updates local state
  const handleMoveAssignment = (assignment: assignment.LocalAssignment, date: Date) => {
    // Optimistically update the assignment's deadline in local state
    const newDeadline = format(date, "yyyy-MM-dd HH:mm:ssxxx")
    
    // Track this as a pending update
    pendingUpdatesRef.current.set(assignment.ID, newDeadline)
    
    // Skip the next sync to prevent overwriting our optimistic update
    skipNextSyncRef.current = true
    
    // Optimistically update local state immediately
    setLocalAssignments(prev => 
      prev.map(a => 
        a.ID === assignment.ID 
          ? { ...a, Deadline: newDeadline } as assignment.LocalAssignment
          : a
      )
    )
    
    // Call the original onMoveAssignment handler
    onMoveAssignment(assignment, date)
  }

  const getDaysInMonth = (date: Date) => {
    const year = date.getFullYear()
    const month = date.getMonth()
    const firstDay = new Date(year, month, 1)
    const startingDayOfWeek = firstDay.getDay()

    const days = []

    // Add days from previous month to fill beginning
    const prevMonth = new Date(year, month - 1)
    const prevMonthYear = prevMonth.getFullYear()
    const prevMonthIndex = prevMonth.getMonth()
    const daysInPrevMonth = new Date(prevMonthYear, prevMonthIndex + 1, 0).getDate()

    for (let i = startingDayOfWeek - 1; i >= 0; i--) {
      days.push(new Date(prevMonthYear, prevMonthIndex, daysInPrevMonth - i))
    }

    // Add all days of current month
    const daysInMonth = new Date(year, month + 1, 0).getDate()
    for (let day = 1; day <= daysInMonth; day++) {
      days.push(new Date(year, month, day))
    }

    // Add days from next month to reach exactly 35 days
    const nextMonth = new Date(year, month + 1)
    const nextMonthYear = nextMonth.getFullYear()
    const nextMonthIndex = nextMonth.getMonth()

    let nextMonthDay = 1
    while (days.length < 35) {
      days.push(new Date(nextMonthYear, nextMonthIndex, nextMonthDay))
      nextMonthDay++
    }

    return days
  }

  const getAssignmentsForDate = (date: Date) => {
    return (localAssignments || []).filter(assignment => {
      if (!assignment.Deadline) return false
      const deadline = parseDeadline(assignment.Deadline)
      try {
        return format(deadline, "MMM d, yyyy") === format(date, "MMM d, yyyy")
      } catch (error) {
        console.warn("Invalid date for assignment:", assignment.ID, assignment.Deadline)
        return false
      }
    }).sort((a, b) => {
      var status: { [key: string]: number } = {
        "Not started": 1,
        "In progress": 0,
        "Done": 2
      }
      return status[a.StatusName] - status[b.StatusName]
    })
  }

  const navigateMonth = (direction: "prev" | "next") => {
    setCurrentDate((prev) => {
      const newDate = new Date(prev)
      if (direction === "prev") {
        newDate.setMonth(prev.getMonth() - 1)
      } else {
        newDate.setMonth(prev.getMonth() + 1)
      }
      return newDate
    })
  }

  const dayNames = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
  const days = getDaysInMonth(currentDate)

  return (
    <DndProvider backend={HTML5Backend}>
      <div className="space-y-6">
        <GlassCard className="border-white/5 bg-white/5 backdrop-blur-xl shadow-lg shadow-black/20">
          <CardHeader className="p-5 pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-500/10 rounded-lg border border-blue-500/10">
                  <Calendar className="h-5 w-5 text-blue-400" />
                </div>
                <CardTitle className="text-lg font-semibold text-white tracking-tight flex items-center gap-2">
                  <span>Assignment Calendar</span>
                  {isLoading && (
                    <span className="flex items-center gap-1.5 text-[10px] font-medium bg-white/5 px-2 py-0.5 rounded-full text-gray-400 uppercase tracking-wider border border-white/5 ml-2">
                      <Loader2 className="h-3 w-3 animate-spin" />
                      Syncing
                    </span>
                  )}
                </CardTitle>
              </div>
              
              <div className="flex items-center bg-white/5 rounded-lg border border-white/5 p-1">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigateMonth("prev")}
                  disabled={isLoading}
                  className="h-7 w-7 p-0 text-gray-400 hover:text-white hover:bg-white/10 rounded-md"
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm font-medium text-white min-w-[100px] text-center">
                  {format(currentDate, "MMMM yyyy")}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigateMonth("next")}
                  disabled={isLoading}
                  className="h-7 w-7 p-0 text-gray-400 hover:text-white hover:bg-white/10 rounded-md"
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-5 pt-2">
            {/* Day Headers */}
            <div className="grid grid-cols-7 gap-1 mb-2">
              {dayNames.map((day) => (
                <div key={day} className="text-center text-[10px] font-semibold text-gray-500 uppercase tracking-wider py-2">
                  {day}
                </div>
              ))}
            </div>

            {/* Calendar Days */}
            <div className="grid grid-cols-7 gap-px p-px rounded-xl overflow-hidden border border-white/10">
              {days.map((day, index) => {
                const isCurrentMonth = day.getMonth() === currentDate.getMonth() && day.getFullYear() === currentDate.getFullYear()
                const isToday = format(day, "MMM d, yyyy") === format(new Date(), "MMM d, yyyy")
                const dayAssignments = getAssignmentsForDate(day)
                return (
                  <CalendarContainer
                    key={index}
                    day={day}
                    dayAssignments={dayAssignments}
                    isCurrentMonth={isCurrentMonth}
                    isToday={isToday}
                    onMoveAssignment={handleMoveAssignment}
                    onDateClick={onDateClick}
                    onEdit={onEdit}
                    onAssignmentClick={onAssignmentClick}
                    index={index}
                  />
                )
              })}
            </div>
          </CardContent>
        </GlassCard>
      </div>
    </DndProvider>
  )
}
