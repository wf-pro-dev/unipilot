"use client"

import { Assignment } from "@/types/models"
import { AssignmentView } from "./assignment-view"

interface ExamAssignmentsProps {
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading?: boolean
}

export function ExamAssignments({ assignments, onToggleComplete, onAssignmentClick, isLoading }: ExamAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const examAssignments = assignments || []

  return (
    <AssignmentView 
      title="Exams" 
      assignments={examAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      isLoading={isLoading}
    />
  )
}