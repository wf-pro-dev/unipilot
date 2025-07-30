"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Clock, AlertTriangle } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { useAssignments } from "@/hooks/use-assignments"
import { formatDeadline } from "@/lib/date-utils"

interface UpcomingDeadlinesProps {
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
}

export function UpcomingDeadlines({ onAssignmentClick }: UpcomingDeadlinesProps) {
  const { data: assignments = [] } = useAssignments()
  
  return (
    <Card className="glass border-0">
      <CardHeader>
        <CardTitle className="text-sm font-medium text-white flex items-center space-x-2">
          <Clock className="h-4 w-4 text-orange-400" />
          <span>Upcoming Deadlines</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {(assignments || []).slice(0, 5).map((assignment) => (
            <div key={assignment.ID} onClick={() => onAssignmentClick(assignment)} className="flex items-center space-x-3 p-3 rounded-lg glass-dark">
              <div className={`w-2 h-2 rounded-full ${assignment.Course?.Color}`} />
              <div className="flex-1">
                <p className="text-sm font-medium text-white">{assignment.Title}</p>
                <p className="text-xs text-gray-400">{assignment.Course?.Code}</p>
              </div>
              <div className="text-right">
                <p className="text-xs text-gray-300">{formatDeadline(assignment.Deadline)}</p>
                {assignment.Priority === "high" && <AlertTriangle className="h-3 w-3 text-red-400 ml-auto mt-1" />}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
