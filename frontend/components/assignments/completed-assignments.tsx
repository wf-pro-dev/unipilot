"use client"

import { assignment } from "@/wailsjs/go/models"
import { AssignmentView } from "./assignment-view"

interface CompletedAssignmentsProps {
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
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
