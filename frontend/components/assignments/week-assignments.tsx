"use client"

import { Assignment } from "@/types/models"
import { AssignmentView } from "./assignment-view"

interface WeekAssignmentsProps {
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading?: boolean
}

export function WeekAssignments({ assignments, onToggleComplete, onAssignmentClick, isLoading }: WeekAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const weekAssignments = assignments || []

  return (
    <AssignmentView 
      title="Due This Week" 
      assignments={weekAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      isLoading={isLoading}
    />
  )
}
