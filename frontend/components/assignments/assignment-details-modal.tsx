"use client"

import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Calendar, Clock, MoreVertical, Edit, Trash2, Flag, User, FileText, Award, ExternalLink } from "lucide-react"
import { format } from "date-fns"
import { Assignment } from "@/types/models"
import { parseDeadline, calculateDaysDifference, isOverdue, getDueDescription } from "@/lib/date-utils"
import { StatusTag } from "../utils/status-tag"

interface AssignmentDetailsModalProps {
  isOpen: boolean
  onClose: () => void
  assignment: Assignment | null
  onEdit: (assignment: Assignment, column: string, value: string) => void
  onDelete: (id: number) => void
  onToggleComplete: (assignment: Assignment) => void
  isLoading?: boolean
}

const priorityColors = {
  low: "text-green-400 border-green-400 bg-green-500/10",
  medium: "text-yellow-400 border-yellow-400 bg-yellow-500/10",
  high: "text-red-400 border-red-400 bg-red-500/10",
}

const statusColors = {
  pending: "bg-gray-500/20 text-gray-400",
  "in-progress": "bg-blue-500/20 text-blue-400",
  completed: "bg-green-500/20 text-green-400",
  overdue: "bg-red-500/20 text-red-400",
}

const typeColors = {
  "HW": "bg-blue-500/20 text-blue-400",
  "Exam": "bg-red-500/20 text-red-400",
}

export function AssignmentDetailsModal({
  isOpen,
  onClose,
  assignment,
  onEdit,
  onDelete,
  onToggleComplete,
}: AssignmentDetailsModalProps) {
  const [notes, setNotes] = useState("")

  if (!assignment) return null

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(assignment.Deadline)

  const isOverdueStatus = isOverdue(deadline, assignment.StatusName)
  const daysUntilDue = calculateDaysDifference(deadline)
  const scorePercentage =
    assignment.EarnedPoints && assignment.MaxPoints
      ? (assignment.EarnedPoints / assignment.MaxPoints) * 100
      : 0



  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass-dark border-gray-600 max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-start justify-between">
            <DialogTitle className="text-xl font-semibold text-white">
              {assignment.Title}
            </DialogTitle>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>

              <DropdownMenuContent align="end" className="glass border-gray-600">
                <DropdownMenuItem onClick={() => onEdit(assignment, "title", assignment.Title)}>
                  <Edit className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => onDelete(assignment.ID)}
                  className="text-red-400 hover:text-red-300"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

          </div>
        </DialogHeader>

        <div className="space-y-6">
          {/* Status and Priority */}
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">

              <StatusTag assignment={assignment} onEdit={onEdit} />

              <Badge variant="outline" className={`text-sm ${priorityColors[assignment.Priority]}`}>
                {assignment.Priority}
              </Badge>

              <Badge variant="outline" className={`text-sm font-semibold ${typeColors[assignment.TypeName as keyof typeof typeColors]}`}>
                {assignment.TypeName}
              </Badge>


            </div>
          </div>

          {/* Course and Type */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <div className={`w-3 h-3 rounded-full ${assignment.Course?.Color}`} />
                <span>Course</span>
              </div>
              <p className="text-white font-medium">{assignment.Course?.Name}</p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <Calendar className="h-4 w-4" />
                <span>Deadline</span>
              </div>
              <div className="flex items-center space-x-4">
                <p className="text-white font-medium">{format(deadline, "EEEE, MMMM d, yyyy")}</p>
                <div className={`flex items-center space-x-1 text-sm ${isOverdueStatus ? "text-red-400" : daysUntilDue <= 1 ? "text-yellow-400" : "text-gray-400"}`}>
                  <Clock className="h-4 w-4" />
                  <span>
                    {getDueDescription(deadline, assignment.StatusName)}
                  </span>
                </div>
              </div>
            </div>

          </div>

          {/* Deadline */}


          {/* Description */}
          {assignment.Todo && (
            <div className="space-y-2">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <FileText className="h-4 w-4" />
                <span>Description</span>
              </div>
              <p className="text-gray-300 leading-relaxed">{assignment.Todo}</p>
            </div>
          )}




          {/* Link */}
          {assignment.Link && (
            <div className="space-y-2">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <ExternalLink className="h-4 w-4" />
                <span>Link</span>
              </div>
              <a
                href={assignment.Link}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-400 text-sm hover:text-blue-300 underline break-all"
                onClick={() => {
                  window.open(assignment.Link, "_self")
                }}
              >
                {assignment.Link}
              </a>
            </div>
          )}

          {/* Notes */}
          <div className="space-y-2">
            <label className="text-sm text-gray-400">Notes</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add your notes here..."
              className="w-full h-24 p-3 bg-gray-800 border border-gray-600 rounded-lg text-white placeholder-gray-400 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
