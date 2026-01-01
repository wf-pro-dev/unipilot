"use client"

import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Calendar, Edit, Trash2, FileText, ExternalLink, Info, Bot, Clock, Link as LinkIcon, CopyPlus } from "lucide-react"
import { format } from "date-fns"
import { models as goModels } from "@/wailsjs/go/models"
import { parseDeadline, calculateDaysDifference, isOverdue, getDueDescription } from "@/lib/date-utils"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { StatusTag } from "@/components/assignments/tags/status-tag"
import { AssignmentDocuments } from "./assignment-documents"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { Separator } from "../ui/separator"
import { TypeTag } from "./tags/type-tag"
import { useAssignments } from "@/hooks/use-assignments"
import { PriorityTag } from "./tags/priority-tag"
import { useRouter } from "next/navigation"

interface AssignmentDetailsModalProps {
  isOpen: boolean
  onClose: () => void
  assignment_id: number | undefined
  assignmentProp?: goModels.LocalAssignment | null
  onOpenEdit?: (assignment: goModels.LocalAssignment) => void
  onEdit?: (assignment: goModels.LocalAssignment, column: string, value: string) => void
  onDelete?: (assignment: goModels.LocalAssignment) => void
  onCopy?: (assignment: goModels.LocalAssignment) => void
  isLoading?: boolean
}

export function AssignmentDetailsModal({
  isOpen,
  onClose,
  assignment_id,
  assignmentProp,
  onOpenEdit,
  onEdit,
  onDelete,
  onCopy,
}: AssignmentDetailsModalProps) {

  if (!assignment_id) return null

  const router = useRouter()

  const [activeView, setActiveView] = useState("info")

  const { data: assignments } = useAssignments()
  const assignment = assignmentProp || assignments?.find((a) => a.ID === assignment_id)
  if (!assignment) return null

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(assignment.Deadline)

  const isOverdueStatus = isOverdue(deadline, assignment.StatusName)
  const daysUntilDue = calculateDaysDifference(deadline)
  const handleOpenLink = () => {
    BrowserOpenURL(assignment.Link)
  }


  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-xl max-h-[95vh] overflow-y-auto p-0 overflow-hidden gap-0">
        <div className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <div className="flex justify-between items-start">
            <div className="space-y-1">
              <div className="flex items-center space-x-2 mb-1">
                <div className={`w-2 h-2 rounded-full ${assignment.Course?.Color}`} />
                <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">{assignment.Course?.Name}</span>
              </div>
              <DialogTitle className="text-xl font-semibold text-white leading-tight">
                {assignment.Title}
              </DialogTitle>
            </div>
          </div>
        </div>


        <div className="p-6">

          <Tabs value={activeView} onValueChange={setActiveView} className="w-full space-y-6">

            <TabsList className="flex flex-row bg-white/5 p-1 rounded-xl w-full border border-white/5">
              <TabsTrigger
                value="info"
                className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
              >
                <Info className="w-4 h-4" />
                <span className="text-sm font-medium">Information</span>
              </TabsTrigger>
              <TabsTrigger
                value="documents"
                className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
              >
                <FileText className="w-4 h-4" />
                <span className="text-sm font-medium">Documents</span>
              </TabsTrigger>
            </TabsList>


            {/* Status and Priority */}
            <TabsContent value="info" className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-300">

              {/* Deadline Section */}
              <div className="p-4 rounded-xl bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 group">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                    <Clock className="w-3.5 h-3.5 text-blue-400" />
                    <span>Deadline</span>
                  </div>
                  <Badge variant="outline" className={`border-white/10 bg-white/5 ${isOverdueStatus ? "text-red-400" : daysUntilDue < 0 ? "text-gray-400" : "text-yellow-400"}`}>
                    {getDueDescription(deadline, assignment.StatusName)}
                  </Badge>
                </div>
                <p className="font-medium text-white text-lg">{format(deadline, "EEEE, MMMM d, yyyy")}</p>
                <p className="text-gray-400 text-sm mt-1">{format(deadline, "h:mm a")}</p>

              </div>

              {/* Tags Grid */}
              <div className="grid grid-cols-3 gap-3">
                <div className="flex flex-col items-center space-y-2 border border-white/5 p-3 rounded-xl bg-white/5">
                  <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Status</span>
                  <StatusTag assignment={assignment} onEdit={onEdit!} />
                </div>

                <div className="flex flex-col items-center space-y-2 border border-white/5 p-3 rounded-xl bg-white/5">
                  <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Priority</span>
                  <PriorityTag assignment={assignment} onEdit={onEdit!} />
                </div>

                <div className="flex flex-col items-center space-y-2 border border-white/5 p-3 rounded-xl bg-white/5">
                  <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">Type</span>
                  <TypeTag assignment={assignment} onEdit={onEdit!} />
                </div>
              </div>

              {/* Description */}
              {assignment.Todo && (
                <div className="space-y-2">
                  <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                    <FileText className="w-3.5 h-3.5" />
                    <span>Description & Notes</span>
                  </div>
                  <div className="bg-white/5 border border-white/5 p-4 rounded-xl  overflow-y-auto custom-scrollbar max-h-[100px] hover:max-h-[200px] transition-all duration-300 ease-in-out">
                    <p className="whitespace-pre-wrap leading-relaxed text-sm text-gray-200">{assignment.Todo}</p>
                  </div>
                </div>
              )}
            </TabsContent>

            <TabsContent value="documents" className="animate-in fade-in slide-in-from-bottom-4 duration-300">
              <AssignmentDocuments assignment={assignment} documents={assignment.Documents} viewMode={!!assignmentProp} />
            </TabsContent>

          </Tabs>

          {/* Actions */}

          <div className="grid grid-cols-2 gap-4 mt-6">

            {!assignmentProp && (

              <Button
                variant="outline"
                size="sm"
                onClick={() => router.push(`/chat?assignment=${assignment.ID}`)}
                className="bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10"
              >
                <Bot className="h-4 w-4 mr-2" />
                <span className="text-sm">AI Help</span>
              </Button>

            )}


            {onOpenEdit && (
              <Button
                variant="outline"
                size="sm"
                className="bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10"
                onClick={(e) => {
                  e.stopPropagation()
                  onOpenEdit(assignment)
                }}
              >
                <Edit className="mr-2 w-4 h-4" />
                Edit
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              className="bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10"
              onClick={(e) => {
                e.stopPropagation()
                handleOpenLink()
              }}
            >
              <ExternalLink className="mr-2 w-4 h-4" />
              Open Link
            </Button>

            {onCopy && (
              <Button
                variant="outline"
                size="sm"
                className="text-blue-400 bg-blue-500/5 border-blue-500/20 hover:bg-blue-500/10 hover:text-blue-300 h-10"
                onClick={(e) => {
                  e.stopPropagation()
                  onCopy?.(assignment)
                }}
              >
                <CopyPlus className="h-4 w-4 mr-2" />
                <span className="text-sm">Copy</span>
              </Button>
            )}
            {onDelete && (
              <Button
                variant="outline"
                size="sm"
                className="text-red-400 bg-red-500/5 border-red-500/20 hover:bg-red-500/10 hover:text-red-300 h-10"
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete(assignment)
                }}
              >
                <Trash2 className="mr-2 w-4 h-4" />
                Delete
              </Button>
            )}
          </div>
        </div>

      </DialogContent>
    </Dialog>
  )
}
