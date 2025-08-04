"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Card, CardContent } from "@/components/ui/card"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Clock, MoreVertical, Edit, Trash2, Flag, BookOpen } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { parseDeadline, calculateDaysDifference, isOverdue, getDueDescription } from "@/lib/date-utils"
import { useState } from "react"
import { StatusTag } from "./utils/status-tag"
import { CourseTag } from "./utils/course-tag"
import { TypeTag } from "./utils/type-tag"
import { AssignmentEditDialog } from "./assignment-edit-dialog"

interface AssignmentItemProps {
  assignment: assignment.LocalAssignment
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick?: (assignment: assignment.LocalAssignment) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  disabled?: boolean
}

const priorityColors = {
  low: "text-green-400 border-green-400",
  medium: "text-yellow-400 border-yellow-400",
  high: "text-red-400 border-red-400",
}

const typeColors = {
  HW: "text-blue-400 border-blue-400",
  Exam: "text-red-400 border-red-400",
}

export function AssignmentItem({
  assignment,
  onEdit,
  onDelete,
  onToggleComplete,
  onAssignmentClick,
  disabled = false
}: AssignmentItemProps) {
  const [checked, setChecked] = useState(assignment.StatusName === "Done")
  const [open, setOpen] = useState(false)

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(assignment.Deadline)
  const daysUntilDue = calculateDaysDifference(deadline)
  const isOverdueStatus = isOverdue(deadline, assignment.StatusName)

  function handleToggleComplete() {
    if (disabled) return
    setChecked(!checked)
    onToggleComplete(assignment)
  }

  const handleCardClick = () => {
    if (onAssignmentClick && !disabled) {
      onAssignmentClick(assignment)
    }
  }

  const handleEditOpen = (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation()
    setOpen(true)
  }

  return (
    <div>
      <Card
        className={`glass border-0 hover:bg-white/5 transition-colors ${!disabled && onAssignmentClick ? 'cursor-pointer' : ''
          } ${disabled ? 'opacity-50' : ''}`}
        onClick={handleCardClick}
      >
        <CardContent className="p-4">
          <div className="flex items-start space-x-3">
            <div onClick={(e) => e.stopPropagation()}>
              <Checkbox
                checked={checked}
                onCheckedChange={handleToggleComplete}
                disabled={disabled}
                className="mt-1"
              />
            </div>
            <div className="flex-1 space-y-2">
              <div className="flex items-start justify-between gap-2">
                <div className="space-y-1">
                  <h3 className="font-medium text-white line-clamp-1">{assignment.Title}</h3>
                  <p className="text-sm text-gray-400 line-clamp-1">{assignment.Todo}</p>
                </div>
                <div className="flex items-center">
                  <div>
                    <Flag className={`h-5 w-5 ${priorityColors[assignment.Priority as keyof typeof priorityColors]}`} />
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-gray-400 hover:text-white"
                        disabled={disabled}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="glass border-gray-700">
                      <DropdownMenuItem
                        onClick={handleEditOpen}
                        disabled={disabled}
                      >
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation()
                          onDelete(assignment)
                        }}
                        disabled={disabled}
                        className="text-red-400"
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <div onClick={(e) => e.stopPropagation()}>
                    <CourseTag assignment={assignment} onEdit={onEdit} />
                  </div>

                  <div onClick={(e) => e.stopPropagation()}>
                    <StatusTag assignment={assignment} onEdit={onEdit} />
                  </div>

                  <div onClick={(e) => e.stopPropagation()}>
                    <TypeTag assignment={assignment} onEdit={onEdit} />
                  </div>


                </div>

                <div className="flex items-center space-x-2 text-xs text-gray-400">
                  <Clock className="h-3 w-3" />
                  <span className={isOverdueStatus ? "text-red-400" : ""}>
                    {getDueDescription(deadline, assignment.StatusName)}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>

      </Card>
      <AssignmentEditDialog
        open={open}
        setOpen={setOpen}
        assignment={assignment}
        onEdit={onEdit}
      />
    </div>
  )
}
