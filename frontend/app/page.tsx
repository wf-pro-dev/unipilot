"use client"

import { useState } from "react"
import { WelcomeSection } from "@/components/dashboard/welcome-section"
import { StatsCards } from "@/components/dashboard/stats-cards"
import { CoursesCard } from "@/components/dashboard/courses-card"
import { UpcomingDeadlines } from "@/components/dashboard/upcoming-deadlines"
import { RecentActivity } from "@/components/dashboard/recent-activity"
import { AssignmentsDueToday } from "@/components/dashboard/assignments-due-today"
import { AssignmentsThisWeek } from "@/components/dashboard/assignments-this-week"
import { AssignmentDetailsModal } from "@/components/assignments/assignment-details-modal"
import { OfflineBanner } from "@/components/ui/offline-banner"
import { assignment, user } from "@/wailsjs/go/models"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { useUpdateAssignment } from "@/hooks/use-assignments"
import { Dashboard } from "@/components/dashboard/dashboard"

export default function DashboardPage() {
  const [selectedAssignment, setSelectedAssignment] = useState<assignment.LocalAssignment | null>(null)

  const updateMutation = useUpdateAssignment()

  const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
    console.log("Editing assignment:", assignment)
    const message = "assignment " + assignment.ID + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      assignment,
      column,
      value
    })
  }

  const handleDeleteAssignment = (assignment: assignment.LocalAssignment) => {
    console.log("Deleting assignment:", assignment.ID)
  }

  return (
    <div className=" ">

      {/* Floating background elements */}
     

      <Dashboard />

    </div>
  )
}
