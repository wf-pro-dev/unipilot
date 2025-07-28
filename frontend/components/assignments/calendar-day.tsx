"use client"

import { useDrop } from "react-dnd"
import { AssignmentItem } from "./assignment-item"
import { DayAssignmentsModal } from "./day-assignments-modal"
import { useState } from "react"
import { Assignment } from "@/types/models"
import { CalendarItem } from "./calendar-item"
import { format } from "date-fns"

interface CalendarDayProps {
  date: Date | null
  assignments: Assignment[]
  isCurrentMonth: boolean
  onToggleComplete: (assignment: Assignment) => void
  onAddAssignment: () => void
}

export function CalendarDay({ date, assignments, isCurrentMonth, onToggleComplete, onAddAssignment }: CalendarDayProps) {
  const [showModal, setShowModal] = useState(false)

  const handleEdit = (assignment: Assignment) => {
    console.log("Editing assignment:", assignment)
  }
  const handleDelete = (id: number) => {
    console.log("Deleting assignment:", id)
  }

  const [{ isOver }, drop] = useDrop({
    accept: "assignment",
    drop: (item: { id: number }) => {
      if (date) {
        const newDate = date.toISOString().split("T")[0]
      }
    },
    collect: (monitor) => ({
      isOver: monitor.isOver(),
    }),
  })

  if (!date) {
    return <div className="h-32 p-1"></div>
  }

  const isToday = format(date, "MMM d, yyyy") === format(new Date(), "MMM d, yyyy")
  const hasMultipleAssignments = assignments.length > 1

  const handleDayClick = () => {
    if (assignments.length > 0) {
      setShowModal(true)
    }
  }

  return (
    <>
      <div
        ref={drop}
        className={`
          h-32 p-1 border rounded-lg transition-all duration-200 overflow-hidden cursor-pointer
          ${isCurrentMonth ? "border-gray-700" : "border-gray-800 opacity-50"}
          ${isOver ? "border-blue-400 bg-blue-500/10" : ""}
          ${isToday ? "border-blue-500 bg-blue-500/5" : ""}
          hover:border-gray-600
        `}
        onClick={handleDayClick}
      >
        <div className="h-full flex flex-col">
          <div className={`text-xs font-medium mb-1 ${isToday ? "text-blue-400" : "text-gray-400"}`}>
            {format(date, "MMM d, yyyy")}
          </div>

          <div className="flex-1 space-y-1 overflow-hidden">
            {hasMultipleAssignments ? (
              <div className="text-center py-4">
                <div className="text-lg font-bold text-blue-400">{assignments.length}</div>
                <div className="text-xs text-gray-400">assignments</div>
              </div>
            ) : (
              assignments.map((assignment) => (
                <div ref={drop} key={assignment.ID} className={`${isOver ? "border-blue-400 bg-blue-500/10" : ""}`}>
                  <CalendarItem assignment={assignment} /> 
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      <DayAssignmentsModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        date={date}
        assignments={assignments}
        onToggleComplete={onToggleComplete}
        onAddAssignment={onAddAssignment}
      />
    </>
  )
}
