"use client"

import { assignment } from "@/wailsjs/go/models"
import { AssignmentView } from "./assignment-view"

interface OverdueAssignmentsProps {
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  isLoading?: boolean
}

export function OverdueAssignments({ assignments, onToggleComplete, onAssignmentClick, onDelete, isLoading }: OverdueAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const overdueAssignments = assignments || []

  return (
    <AssignmentView 
      title="Overdue" 
      assignments={overdueAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      onDelete={onDelete}
      isLoading={isLoading}
    />
  )
}
