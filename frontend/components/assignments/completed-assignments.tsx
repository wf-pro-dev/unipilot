"use client"

import { Assignment } from "@/types/models"
import { AssignmentView } from "./assignment-view"

interface CompletedAssignmentsProps {
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading?: boolean
}

export function CompletedAssignments({ assignments, onToggleComplete, onAssignmentClick, isLoading }: CompletedAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const completedAssignments = assignments || []

  return (
    <AssignmentView 
      title="Completed Assignments" 
      assignments={completedAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      isLoading={isLoading}
    />
  )
}
