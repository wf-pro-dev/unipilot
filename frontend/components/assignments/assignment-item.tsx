"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Card, CardContent } from "@/components/ui/card"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Clock, MoreVertical, Edit, Trash2, Flag, Book, BookOpen } from "lucide-react"
import { Assignment } from "@/types/models"
import { parseDeadline, calculateDaysDifference, isOverdue, getDueDescription } from "@/lib/date-utils"
import { StatusTag } from "../utils/status-tag"
import { useState } from "react"

interface AssignmentItemProps {
  assignment: Assignment
  onEdit: (assignment: Assignment, column: string, value: string) => void
  onDelete: (id: number) => void
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick?: (assignment: Assignment) => void
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

  return (
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
                <h3 className="font-medium text-white">{assignment.Title}</h3>
                <p className="text-sm text-gray-400">{assignment.Todo}</p>
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
                      onClick={(e) => {
                        e.stopPropagation()
                        // Handle edit
                      }}
                      disabled={disabled}
                    >
                      <Edit className="h-4 w-4 mr-2" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.stopPropagation()
                        onDelete(assignment.ID)
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
                <Badge variant="secondary" className="text-xs flex flex-row gap-2">
                  <div className={`h-2 w-2  rounded-full ${assignment.Course?.Color}`} />
                  {assignment.Course?.Code || 'No Course'}
                </Badge>
                <div onClick={(e) => e.stopPropagation()}>
                  <StatusTag assignment={assignment} onEdit={onEdit} />
                </div>

                <Badge
                  variant="outline"
                  className={`text-xs ${typeColors[assignment.TypeName as keyof typeof typeColors]}`}
                >
                  <BookOpen className="h-3 w-3 mr-1" />
                  {assignment.TypeName}
                </Badge>
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
  )
}
