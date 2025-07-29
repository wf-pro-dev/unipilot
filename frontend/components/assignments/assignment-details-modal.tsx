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
import { AssignmentDocuments } from "./assignment-documents"

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


  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass-dark border-gray-600 max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex justify-between items-start">
            <DialogTitle className="text-xl font-semibold text-white">
              {assignment.Title}
            </DialogTitle>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="p-0 w-6 h-6">
                  <MoreVertical className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>

              <DropdownMenuContent align="end" className="border-gray-600 glass">
                <DropdownMenuItem onClick={() => onEdit(assignment, "title", assignment.Title)}>
                  <Edit className="mr-2 w-4 h-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => onDelete(assignment.ID)}
                  className="text-red-400 hover:text-red-300"
                >
                  <Trash2 className="mr-2 w-4 h-4" />
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
              <p className="font-medium text-white">{assignment.Course?.Name}</p>
            </div>
            <div className="space-y-2">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <Calendar className="w-4 h-4" />
                <span>Deadline</span>
              </div>
              <div className="flex items-center space-x-4">
                <p className="font-medium text-white">{format(deadline, "EEEE, MMMM d, yyyy")}</p>
                <div className={`flex items-center space-x-1 text-sm ${isOverdueStatus ? "text-red-400" : daysUntilDue <= 1 ? "text-yellow-400" : "text-gray-400"}`}>
                  <Clock className="w-4 h-4" />
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
                <FileText className="w-4 h-4" />
                <span>Description</span>
              </div>
              <p className="leading-relaxed text-gray-300">{assignment.Todo}</p>
            </div>
          )}


          {/* Link 
          {assignment.Link && (
            <div className="space-y-2">
              <div className="flex items-center space-x-2 text-sm text-gray-400">
                <ExternalLink className="w-4 h-4" />
                <span>Link</span>
              </div>
              <a
                href={assignment.Link}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-400 underline break-all hover:text-blue-300"
                onClick={() => {
                  window.open(assignment.Link, "_self")
                }}
              >
                {assignment.Link}
              </a>
            </div>
          )}

          {/* Notes *
          <div className="space-y-2">
            <label className="text-sm text-gray-400">Notes</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add your notes here..."
              className="p-3 w-full h-24 placeholder-gray-400 text-white bg-gray-800 rounded-lg border border-gray-600 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div> */}

          {/* Documents Section */}
          <div className="pt-4 border-t border-gray-600">
            <AssignmentDocuments assignment={assignment} />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
