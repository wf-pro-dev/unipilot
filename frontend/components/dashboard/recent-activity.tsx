"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Activity, CheckCircle, FileText, BookOpen } from "lucide-react"

const activities = [
  {
    id: 1,
    type: "assignment",
    title: "Completed Homework #3",
    course: "CS 201",
    time: "2 hours ago",
    icon: CheckCircle,
    color: "text-green-400",
  },
  {
    id: 2,
    type: "note",
    title: "Created study notes",
    course: "MATH 301",
    time: "4 hours ago",
    icon: FileText,
    color: "text-blue-400",
  },
  {
    id: 3,
    type: "course",
    title: "Added new course",
    course: "AI 401",
    time: "1 day ago",
    icon: BookOpen,
    color: "text-purple-400",
  },
]

export function RecentActivity() {
  return (
    <Card className="glass border-0">
      <CardHeader>
        <CardTitle className="text-sm font-medium text-white flex items-center space-x-2">
          <Activity className="h-4 w-4 text-green-400" />
          <span>Recent Activity</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {activities.map((activity) => {
            const Icon = activity.icon
            return (
              <div key={activity.id} className="flex items-center space-x-3 p-3 rounded-lg glass-dark">
                <Icon className={`h-4 w-4 ${activity.color}`} />
                <div className="flex-1">
                  <p className="text-sm font-medium text-white">{activity.title}</p>
                  <p className="text-xs text-gray-400">{activity.course}</p>
                </div>
                <p className="text-xs text-gray-400">{activity.time}</p>
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
