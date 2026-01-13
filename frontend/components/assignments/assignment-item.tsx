"use client"

import { CardContent, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { GlassCard, GlassCardVariants } from "@/components/ui/glass-card"
import { models } from "@/wailsjs/go/models"
import { parseDeadline, getDueDescription } from "@/lib/date-utils"
import { useState } from "react"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { StatusTag } from "./tags/status-tag"
import { TypeTag } from "./tags/type-tag"
import { PriorityTag } from "./tags/priority-tag"
import { Bot, Clock, CopyPlus, Edit, Link, MoreHorizontal, Trash2 } from "lucide-react"
import { Button } from "../ui/button"
import { useRouter } from "next/navigation"
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "../ui/dropdown-menu"
import { cn } from "@/lib/utils"
import { Avatar, AvatarImage, AvatarFallback } from "@radix-ui/react-avatar"
import { AssignmentDetailsModal } from "./assignment-details-modal"

interface AssignmentItemProps<T extends models.LocalAssignment | models.Assignment> {
  assignment: T
  onEdit?: (assignment: models.LocalAssignment, column: string, value: string) => void
  onDelete?: (assignment: models.LocalAssignment) => void
  onOpenEdit?: (assignment: models.LocalAssignment) => void
  size?: "default" | "sm"
  disabled?: boolean
  variant?: GlassCardVariants
  mode?: "default" | "user"
  user?: models.User
  onCopy?: (assignment: models.Assignment) => void
}

interface SideActionsDropDownProps {
  isOpen: boolean
  setIsOpen: (isOpen: boolean) => void
}


export function AssignmentItem({
  assignment,
  user,
  size = "default",
  disabled = false,
  variant = "default",
  mode = "default",
  onEdit,
  onDelete,
  onOpenEdit,
  onCopy,
}: AssignmentItemProps<models.LocalAssignment | models.Assignment>) {


  const router = useRouter()
  // Parse deadline with timezone awareness
  const { Deadline, Status } = assignment
  const deadline = parseDeadline(Deadline)
  const [isDetailsOpen, setIsDetailsOpen] = useState(false)
  const [isActionsOpen, setIsActionsOpen] = useState(false)

  const handleCardClick = () => {
    setIsDetailsOpen(true)
  }

  const handleOpenLink = (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation()
    BrowserOpenURL(assignment.Link)
  }


  const statusColors = {
    "Not started": "bg-gray-500",
    "In progress": "bg-accent-amber",
    "Done": "bg-green-500",
  }



  const DefaultssignmentItem = ({ assignment, onEdit, variant, size }: AssignmentItemProps<models.LocalAssignment>) => {

    if (!assignment || !onOpenEdit || !onDelete || !onEdit) return null

    const handleEditOpen = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      onOpenEdit(assignment)
    }


    const handleOpenAIHelp = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      router.push(`/chat?assignment=${assignment.ID}`)
    }

    const handleDelete = (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation()
      onDelete(assignment)
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
        onClick: handleEditOpen
      },
      {
        label: "Delete",
        icon: Trash2,
        onClick: handleDelete
      }
    ]



    const SideActionDropdown = ({ isOpen, setIsOpen }: SideActionsDropDownProps) => {

      return (
        <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
          <DropdownMenuTrigger asChild>
            <Button variant={variant == "default" ? "default" : "outline"} size="icon" className="rounded-full w-7 h-7">
              <MoreHorizontal className="w-3.5 h-3.5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="glass border-gray-600">
            {SideActions.map((action) => (
              <DropdownMenuItem key={action.icon.name} onClick={(e) => action.onClick(e)}>
                <action.icon className="w-3.5 h-3.5" />
                {action.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      )
    }

    const handleActionOpenChange = (e: React.MouseEvent<HTMLDivElement>, isOpen: boolean) => {
      e.stopPropagation()
      setIsDetailsOpen(isOpen)
    }

    return (
      <GlassCard
        variant={variant}
        onClick={handleCardClick}
      >
        <CardContent className="flex flex-col flex-1 p-5 gap-4 ">

          {/* Right Column: Main Content */}
          <div className="flex flex-1 items-center gap-2">

            {/* Status Indicator - Vertical Bar */}
            <div className={cn("w-1 h-10 rounded-full shrink-0", statusColors[assignment.Status as keyof typeof statusColors])} />

            <div className="flex-1 flex flex-col gap-2">

              {/* 2. Main Info: Title & Description */}
              <div className="flex items-center justify-between">


                {size === "default" && (
                  <h5 className={`text-h5 line-clamp-1 tracking-tight`}>
                    {assignment.Title}
                  </h5>

                )}
                {size === "sm" && (
                  <p className={`text-body line-clamp-1 tracking-tight`}>
                    {assignment.Title}
                  </p>
                )}

                <SideActionDropdown isOpen={isActionsOpen} setIsOpen={setIsActionsOpen} />
              </div>


              <div className="flex flex-1 items-center justify-between">
                <p className={`text-caption text-text-caption flex items-center gap-1 line-clamp-1 leading-relaxed`}  >
                  <Clock className="w-3.5 h-3.5" />
                  {getDueDescription(deadline, Status)}
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

          <CardFooter className="grid grid-cols-3 gap-4">
            <StatusTag assignment={assignment} onEdit={onEdit} variant={variant} />
            <TypeTag assignment={assignment} onEdit={onEdit} variant={variant} />
            <PriorityTag assignment={assignment} onEdit={onEdit} variant={variant} />
          </CardFooter>
        )}

        <AssignmentDetailsModal
          isOpen={isDetailsOpen}
          onClose={() => setIsDetailsOpen(false)}
          assignment={assignment}
          onEdit={onEdit}
          onDelete={onDelete}
          onOpenEdit={onOpenEdit}
        />
      </GlassCard >
    )
  }

  const UserAssignmentItem = ({ assignment, variant, size, onCopy, user }: AssignmentItemProps<models.Assignment>) => {

    return (
      <GlassCard
        variant={variant}
        onClick={handleCardClick}
      >
        <CardContent className="flex flex-col flex-1 p-5 gap-4 ">

          {/* Right Column: Main Content */}
          <div className="flex flex-1 items-center gap-3">

            {/* Status Indicator - Vertical Bar */}
            <div className={cn("w-1 h-10 rounded-full shrink-0", statusColors[assignment.Status as keyof typeof statusColors])} />

            <div className="flex-1 flex flex-col gap-2">

              {/* 2. Main Info: Title & Description */}

              <div className="flex items-center justify-between">
                <h5 className={`text-h5 line-clamp-1 tracking-tight`}>
                  {assignment.Title}
                </h5>

                <Button variant={variant == "default" ? "default" : "outline"} size="icon" className="rounded-full w-7  h-7" onClick={(e) => {
                  e.stopPropagation()
                  onCopy?.(assignment)
                }}>
                  <CopyPlus className="w-3.5 h-3.5" />
                </Button>


              </div>


              <div className="flex flex-1 items-center justify-between">
                <p className={`text-caption text-text-caption flex items-center gap-1 line-clamp-1 leading-relaxed`}  >
                  <Clock className="w-3.5 h-3.5" />
                  {getDueDescription(deadline, assignment.Status)}
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

        <AssignmentDetailsModal
          isRemote
          isOpen={isDetailsOpen}
          onClose={() => setIsDetailsOpen(false)}
          assignment={assignment}
          onCopy={onCopy}
        />

      </GlassCard >
    )
  }


  switch (mode) {
    case "user":
      return <UserAssignmentItem assignment={assignment as models.Assignment} disabled={disabled} variant={variant} mode={mode} onCopy={onCopy} size={size} user={user} />
    default:
      return <DefaultssignmentItem assignment={assignment as models.LocalAssignment} onEdit={onEdit} onDelete={onDelete} onOpenEdit={onOpenEdit} disabled={disabled} variant={variant} mode={mode} size={size} />
  }
}
