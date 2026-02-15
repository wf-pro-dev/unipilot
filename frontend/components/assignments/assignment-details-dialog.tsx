"use client"

import { memo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import {
  Edit,
  Trash2,
  FileText,
  ExternalLink,
  Bot,
  Clock,
  CopyPlus,
  Calendar,
  Flag,
  Tag,
  CheckCircle2,
  Circle,
  AlertCircle
} from "lucide-react"
import { format } from "date-fns"
import { models } from "@/wailsjs/go/models"
import { parseDeadline, calculateDaysDifference, isOverdue, getDueDescription } from "@/lib/date-utils"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { StatusTag } from "@/components/assignments/tags/status-tag"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { TypeTag } from "./tags/type-tag"
import { useAssignment, useDeleteAssignment } from "@/hooks/use-assignments"
import { PriorityTag } from "./tags/priority-tag"
import { useRouter } from "next/navigation"
import FileUpload05 from "../file-upload-05"
import { cn } from "@/lib/utils"
import { EmptyState } from "../ui/empty-state"
import { Action } from "../core/action"
import { GlassCard } from "../ui/glass-card"

interface AssignmentDetailsDialogProps {
  isOpen: boolean
  onClose: () => void
  assignmentId: string
  assignmentRO?: models.Assignment
  handleEditOpen?: () => void
  onCopy?: (assignment: models.Assignment, includeDocuments: boolean) => void
  mode?: "default" | "readonly"
}

const BaseAssignmentDetailsDialog = ({
  isOpen,
  onClose,
  assignmentId,
  assignmentRO,
  handleEditOpen,
  onCopy,
  mode = "default",
}: AssignmentDetailsDialogProps) => {
  const { data: assignmentData } = useAssignment(assignmentId)
  const deleteMutation = useDeleteAssignment()

  var assignment = mode == "default" ? assignmentData as models.LocalAssignment : assignmentRO

  if (!assignment) return null

  const [includeDocuments, setIncludeDocuments] = useState(true)
  const [activeView, setActiveView] = useState("overview")

  const { Title, Deadline, Status, Priority, Type, Todo, Course, Link } = assignment

  const router = useRouter()

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(Deadline)
  const isOverdueStatus = isOverdue(deadline, Status)
  const daysUntilDue = calculateDaysDifference(deadline)

  const handleOpenLink = () => {
    BrowserOpenURL(Link)
  }

  // Status icon helper
  const getStatusIcon = () => {
    switch (Status) {
      case "Done":
        return <CheckCircle2 className="w-5 h-5 text-status-success" />
      case "In progress":
        return <Circle className="w-5 h-5 text-primary-blue-400" />
      case "To do":
        return <Circle className="w-5 h-5 text-text-caption" />
      default:
        return <Circle className="w-5 h-5 text-text-caption" />
    }
  }

  // Priority color helper
  const getPriorityColor = () => {
    switch (Priority) {
      case "High":
        return "text-status-danger"
      case "Medium":
        return "text-primary-yellow-400"
      case "Low":
        return "text-status-success"
      default:
        return "text-text-caption"
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-4xl p-0 gap-0 max-h-[90vh] flex flex-col md:flex-row">

        {/* === LEFT SIDEBAR: Assignment Identity === */}
        <div className="md:w-80 bg-white/5 border-r border-white/5 relative flex flex-col overflow-y-auto shrink-0">
          <div className="p-6 flex flex-col">

            {/* Header: Course & Title */}
            <div className="flex flex-col items-center mb-6 mt-2 text-center">
              <div className={cn(
                "w-20 h-20 rounded-2xl flex items-center justify-center shadow-2xl shadow-black/50 mb-5 ring-2 ring-white/10",
                Course?.Color
              )}>
                <FileText className="w-8 h-8 text-white" />
              </div>

              <div className="flex items-center gap-2 mb-2">
                <Badge variant="outline" className="text-caption border-white/20 bg-white/5 text-text-caption px-2 py-0.5 h-6">
                  {Course?.Code}
                </Badge>
              </div>

              <DialogTitle className="text-h3 text-text-title mb-1 leading-tight">
                {Title}
              </DialogTitle>

              <p className="text-caption text-text-caption font-medium">
                {Course?.Name}
              </p>
            </div>

            <Separator className="bg-white/10 mb-6" />

            {/* Deadline Section */}
            <div className="mb-6 bg-white/5 rounded-xl p-4 border border-white/5">
              <div className="flex items-center gap-2 mb-3">
                <Clock className="w-4 h-4 text-primary-blue-400" />
                <span className="text-caption text-text-caption uppercase tracking-wider font-semibold">Deadline</span>
              </div>

              <div className="space-y-2">
                <div className="flex items-baseline justify-between">
                  <p className="text-h5 font-bold text-white">{format(deadline, "MMM d, yyyy")}</p>
                  <Badge
                    variant="outline"
                    className={cn(
                      "border-white/10 bg-white/5 text-xs",
                      isOverdueStatus ? "text-status-danger" : daysUntilDue < 0 ? "text-text-caption" : "text-primary-yellow-400"
                    )}
                  >
                    {getDueDescription(deadline, Status)}
                  </Badge>
                </div>
                <p className="text-sm text-text-caption">{format(deadline, "EEEE 'at' h:mm a")}</p>
              </div>
            </div>

            {/* Status, Priority, Type Grid */}
            <div className="space-y-3 mb-6">
              {/* Status */}
              <div className="bg-white/5 border border-white/5 rounded-xl p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {getStatusIcon()}
                    <span className="text-sm font-medium text-text-caption uppercase tracking-wider">Status</span>
                  </div>
                  {mode === "default" ? (
                    <StatusTag assignment={assignment as models.LocalAssignment} variant="outline" />
                  ) : (
                    <Badge variant="outline" className="text-caption border-white/10 bg-white/5">
                      {Status}
                    </Badge>
                  )}
                </div>
              </div>

              {/* Priority */}
              <div className="bg-white/5 border border-white/5 rounded-xl p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Flag className={cn("w-5 h-5", getPriorityColor())} />
                    <span className="text-sm font-medium text-text-caption uppercase tracking-wider">Priority</span>
                  </div>
                  {mode === "default" ? (
                    <PriorityTag assignment={assignment as models.LocalAssignment} variant="outline" />
                  ) : (
                    <Badge variant="outline" className="text-caption border-white/10 bg-white/5">
                      {Priority}
                    </Badge>
                  )}
                </div>
              </div>

              {/* Type */}
              <div className="bg-white/5 border border-white/5 rounded-xl p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Tag className="w-5 h-5 text-text-caption" />
                    <span className="text-sm font-medium text-text-caption uppercase tracking-wider">Type</span>
                  </div>
                  {mode === "default" ? (
                    <TypeTag assignment={assignment as models.LocalAssignment} variant="outline" />
                  ) : (
                    <Badge variant="outline" className="text-caption border-white/10 bg-white/5">
                      {Type}
                    </Badge>
                  )}
                </div>
              </div>
            </div>

          </div>
        </div>

        {/* === RIGHT CONTENT AREA === */}
        <div className="flex-1 bg-bg-base/30">
          <Tabs value={activeView} onValueChange={setActiveView} className="flex flex-col h-full min-h-0">

            {/* Tabs Header */}
            <TabsList className="flex border-b border-white/5 px-6 pt-6 pb-0 shrink-0">
              <TabsTrigger
                value="overview"
                className="px-4 py-3 text-sm font-medium text-text-caption uppercase tracking-wider data-[state=active]:text-white data-[state=active]:border-b-2 data-[state=active]:border-primary-blue-400 transition-all duration-200"
              >
                Overview
              </TabsTrigger>
              <TabsTrigger
                value="documents"
                className="px-4 py-3 text-sm font-medium text-text-caption uppercase tracking-wider data-[state=active]:text-white data-[state=active]:border-b-2 data-[state=active]:border-primary-blue-400 transition-all duration-200"
              >
                Documents
              </TabsTrigger>
            </TabsList>

            {/* Tabs Content */}
            <div className="h-full overflow-y-auto p-6">

              {/* OVERVIEW TAB */}
              <TabsContent value="overview" className="mt-0 space-y-6 outline-none animate-in fade-in-50 duration-300">
                {/* Urgency Alert */}
                {isOverdueStatus && (
                  <div className="bg-status-danger/10 border border-status-danger/20 rounded-xl p-4">
                    <div className="flex items-start gap-3">
                      <AlertCircle className="w-5 h-5 text-status-danger mt-0.5 flex-shrink-0" />
                      <div>
                        <p className="text-sm font-medium text-status-danger mb-1">Assignment Overdue</p>
                        <p className="text-xs text-status-danger/80">
                          This assignment is past its deadline. Consider updating the status or extending the deadline.
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {/* Actions */}
                <div className="grid grid-cols-2 gap-3">
                  <Action
                    label="AI Help"
                    Icon={Bot}
                    onClick={() => router.push(`/chat?assignment=${assignmentId}`)}
                  />
                  <Action
                    label="Open Link"
                    Icon={ExternalLink}
                    onClick={() => handleOpenLink()}
                  />
                  <Action
                    label="Edit Assignment"
                    Icon={Edit}
                    onClick={() => handleEditOpen?.()}
                  />
                  <Action
                    label="Delete Assignment"
                    Icon={Trash2}
                    onClick={() => deleteMutation.mutate(assignment as models.LocalAssignment)}
                  />

                </div>

                {/* Quick Info Grid */}
                <div className="space-y-2">
                  <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider">
                    Quick Information
                  </h3>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="bg-white/5 border border-white/5 rounded-xl p-4 text-center">
                      <Calendar className="w-5 h-5 text-white mx-auto mb-2" />
                      <p className="text-xs text-text-caption uppercase mb-1">Created</p>
                      <p className="text-sm font-medium text-white">
                        {format(new Date(assignment.CreatedAt), "MMM d, yyyy")}
                      </p>
                    </div>
                    <div className="bg-white/5 border border-white/5 rounded-xl p-4 text-center">
                      <Clock className="w-5 h-5 text-primary-purple-400 mx-auto mb-2" />
                      <p className="text-xs text-text-caption uppercase mb-1">Last Updated</p>
                      <p className="text-sm font-medium text-white">
                        {format(new Date(assignment.UpdatedAt), "MMM d, yyyy")}
                      </p>
                    </div>

                    <div className="col-span-2">
                      {Todo && (
                        <GlassCard
                          className="p-4"
                          variant="board"
                        >
                          <p className="whitespace-pre-wrap leading-relaxed text-sm text-text-body">
                            {Todo}
                          </p>
                        </GlassCard>
                      )}
                      {!Todo && (
                        <div className="flex flex-col items-center justify-center text-center p-12 border border-white/10 rounded-xl bg-white/5">
                          <EmptyState
                            icon={FileText}
                            title="No description"
                            description="Add a description to keep track of important details and requirements."
                          />
                        </div>
                      )}
                    </div>
                  </div>
                </div>


                {/* Description Section */}


              </TabsContent>

              {/* DOCUMENTS TAB */}
              <TabsContent value="documents" className="mt-0 space-y-6 outline-none animate-in fade-in-50 duration-300">
                <FileUpload05
                  assignment={assignment as models.LocalAssignment}
                  mode={mode}
                  includeDocuments={includeDocuments}
                  setIncludeDocuments={setIncludeDocuments}
                />


              </TabsContent>

            </div>
          </Tabs >
        </div >

      </DialogContent >
    </Dialog >
  )
}

export const AssignmentDetailsDialog = memo(BaseAssignmentDetailsDialog, (prevProps, nextProps) => {
  return prevProps.assignmentId === nextProps.assignmentId &&
    prevProps.isOpen === nextProps.isOpen
})