"use client"

import { GlassCard } from "@/components/ui/glass-card"
import { Button } from "@/components/ui/button"
import { Clock, CopyPlus } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { parseDeadline, getDueDescription, isOverdue } from "@/lib/date-utils"
import { cn } from "@/lib/utils"
import { toast } from "sonner"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { format } from "date-fns"

interface AssignmentItemCompactProps {
  assignment: assignment.LocalAssignment
  onClick?: (assignment: assignment.LocalAssignment) => void
  onCopy?: (assignment: assignment.LocalAssignment) => void
  disabled?: boolean
  className?: string
}

const priorityColors = {
  high: "bg-red-500",
  medium: "bg-yellow-500",
  low: "bg-green-500",
}

export function AssignmentItemCompact({
  assignment,
  onClick,
  onCopy,
  disabled = false,
  className
}: AssignmentItemCompactProps) {
  // Parse deadline
  const deadline = parseDeadline(assignment.Deadline)

  // Mock user data since assignment model doesn't have it directly yet
  // In a real app, this would come from assignment.User or similar
  const user = {
    username: "Instructor",
    avatar: "/placeholder-user.jpg"
  }

 

  const priorityColor = priorityColors[assignment.Priority as keyof typeof priorityColors] || "bg-gray-500"

  return (
    <div className={className} onClick={() => onClick?.(assignment)}>
      <GlassCard
        variant="board"
        className={`p-4 ${disabled ? 'opacity-50' : ''}`}
      >
        <div className="flex items-start gap-3">
          {/* Priority Indicator - Vertical Bar */}
          <div className={cn("w-1 h-10 rounded-full shrink-0", 
            priorityColor
          )} />

          <div className="flex-1 min-w-0 space-y-4">
            <div className="space-y-2">
              <h3 className="text-sm font-medium text-gray-200 truncate leading-tight">
                {assignment.Title}
              </h3>
              <p className="text-xs text-gray-400 font-medium">
                {assignment.Course?.Code || assignment.CourseCode}
              </p>
            </div>

            <div className="h-px w-full bg-white/20" />

            <div className="flex items-center justify-between">
              {/* User Info */}
              <div className="flex items-center gap-2">
                <Avatar className="h-5 w-5 border border-white/10">
                  <AvatarImage src={assignment.User?.Avatar || "/placeholder-user.jpg"} />
                  <AvatarFallback className="text-[10px]">IN</AvatarFallback>
                </Avatar>
                <span className="text-xs text-gray-400 truncate max-w-[100px]">
                  {assignment.User?.Username || assignment.User?.Email}
                </span>
              </div>

              {/* Date */}
              <div className="flex items-center gap-1.5 text-xs">
                <Clock className="h-3 w-3 text-gray-500" />
                <span className="text-gray-400 font-medium">
                  {format(deadline, "MMM d, yyyy")}
                </span>
              </div>
            </div>
          </div>

          <Button
            size="icon"
            variant="ghost"
            className="h-8 w-8 -mt-1 -mr-1 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg shrink-0"
            onClick={(e) => {
              e.stopPropagation()
              onCopy?.(assignment)
            }}
            disabled={disabled}
            title="Create personal copy"
          >
            <CopyPlus className="h-4 w-4" />
          </Button>
        </div>
      </GlassCard>
    </div>
  )
}
