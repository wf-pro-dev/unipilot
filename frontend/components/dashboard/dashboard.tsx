import { useCallback, useEffect, useMemo, useState } from "react"
import { StatsCards } from "../dashboard/stats-cards"
import { useAuthContext } from "../provider/auth-provider"
import { BookOpen, Calendar, FileText, MapPin, User, Clock, ArrowRight, StickyNote } from "lucide-react"
import { getNextCourse, parseDeadline } from "@/lib/date-utils"
import { useCoursesBySemester } from "@/hooks/use-courses"
import { CardContent, CardHeader } from "../ui/card"
import { GlassCard } from "../ui/glass-card"
import { EmptyState, HorizontalEmptyState } from "../ui/empty-state"
import { Button } from "../ui/button"
import { useRouter } from "next/navigation"
import { Badge } from "../ui/badge"
import { useAssignments, useDeleteAssignment, useExamAssignments, useUpdateAssignment } from "@/hooks/use-assignments"
import { toast } from "sonner"
import { LogInfo } from "@/wailsjs/runtime"
import { assignment, note } from "@/wailsjs/go/models"
import { format, isAfter, isSameDay } from "date-fns"
import { useNotes } from "@/hooks/use-notes"
import useEmblaCarousel from 'embla-carousel-react'
import { AssignmentEditDialog } from "../assignments/assignment-edit-dialog"
import { AssignmentItemCompact } from "../assignments/assignment-item-compact"

// Helper function to get gradient classes for course colors
// This ensures Tailwind can detect all possible class combinations
const getCourseGradientClasses = (color: string | undefined, isOn: boolean | null) => {
    if (!color || !isOn) {
        return {
            bg: "bg-white/5",
            hover: "hover:bg-white/10 hover:border-white/10"
        }
    }

    // Map color values to gradient classes that Tailwind can detect
    const colorMap: Record<string, { bg: string; hover: string }> = {
        "bg-blue-500": {
            bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-blue-500/20 before:via-blue-500/5 before:to-transparent before:transition-opacity before:duration-500",
            hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-blue-500/40 after:via-blue-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
        },
        "bg-green-500": {
            bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-green-500/20 before:via-green-500/5 before:to-transparent before:transition-opacity before:duration-500",
            hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-green-500/40 after:via-green-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
        },
        "bg-purple-500": {
            bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-purple-500/20 before:via-purple-500/5 before:to-transparent before:transition-opacity before:duration-500",
            hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-purple-500/40 after:via-purple-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
        },
        "bg-red-500": {
            bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-red-500/20 before:via-red-500/5 before:to-transparent before:transition-opacity before:duration-500",
            hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-red-500/40 after:via-red-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
        },
        "bg-orange-500": {
            bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-orange-500/20 before:via-orange-500/5 before:to-transparent before:transition-opacity before:duration-500",
            hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-orange-500/40 after:via-orange-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
        },
        "bg-pink-500": {
            bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-pink-500/20 before:via-pink-500/5 before:to-transparent before:transition-opacity before:duration-500",
            hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-pink-500/40 after:via-pink-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
        }
    }

    return colorMap[color] || {
        bg: "bg-white/5",
        hover: "hover:bg-white/10 hover:border-white/10"
    }
}

