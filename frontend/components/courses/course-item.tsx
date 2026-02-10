import { memo, useMemo } from "react"
import { CardContent } from "../ui/card"
import { GlassCard, GlassCardVariants } from "../ui/glass-card"
import { Badge } from "../ui/badge"
import { Button } from "../ui/button"
import {
  Users,
  Clock,
  Edit,
  Trash2,
  MapPin,
  Calendar,
  BookOpen,
  MoreHorizontal,
  Share2,
  LucideIcon,
  LogOut,
  UserPlus,
  UserMinus
} from "lucide-react"
import { client, models } from "@/wailsjs/go/models"
import { useCourseAssignments, useCourse, useDeleteCourse, useGetClusterStatus, useAcceptCourseInvitation } from "@/hooks/use-courses"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "../ui/dropdown-menu"
import { useDialogContext } from "../provider/dialog-provider"
import { cn } from "@/lib/utils"
import { useSendClusterRequest } from "@/hooks/use-courses"
import { useFriendShipStatus } from "@/hooks/use-friends"
import { parseSchedule } from "@/lib/date-utils"
import { differenceInMinutes, format, isBefore, isAfter } from "date-fns"


interface DefaultCourseItemProps {
  disabled?: boolean
  variant?: GlassCardVariants
  size?: "default" | "compact"
  mode?: "default" | "schedule" | "readonly"
}


interface ReadonlyCourseItemProps {
  courseRO?: models.Course
  user?: models.User
  onAccept?: () => void
  onDecline?: () => void
}

interface ScheduleCourseItemProps {
  timeSlots: number[]
  day: string
}



interface CourseItemProps extends Partial<ScheduleCourseItemProps>, Partial<ReadonlyCourseItemProps>, Partial<DefaultCourseItemProps> {
  courseId: string
}

type ClusterActionProps = {
  label: string
  variant: "default" | "outline"
  Icon: LucideIcon
  onClick: (e: React.MouseEvent) => void
  className: string
}



