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
import { models } from "@/wailsjs/go/models"
import { format, isAfter, isSameDay } from "date-fns"
import { useNotes } from "@/hooks/use-notes"
import useEmblaCarousel from 'embla-carousel-react'
import { AssignmentEditDialog } from "../assignments/assignment-edit-dialog"
import { AssignmentItemCompact } from "../assignments/assignment-item-compact"
import { AssignmentItem } from "../assignments/assignment-item"
import { cn } from "@/lib/utils"

// Helper function to get gradient classes for course colors
// This ensures Tailwind can detect all possible class combinations
const getCourseGradientClasses = (color: string | undefined, isOn: boolean | null) => {
    if (!color || !isOn) {
        return {
            bg: "",
            hover: ""
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
        hover: "over:bg-white/10 hover:border-white/10"
    }
}

export function Dashboard() {
    const router = useRouter()
    const { user } = useAuthContext()
    const { data: coursesBySemester } = useCoursesBySemester(user?.Semester || "")
    const { data: assignments } = useAssignments()
    const { data: exams } = useExamAssignments()
    const { data: notes } = useNotes()

    // Filter priority assignments (Not Done, sorted by deadline)
    const dashboardAssignmentIDs = useMemo(() =>
        assignments
            ?.filter(a => a.Status !== "Done")
            .slice(0, 6)
            .sort((a, b) => new Date(a.Deadline).getTime() - new Date(b.Deadline).getTime())
            .map(a => a.ID)
        || []
        , [assignments])

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

    const SectionTitle = ({ title }: { title: string }) => {
        return (
            <div className="flex items-center">
                <p className="text-body font-medium text-gray-400 uppercase tracking-wider">{title}</p>
            </div>
        )
    }


    return (

        <div className="flex flex-col flex-1 gap-6">

            <div className="shrink-0">
                <StatsCards />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1">
                {/* Main Column (Left 2/3) */}
                <div className="lg:col-span-2 flex flex-col flex-1 gap-6">

                    {/* Next Class Card */}
                    <div className={cn("flex", course ? "flex-shrink-0" : "flex-1")}>
                        <GlassCard
                            variant={course ? "outline" : "board"}
                            onClick={() => course && router.push(`/courses?view=schedule?course=${course.Code}`)}
                            className={cn("relative group", getCourseGradientClasses(course?.Color, course && isOn).bg, getCourseGradientClasses(course?.Color, course && isOn).hover)}
                        >

                            <CardHeader className="flex flex-row items-center justify-between pb-4 z-10 relative">
                                <SectionTitle title="Next Class" />
                                {isOn && (
                                    <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30 animate-pulse">
                                        Happening Now
                                    </Badge>
                                )}
                            </CardHeader>

                            <CardContent className="flex flex-col flex-1 z-10 relative">
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
                                    <HorizontalEmptyState
                                        icon={BookOpen}
                                        title="No upcoming classes"
                                        description="Enjoy your free time!"
                                        className="flex-1 items-center"
                                    />
                                )}
                            </CardContent>
                        </GlassCard>
                    </div>

                    {/* Priority Assignments */}
                    <div className="flex flex-1">
                        <div className="flex flex-1 flex-col gap-4">
                            <div className="flex items-center justify-between px-1 shrink-0">
                                <SectionTitle title="Priority Tasks" />
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="text-body text-text-caption hover:text-white hover:bg-white/10"
                                    onClick={() => router.push('/assignments')}
                                >
                                    View All <ArrowRight className="ml-1 w-4 h-4" />
                                </Button>
                            </div>

                            <GlassCard variant="board">
                                {dashboardAssignmentIDs.length > 0 ? (
                                    <div className="grid grid-cols-2 gap-3 overflow-y-auto">
                                        {dashboardAssignmentIDs.map(assignmentId => (
                                            <AssignmentItem
                                                key={assignmentId}
                                                mode="ghost"
                                                assignmentId={assignmentId}
                                            />
                                        ))}
                                    </div>
                                ) : (
                                    <HorizontalEmptyState
                                        icon={Clock}
                                        title="All caught up!"
                                        description="No pending assignments due soon."
                                        className="flex-1"
                                    />

                                )}
                            </GlassCard>
                        </div>
                    </div>

                </div>

                {/* Side Column (Right 1/3) */}
                <div className="flex flex-col flex-1 gap-6">

                    {/* Recent Notes */}
                    <div className="flex flex-1 flex-col gap-4">
                        <div className="flex items-center justify-between px-1">
                            <SectionTitle title="Recent Notes" />
                            <div className="flex gap-1">
                                <Button variant="ghost" size="icon" className="h-6 w-6 hover:bg-white/10" onClick={() => emblaApi?.scrollPrev()}>
                                    <ArrowRight className="w-4 h-4 rotate-180" />
                                </Button>
                                <Button variant="ghost" size="icon" className="h-6 w-6 hover:bg-white/10" onClick={() => emblaApi?.scrollNext()}>
                                    <ArrowRight className="w-4 h-4" />
                                </Button>
                            </div>
                        </div>

                        <div className="flex flex-1 overflow-hidden" ref={emblaRef}>

                            {notes && notes.length > 0 ? (
                                notes.slice(0, 5).map((note) => (
                                    <div key={note.ID} className="flex-[0_0_100%] min-w-0 px-1">
                                        <GlassCard
                                            variant="board"
                                            onClick={() => router.push('/notes')}
                                            className="border-white/5 bg-white/5 hover:bg-white/10 p-4 h-full flex flex-col justify-between"
                                        >
                                            <div className="flex items-start justify-between mb-3">
                                                <div className={`w-8 h-8 rounded-full flex items-center justify-center bg-white/5 border border-white/5 ${note.Course.Color ? `text-${note.Course.Color.replace('bg-', '').replace('-500', '-400')}` : 'text-gray-400'}`}>
                                                    <FileText className="w-4 h-4" />
                                                </div>
                                                <Badge variant="outline" className="border-white/10 text-gray-400 text-[10px]">{note.CourseCode}</Badge>
                                            </div>
                                            <div>
                                                <h4 className="font-medium text-white line-clamp-2 mb-2">{note.Title}</h4>
                                                <p className="text-xs text-gray-400 line-clamp-2">{note.Subject}</p>
                                            </div>
                                        </GlassCard>
                                    </div>
                                ))
                            ) : (

                                <GlassCard variant="board">
                                    <EmptyState
                                        icon={FileText}
                                        title="No notes yet"
                                        description="Create your first note."
                                        className="flex-1"
                                    />
                                </GlassCard>
                            )}
                        </div>

                    </div>

                    {/* Upcoming Exams */}
                    <div className="flex flex-col flex-1 gap-4">
                        <SectionTitle title="Upcoming Exams" />
                        <div className="flex flex-1 overflow-y-auto">
                            <GlassCard variant="board">
                                {upcomingExams && upcomingExams.length > 0 ? (

                                    upcomingExams.map(exam => (
                                        <AssignmentItem key={exam.ID} mode="ghost" assignmentId={exam.ID} />
                                    ))
                                ) : (

                                    <EmptyState
                                        icon={Calendar}
                                        title="No exams"
                                        description="Time to relax!"
                                        className="flex-1"
                                    />

                                )}
                            </GlassCard>
                        </div>
                    </div>




                </div>
            </div>



        </div>
    )
}
