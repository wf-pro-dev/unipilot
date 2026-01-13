"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Calendar, BookOpen, CheckCircle2, Plus } from "lucide-react"
import { format, isSameDay } from "date-fns"
import { models } from "@/wailsjs/go/models"
import { AssignmentItem } from "./assignment-item"

interface DayAssignmentsModalProps {
  isOpen: boolean
  onClose: () => void
  date: Date | null
  assignments: models.LocalAssignment[]
  onAddAssignment: () => void
  onEdit: (assignment: models.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: models.LocalAssignment) => void
  onOpenEdit: (assignment: models.LocalAssignment) => void
  isLoading: boolean
}

export function DayAssignmentsModal({
  isOpen,
  onClose,
  date,
  assignments,
  onAddAssignment,
  onEdit,
  onDelete,
  onOpenEdit,
  isLoading,
  }: DayAssignmentsModalProps) {
  if (!date) return null
  const dayAssignments = assignments.filter((a) => isSameDay(a.Deadline, date))
  const completedCount = dayAssignments.filter((a) => a.Status === "Done").length
  const totalCount = dayAssignments.length
  const isDayComplete = completedCount === totalCount

    
  return (
    <Dialog open={isOpen} onOpenChange={onClose}  >
      <DialogContent className={`glass border-white/10 text-white max-w-2xl max-h-[80vh] overflow-y-auto p-0 overflow-hidden gap-0`}>
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <DialogTitle className="flex items-center space-x-2 text-xl font-semibold">
            <Calendar className="h-5 w-5 text-blue-400" />
            <span>Assignments for {format(date, "MMMM d, yyyy")}</span>
          </DialogTitle>
        </DialogHeader>

        <div className="p-6 space-y-6">
          {/* Summary */}
          <div className={`flex items-center justify-between p-4 rounded-xl bg-white/5 border border-white/5`}>
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
                  <span className={`${isDayComplete ? "text-green-400" : "text-gray-400"} font-medium text-sm`}>{completedCount} completed</span>
                </div>
              )}
            </div>
            <Button
              onClick={onAddAssignment}
              size="sm"
              className="bg-blue-600 hover:bg-blue-500 text-white shadow-[0_0_15px_rgba(37,99,235,0.2)] border-0"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Assignment
            </Button>
          </div>

          {/* Assignments List */}
          {dayAssignments.length === 0 ? (
            <div className="text-center py-12 border border-dashed border-white/10 rounded-xl bg-white/5">
              <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
                <Calendar className="h-8 w-8 text-gray-500" />
              </div>
              <h3 className="text-lg font-medium text-white mb-1">No assignments</h3>
              <p className="text-gray-400 mb-4 text-sm">No assignments are due on this date</p>
              <Button 
                onClick={onAddAssignment} 
                variant="outline" 
                className="border-white/10 bg-white/5 hover:bg-white/10 text-white"
              >
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
                  variant="outline"
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onOpenEdit={onOpenEdit}
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
