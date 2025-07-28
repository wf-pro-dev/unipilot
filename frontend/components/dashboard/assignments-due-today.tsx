"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Clock, ChevronRight } from "lucide-react"
import { isToday } from "date-fns"
import { Assignment } from "@/types/models"
import { parseDeadline } from "@/lib/date-utils"
import { useAssignments } from "@/hooks/use-assignments"
import Link from "next/link"

export function AssignmentsDueToday() {
  const { data: assignments = [] } = useAssignments()
  
  const todayAssignments = (assignments || []).filter((a) => isToday(parseDeadline(a.Deadline)))
  const pendingCount = todayAssignments.filter((a) => a.StatusName !== "Done").length

  return (
    <Card className="glass border-0">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-white">Due Today</CardTitle>
        <Clock className="h-4 w-4 text-blue-400" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold text-white">{pendingCount}</div>
        <p className="text-xs text-gray-400">
          {pendingCount === 1 ? "assignment" : "assignments"} pending
        </p>
        {todayAssignments.length > 0 && (
          <div className="mt-4 space-y-2">
            {todayAssignments.slice(0, 2).map((assignment) => (
              <div key={assignment.ID} className="flex items-center justify-between text-sm">
                <span className="text-gray-300 truncate">{assignment.Title}</span>
                <Badge 
                  variant={assignment.StatusName === "Done" ? "default" : "secondary"}
                  className="text-xs"
                >
                  {assignment.StatusName}
                </Badge>
              </div>
            ))}
            {todayAssignments.length > 2 && (
              <Link href="/assignments?view=today" className="flex items-center text-xs text-blue-400 hover:text-blue-300">
                <span>View all {todayAssignments.length} assignments</span>
                <ChevronRight className="h-3 w-3 ml-1" />
              </Link>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
