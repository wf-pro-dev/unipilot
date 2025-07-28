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
import { Assignment } from "@/types/models"

export default function DashboardPage() {
  const [selectedAssignment, setSelectedAssignment] = useState<Assignment | null>(null)

  const handleToggleComplete = (id: number) => {
    console.log("Toggling completion for assignment:", id)
  }

  const handleEditAssignment = (assignment: Assignment) => {
    console.log("Editing assignment:", assignment)
  }

  const handleDeleteAssignment = (id: number) => {
    console.log("Deleting assignment:", id)
  }

  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="relative z-10">
        <WelcomeSection />

        <div className="mt-8">
          <StatsCards />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-8">
          <div className="lg:col-span-2 space-y-6">
            <AssignmentsDueToday onAssignmentClick={setSelectedAssignment} />
            <AssignmentsThisWeek onAssignmentClick={setSelectedAssignment}  />
            <RecentActivity />
          </div>

          <div className="space-y-6">
            <CoursesCard />
            <UpcomingDeadlines onAssignmentClick={setSelectedAssignment} />
          </div>
        </div>
      </div>
      <AssignmentDetailsModal
          isOpen={!!selectedAssignment}
          onClose={() => setSelectedAssignment(null)}
          assignment={selectedAssignment}
          onEdit={handleEditAssignment}
          onDelete={handleDeleteAssignment}
          onToggleComplete={handleToggleComplete}
        />
    </div>
  )
}