function BaseCourseItem({
  courseId,
  disabled = false,
  size = "default",
  mode = "default",
  variant = "outline",
  courseRO,
  user,
  timeSlots = [],
  day
}: CourseItemProps) {

  const deleteMutation = useDeleteCourse()
  const { SetDialogState } = useDialogContext()

  const handleDelete = (course: models.LocalCourse) => {
    deleteMutation.mutate(course)
  }

  const handleEdit = (id: string) => {
    SetDialogState({ modelType: "course", dialogType: "edit", id })
  }

  // Default Course Item
  function DefaultCourseItem() {
    const { data: course } = useCourse(courseId)
    const { data: courseAssignments } = useCourseAssignments(courseId)

    if (!course) return null

    const completedCount = useMemo(() => {
      return courseAssignments?.filter((assignment) => assignment.Status === "Done").length || 0
    }, [courseAssignments])

    const completionPercentage = useMemo(() => {
      return courseAssignments?.length > 0
        ? (completedCount / courseAssignments.length) * 100
        : 0
    }, [courseAssignments, completedCount])

    const sideActions = [
      {
        label: "Edit",
        icon: Edit,
        onClick: () => handleEdit(course.ID)
      },
      {
        label: "Delete",
        icon: Trash2,
        onClick: () => handleDelete(course),
        variant: "danger" as const
      }
    ]

    // Compact size
    if (size === "compact") {
      return (
        <GlassCard
          variant={variant}
          className={cn("group cursor-pointer", disabled && "opacity-50")}
          onClick={() => SetDialogState({ modelType: "course", dialogType: "details", id: courseId })}
        >
          <CardContent className="p-4">
            <div className="flex items-center gap-4">

              {/* Color indicator */}
              <div className="flex-shrink-0">
                <div className={cn("w-12 h-12 rounded-lg flex items-center justify-center", course.Color)}>
                  <BookOpen className="w-6 h-6 text-white" />
                </div>
              </div>

              {/* Course info */}
              <div className="flex-1 min-w-0">
                <h4 className="text-body font-bold text-white truncate mb-1">
                  {course.Code}
                </h4>
                <p className="text-caption text-text-caption truncate">
                  {course.Name}
                </p>
              </div>

              {/* Credits badge */}
              <Badge variant="outline" className="text-caption border-white/10 bg-white/5 text-text-caption px-2 py-1">
                {course.Credits} CR
              </Badge>

            </div>
          </CardContent>
        </GlassCard>
      )
    }

    // Default size
    return (
      <GlassCard
        variant={variant}
        className={cn("group cursor-pointer", disabled && "opacity-50")}
        onClick={() => SetDialogState({ modelType: "course", dialogType: "details", id: courseId })}
      >
        <CardContent className="flex flex-col p-5 gap-4">

          {/* Header with course code and color */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3 flex-1 min-w-0">

              {/* Color indicator */}
              <div className={cn("w-14 h-14 rounded-xl flex items-center justify-center flex-shrink-0", course.Color)}>
                <BookOpen className="w-7 h-7 text-white" />
              </div>

              {/* Course code and name */}
              <div className="flex-1 min-w-0">
                <h3 className="text-h5 font-bold text-white tracking-tight mb-1">
                  {course.Code}
                </h3>
                <p className="text-caption text-text-caption line-clamp-1">
                  {course.Name}
                </p>
              </div>
            </div>

            {/* Actions dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  size="icon"
                  className="rounded-full w-8 h-8"
                  onClick={(e) => e.stopPropagation()}
                >
                  <MoreHorizontal className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="glass border-white/10">
                {sideActions.map((action) => (
                  <DropdownMenuItem
                    key={action.label}
                    onClick={(e) => {
                      e.stopPropagation()
                      action.onClick()
                    }}
                    className={cn(
                      action.variant === "danger" && "text-red-400 hover:text-red-300"
                    )}
                  >
                    <action.icon className="w-4 h-4" />
                    {action.label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Course details */}
          <div className="space-y-3">

            {/* Instructor */}
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                <Users className="w-4 h-4 text-text-caption" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-caption text-text-caption uppercase tracking-wider mb-0.5">
                  Instructor
                </p>
                <p className="text-body text-white font-medium truncate">
                  {course.Instructor}
                </p>
              </div>
            </div>

            {/* Schedule */}
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                <Clock className="w-4 h-4 text-text-caption" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-caption text-text-caption uppercase tracking-wider mb-0.5">
                  Schedule
                </p>
                <p className="text-body text-white font-medium truncate">
                  {course.Schedule}
                </p>
              </div>
            </div>

            {/* Location */}
            {course.Location && (
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                  <MapPin className="w-4 h-4 text-text-caption" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-caption text-text-caption uppercase tracking-wider mb-0.5">
                    Location
                  </p>
                  <p className="text-body text-white font-medium truncate">
                    {course.Location}
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Progress section */}
          {/* {courseAssignments && courseAssignments.length > 0 && (
            <div className="p-4 rounded-xl bg-white/5 border border-white/5 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-caption text-text-caption uppercase tracking-wider">
                  Progress
                </span>
                <span className="text-body font-bold text-white">
                  {completedCount}/{courseAssignments.length}
                </span>
              </div>
              <Progress value={completionPercentage} className="h-2" />
              <div className="flex items-center justify-between text-caption">
                <span className="text-text-caption">
                  {Math.round(completionPercentage)}% Complete
                </span>
                <span className="text-text-caption">
                  {courseAssignments.length - completedCount} remaining
                </span>
              </div>
            </div>
          )} */}

          {/* Footer with metadata */}
          <div className="flex items-center justify-between">
            <Badge variant="outline" className="text-caption border-white/10 bg-white/5 text-text-caption">
              {course.Credits} Credits
            </Badge>
            <Badge variant="outline" className="text-caption border-white/10 bg-white/5 text-text-caption">
              <Calendar className="w-3 h-3 mr-1" />
              {course.Semester}
            </Badge>
          </div>

        </CardContent>
      </GlassCard>
    )
  }


  // Readonly Course Item (for user courses)
  function ReadonlyCourseItem() {
    if (!courseRO) return null

    const { data: clusterStatus } = useGetClusterStatus(courseRO.ID)
    const { data: friendshipStatus } = useFriendShipStatus(courseRO.UserID)

    const { mutate: sendClusterRequest } = useSendClusterRequest()
    const { mutate: acceptClusterRequest } = useAcceptCourseInvitation()


    const ClusterActions = memo(({ clusterStatus }: { clusterStatus: client.CourseStatusResponse }) => {

      var props: ClusterActionProps = {
        label: "Join",
        variant: "outline" as const,
        Icon: LogOut,
        onClick: (e: React.MouseEvent) => {
          e.stopPropagation()
        },
        className: "bg-primary-blue-500 hover:bg-primary-blue-600 text-white"
      }
      if (clusterStatus?.status === "accepted") {
        props = {
          label: "Quit",
          variant: "outline" as const,
          Icon: LogOut,
          onClick: (e: React.MouseEvent) => {
            e.stopPropagation()
            sendClusterRequest({
              courseID: courseRO.ID
            })
          },
          className: "text-red-400 hover:text-red-300"
        }
      }
      if (clusterStatus?.status === "pending") {
        if (clusterStatus?.is_pending_for_you) {
          props = {
            label: "Accept",
            variant: "default" as const,
            Icon: UserPlus,
            onClick: (e: React.MouseEvent) => {
              e.stopPropagation()
            },
            className: "bg-primary-blue-500 hover:bg-primary-blue-600 text-white"
          }
        } else {
          props = {
            label: "Cancel Request",
            variant: "outline" as const,
            Icon: UserMinus,
            onClick: (e: React.MouseEvent) => {
              e.stopPropagation()
              acceptClusterRequest({
                invitation: {
                  ID: clusterStatus.id,
                  Course: courseRO as models.Course
                } as models.CourseInvitation
              })
            },
            className: "text-red-400 hover:text-red-300"
          }
        }
      }


      return (
        <Button
          variant={props.variant}
          size="sm"
          className={cn("flex-1 text-xs font-medium h-9", props.className)}
          onClick={props.onClick}
        >
          <props.Icon className="w-3.5 h-3.5 mr-1.5" />
          {props.label}
        </Button>
      )
    }, (prevProps, nextProps) => {
      return prevProps.clusterStatus?.status === nextProps.clusterStatus?.status &&
        prevProps.clusterStatus?.is_pending_for_you === nextProps.clusterStatus?.is_pending_for_you
    })

    return (
      <GlassCard
        variant={variant}
        className={cn("group", disabled && "opacity-50")}
        onClick={() => SetDialogState({
          modelType: "course",
          dialogType: "details",
          id: courseRO.ID,
          item: courseRO as models.Course,
          viewMode: "readonly"
        })}
      >
        <CardContent className="flex flex-col p-5 gap-4">

          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3 flex-1 min-w-0">

              {/* Color indicator */}
              <div className={cn("w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0", courseRO.Color)}>
                <BookOpen className="w-6 h-6 text-white" />
              </div>

              {/* Course info */}
              <div className="flex-1 min-w-0">
                <h3 className="text-h5 font-bold text-white tracking-tight mb-1">
                  {courseRO.Code}
                </h3>
                <p className="text-caption text-text-caption line-clamp-1">
                  {courseRO.Name}
                </p>
              </div>
            </div>

            {/* User badge */}
            {user && (
              <Badge variant="outline" className="text-caption border-white/10 bg-white/5 text-text-caption">
                {user.Username}
              </Badge>
            )}
          </div>

          {/* Course details */}
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-caption">
              <Users className="w-3.5 h-3.5 text-text-caption" />
              <span className="text-text-caption truncate">{courseRO.Instructor}</span>
            </div>
            <div className="flex items-center gap-2 text-caption">
              <Clock className="w-3.5 h-3.5 text-text-caption" />
              <span className="text-text-caption truncate">{courseRO.Schedule}</span>
            </div>
          </div>

          {/* Actions */}

          <div className="flex gap-2 border-t border-white/5">
            {clusterStatus && friendshipStatus?.status == "accepted" && <ClusterActions clusterStatus={clusterStatus} />}
          </div>


        </CardContent>
      </GlassCard>
    )
  }



  function ScheduleCourseItem() {

    const { data: course } = useCourse(courseId)
    if (!course) return null
    const parsedSchedule = parseSchedule(course?.Schedule)
    if (!parsedSchedule) return null

    var isOn = false
    const startHour = parsedSchedule?.startHour || 0
    const startMinute = parsedSchedule?.startMinute || 0
    const endHour = parsedSchedule?.endHour || 0
    const endMinute = parsedSchedule?.endMinute || 0

    const today = new Date()
    const startDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), startHour, startMinute)
    const endDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), endHour, endMinute)

    isOn = day == format(today, 'EEEE') && isBefore(startDate, today) && isAfter(endDate, today)
    const duration = differenceInMinutes(endDate, startDate)

    // Calculate position
    var topPosition = ((startHour - timeSlots[0]) * 60)
    if (startMinute != 0) {
      topPosition = topPosition + (60 / (60 / startMinute))
    }
    const height = duration // Exact duration

    return (
      <div
        key={course.ID}
        className={` absolute left-1 right-1 border rounded-lg hover:translate-y-0.5 backdrop-blur-lg transition-all duration-300 overflow-hidden  cursor-pointer group 
          ${isOn
          ? 'border-blue-400/50 ring-2 ring-blue-500/30 shadow-[0_0_15px_rgba(59,130,246,0.3)]'
          : 'bg-white/5 border-white/5 shadow-lg shadow-black/60'
          }`}
        style={{
          top: `${topPosition}px`,
          height: `${height}px`,
        }}
        onClick={() => SetDialogState({ modelType: "course", dialogType: "details", id: courseId })}
      >
        <div className="h-full w-full p-2.5 flex flex-col relative overflow-hidden gap-2">
          {/* Shine effect on hover */}
          <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

          <div className="flex items-center relative z-10">
            <div className="font-semibold text-sm text-white drop-shadow-md truncate">
              {course?.Code}
            </div>
          </div>

          {course.Name && height > 50 && (
            <div className="text-[10px] text-muted-foreground truncate relative z-10">
              {course?.Name}
            </div>
          )}

          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3 text-gray-500" />
            <div className="flex-1 min-w-0">
              <div className="text-[10px] font-medium text-gray-400 truncate relative z-10 group-hover:text-white">
                {parsedSchedule?.startTimeString} - {parsedSchedule?.endTimeString}
              </div>
            </div>
          </div>


          {/* Location */}
          {course.Location && (
            <div className="flex items-center gap-1">
              <MapPin className="w-3 h-3 text-gray-500" />
              <div className="flex-1 min-w-0">
                <p className="text-[10px] text-gray-400 font-medium truncate group-hover:text-white">
                  {course.Location}
                </p>
              </div>
            </div>
          )}

        </div>
      </div>


    )
  }


  switch (mode) {
    case "readonly":
      return <ReadonlyCourseItem />
    case "schedule":
      return <ScheduleCourseItem />
    default:
      return <DefaultCourseItem />
  }
}

export const CourseItem = memo(BaseCourseItem, (prevProps, nextProps) => {
  return (
    prevProps.courseId === nextProps.courseId &&
    prevProps.size === nextProps.size &&
    prevProps.mode === nextProps.mode
  )
})