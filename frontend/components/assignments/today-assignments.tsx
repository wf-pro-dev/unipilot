"use client"

import { Assignment } from "@/types/models"
import { AssignmentView } from "./assignment-view"

interface TodayAssignmentsProps {
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading?: boolean
}

export function TodayAssignments({ assignments, onToggleComplete, onAssignmentClick, isLoading }: TodayAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const todayAssignments = assignments || []

  return (
    <AssignmentView 
      title="Due Today" 
      assignments={todayAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      isLoading={isLoading}
    />
  )
}
