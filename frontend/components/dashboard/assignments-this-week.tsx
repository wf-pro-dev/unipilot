"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { CalendarDays } from "lucide-react"
import { isWithinInterval, startOfWeek, endOfWeek } from "date-fns"
import { Assignment } from "@/types/models"
import { parseDeadline, isOverdue } from "@/lib/date-utils"
import { useAssignments } from "@/hooks/use-assignments"

export function AssignmentsThisWeek() {
  const { data: assignments = [] } = useAssignments()
  
  const upcomingAssignments = (assignments || []).filter((a) => isWithinInterval(parseDeadline(a.Deadline), { start: startOfWeek(new Date()), end: endOfWeek(new Date()) }))
  
  const stats = {
    completed: upcomingAssignments.filter((a) => a.StatusName === "Done").length,
    pending: upcomingAssignments.filter((a) => a.StatusName !== "Done").length,
    overdue: upcomingAssignments.filter((a) => isOverdue(parseDeadline(a.Deadline), a.StatusName)).length,
  }

  return (
    <Card className="glass border-0">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-white">This Week</CardTitle>
        <CalendarDays className="h-4 w-4 text-blue-400" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold text-white">{upcomingAssignments.length}</div>
        <p className="text-xs text-gray-400">
          {stats.pending} pending, {stats.completed} completed
        </p>
        {stats.overdue > 0 && (
          <p className="text-xs text-red-400 mt-1">
            {stats.overdue} overdue
          </p>
        )}
      </CardContent>
    </Card>
  )
}
