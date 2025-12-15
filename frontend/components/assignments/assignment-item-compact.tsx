"use client"

import { Checkbox } from "@/components/ui/checkbox"
import { GlassCard } from "@/components/ui/glass-card"
import { Clock, Flag } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { parseDeadline, isOverdue, getDueDescription } from "@/lib/date-utils"
import { useState } from "react"
import { cn } from "@/lib/utils"

interface AssignmentItemCompactProps {
  assignment: assignment.LocalAssignment
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick?: (assignment: assignment.LocalAssignment) => void
  disabled?: boolean
  className?: string
}

const priorityColors = {
  low: "text-green-400",
  medium: "text-yellow-400",
  high: "text-red-400",
}

export function AssignmentItemCompact({
  assignment,
  onToggleComplete,
  onAssignmentClick,
  disabled = false,
  className
}: AssignmentItemCompactProps) {
  const [checked, setChecked] = useState(assignment.StatusName === "Done")

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(assignment.Deadline)
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
    <div className={className}>
      <GlassCard
        variant={!disabled && onAssignmentClick ? "interactive" : "default"}
        className={`border-white/5 bg-white/5 hover:bg-white/10 transition-all duration-300 ${disabled ? 'opacity-50' : ''}`}
        onClick={handleCardClick}
      >
        <div className="flex items-center p-3 gap-3">
          {/* Checkbox */}
          <div onClick={(e) => e.stopPropagation()}>
            <Checkbox
              checked={checked}
              onCheckedChange={handleToggleComplete}
              disabled={disabled}
              className="h-4 w-4 border-white/20 data-[state=checked]:bg-blue-500 data-[state=checked]:border-blue-500"
            />
          </div>

          {/* Main Content: Title & Course */}
          <div className="flex-1 min-w-0 flex flex-col justify-center">
            <h3 className={cn(
              "text-sm font-medium leading-none truncate",
              assignment.StatusName === "Done" ? "line-through text-gray-500" : "text-gray-200"
            )}>
              {assignment.Title}
            </h3>
            {(assignment.Course?.Code || assignment.CourseCode) && (
              <p className="text-xs text-gray-500 truncate mt-1">
                {assignment.Course?.Code || assignment.CourseCode}
              </p>
            )}
          </div>

          {/* Right Side: Date & Priority */}
          <div className="flex items-center gap-3 shrink-0">
            {/* Due Date */}
            <div className="flex items-center gap-1.5 text-xs">
              <Clock className="h-3 w-3 text-gray-500" />
              <span className={cn(
                "font-medium",
                isOverdueStatus ? "text-red-400" : "text-gray-400"
              )}>
                {getDueDescription(deadline, assignment.StatusName)}
              </span>
            </div>

            {/* Priority Flag */}
            {assignment.Priority && assignment.Priority !== "none" && (
              <Flag className={cn(
                "h-3 w-3",
                priorityColors[assignment.Priority as keyof typeof priorityColors] || "text-gray-500"
              )} />
            )}
          </div>
        </div>
      </GlassCard>
    </div>
  )
}

