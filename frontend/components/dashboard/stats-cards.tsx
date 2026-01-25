"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Calendar, Clock, AlertTriangle, CheckCircle2 } from "lucide-react"
import { useAssignments, useCompletedAssignments, useExams, useOverdueAssignments, useTodayAssignments, useWeekAssignments } from "@/hooks/use-assignments"
import Link from "next/link"
import { GlassCard } from "../ui/glass-card"

export function StatsCards() {
  const { data: exams } = useExams()
  const { data: weekAssignments } = useWeekAssignments()
  const { data: overdueAssignments } = useOverdueAssignments()
  const { data: completedAssignments } = useCompletedAssignments()


  const stats = [
    {
      title: "This Week",
      value: weekAssignments.length,
      change: `${weekAssignments.filter((a) => a.Status !== "Done").length} pending`,
      icon: Calendar,
      link: "/assignments?view=week"
    },
    {
      title: "Exams",
      value: exams.filter((a) => a.Status !== "Done").length,
      change: "Upcoming",
      icon: Clock,
      link: "/assignments?view=exam"
    },
    {
      title: "Overdue",
      value: overdueAssignments.length,
      change: "Need attention",
      icon: AlertTriangle,
      link: "/assignments?view=overdue"
    },
    {
      title: "Completed",
      value: completedAssignments.length,
      change: `+ ${completedAssignments.filter((a) => a.Course?.Semester === "SUMMER 2025").length} this semester`,
      icon: CheckCircle2,
      link: "/assignments?view=list&status=Done"
    },
  ]

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat, index) => (
        <Link href={stat.link} key={index}>
          <GlassCard variant="outline">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
              <CardTitle className="text-h5 font-medium uppercase tracking-wider">
                {stat.title}
              </CardTitle>
              <div className={`p-2 rounded-xl bg-primary-blue-500/30 border border-primary-blue-400/50`}>
                <stat.icon className={`h-4 w-4 text-primary-blue-400`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white mb-1 tracking-tight">{stat.value}</div>
              <p className="text-caption text-gray-400">
                {stat.change}
              </p>
            </CardContent>
          </GlassCard>
        </Link>
      ))}
    </div>
  )
}
