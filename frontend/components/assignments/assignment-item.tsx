"use client"

import { CardContent, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { GlassCard, GlassCardVariants } from "@/components/ui/glass-card"
import { models } from "@/wailsjs/go/models"
import { parseDeadline, getDueDescription } from "@/lib/date-utils"
import { memo, useCallback, useRef, useState } from "react"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { StatusTag } from "./tags/status-tag"
import { TypeTag } from "./tags/type-tag"
import { PriorityTag } from "./tags/priority-tag"
import { Bot, Clock, CopyPlus, Edit, Link, LucideIcon, MoreHorizontal, Trash2 } from "lucide-react"
import { Button } from "../ui/button"
import { useRouter } from "next/navigation"
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "../ui/dropdown-menu"
import { cn } from "@/lib/utils"
import { Avatar, AvatarImage, AvatarFallback } from "@radix-ui/react-avatar"
import { AssignmentDetailsDialog } from "./assignment-details-dialog"
import { useAssignment, useDeleteAssignment, useUpdateAssignment } from "@/hooks/use-assignments"
import { format } from "date-fns"
import { useDialogContext } from "../provider/dialog-provider"

interface AssignmentItemProps {
  assignmentId: string
  size?: "default" | "sm"
  disabled?: boolean
  variant?: GlassCardVariants
  mode?: "default" | "ghost" | "user"

  assignment?: models.Assignment
  user?: models.User
  onCopy?: (assignment: models.Assignment, includeDocuments: boolean) => void
}

interface SideAction {
  label: string
  icon: LucideIcon
  onClick: (e: React.MouseEvent<HTMLDivElement>) => void
}

interface SideActionsDropDownProps {
  isOpen: boolean
  setIsOpen: (isOpen: boolean) => void
  sideActions: SideAction[]
  variant?: GlassCardVariants
}


const BaseAssignmentItem = ({
  assignmentId,
  user,
  size = "default",
  disabled = false,
  variant = "default",

  assignment,
  mode = "default",
  onCopy,
}: AssignmentItemProps) => {

  const [isActionsOpen, setIsActionsOpen] = useState(false)
  
  const [isDetailsOpen, setIsDetailsOpen] = useState(false)
  const [isEditOpen, setIsEditOpen] = useState(false)

  const router = useRouter()

  const deleteMutation = useDeleteAssignment()
  const { SetDialogState } = useDialogContext()

  const handleEditDialog = useCallback(() => {
    SetDialogState({ modelType: "assignment", dialogType: "edit", id: assignmentId })
  }, [])

  const handleDetailsDialog = useCallback(() => {
    SetDialogState({ modelType: "assignment", dialogType: "details", id: assignmentId })
  }, [])  


  const statusColors = {
    "Not started": "bg-gray-500",
    "In progress": "bg-accent-amber",
    "Done": "bg-green-500",
  }

  const SideActionDropdown = ({ isOpen, setIsOpen, sideActions, variant }: SideActionsDropDownProps) => {

    return (
      <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant={variant == "default" ? "default" : "outline"} size="icon" className="rounded-full w-7 h-7">
            <MoreHorizontal className="w-3.5 h-3.5" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="glass border-gray-600">
          {sideActions.map((action) => (
            <DropdownMenuItem key={action.label} onClick={(e) => action.onClick(e)}>
              <action.icon className="w-3.5 h-3.5" />
              {action.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    )
  }



  const DefaultssignmentItem = ({ assignmentId, variant, size }: AssignmentItemProps) => {

    const { data: assignment } = useAssignment(assignmentId)

    if (!assignment) return null


    const handleOpenLink = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      BrowserOpenURL(assignment.Link)
    }





    const handleOpenAIHelp = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      router.push(`/chat?assignment=${assignment.ID}`)
    }

    const handleDelete = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      deleteMutation.mutate(assignment)
    }


    const SideActions = [
      {
        label: "AI Help",
        icon: Bot,
        onClick: handleOpenAIHelp
      },
      {
        label: "Open Link",
        icon: Link,
        onClick: handleOpenLink
      },

      {
        label: "Edit",
        icon: Edit,
        onClick: (e: React.MouseEvent<HTMLDivElement>) => {
          e.stopPropagation()
          SetDialogState({ modelType: "assignment", dialogType: "edit", id: assignmentId })
        }
      },
      {
        label: "Delete",
        icon: Trash2,
        onClick: handleDelete
      }
    ]



    return (
      <div className="flex flex-1">
        <GlassCard
          key={assignment.ID}
          variant={variant}
          onClick={() => SetDialogState({ modelType: "assignment", dialogType: "details", id: assignmentId })}
          className="min-w-0"
        >
          <CardContent className="flex flex-col flex-1 p-4 gap-4 ">

            {/* Right Column: Main Content */}
            <div className="flex flex-1 items-center gap-4 min-w-0">
              <div className="flex flex-1 items-center gap-3 min-w-0">

                <div className={cn("w-1 h-10 rounded-full shrink-0", statusColors[assignment.Status as keyof typeof statusColors])} />

                <div className="relative bg-white/5 rounded-lg border shadow-lg shadow-black/60 border-white/10 group-hover:border-white/15 transition-colors shrink-0">
                  <div className="absolute inset-0 bg-gradient-to-br from-white/15 to-transparent pointer-events-none" />
                  <div className="flex flex-col items-center my-2 mx-2.5">
                    <span className="text-caption font-semibold text-gray-400 uppercase">{format(parseDeadline(assignment.Deadline), "MMM")}</span>
                    <span className="text-h5 font-bold text-white">{format(parseDeadline(assignment.Deadline), "d")}</span>
                  </div>
                </div>

                <div className="flex flex-col gap-1 min-w-0 max-w-3/4 flex-1" >
                  <span className="text-caption font-semibold text-gray-400 uppercase tracking-wider">{assignment.Course?.Code}</span>
                  <p className="text-body font-medium text-white truncate leading-tight mr-2">{assignment.Title}</p>
                  <div className="text-caption text-gray-400 flex items-center gap-1.5">
                    <Clock className="w-3 h-3" />
                    {getDueDescription(parseDeadline(assignment.Deadline), assignment.Status)}
                  </div>
                </div>
              </div>

              <SideActionDropdown isOpen={isActionsOpen} setIsOpen={setIsActionsOpen} sideActions={SideActions} variant={"outline"} />
            </div>

            {/* Left Column: Checkbox */}

          </CardContent>

          {size === "default" && (
            <CardFooter className="grid grid-cols-3 gap-4">
              <StatusTag assignment={assignment} variant={variant} />
              <TypeTag assignment={assignment} variant={variant} />
              <PriorityTag assignment={assignment} variant={variant} />
            </CardFooter>
          )}


        </GlassCard >
       
      </div>
    )
  }

  const GhostAssignmentItem = ({ assignmentId }: AssignmentItemProps) => {
    const { data: assignment } = useAssignment(assignmentId)

    if (!assignment) return null

    const [isActionsOpen, setIsActionsOpen] = useState(false)

    const handleOpenAIHelp = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      router.push(`/chat?assignment=${assignment.ID}`)
    }

    const SideActions = [
      {
        label: "AI Help",
        icon: Bot,
        onClick: handleOpenAIHelp
      },
    ]



    return (
      <div
        key={assignment.ID}
        className="flex hover:bg-white/10 p-4 transition-all group/exam"
        onClick={() => SetDialogState({ modelType: "assignment", dialogType: "details", id: assignmentId })}
      >
        <div className="flex flex-1 items-center gap-4 min-w-0">
          <div className="flex flex-1 items-center gap-3 min-w-0">

            <div className={cn("w-1 h-10 rounded-full shrink-0", statusColors[assignment.Status as keyof typeof statusColors])} />

            <div className="relative bg-white/5 rounded-lg border shadow-lg shadow-black/60 border-white/10 group-hover:border-white/15 transition-colors shrink-0">
              <div className="absolute inset-0 bg-gradient-to-br from-white/15 to-transparent pointer-events-none" />
              <div className="flex flex-col items-center my-2 mx-2.5">
                <span className="text-caption font-semibold text-gray-400 uppercase">{format(parseDeadline(assignment.Deadline), "MMM")}</span>
                <span className="text-h5 font-bold text-white">{format(parseDeadline(assignment.Deadline), "d")}</span>
              </div>
            </div>

            <div className="flex flex-col gap-1 min-w-0 max-w-3/4 flex-1" >
              <span className="text-caption font-semibold text-gray-400 uppercase tracking-wider">{assignment.Course?.Code}</span>
              <p className="text-body font-medium text-white truncate leading-tight">{assignment.Title}</p>
              <div className="text-caption text-gray-400 flex items-center gap-1.5">
                <Clock className="w-3 h-3" />
                {getDueDescription(parseDeadline(assignment.Deadline), assignment.Status)}
              </div>
            </div>
          </div>

          <SideActionDropdown isOpen={isActionsOpen} setIsOpen={setIsActionsOpen} sideActions={SideActions} variant={"outline"} />
        </div>
        
      </div>
    )
  }

  const UserAssignmentItem = ({ assignmentId, assignment, variant, size, onCopy, user }: AssignmentItemProps) => {

    if (!assignment) return null

    return (
      <div className="flex flex-1">
        <GlassCard
          key={assignment.ID}
          variant={variant}
          onClick={() => SetDialogState({ 
            modelType: "assignment", 
            dialogType: "details", 
            id: assignmentId,
            item: assignment,
            viewMode: "readonly"
          })}
        >
          <CardContent className="flex flex-col flex-1 p-5 gap-4 ">

            {/* Right Column: Main Content */}
            <div className="flex flex-1 items-center gap-3">

              {/* Status Indicator - Vertical Bar */}
              <div className={cn("w-1 h-10 rounded-full shrink-0", statusColors[assignment.Status as keyof typeof statusColors])} />

              <div className="flex-1 flex flex-col gap-2">

                {/* 2. Main Info: Title & Description */}

                <div className="flex items-center justify-between">
                  <h5 className={`text-h5 font-medium line-clamp-1 tracking-tight`}>
                    {assignment.Title}
                  </h5>

                  <Button variant={variant == "default" ? "default" : "outline"} size="icon" className="rounded-full w-7  h-7" onClick={(e) => {
                    e.stopPropagation()
                    onCopy?.(assignment, true)
                  }}>
                    <CopyPlus className="w-3.5 h-3.5" />
                  </Button>


                </div>


                <div className="flex flex-1 items-center justify-between">
                  <p className={`text-caption text-text-caption flex items-center gap-1 line-clamp-1 leading-relaxed`}  >
                    <Clock className="w-3.5 h-3.5" />
                    {getDueDescription(parseDeadline(assignment.Deadline), assignment.Status)}
                  </p>

                  <p className="text-caption flex items-center gap-1">
                    <div className={cn("w-2 h-2 rounded-full shrink-0", assignment.Course?.Color)} />
                    <span className="text-white">{assignment.Course?.Code}</span>
                  </p>
                </div>
              </div>

            </div>

            {/* Left Column: Checkbox */}

          </CardContent>

          {size === "default" && (

            <CardFooter className="flex-row-reverse p-4 pt-0 gap-2">
              <Badge variant="outline" className="gap-2">
                <span className="text-caption text-text-body">
                  {user?.Username || user?.Email}
                </span>
                <Avatar className="h-5 w-5 rounded-full overflow-hidden border border-white/10">
                  <AvatarImage src={user?.Avatar || "/placeholder-user.jpg"} />
                  <AvatarFallback className="text-[10px]">IN</AvatarFallback>
                </Avatar>
              </Badge>
            </CardFooter>
          )}

        </GlassCard >
       
      </div>
    )
  }


  switch (mode) {
    case "ghost":
      return <GhostAssignmentItem assignmentId={assignmentId} />
    case "user":
      return <UserAssignmentItem assignmentId={assignmentId} assignment={assignment} disabled={disabled} variant={variant} mode={mode} onCopy={onCopy} size={size} user={user} />
    default:
      return <DefaultssignmentItem assignmentId={assignmentId} disabled={disabled} variant={variant} mode={mode} size={size} />
  }
}

export const AssignmentItem = memo(BaseAssignmentItem, (prevProps, nextProps) => {
  return prevProps.assignmentId === nextProps.assignmentId
})
