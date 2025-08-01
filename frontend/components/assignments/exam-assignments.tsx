"use client"

import { assignment } from "@/wailsjs/go/models"
import { AssignmentView } from "./assignment-view"

interface ExamAssignmentsProps {
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  isLoading?: boolean
}

export function ExamAssignments({ assignments, onToggleComplete, onAssignmentClick, onEdit, onDelete, isLoading }: ExamAssignmentsProps) {
  // No need to filter here - the hook already provides filtered data
  const examAssignments = assignments || []

  return (
    <AssignmentView 
      title="Exams" 
      assignments={examAssignments} 
      onToggleComplete={onToggleComplete} 
      onAssignmentClick={onAssignmentClick}
      onEdit={onEdit}
      onDelete={onDelete}
      isLoading={isLoading}
    />
  )
}