export function Dashboard() {
    const router = useRouter()
    const { user } = useAuthContext()
    const { data: coursesBySemester } = useCoursesBySemester(user?.Semester || "")
    const { data: assignments } = useAssignments()
    const { data: exams } = useExamAssignments()
    const { data: notes } = useNotes()

    const deleteMutation = useDeleteAssignment()
    const updateMutation = useUpdateAssignment()

    const [editAssignment, setEditAssignment] = useState<assignment.LocalAssignment | null>(null)
    const [editDialogOpen, setEditDialogOpen] = useState(false)

    // Filter priority assignments (Not Done, sorted by deadline)
    const priorityAssignments = useMemo(() => {
        if (!assignments) return []
        return assignments
            .filter(a => a.StatusName !== "Done")
            .sort((a, b) => new Date(a.Deadline).getTime() - new Date(b.Deadline).getTime())
            .slice(0, 6)
    }, [assignments])

    const upcomingExams = useMemo(() => {
        return exams?.filter((exam) => isAfter(exam.Deadline, new Date()) || isSameDay(exam.Deadline, new Date()))
            .sort((a, b) => new Date(a.Deadline).getTime() - new Date(b.Deadline).getTime())
            .slice(0, 5)
    }, [exams])

    const { course, isOn, until } = getNextCourse(coursesBySemester)

    // Calculate time until class
    const daysUntil = until ? Math.floor(until / 1440) : 0
    const hoursUntil = until ? Math.floor((until % 1440) / 60) : 0
    const minutesUntil = until ? (until % 60) + 1 : 0

    // Notes carousel
    const [emblaRef, emblaApi] = useEmblaCarousel({ align: 'start', containScroll: 'trimSnaps' })

    const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
        const message = "[Frontend] assignment " + assignment.ID + " remote_id " + assignment.RemoteID + " " + column + " changed to " + value
        LogInfo(format(new Date(), "yyyy/MM/dd HH:mm:ssxxx") + " " + message)

        updateMutation.mutate({ assignment, column, value }, {
            onSuccess: () => toast.success("Assignment updated"),
            onError: () => toast.error("Update failed")
        })
    }

    const handleToggleComplete = (assignment: assignment.LocalAssignment) => {
        const newStatus = assignment.StatusName === "Done" ? "Not started" : "Done"
        handleEditAssignment(assignment, "status_name", newStatus)
    }

    const handleDelete = (assignment: assignment.LocalAssignment) => {
        deleteMutation.mutate(assignment, {
            onSuccess: () => toast.success("Assignment deleted"),
            onError: () => toast.error("Delete failed")
        })
    }

    const handleOpenEdit = (assignment: assignment.LocalAssignment) => {
        setEditAssignment(assignment)
        setEditDialogOpen(true)
    }

    return (
        <div className="flex flex-col h-full gap-6">
            <div className="shrink-0">
                <StatsCards />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1 min-h-0">
                {/* Main Column (Left 2/3) */}
                <div className="lg:col-span-2 flex flex-col gap-6 h-full min-h-0">

                    {/* Next Class Card */}
                    <div className="shrink-0">
                        <GlassCard
                            variant="interactive"
                            onClick={() => course && router.push(`/courses?view=schedule`)}
                            className={`${getCourseGradientClasses(course?.Color, course && isOn).bg} ${getCourseGradientClasses(course?.Color, course && isOn).hover} border border-white/5 transition-all duration-300 group overflow-hidden relative`}
                        >

                            <CardHeader className="flex flex-row items-center justify-between pb-4 z-10 relative">
                                <div className="flex items-center space-x-3">
                                    <div className="p-2 rounded-xl bg-white/10 backdrop-blur-md border border-white/5 shadow-inner">
                                        <BookOpen className="h-5 w-5 text-white" />
                                    </div>
                                    <h3 className="text-lg font-semibold text-white">Next Class</h3>
                                </div>
                                {isOn && (
                                    <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30 animate-pulse">
                                        Happening Now
                                    </Badge>
                                )}
                            </CardHeader>

                            <CardContent className="z-10 relative">
                                {course ? (
                                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 group/course">
                                        <div className="space-y-4">
                                            <div className="space-y-1">
                                                <div className="flex items-center gap-2">
                                                    <div className={`w-2 h-2 rounded-full ${course.Color} shadow-[0_0_8px_currentColor]`} />
                                                    <span className="text-sm font-medium text-gray-400 tracking-wider">{course.Code}</span>
                                                </div>
                                                <h2 className="text-3xl font-bold text-white tracking-tight">{course.Name}</h2>
                                            </div>

                                            <div className="flex flex-wrap gap-4 text-sm text-gray-300">
                                                <div className="flex items-center gap-2 bg-white/5 px-3 py-1.5 rounded-lg border border-white/5">
                                                    <Clock className="w-4 h-4 text-blue-400" />
                                                    <span>{course.Schedule}</span>
                                                </div>
                                                <div className="flex items-center gap-2 bg-white/5 px-3 py-1.5 rounded-lg border border-white/5">
                                                    <MapPin className="w-4 h-4 text-purple-400" />
                                                    <span>{course.Location || "Online"}</span>
                                                </div>

                                            </div>
                                        </div>

                                        {until && (
                                            <div className="relative border border-white/10 group-hover:border-white/20 shadow-lg shadow-black/60 rounded-xl flex flex-col items-end justify-center min-w-[140px] p-4">
                                                <div className="absolute inset-0 bg-gradient-to-br from-white/15 to-transparent pointer-events-none group-hover/course:opacity-0 transition-opacity duration-300" />
                                                <span className="text-xs text-gray-400 uppercase tracking-wider font-medium mb-1">{isOn ? "Ends in" : "Starts in"}</span>
                                                <div className="flex items-baseline gap-1">
                                                    {daysUntil > 0 && <span className="text-2xl font-bold text-white">{daysUntil}<span className="text-sm font-normal text-gray-400 ml-1">d</span></span>}
                                                    {hoursUntil > 0 && <span className="text-2xl font-bold text-white">{hoursUntil}<span className="text-sm font-normal text-gray-400 ml-1">h</span></span>}
                                                    <span className="text-2xl font-bold text-white">{minutesUntil}<span className="text-sm font-normal text-gray-400 ml-1">m</span></span>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                ) : (
                                    <EmptyState
                                        icon={BookOpen}
                                        title="No upcoming classes"
                                        description="Enjoy your free time!"
                                        className="py-8"
                                    />
                                )}
                            </CardContent>
                        </GlassCard>
                    </div>

                    {/* Priority Assignments */}
                    <div className="flex flex-col gap-4 flex-1 min-h-0">
                        <div className="flex items-center justify-between px-1 shrink-0">
                            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                                <Clock className="w-5 h-5 text-orange-400" />
                                Priority Tasks
                            </h3>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="text-gray-400 hover:text-white hover:bg-white/10"
                                onClick={() => router.push('/assignments')}
                            >
                                View All <ArrowRight className="ml-1 w-4 h-4" />
                            </Button>
                        </div>

                        {priorityAssignments.length > 0 ? (
                            <div className="grid grid-cols-2 gap-3 overflow-y-auto pr-2 pb-20">
                                {priorityAssignments.map(assignment => (
                                    <AssignmentItemCompact
                                        key={assignment.ID}
                                        assignment={assignment}
                                        onToggleComplete={handleToggleComplete}
                                        onAssignmentClick={(a) => router.push(`/assignments?view=assignment&assignment=${a.ID}`)}
                                    />
                                ))}
                            </div>
                        ) : (
                            <GlassCard className="border-white/5 bg-white/5 py-12 flex-1 flex flex-col justify-center">
                                <HorizontalEmptyState
                                    icon={Clock}
                                    title="All caught up!"
                                    description="No pending assignments due soon."
                                    className="bg-transparent border-0"
                                />
                            </GlassCard>
                        )}
                    </div>
                </div>

                {/* Side Column (Right 1/3) */}
                <div className="flex flex-col gap-6 h-full">

                    {/* Recent Notes */}
                    <div className="space-y-4">
                        <div className="flex items-center justify-between px-1">
                            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                                <StickyNote className="w-5 h-5 text-yellow-400" />
                                Recent Notes
                            </h3>
                            <div className="flex gap-1">
                                <Button variant="ghost" size="icon" className="h-6 w-6 hover:bg-white/10" onClick={() => emblaApi?.scrollPrev()}>
                                    <ArrowRight className="w-4 h-4 rotate-180" />
                                </Button>
                                <Button variant="ghost" size="icon" className="h-6 w-6 hover:bg-white/10" onClick={() => emblaApi?.scrollNext()}>
                                    <ArrowRight className="w-4 h-4" />
                                </Button>
                            </div>
                        </div>

                        <div className="overflow-hidden" ref={emblaRef}>
                            <div className="flex gap-4">
                                {notes && notes.length > 0 ? (
                                    notes.slice(0, 5).map((note) => (
                                        <div key={note.ID} className="flex-[0_0_100%] min-w-0">
                                            <GlassCard
                                                variant="interactive"
                                                onClick={() => router.push('/notes')}
                                                className="border-white/5 bg-white/5 hover:bg-white/10 p-4 h-full"
                                            >
                                                <div className="flex items-start justify-between mb-3">
                                                    <div className={`w-8 h-8 rounded-full flex items-center justify-center bg-white/5 border border-white/5 ${note.Course.Color ? `text-${note.Course.Color.replace('bg-', '').replace('-500', '-400')}` : 'text-gray-400'}`}>
                                                        <FileText className="w-4 h-4" />
                                                    </div>
                                                    <Badge variant="outline" className="border-white/10 text-gray-400 text-[10px]">{note.course_code}</Badge>
                                                </div>
                                                <h4 className="font-medium text-white line-clamp-2 mb-2">{note.title}</h4>
                                                <p className="text-xs text-gray-400 line-clamp-2">{note.subject}</p>
                                            </GlassCard>
                                        </div>
                                    ))
                                ) : (
                                    <div className="w-full">
                                        <GlassCard className="border-white/5 bg-white/5 py-8">
                                            <EmptyState
                                                icon={FileText}
                                                title="No notes yet"
                                                description="Create your first note."
                                                className="bg-transparent border-0 p-0"
                                            />
                                        </GlassCard>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Upcoming Exams */}
                    <div className="flex flex-col gap-4">
                        <h3 className="text-lg font-semibold text-white flex items-center gap-2 px-1 shrink-0">
                            <Calendar className="w-5 h-5 text-red-400" />
                            Upcoming Exams
                        </h3>
                        <div className="space-y-3  overflow-y-auto">
                            {upcomingExams && upcomingExams.length > 0 ? (
                                upcomingExams.map(exam => (
                                    <GlassCard
                                        key={exam.ID}
                                        variant="interactive"
                                        onClick={() => router.push(`/assignments?view=exam&assignment=${exam.ID}`)}
                                        className="border-white/5 bg-white/5 hover:bg-white/10 p-4 transition-all group/exam"
                                    >
                                        <div className="flex items-start gap-3">
                                            <div className="relative aspect-square flex flex-col items-center min-w-[50px] bg-white/5 rounded-lg p-2 border shadow-lg shadow-black/60 border-white/10 group-hover:border-white/15 transition-colors">
                                                <div className="absolute inset-0 bg-gradient-to-br from-white/15 to-transparent pointer-events-none" />

                                                <span className="text-xs font-bold text-red-400 uppercase">{format(parseDeadline(exam.Deadline), "MMM")}</span>
                                                <span className="text-xl font-bold text-white">{format(parseDeadline(exam.Deadline), "d")}</span>
                                            </div>
                                            <div className="flex flex-col min-w-0">
                                                <span className="text-[10px] font-bold text-gray-400 uppercase tracking-wider mb-0.5">{exam.CourseCode}</span>
                                                <h4 className="text-sm font-medium text-white truncate leading-tight mb-1">{exam.Title}</h4>
                                                <div className="flex items-center gap-1.5 text-xs text-gray-400">
                                                    <Clock className="w-3 h-3" />
                                                    {format(parseDeadline(exam.Deadline), "h:mm a")}
                                                </div>
                                            </div>
                                        </div>
                                    </GlassCard>
                                ))
                            ) : (
                                <GlassCard className="border-white/5 bg-white/5 py-4">
                                    <EmptyState
                                        icon={Calendar}
                                        title="No exams"
                                        description="Time to relax!"
                                        className="bg-transparent border-0 p-0"
                                    />
                                </GlassCard>
                            )}
                        </div>
                    </div>




                </div>
            </div>

            <AssignmentEditDialog
                open={editDialogOpen}
                setOpen={setEditDialogOpen}
                assignment={editAssignment}
                onEdit={handleEditAssignment}
            />
        </div>
    )
}
