"use client"

import { useDrag } from "react-dnd"
import { Card, CardContent } from "@/components/ui/card"
import { assignment } from "@/wailsjs/go/models"
import { parseDeadline, calculateDaysDifference } from "@/lib/date-utils"
import { StatusTag } from "../utils/status-tag"
import { format } from "date-fns"

interface CalendarItemProps {
  assignment: assignment.LocalAssignment
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
}

export function CalendarItem({ assignment, onEdit, onAssignmentClick }: CalendarItemProps) {
  
  
  const [{ isDragging }, drag] = useDrag({
    type: "assignment",
    item: { assignment: assignment },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  })

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(assignment.Deadline)

  return (
    <div ref={drag} className={`cursor-move ${isDragging ? "opacity-50" : ""}`} onClick={() => onAssignmentClick(assignment)}>
      <Card className="glass-dark border-0 hover:glass-hover transition-all duration-300 mb-2">
        <CardContent className="p-3">
          <div className="flex flex-col space-y-2">

            <h4 className={`text-sm font-medium truncate ${assignment.StatusName === "Done" ? "line-through text-gray-500" : "text-white"}`}>
              {assignment.Title}
            </h4>

            <div className="flex items-center space-x-2 mt-1">
              <div className={`w-2 h-2 rounded-full ${assignment.Course?.Color}`} />
              <span className="text-xs text-gray-400 truncate">{assignment.Course?.Code}</span>
            </div>

            <span className="text-xs text-gray-400 self-start">
              {format(deadline, "h:mm a")}
            </span>

            <StatusTag assignment={assignment} onEdit={onEdit} />

          </div>
        </CardContent>
      </Card>
    </div>
  )
}
