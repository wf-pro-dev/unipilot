"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { AssignmentItem } from "./assignment-item"
import { CalendarDays, CheckCircle2, Loader2 } from "lucide-react"
import { Assignment } from "@/types/models"

interface AssignmentViewProps {
  title: string
  assignments: Assignment[]
  onToggleComplete: (assignment: Assignment) => void
  onAssignmentClick: (assignment: Assignment) => void
  isLoading?: boolean
}

export function AssignmentView({ title, assignments, onToggleComplete, onAssignmentClick, isLoading }: AssignmentViewProps) {
  const handleEdit = (assignment: Assignment) => {
    console.log("Editing assignment:", assignment)
  }

  const handleDelete = (id: number) => {
    console.log("Deleting assignment:", id)
  }

  return (
    <div className="space-y-6">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <CalendarDays className="h-5 w-5" />
            <span>{title}</span>
            {isLoading && <Loader2 className="h-4 w-4 animate-spin ml-2" />}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {(assignments || []).length === 0 ? (
            <div className="flex items-center justify-center h-32 text-gray-400">
              <div className="text-center">
                <CheckCircle2 className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>No assignments {title.toLowerCase()}</p>
              </div>
            </div>
          ) : (
            <div className="grid md:grid-cols-1 lg:grid-cols-2 gap-4">
              {(assignments || []).map((assignment, index) => (
                <AssignmentItem
                  key={assignment.ID}
                  assignment={assignment}
                  onToggleComplete={onToggleComplete}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  onAssignmentClick={onAssignmentClick}
                  disabled={isLoading}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
