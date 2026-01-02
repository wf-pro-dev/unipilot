"use client"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Clock, MoreVertical, Edit, Trash2, Flag } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { parseDeadline, isOverdue, getDueDescription } from "@/lib/date-utils"
import { useState } from "react"
import { StatusTag } from "./tags/status-tag"
import { CourseTag } from "./tags/course-tag"
import { TypeTag } from "./tags/type-tag"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"

interface AssignmentItemProps {
  assignment: assignment.LocalAssignment
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick?: (assignment: assignment.LocalAssignment) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  onOpenEdit: (assignment: assignment.LocalAssignment) => void
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
  onOpenEdit,
  disabled = false
}: AssignmentItemProps) {
  const [checked, setChecked] = useState(assignment.Status === "Done")

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(assignment.Deadline)
  const isOverdueStatus = isOverdue(deadline, assignment.Status)

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
    onOpenEdit(assignment)
  }

  const handleOpenLink = (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation()
    BrowserOpenURL(assignment.Link)
  }

  return (
    <div>
      <GlassCard
        variant={!disabled && onAssignmentClick ? "interactive" : "default"}
        className={`h-full border-white/5 bg-white/5 hover:bg-white/10 transition-all duration-300 ${disabled ? 'opacity-50' : ''}`}
        onClick={handleCardClick}
      >
        <CardContent className="p-5">
          <div className="flex  gap-4">
            {/* Left Column: Checkbox */}
            <div onClick={(e) => e.stopPropagation()} className="pt-1">
              <Checkbox
                checked={checked}
                onCheckedChange={handleToggleComplete}
                disabled={disabled}
                className="border-white/20 data-[state=checked]:bg-blue-500 data-[state=checked]:border-blue-500"
              />
            </div>

            {/* Right Column: Main Content */}
            <div className="flex-1 min-w-0 flex flex-col gap-3">

              {/* 1. Context Header: Course & Type tags + Actions */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
                  <CourseTag assignment={assignment} onEdit={onEdit} />
                  <TypeTag assignment={assignment} onEdit={onEdit} />
                </div>

                {/* Actions Group */}
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-lg bg-white/5 border border-white/5">
                    <Flag className={`h-4 w-4 ${priorityColors[assignment.Priority as keyof typeof priorityColors]}`} />
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 w-7 p-0 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                        disabled={disabled}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    {/* ... existing dropdown content ... */}
                    <DropdownMenuContent align="end" className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                      <DropdownMenuItem
                        onClick={handleEditOpen}
                        disabled={disabled}
                        className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
                      >
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={handleOpenLink}
                        disabled={disabled}
                        className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
                      >
                        <Edit className="h-4 w-4 mr-2" />
                        Open link
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation()
                          onDelete(assignment)
                        }}
                        disabled={disabled}
                        className="text-red-400 focus:text-red-300 focus:bg-red-500/10 cursor-pointer"
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>

              {/* 2. Main Info: Title & Description */}
              <div className="space-y-1">
                <h3 className={`text-h4 ${assignment.Status === "Done" ? "line-through text-gray-400" : "text-white"} line-clamp-1 tracking-tight`}>
                  {assignment.Title}
                </h3>
                {assignment.Todo ? (
                  <p className={`text-caption ${assignment.Status === "Done" ? "text-gray-400" : "text-white"} line-clamp-1 leading-relaxed`}  >{assignment.Todo}</p>
                ) : (
                  <p className="text-caption text-gray-400 line-clamp-1 leading-relaxed">{"No description yet..."}</p>
                )}
              </div>

              {/* 3. Footer: Status & Date */}
              <div className="flex items-center justify-between pt-1 border-t border-white/5 mt-1">
                <div onClick={(e) => e.stopPropagation()}>
                  <StatusTag assignment={assignment} onEdit={onEdit} />
                </div>

                <div className="flex items-center space-x-1.5 text-xs text-gray-400 bg-black/20 px-2.5 py-1.5 rounded-lg border border-white/5">
                  <Clock className="h-3 w-3" />
                  <span className={isOverdueStatus ? "text-red-400 font-bold" : "font-medium text-gray-300"}>
                    {getDueDescription(deadline, assignment.Status)}
                  </span>
                </div>
              </div>

            </div>
          </div>
        </CardContent>
      </GlassCard>
    </div>
  )
}
