"use client"

import { assignment } from "@/wailsjs/go/models"
import { AssignmentView } from "./assignment-view"

interface WeekAssignmentsProps {
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void 
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  isLoading?: boolean
}

export function WeekAssignments({ assignments, onToggleComplete, onAssignmentClick, onEdit, onDelete, isLoading }: WeekAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const weekAssignments = assignments || []

  return (
    <AssignmentView 
      title="Due This Week" 
      assignments={weekAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      onEdit={onEdit}
      onDelete={onDelete}
      isLoading={isLoading}
    />
  )
}
