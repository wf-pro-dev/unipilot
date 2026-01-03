"use client"

import { GlassCard } from "@/components/ui/glass-card"
import { AssignmentItem } from "./assignment-item"
import { CalendarDays, CheckCircle2, Loader2 } from "lucide-react"
import { EmptyState } from "@/components/ui/empty-state"
import { models } from "@/wailsjs/go/models"

interface AssignmentViewProps {
  title: string
  assignments: models.LocalAssignment[]
  onAssignmentClick: (assignment: models.LocalAssignment) => void
  onOpenEdit: (assignment: models.LocalAssignment) => void
  onEdit: (assignment: models.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: models.LocalAssignment) => void
  onEmptyClick: () => void
  isLoading?: boolean
}

export function AssignmentView({ 
  title, 
  assignments, 
  onAssignmentClick, 
  onDelete, 
  onEdit, 
  onOpenEdit, 
  onEmptyClick,
  isLoading }: AssignmentViewProps) {

  if (assignments.length === 0) {
    return (
      <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
        <EmptyState
          icon={CheckCircle2}
          title="No assignments found"
          description="You're all caught up! create a new assignment to get started tracking your work."
          className="flex-1 items-center"
          buttonText="Create Assignment"
          onClick={onEmptyClick}
        />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <GlassCard variant="board" className="p-4 flex items-center justify-between">
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


      <div className="grid md:grid-cols-1 lg:grid-cols-3 gap-4">
        {(assignments || []).map((assignment, index) => (
          <AssignmentItem
            key={assignment.ID}
            assignment={assignment}
            onEdit={onEdit}
            onDelete={onDelete}
            onAssignmentClick={onAssignmentClick}
            disabled={isLoading}
            onOpenEdit={onOpenEdit}
            variant="outline"
          />
        ))}
      </div>


    </div>
  )
}
