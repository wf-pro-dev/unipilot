"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { AssignmentItem } from "./assignment-item"
import { CalendarDays, CheckCircle2, Loader2 } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { EmptyState } from "@/components/ui/empty-state"

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

export function AssignmentView({ 
  title, 
  assignments, 
  onToggleComplete,
  onAssignmentClick, 
  onDelete, 
  onEdit, 
  onOpenEdit, 
  isLoading }: AssignmentViewProps) {

  if (assignments.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh]">
        <EmptyState
          icon={CheckCircle2}
          title="No assignments found"
          description="You're all caught up! create a new assignment to get started tracking your work."
        />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <GlassCard className="p-4 border-white/5 bg-white/5 flex items-center justify-between shadow-lg shadow-black/20 backdrop-blur-xl rounded-xl">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-blue-500/10 rounded-lg border border-blue-500/10">
            <CalendarDays className="h-5 w-5 text-blue-400" />
          </div>
          <h2 className="text-h3 text-white tracking-tight">{title}</h2>
        </div>
        {isLoading && (
          <div className="flex items-center gap-2 text-gray-400 px-3 py-1.5 rounded-lg bg-white/5 border border-white/5">
            <Loader2 className="h-3.5 w-3.5 animate-spin text-blue-400" />
            <span className="text-[10px] uppercase tracking-wider font-medium">Syncing</span>
          </div>
        )}
      </GlassCard>


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
