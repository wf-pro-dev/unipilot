"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Calendar, Clock, AlertTriangle, CheckCircle2 } from "lucide-react"
import { useAssignments, useCompletedAssignments, useOverdueAssignments, useTodayAssignments, useWeekAssignments } from "@/hooks/use-assignments"
import Link from "next/link"

export function StatsCards() {
  const { data: assignments = [] } = useAssignments()
  const { data: todayAssignments } = useTodayAssignments()
  const { data: weekAssignments } = useWeekAssignments()
  const { data: overdueAssignments } = useOverdueAssignments()
  const { data: completedAssignments } = useCompletedAssignments()
 

  const stats = [
    {
      title: "Due Today",
      value: todayAssignments.length,
      change: `${todayAssignments.filter((a) => a.StatusName !== "Done").length} pending`,
      icon: Clock,
      color: "text-orange-400",
      bgColor: "bg-orange-500/10",
      link: "/assignments?view=today"
    },
    {
      title: "This Week", 
      value: weekAssignments.length,
      change: `${weekAssignments.filter((a) => a.StatusName !== "Done").length} pending`,
      icon: Calendar,
      color: "text-blue-400",
      bgColor: "bg-blue-500/10",
      link: "/assignments?view=week"
    },
    {
      title: "Overdue",
      value: overdueAssignments.length,
      change: "Need attention",
      icon: AlertTriangle,
      color: "text-red-400",
      bgColor: "bg-red-500/10",
      link: "/assignments?view=overdue"
    },
    {
      title: "Completed",
      value: completedAssignments.length,
      change: `+${completedAssignments.filter((a) => a.Course?.Semester === "SUMMER 2025").length} this semester`,
      icon: CheckCircle2,
      color: "text-green-400",
      bgColor: "bg-green-500/10",
      link: "/assignments?view=list&status=Done"
    },
  ]

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat, index) => (
        <Link href={stat.link} key={index}>
        <Card className="bg-white/5 border-white/5 hover:bg-white/10 hover:scale-[1.02] hover:shadow-xl shadow-lg shadow-black/20 transition-all duration-300 backdrop-blur-xl">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <CardTitle className="text-h5">
              {stat.title}
            </CardTitle>
            <div className={`p-2 rounded-xl ${stat.bgColor} border border-white/5`}>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white mb-1 tracking-tight">{stat.value}</div>
            <p className="text-caption">
              {stat.change}
            </p>
          </CardContent>
        </Card>
        </Link>
      ))}
    </div>
  )
}
