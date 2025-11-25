"use client"

import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Calendar, Edit, Trash2, FileText, ExternalLink, Info, Bot } from "lucide-react"
import { format } from "date-fns"
import { assignment } from "@/wailsjs/go/models"
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
  onOpenEdit: (assignment: assignment.LocalAssignment) => void
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  isLoading?: boolean
}

export function AssignmentDetailsModal({
  isOpen,
  onClose,
  assignment_id,
  onOpenEdit,
  onEdit,
  onDelete,
}: AssignmentDetailsModalProps) {

  if (!assignment_id) return null

  const router = useRouter()

  const [activeView, setActiveView] = useState("info")

  const { data: assignments } = useAssignments()
  const assignment = assignments?.find((a) => a.ID === assignment_id)
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
      <DialogContent className="glass border-0 text-white max-w-xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex justify-between items-start">
            <DialogTitle className="text-xl font-semibold text-white">
              {assignment.Title}
            </DialogTitle>
          </div>
        </DialogHeader>


        <div className="relative z-10 space-y-4" >

          <Tabs value={activeView} onValueChange={setActiveView} className="w-full space-y-4">

            <TabsList className="flex flex-row border-0 w-full">
              <TabsTrigger value="info" className="flex w-full justify-center items-center space-y-1 space-x-1 py-2 text-gray-400 hover:text-white data-[state=active]:text-white">
                <Info className="w-4 h-4" />
                <span className="text-sm">Information</span>
              </TabsTrigger>
              <Separator orientation="vertical" className="bg-gray-700 W-1 h-full" />
              <TabsTrigger value="documents" className="flex w-full justify-center items-center space-y-1 space-x-1 py-2 text-gray-400 hover:text-white data-[state=active]:text-white">
                <FileText className="w-4 h-4" />
                <span className="text-sm">Documents</span>
              </TabsTrigger>
            </TabsList>

            <Separator className="bg-gray-700" />


            {/* Status and Priority */}
            <TabsContent value="info" className="space-y-4">

              {/* Course and Type */}
              <div className="flex flex-col space-y-2">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <div className={`w-2 h-2 rounded-full ${assignment.Course?.Color}`} />
                      <span>Course</span>
                    </div>
                    <div className="bg-gray-800/50 border border-gray-600 p-3 rounded-lg">
                      <p className="font-medium text-white text-sm">{assignment.Course?.Name}</p>
                    </div>
                  </div>
                  <div className="space-y-2">
                    <div className="flex items-center space-x-4 text-sm">
                      <div className="flex items-center space-x-2 text-gray-400">
                        <Calendar className="w-4 h-4" />
                        <span>Deadline</span>
                      </div>
                      <Badge variant="outline" className={`${isOverdueStatus ? "text-red-400" : daysUntilDue < 0 ? "text-gray-400" : "text-yellow-400"}`}>
                        {getDueDescription(deadline, assignment.StatusName)}
                      </Badge>
                    </div>
                    <div className="bg-gray-800/50 border border-gray-600 p-3 rounded-lg">
                      <p className="font-medium text-white text-sm">{format(deadline, "EEEE, MMMM d, yyyy")}</p>
                    </div>
                  </div>

                </div>

                {/* Description */}
                {assignment.Todo && (
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <div className="flex items-center space-x-2 text-sm text-gray-400">
                        <FileText className="w-4 h-4" />
                        <span>Description</span>
                      </div>
                    </div>
                    <div className="bg-gray-800/50 border border-gray-600 p-3 rounded-lg max-h-[200px] overflow-y-auto">
                      <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{assignment.Todo}</p>
                    </div>
                  </div>
                )}
              </div>
              <Separator className="bg-gray-700 w-[80%] mx-auto" />

              <div className="flex flex-col items-center w-full">

                <div className="w-full grid grid-cols-1 gap-2 md:grid-cols-3">

                  <div className="flex flex-col w-full items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <span className="text-sm text-gray-400">Status</span>
                    <StatusTag assignment={assignment} onEdit={onEdit} />
                  </div>

                  <div className="flex flex-col w-full items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <span className="text-sm text-gray-400">Priority</span>
                    <PriorityTag assignment={assignment} onEdit={onEdit} />
                  </div>

                  <div className="flex flex-col w-full items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <span className="text-sm text-gray-400">Type</span>
                    <TypeTag assignment={assignment} onEdit={onEdit} />
                  </div>
                </div>

              </div>


            </TabsContent>

            <TabsContent value="documents">
              <AssignmentDocuments assignment={assignment} />
            </TabsContent>



          </Tabs>

          <Separator className="bg-gray-700" />

          {/* Actions */}
          
          <div className="grid grid-cols-2 gap-2">
        
            <Button
              variant="outline"
              size="sm"
              onClick={() => router.push(`/chat?assignment=${assignment.ID}`)}
              className="flex-1 bg-transparent border-gray-600"
            >
              <Bot className="h-4 w-4" />
              <span className="text-sm">AI Help</span>
            </Button>
    
            <Button
              variant="outline"
              size="sm"
              className="flex-1 bg-transparent border-gray-600"
              onClick={(e) => {
                e.stopPropagation()
                onOpenEdit(assignment)
              }}
            >
              <Edit className="mr-1 w-3 h-3" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="flex-1 bg-transparent border-gray-600"
              onClick={(e) => {
                e.stopPropagation()
                handleOpenLink()
              }}
            >
              <ExternalLink className="mr-1 w-3 h-3" />
              Open Link
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10"
              onClick={(e) => {
                e.stopPropagation()
                onDelete(assignment)
              }}
            >
              <Trash2 className="mr-1 w-3 h-3" />
              Delete
            </Button>

          </div>
        </div>

      </DialogContent>
    </Dialog>
  )
}
