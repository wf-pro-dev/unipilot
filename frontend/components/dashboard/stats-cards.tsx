"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Calendar, Clock, AlertTriangle, CheckCircle2 } from "lucide-react"
import { isToday, isWithinInterval, startOfWeek, endOfWeek } from "date-fns"
import { Assignment } from "@/types/models"
import { parseDeadline, isOverdue } from "@/lib/date-utils"
import { useAssignments } from "@/hooks/use-assignments"

export function StatsCards() {
  const { data: assignments = [] } = useAssignments()
  
  const todayAssignments = (assignments || []).filter((a) => isToday(parseDeadline(a.Deadline)))
  
  const thisWeekAssignments = (assignments || []).filter((a) => isWithinInterval(parseDeadline(a.Deadline), { start: startOfWeek(new Date()), end: endOfWeek(new Date()) }))
  const overdueAssignments = (assignments || []).filter((a) => isOverdue(parseDeadline(a.Deadline), a.StatusName))
  const completedAssignments = (assignments || []).filter((a) => a.StatusName === "Done")

  const stats = [
    {
      title: "Due Today",
      value: todayAssignments.length,
      change: `${todayAssignments.filter((a) => a.StatusName !== "Done").length} pending`,
      icon: Clock,
      color: "text-orange-400",
      bgColor: "bg-orange-500/10",
    },
    {
      title: "This Week", 
      value: thisWeekAssignments.length,
      change: `${thisWeekAssignments.filter((a) => a.StatusName !== "Done").length} pending`,
      icon: Calendar,
      color: "text-blue-400",
      bgColor: "bg-blue-500/10",
    },
    {
      title: "Overdue",
      value: overdueAssignments.length,
      change: "Need attention",
      icon: AlertTriangle,
      color: "text-red-400",
      bgColor: "bg-red-500/10",
    },
    {
      title: "Completed",
      value: completedAssignments.length,
      change: `+${completedAssignments.filter((a) => a.Course?.Semester === "SUMMER 2025").length} this semester`,
      icon: CheckCircle2,
      color: "text-green-400",
      bgColor: "bg-green-500/10",
    },
  ]

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat, index) => (
        <Card key={index} className="glass border-0">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-white">
              {stat.title}
            </CardTitle>
            <div className={`p-2 rounded-lg ${stat.bgColor}`}>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stat.value}</div>
            <p className="text-xs text-gray-400">
              {stat.change}
            </p>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
