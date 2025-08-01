"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Calendar, ChevronLeft, ChevronRight, Loader2, Plus } from "lucide-react"
import { format } from "date-fns"
import { assignment } from "@/wailsjs/go/models"
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
    return (assignments || []).filter(assignment => {
      if (!assignment.Deadline) return false
      const deadline = parseDeadline(assignment.Deadline)
      try {
        return format(deadline, "MMM d, yyyy") === format(date, "MMM d, yyyy")
      } catch (error) {
        console.warn("Invalid date for assignment:", assignment.ID, assignment.Deadline)
        return false
      }
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
        <Card className="glass border-0">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center space-x-2 text-white">
                <Calendar className="h-5 w-5" />
                <span>Assignment Calendar</span>
                {isLoading && <Loader2 className="h-4 w-4 animate-spin ml-2" />}
              </CardTitle>
              <div className="flex items-center space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigateMonth("prev")}
                  disabled={isLoading}
                  className="text-gray-400 hover:text-white"
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-white font-medium min-w-[120px] text-center">
                  {format(currentDate, "MMMM yyyy")}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigateMonth("next")}
                  disabled={isLoading}
                  className="text-gray-400 hover:text-white"
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
                <Button
                  onClick={onAddAssignment}
                  className="bg-blue-600 hover:bg-blue-700"
                  disabled={isLoading}
                >
                  <Plus className="h-4 w-4 mr-2" />
                  Add
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-6">
            {/* Day Headers */}
            <div className="grid grid-cols-7 gap-1 mb-4">
              {dayNames.map((day) => (
                <div key={day} className="text-center text-sm font-medium text-gray-400 py-2">
                  {day}
                </div>
              ))}
            </div>

            {/* Calendar Days */}
            <div className="grid grid-cols-7 gap-1">
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
                    onMoveAssignment={onMoveAssignment}
                    onDateClick={onDateClick}
                    onEdit={onEdit}
                    onAssignmentClick={onAssignmentClick}
                    index={index}
                    />
                )
              })}
            </div>
          </CardContent>
        </Card>
      </div>
    </DndProvider>
  )
}
