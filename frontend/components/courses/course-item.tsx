import { memo, useMemo } from "react"
import { CardContent } from "../ui/card"
import { GlassCard, GlassCardVariants } from "../ui/glass-card"
import { Progress } from "../ui/progress"
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
  MoreHorizontal 
} from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useCourseAssignments, useCourse, useDeleteCourse } from "@/hooks/use-courses"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "../ui/dropdown-menu"
import { useDialogContext } from "../provider/dialog-provider"
import { cn } from "@/lib/utils"
import { format } from "date-fns"

interface CourseItemProps {
  courseId: string
  disabled?: boolean
  variant?: GlassCardVariants
  size?: "default" | "compact"
  mode?: "default" | "readonly"
  courseRO?: models.Course | models.LocalCourse
  user?: models.User
  onAccept?: () => void
  onDecline?: () => void
}

function BaseCourseItem({
  courseId,
  disabled = false,
  size = "default",
  mode = "default",
  variant = "outline",
  courseRO,
  user,
  onAccept,
  onDecline
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
          {(onAccept || onDecline) && (
            <div className="flex gap-2 border-t border-white/5">
              {onAccept && (
                <Button 
                  onClick={(e) => {
                    e.stopPropagation()
                    onAccept()
                  }} 
                  variant="outline" 
                  size="sm" 
                  className="flex-1 rounded-full"
                >
                  Accept
                </Button>
              )}
              {onDecline && (
                <Button 
                  onClick={(e) => {
                    e.stopPropagation()
                    onDecline()
                  }} 
                  variant="outline" 
                  size="sm" 
                  className="flex-1 rounded-full"
                >
                  Decline
                </Button>
              )}
            </div>
          )}

        </CardContent>
      </GlassCard>
    )
  }

  return mode === "readonly" ? <ReadonlyCourseItem /> : <DefaultCourseItem />
}

export const CourseItem = memo(BaseCourseItem, (prevProps, nextProps) => {
  return (
    prevProps.courseId === nextProps.courseId &&
    prevProps.size === nextProps.size &&
    prevProps.mode === nextProps.mode
  )
})