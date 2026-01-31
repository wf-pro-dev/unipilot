import { memo, useMemo, useState } from "react"
import { CardContent } from "../ui/card"
import { GlassCard } from "../ui/glass-card"
import { Progress } from "../ui/progress"
import { Badge } from "../ui/badge"
import { Users, Clock, Edit, MoreVertical, Trash2, ChevronLeft } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useCourseAssignments } from "@/hooks/use-courses"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { Button } from "../ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "../ui/dropdown-menu"
import { useDeleteCourse, useUpdateCourse } from "@/hooks/use-courses"
import { GlassCardVariants } from "../ui/glass-card"
import { useCourse } from "@/hooks/use-courses"
import { parseSchedule } from "@/lib/date-utils"
import { differenceInMinutes, format, isAfter, isBefore } from "date-fns"
import { useDialogContext } from "../provider/dialog-provider"


interface CourseItemProps {
    courseId: number
    disabled?: boolean
    variant?: GlassCardVariants
    size?: "default" | "sm"
    mode?: "default" | "readonly" | "schedule"

    timeSlots?: number[]
    day?: string

    courseRO?: models.Course
    user?: models.User
    onClick?: () => void
    onAccept?: () => void
    onDecline?: () => void
}

function BaseCourseItem({
    courseId,
    disabled = false,
    size = "default",
    mode = "default",
    variant = "default",
    courseRO,
    user,
    timeSlots,
    day,
    onClick,
    onAccept,
    onDecline
}: CourseItemProps) {

    const updateMutation = useUpdateCourse()
    const deleteMutation = useDeleteCourse()

    const { SetDialogState } = useDialogContext()


    /**
   * Handles course field updates with optimistic UI updates.
   * 
   * Updates a specific field of a course and provides immediate UI feedback
   * through optimistic updates. Logs the change for audit purposes.
   * 
   * @param {course.LocalCourse} courseData - The course to update
   * @param {string} column - The field name to update
   * @param {string} value - The new value for the field
   * @returns {Promise<void>}
   */
    const handleEditCourse = async (courseData: models.LocalCourse, column: string, value: string) => {
        const message = "course " + courseData.Code + " " + column + " changed to " + value
        LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

        // Use the optimistic update mutation
        updateMutation.mutate({
            course: courseData,
            column,
            value
        })
    }

    /**
     * Handles course deletion with optimistic UI updates.
     * 
     * Deletes a course and provides immediate UI feedback. Logs the deletion
     * for audit purposes. Note: This triggers the delete confirmation dialog.
     * 
     * @param {course.LocalCourse} course - The course to delete
     * @returns {Promise<void>}
     */
    const handleDeleteCourse = async (course: models.LocalCourse) => {
        const message = "course " + course.Code + " deleted"
        LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
        deleteMutation.mutate(course)
    }



    // Helper function to get gradient classes for course colors
    // This ensures Tailwind can detect all possible class combinations
    const getCourseGradientClasses = (color: string | undefined, isOn: boolean | null) => {
        if (!color || !isOn) {
            return {
                bg: "bg-white/5",
                hover: "group-hover/item:bg-white/10 group-hover/item:border-white/10"
            }
        }

        // Map color values to gradient classes that Tailwind can detect
        const colorMap: Record<string, { bg: string; hover: string }> = {
            "bg-blue-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-blue-500/20 before:via-blue-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "group-hover/item:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-blue-500/40 after:via-blue-500/15 after:to-transparent after:opacity-0 group-hover/item:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-green-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-green-500/20 before:via-green-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "group-hover/item:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-green-500/40 after:via-green-500/15 after:to-transparent after:opacity-0 group-hover/item:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-purple-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-purple-500/20 before:via-purple-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "group-hover/item:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-purple-500/40 after:via-purple-500/15 after:to-transparent after:opacity-0 group-hover/item:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-red-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-red-500/20 before:via-red-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "group-hover/item:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-red-500/40 after:via-red-500/15 after:to-transparent after:opacity-0 group-hover/item:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-orange-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-orange-500/20 before:via-orange-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "group-hover/item:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-orange-500/40 after:via-orange-500/15 after:to-transparent after:opacity-0 group-hover/item:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-pink-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-pink-500/20 before:via-pink-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "group-hover/item:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-pink-500/40 after:via-pink-500/15 after:to-transparent after:opacity-0 group-hover/item:after:opacity-100 after:transition-opacity after:duration-500",
            }
        }

        return colorMap[color] || {
            bg: "bg-white/5",
            hover: "hover:bg-white/10 hover:border-white/10"
        }
    }


    function DefaultCourseItem({
        courseId,
        variant,
        size
    }: CourseItemProps) {

        const { data: course } = useCourse(courseId)

        if (!course) return null

        const { data: course_assignments } = useCourseAssignments(courseId)

        const completed_assignments_count = useMemo(() => {
            return course_assignments?.filter((assignment) => assignment.Status === "Done").length || 0
        }, [course_assignments])

        const completionPercentage = useMemo(() => {
            return course_assignments?.length > 0 ? (completed_assignments_count / course_assignments?.length) * 100 : 0
        }, [course_assignments, completed_assignments_count])


        if (size === "sm") {
            return (
                <div
                    key={course.ID}
                    onClick={() => SetDialogState({ modelType: "course", dialogType: "details", id: courseId })}
                    className="flex flex-1 justify-between items-center border border-white/5  shadow-lg shadow-black/60 rounded-xl py-2 px-4 overflow-hidden relative group/item">

                    <div className={`absolute inset-0 z-0 ${getCourseGradientClasses(course.Color, true).bg} ${getCourseGradientClasses(course.Color, true).hover} transition-colors duration-300`} />

                    <div className="flex items-center justify-center gap-2 z-10">

                        {!onAccept && !onDecline && (

                            <div className=" flex items-center justify-center p-1 rounded-full bg-white/10 border border-white/10 shadow-lg shadow-black/40">
                                <ChevronLeft className="w-4 h-4 text-white" strokeWidth={1.5} />
                            </div>
                        )}


                        <div className={`flex flex-col flex-1 z-10`}>

                            <h3 className="text-caption text-gray-300 line-clamp-1 tracking-tight">{course.Code}</h3>


                            <p className="text-body line-clamp-1 font-medium">{course.Name}</p>

                        </div>
                    </div>

                    <div className="flex flex-col items-center gap-1 z-10">
                        {onAccept && (
                            <Button onClick={onAccept} variant="ghost" size="sm" className="flex-1 text-gray-400 bg-gray-500/10 border-gray-500/20 hover:bg-gray-500/20 hover:text-gray-300 hover:border-gray-500/30 transition-all h-9 text-xs font-medium">
                                Accept
                            </Button>
                        )}
                        {onDecline && (
                            <Button onClick={onDecline} variant="ghost" size="sm" className="flex-1 text-gray-400 bg-gray-500/10 border-gray-500/20 hover:bg-gray-500/20 hover:text-gray-300 hover:border-gray-500/30 transition-all h-9 text-xs font-medium">
                                Decline
                            </Button>
                        )}
                    </div>

                </div>
            )
        }

        return (
            <div>
                <GlassCard
                    variant={"outline"}
                    className={`${disabled ? 'opacity-50' : ''}`}
                    onClick={() => SetDialogState({ modelType: "course", dialogType: "details", id: courseId })}
                    key={course.ID}
                >
                    <CardContent className="p-5">

                        <div className="flex justify-between items-start mb-5 ">
                            <div className={`flex items-center border border-white/5  shadow-lg shadow-black/60 rounded-xl p-3 overflow-hidden  space-x-4 w-full  relative `}>

                                <div className={`absolute inset-0 z-0 ${getCourseGradientClasses(course.Color, true).bg} ${getCourseGradientClasses(course.Color, true).hover} transition-colors duration-300`} />

                                <div className="flex flex-col w-full min-w-0">
                                    <div className="flex items-center justify-between gap-2">
                                        <h3 className="text-base font-bold text-white line-clamp-1 tracking-tight">{course.Code}</h3>
                                        <div className="flex items-center gap-2 flex-shrink-0">
                                            <Badge variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-300 uppercase tracking-wider px-2 py-0.5">
                                                {course.Credits} credits
                                            </Badge>

                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        className="p-0 w-8 h-8 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                                                        disabled={disabled}
                                                        onClick={(e) => e.stopPropagation()}
                                                    >
                                                        <MoreVertical className="w-4 h-4" />
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end" className="border-white/10 bg-black/90 backdrop-blur-xl glass">
                                                    <DropdownMenuItem
                                                        onClick={() => SetDialogState({ modelType: "course", dialogType: "edit", id: courseId })}
                                                        disabled={disabled}
                                                        className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
                                                    >
                                                        <Edit className="mr-2 w-4 h-4" />
                                                        Edit
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem
                                                        onClick={(e) => {
                                                            e.stopPropagation()
                                                            SetDialogState({ modelType: "course", dialogType: "delete", id: courseId })
                                                        }}
                                                        disabled={disabled}
                                                        className="text-red-400 focus:text-red-300 focus:bg-red-500/10 cursor-pointer"
                                                    >
                                                        <Trash2 className="mr-2 w-4 h-4" />
                                                        Delete
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </div>
                                    </div>

                                    <p className="text-xs text-gray-400 line-clamp-1 mt-0.5 font-medium">{course.Name}</p>
                                </div>
                            </div>

                        </div>

                        <div className="mb-5 space-y-2">
                            <div className="flex items-center space-x-3 text-xs">
                                <div className="p-1 bg-blue-500/10 rounded-md">
                                    <Users className="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
                                </div>
                                <span className="text-gray-300 line-clamp-1 font-medium">{course.Instructor}</span>
                            </div>
                            <div className="flex items-center space-x-3 text-xs">
                                <div className="p-1 bg-purple-500/10 rounded-md">
                                    <Clock className="w-3.5 h-3.5 text-purple-400 flex-shrink-0" />
                                </div>
                                <span className="text-gray-300 line-clamp-1 font-medium">{course.Schedule}</span>
                            </div>
                        </div>

                        <div className="flex gap-4 items-end pt-1">
                            <div className="flex flex-col space-y-2 grow">
                                <div className="flex justify-between text-[10px] uppercase tracking-wider font-medium">
                                    <span className="text-gray-500">Progress</span>
                                    <span className="text-gray-400">
                                        {Math.round(completionPercentage)}%
                                    </span>
                                </div>
                                <Progress value={completionPercentage} className="h-1.5 bg-white/10" />
                            </div>
                            <Badge variant="secondary" className="text-[10px] border-white/10 bg-white/5 text-gray-400 flex-shrink-0 font-medium px-2">
                                {course.Semester}
                            </Badge>
                        </div>
                    </CardContent>
                </GlassCard>

            </div>


        )

    }

    function UserCourseItem({

        courseRO,
        size,
        onAccept,
        onDecline
    }: CourseItemProps) {
        
        if (!courseRO) return null

        if (size === "sm") {
            return (
                <div
                    key={courseRO.ID}
                    onClick={onClick ? onClick : () => SetDialogState({ modelType: "course", dialogType: "details", id: courseRO.ID, item: courseRO, viewMode: "readonly" })}
                    className="flex flex-1 justify-between items-center border border-white/5  shadow-lg shadow-black/60 rounded-xl py-2 px-4 overflow-hidden relative group/item">

                    <div className={`absolute inset-0 z-0 ${getCourseGradientClasses(courseRO.Color, true).bg} ${getCourseGradientClasses(courseRO.Color, true).hover} transition-colors duration-300`} />

                    <div className="flex items-center justify-center gap-2 z-10">

                        {!onAccept && !onDecline && (

                            <div className=" flex items-center justify-center p-1 rounded-full bg-white/10 border border-white/10 shadow-lg shadow-black/40">
                                <ChevronLeft className="w-4 h-4 text-white" strokeWidth={1.5} />
                            </div>
                        )}


                        <div className={`flex flex-col flex-1 z-10`}>

                            <h3 className="text-caption text-gray-300 line-clamp-1 tracking-tight">{courseRO.Code}</h3>


                            <p className="text-body line-clamp-1 font-medium">{courseRO.Name}</p>

                        </div>
                    </div>

                    <div className="flex flex-col items-center gap-1 z-10">
                        {onAccept && (
                            <Button onClick={onAccept} variant="ghost" size="sm" className="flex-1 text-gray-400 bg-gray-500/10 border-gray-500/20 hover:bg-gray-500/20 hover:text-gray-300 hover:border-gray-500/30 transition-all h-9 text-xs font-medium">
                                Accept
                            </Button>
                        )}
                        {onDecline && (
                            <Button onClick={onDecline} variant="ghost" size="sm" className="flex-1 text-gray-400 bg-gray-500/10 border-gray-500/20 hover:bg-gray-500/20 hover:text-gray-300 hover:border-gray-500/30 transition-all h-9 text-xs font-medium">
                                Decline
                            </Button>
                        )}
                    </div>

                </div>
            )
        }

        return (
            <div>
                <GlassCard
                    variant={"outline"}
                    className={`${disabled ? 'opacity-50' : ''}`}
                    onClick={onClick ? onClick : () => SetDialogState({ modelType: "course", dialogType: "details", id: courseRO.ID, item: courseRO, viewMode: "readonly" })}
                    key={courseRO.ID}
                >
                    <CardContent className="p-5">

                        <div className="flex justify-between items-start mb-5 ">
                            <div className={`flex items-center border border-white/5  shadow-lg shadow-black/60 rounded-xl p-3 overflow-hidden  space-x-4 w-full  relative `}>

                                <div className={`absolute inset-0 z-0 ${getCourseGradientClasses(courseRO.Color, true).bg} ${getCourseGradientClasses(courseRO.Color, true).hover} transition-colors duration-300`} />

                                <div className="flex flex-col w-full min-w-0">
                                    <div className="flex items-center justify-between gap-2">
                                        <h3 className="text-base font-bold text-white line-clamp-1 tracking-tight">{courseRO.Code}</h3>
                                        <div className="flex items-center gap-2 flex-shrink-0">
                                            <Badge variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-300 uppercase tracking-wider px-2 py-0.5">
                                                {courseRO.Credits} credits
                                            </Badge>
                                        </div>
                                    </div>

                                    <p className="text-xs text-gray-400 line-clamp-1 mt-0.5 font-medium">{courseRO.Name}</p>
                                </div>
                            </div>

                        </div>

                        <div className="mb-5 space-y-2">
                            <div className="flex items-center space-x-3 text-xs">
                                <div className="p-1 bg-blue-500/10 rounded-md">
                                    <Users className="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
                                </div>
                                <span className="text-gray-300 line-clamp-1 font-medium">{courseRO.Instructor}</span>
                            </div>
                            <div className="flex items-center space-x-3 text-xs">
                                <div className="p-1 bg-purple-500/10 rounded-md">
                                    <Clock className="w-3.5 h-3.5 text-purple-400 flex-shrink-0" />
                                </div>
                                <span className="text-gray-300 line-clamp-1 font-medium">{courseRO.Schedule}</span>
                            </div>
                        </div>


                    </CardContent>
                </GlassCard>

            </div>


        )

    }

    type ScheduleCourseItemProps = CourseItemProps & {
        timeSlots: number[]
        day: string
    }


    function ScheduleCourseItem({
        courseId,
        timeSlots,
        day
    }: ScheduleCourseItemProps) {

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
                className={` absolute left-1 right-1 border rounded-lg hover:translate-y-0.5 backdrop-blur-lg transition-all duration-300 overflow-hidden  cursor-pointer group ${isOn
                    ? 'border-blue-400/50 ring-2 ring-blue-500/30 shadow-[0_0_15px_rgba(59,130,246,0.3)]'
                    : 'bg-white/5 border-white/5 shadow-lg shadow-black/60'
                    }`}
                style={{
                    top: `${topPosition}px`,
                    height: `${height}px`,
                }}
                onClick={() => SetDialogState({ modelType: "course", dialogType: "details", id: courseId })}
            >
                <div className="h-full w-full p-2.5 flex flex-col relative overflow-hidden">
                    {/* Shine effect on hover */}
                    <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                    <div className="flex items-center gap-2 mb-0.5 relative z-10">
                        <div className="font-semibold text-sm text-white drop-shadow-md truncate">
                            {course?.Code}
                        </div>
                    </div>

                    <div className="text-[10px] font-medium text-white/90 truncate relative z-10">
                        {parsedSchedule?.startTimeString} - {parsedSchedule?.endTimeString}
                    </div>

                    {course.Name && height > 50 && (
                        <div className="text-[10px] text-white/80 truncate mt-1 relative z-10">
                            {course?.Name}
                        </div>
                    )}
                </div>
            </div>

        )
    }

    switch (mode) {
        case "default":
            return <DefaultCourseItem courseId={courseId} variant={variant} size={size} />
        case "readonly":
            return <UserCourseItem courseId={courseRO?.ID!} courseRO={courseRO} size={size} onAccept={onAccept} onDecline={onDecline} />
        case "schedule":
            return <ScheduleCourseItem courseId={courseId} timeSlots={timeSlots!} day={day!} />

    }


}

export const CourseItem = memo(BaseCourseItem, (prevProps, nextProps) => {
    return prevProps.courseId === nextProps.courseId
})