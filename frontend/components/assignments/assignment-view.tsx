"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { AssignmentItem } from "./assignment-item"
import { CalendarDays, CheckCircle2, Loader2 } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"

interface AssignmentViewProps {
  title: string
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  onOpenEdit: (assignment: assignment.LocalAssignment) => void
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  isLoading?: boolean
}

export function AssignmentView({ title, assignments, onToggleComplete, onAssignmentClick, onDelete, onEdit, onOpenEdit, isLoading }: AssignmentViewProps) {

  if (assignments.length === 0) {
    return (
      <div className="py-32 text-center">
        <CheckCircle2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-white mb-2">No assignments found</h3>
        <p className="text-gray-400">Create an assignment to get started</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <CalendarDays className="h-5 w-5" />
            <span className="text-lg font-medium">{title}</span>
            {isLoading && <Loader2 className="h-4 w-4 animate-spin ml-2" />}
          </CardTitle>
        </CardHeader>
      </Card>


      <div className="grid md:grid-cols-1 lg:grid-cols-2 gap-4">
        {(assignments || []).map((assignment, index) => (
          <AssignmentItem
            key={assignment.ID}
            assignment={assignment}
            onToggleComplete={onToggleComplete}
            onEdit={onEdit}
            onDelete={onDelete}
            onAssignmentClick={onAssignmentClick}
            disabled={isLoading}
            onOpenEdit={onOpenEdit}
          />
        ))}
      </div>


    </div>
  )
}
