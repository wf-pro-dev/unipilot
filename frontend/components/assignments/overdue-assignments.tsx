"use client"

import { Assignment } from "@/types/models"
import { AssignmentView } from "./assignment-view"

interface OverdueAssignmentsProps {
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading?: boolean
}

export function OverdueAssignments({ assignments, onToggleComplete, onAssignmentClick, isLoading }: OverdueAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const overdueAssignments = assignments || []

  return (
    <AssignmentView 
      title="Overdue" 
      assignments={overdueAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      isLoading={isLoading}
    />
  )
}
