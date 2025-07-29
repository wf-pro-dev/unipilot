"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Calendar, BookOpen, CheckCircle2, Plus } from "lucide-react"
import { format, isSameDay } from "date-fns"
import { Assignment } from "@/types/models"
import { AssignmentItem } from "./assignment-item"

interface DayAssignmentsModalProps {
  isOpen: boolean
  onClose: () => void
  date: Date | null
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAddAssignment: () => void
  onEdit: (assignment: Assignment, column: string, value: string) => void
  onDelete: (id: number) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading: boolean
}

export function DayAssignmentsModal({
  isOpen,
  onClose,
  date,
  assignments,
  onToggleComplete,
  onAddAssignment,
  onEdit,
  onDelete,
  onAssignmentClick,

  isLoading,
  }: DayAssignmentsModalProps) {
  if (!date) return null
  const dayAssignments = assignments.filter((a) => isSameDay(a.Deadline, date))
  const completedCount = dayAssignments.filter((a) => a.StatusName === "Done").length
  const totalCount = dayAssignments.length
  const isDayComplete = completedCount === totalCount

    
  return (
    <Dialog open={isOpen} onOpenChange={onClose}  >
      <DialogContent className={`glass border-0 text-white max-w-2xl max-h-[80vh] overflow-y-auto`}>
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <Calendar className="h-5 w-5 text-blue-400" />
            <span>Assignments for {format(date, "MMMM d, yyyy")}</span>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6 mt-6">
          {/* Summary */}
          <div className={`flex items-center justify-between p-4 rounded-lg glass-dark`}>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <BookOpen className="h-4 w-4 text-blue-400" />
                <span className="text-white font-medium">
                  {totalCount} assignment{totalCount !== 1 ? "s" : ""}
                </span>
              </div>
              {totalCount > 0 && (
                <div className="flex items-center space-x-2">
                  <CheckCircle2 className={`h-4 w-4 ${isDayComplete ? "text-green-400" : "text-gray-400"}`} />
                  <span className={`${isDayComplete ? "text-green-400" : "text-gray-400"}`}>{completedCount} completed</span>
                </div>
              )}
            </div>
            <Button
              onClick={onAddAssignment}
              size="sm"
              className="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700"
            >
              <Plus className="h-4 w-4 mr-1" />
              Add Assignment
            </Button>
          </div>

          {/* Assignments List */}
          {dayAssignments.length === 0 ? (
            <div className="text-center py-12">
              <Calendar className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-300 mb-2">No assignments</h3>
              <p className="text-gray-400 mb-4">No assignments are due on this date</p>
              <Button onClick={onAddAssignment} variant="outline" className="border-gray-600 bg-transparent">
                <Plus className="h-4 w-4 mr-2" />
                Add Assignment
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {dayAssignments.map((assignment) => (
                <AssignmentItem
                  key={assignment.ID}
                  assignment={assignment}
                  onToggleComplete={onToggleComplete}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onAssignmentClick={onAssignmentClick}
                  disabled={isLoading}
                />
              ))}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
