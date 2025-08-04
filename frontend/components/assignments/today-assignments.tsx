"use client"

import { assignment } from "@/wailsjs/go/models"
import { AssignmentView } from "./assignment-view"

interface TodayAssignmentsProps {
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  isLoading?: boolean
}

export function TodayAssignments({ assignments, onToggleComplete, onAssignmentClick, onEdit, onDelete, isLoading }: TodayAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const todayAssignments = assignments || []

  return (
    <AssignmentView 
      title="Due Today" 
      assignments={todayAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      onEdit={onEdit}
      onDelete={onDelete}
      isLoading={isLoading}
    />
  )
}
