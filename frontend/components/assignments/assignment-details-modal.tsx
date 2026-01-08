"use client"

import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Calendar, Edit, Trash2, FileText, ExternalLink, Info, Bot, Clock, Link as LinkIcon, CopyPlus } from "lucide-react"
import { format } from "date-fns"
import { models } from "@/wailsjs/go/models"
import { parseDeadline, calculateDaysDifference, isOverdue, getDueDescription } from "@/lib/date-utils"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { StatusTag } from "@/components/assignments/tags/status-tag"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { TypeTag } from "./tags/type-tag"
import { useAssignments } from "@/hooks/use-assignments"
import { PriorityTag } from "./tags/priority-tag"
import { useRouter } from "next/navigation"
import FileUpload05 from "../file-upload-05"
import { AnimatePresence, motion } from "framer-motion"

interface AssignmentDetailsModalProps<T extends models.LocalAssignment | models.Assignment> {
  isOpen: boolean
  onClose: () => void
  assignment: T
  onOpenEdit?: (assignment: models.LocalAssignment) => void
  onEdit?: (assignment: models.LocalAssignment, column: string, value: string) => void
  onDelete?: (assignment: models.LocalAssignment) => void
  onCopy?: (assignment: models.Assignment) => void
  isRemote?: boolean
}

export function AssignmentDetailsModal({
  isOpen,
  onClose,
  assignment,
  onOpenEdit,
  onEdit,
  onDelete,
  onCopy,
  isRemote = false,
}: AssignmentDetailsModalProps<models.LocalAssignment | models.Assignment>) {

  if (!assignment) return null

  const { Title, Deadline, Status, Priority, Type, Todo, Documents, Course, Link } = assignment

  const router = useRouter()

  const [activeView, setActiveView] = useState("info")

  // Parse deadline with timezone awareness
  const deadline = parseDeadline(Deadline)

  const isOverdueStatus = isOverdue(deadline, Status)

  const daysUntilDue = calculateDaysDifference(deadline)
  
  const handleOpenLink = () => {
    BrowserOpenURL(Link)
  }


  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-lg p-0 overflow-hidden gap-0">

        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">

          <div className="flex items-center space-x-2 mb-1">
            <div className={`w-2 h-2 rounded-full ${Course?.Color}`} />
            <span className="text-caption uppercase tracking-wider">{Course?.Name}</span>
          </div>
          <DialogTitle className="text-h3">{Title}</DialogTitle>

        </DialogHeader>


        <div className="p-6">
          <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl pointer-events-none" />

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
            <AnimatePresence mode="wait" >


              <TabsContent value="info" key="info" className="space-y-6" asChild>
                <motion.div
                  key="info"
                  initial={{ x: -20, opacity: 0 }}
                  animate={{ x: 0, opacity: 1 }}
                  exit={{ x: -20, opacity: 0 }}
                  transition={{ duration: 0.2 }}
                >

                  {/* Deadline Section */}
                  <div className="p-4 rounded-xl bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 group">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                        <Clock className="w-3.5 h-3.5 text-blue-400" />
                        <span>Deadline</span>
                      </div>
                      <Badge variant="outline" className={`border-white/10 bg-white/5 ${isOverdueStatus ? "text-red-400" : daysUntilDue < 0 ? "text-gray-400" : "text-yellow-400"}`}>
                        {getDueDescription(deadline, Status)}
                      </Badge>
                    </div>
                    <p className="font-medium text-white text-lg">{format(deadline, "EEEE MMMM d, yyyy")}</p>
                    <p className="text-gray-400 text-sm mt-1">{format(deadline, "h:mm a")}</p>

                  </div>

                  {/* Tags Grid */}
                  <div className="grid grid-cols-3 gap-3">
                    <div className="flex flex-col items-center space-y-2 border border-white/5 p-3 rounded-xl bg-white/5">
                      <span className="text-xs font-medium uppercase tracking-wider">Status</span>
                      {!isRemote ? (
                        <StatusTag assignment={assignment as models.LocalAssignment} onEdit={onEdit!} variant="outline" />
                      ) : (
                        < Badge variant="outline" className={`text-caption font-normal`}>
                          {Status}
                        </Badge>
                      )}

                    </div>

                    <div className="flex flex-col items-center space-y-2 border border-white/5 p-3 rounded-xl bg-white/5">
                      <span className="text-xs font-medium uppercase tracking-wider">Priority</span>
                      {!isRemote ? (
                        <PriorityTag assignment={assignment as models.LocalAssignment} onEdit={onEdit!} variant="outline" />
                      ) : (
                        < Badge variant="outline" className={`text-caption font-normal`}>
                          {Priority}
                        </Badge>
                      )}
                    </div>

                    <div className="flex flex-col items-center space-y-2 border border-white/5 p-3 rounded-xl bg-white/5">
                      <span className="text-xs font-medium uppercase tracking-wider">Type</span>
                      {!isRemote ? (
                        <TypeTag assignment={assignment as models.LocalAssignment} onEdit={onEdit!} variant="outline" />
                      ) : (
                        < Badge variant="outline" className={`text-caption font-normal`}>
                          {Type}
                        </Badge>
                      )}
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
                        <p className="whitespace-pre-wrap leading-relaxed text-sm text-gray-200">{Todo}</p>
                      </div>
                    </div>
                  )}
                </motion.div>
              </TabsContent>




              <TabsContent value="documents" key="documents" className="flex flex-col flex-1" asChild>
                <motion.div
                  key="documents"
                  initial={{ x: 20, opacity: 0 }}
                  animate={{ x: 0, opacity: 1 }}
                  exit={{ x: 20, opacity: 0 }}
                  transition={{ duration: 0.2 }}
                >
                  <FileUpload05
                    assignment={assignment as models.LocalAssignment}
                  />
                </motion.div>
              </TabsContent>

            </AnimatePresence>
          </Tabs>

          {/* Actions */}

          <div className="grid grid-cols-4 gap-3 mt-6">

            {!isRemote && (

              <Button
                variant="outline"
                size="sm"
                className="rounded-full"
                onClick={() => router.push(`/chat?assignment=${assignment.ID}`)}
              >
                <Bot className="w-4 h-4" />
                <span>AI Help</span>
              </Button>


            )}


              {!isRemote && (
              <Button
                variant="outline"
                size="sm"
                className="rounded-full"
                onClick={(e) => {
                  e.stopPropagation()
                  onOpenEdit?.(assignment as models.LocalAssignment)
                }}
              >
                <Edit className="w-4 h-4" />
                <span>Edit</span>
              </Button>
            )}

            <Button
              variant="outline"
              size="sm"
              className="rounded-full"
              onClick={(e) => {
                e.stopPropagation()
                handleOpenLink()
              }}
            >
              <ExternalLink className="w-4 h-4" />
              <span>Link</span>
            </Button>

            {isRemote && (
              <Button
                variant="primary"
                size="sm"
                className=" rounded-full"
                onClick={(e) => {
                  e.stopPropagation()
                  onCopy?.(assignment as models.Assignment)
                }}
              >

                <CopyPlus className="h-4 w-4" />
                <span>Copy</span>
              </Button>
            )}
            {!isRemote && (
              <Button
                variant="danger"
                size="sm"
                className="rounded-full"
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete?.(assignment as models.LocalAssignment)
                }}
              >
                <Trash2 className="w-4 h-4" />
                <span>Delete</span>
              </Button>
            )}
          </div>
        </div>

      </DialogContent>
    </Dialog >
  )
}
