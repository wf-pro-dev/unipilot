"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Separator } from "@/components/ui/separator"
import { Calendar, Clock, Flag, BookOpen, CheckCircle2, Plus } from "lucide-react"
import { format } from "date-fns"
import { Assignment } from "@/types/models"

interface DayAssignmentsModalProps {
  isOpen: boolean
  onClose: () => void
  date: Date | null
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAddAssignment: () => void
}

const priorityColors = {
  low: "text-green-400 border-green-400",
  medium: "text-yellow-400 border-yellow-400",
  high: "text-red-400 border-red-400",
}

const statusColors = {
  pending: "bg-gray-500/20 text-gray-400",
  "in-progress": "bg-blue-500/20 text-blue-400",
  completed: "bg-green-500/20 text-green-400",
  overdue: "bg-red-500/20 text-red-400",
}

export function DayAssignmentsModal({
  isOpen,
  onClose,
  date,
  assignments,
  onToggleComplete,
  onAddAssignment,
}: DayAssignmentsModalProps) {
  if (!date) return null

  const completedCount = assignments.filter((a) => a.Completed).length
  const totalCount = assignments.length

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-0 text-white max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <Calendar className="h-5 w-5 text-blue-400" />
            <span>Assignments for {format(date, "MMMM d, yyyy")}</span>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6 mt-6">
          {/* Summary */}
          <div className="flex items-center justify-between p-4 rounded-lg glass-dark">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <BookOpen className="h-4 w-4 text-blue-400" />
                <span className="text-white font-medium">
                  {totalCount} assignment{totalCount !== 1 ? "s" : ""}
                </span>
              </div>
              {totalCount > 0 && (
                <div className="flex items-center space-x-2">
                  <CheckCircle2 className="h-4 w-4 text-green-400" />
                  <span className="text-green-400">{completedCount} completed</span>
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
          {assignments.length === 0 ? (
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
              {assignments.map((assignment, index) => (
                  <div key={assignment.ID}>
                  <div className="flex items-start space-x-4 p-4 rounded-lg glass-dark">
                    <Checkbox
                      checked={assignment.StatusName == "Done"}
                      onCheckedChange={() => onToggleComplete(assignment)}
                      className="mt-1"
                    />

                    <div className="flex-1 space-y-2">
                      <div className="flex items-start justify-between">
                        <div className="space-y-1">
                          <h3
                            className={`font-semibold ${
                              assignment.StatusName == "Done" ? "line-through text-gray-500" : "text-white"
                            }`}
                          >
                            {assignment.Title}
                          </h3>
                          <div className="flex items-center space-x-3 text-sm">
                            <div className="flex items-center space-x-1">
                              <div className={`w-2 h-2 rounded-full ${assignment.Course?.Color}`} />
                              <span className="text-gray-300">{assignment.Course?.Name}</span>
                            </div>
                            <Badge variant="outline" className="text-xs border-gray-600">
                              {assignment.TypeName}
                            </Badge>
                          </div>
                        </div>

                        <div className="flex items-center space-x-2">
                          <Flag className={`h-4 w-4 ${priorityColors[assignment.Priority]}`} />
                          <Badge variant="outline" className={`text-xs ${statusColors[assignment.StatusName.toLowerCase() as keyof typeof statusColors  ]}`}>
                            {assignment.StatusName === "In Progress" ? "In Progress" : assignment.StatusName}
                          </Badge>
                        </div>
                      </div>

                      <div className="flex items-center space-x-4 text-sm text-gray-400">
                        <div className="flex items-center space-x-1">
                          <Clock className="h-3 w-3" />
                          <span>Due {format(assignment.Deadline, "h:mm a")}</span>
                        </div>
                        <Badge variant="outline" className={`text-xs ${priorityColors[assignment.Priority]}`}>
                          {assignment.Priority} priority
                        </Badge>
                      </div>

                      {assignment.Todo && <p className="text-sm text-gray-400 mt-2">{assignment.Todo}</p>}
                    </div>
                  </div>

                  {index < assignments.length - 1 && <Separator className="bg-gray-700 my-2" />}
                </div>
              ))}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
