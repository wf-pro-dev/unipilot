import { useCallback, useEffect, useMemo, useState } from "react"
import { StatsCards } from "../dashboard/stats-cards"
import { WelcomeSection } from "./welcome-section"
import { useAuthContext } from "../provider/auth-provider"
import { BookOpen, Calendar, ClipboardList, FileText, MapPin, Users, CheckCircle2, Dot, ChevronLeft, ChevronRight, Clock } from "lucide-react"
import { getNextCourse } from "@/lib/date-utils"
import { useCoursesBySemester } from "@/hooks/use-courses"
import { Card, CardContent, CardHeader, CardFooter } from "../ui/card"
import { Button } from "../ui/button"
import { useRouter } from "next/navigation"
import { Badge } from "../ui/badge"
import { useExamAssignments, useWeekAssignments } from "@/hooks/use-assignments"
import { CalendarItem } from "../assignments/calendar-item"
import { useUpdateAssignment } from "@/hooks/use-assignments"
import { toast } from "sonner"
import { LogInfo } from "@/wailsjs/runtime"
import { assignment, note } from "@/wailsjs/go/models"
import { format, isAfter } from "date-fns"
import { DndProvider, useDrop } from "react-dnd"
import { HTML5Backend } from "react-dnd-html5-backend"
import { useNotes } from "@/hooks/use-notes"
import useEmblaCarousel from 'embla-carousel-react'
import Link from "next/link"


export function Dashboard() {
    const router = useRouter()
    const { user } = useAuthContext()
    const { data: coursesBySemester } = useCoursesBySemester(user?.Semester || "")
    const { data: week_assignments } = useWeekAssignments()
    const [selectedIndex, setSelectedIndex] = useState(0)
    const { data: exams } = useExamAssignments()
    const UpcomingExams = useMemo(() => {
        return exams?.filter((exam) => isAfter(exam.Deadline, new Date()))
    }, [exams])
    const { data: notes } = useNotes()
    var notesPerPage = 2
    const notesPages = []
    for (let i = 0; i < (notes || []).length; i += notesPerPage) {
        notesPages.push(notes?.slice(i, i + notesPerPage))
    }

    const { course, isOn, until } = getNextCourse(coursesBySemester)
    var daysUntil = Math.floor(until! / 1440)
    var hoursUntil = Math.floor((until! % 1440) / 60)
    var minutesUntil = (until! % 60) + 1

    const [emblaRef, emblaApi] = useEmblaCarousel({
        align: 'start',
        containScroll: 'trimSnaps',
        dragFree: false,
        skipSnaps: false
    })

    // Carousel navigation functions
    const scrollPrev = useCallback(() => {
        if (emblaApi) emblaApi.scrollPrev()
    }, [emblaApi])

    const scrollNext = useCallback(() => {
        if (emblaApi) emblaApi.scrollNext()
    }, [emblaApi])

    const hour = new Date().getHours()

    let greeting = "Good morning"
    if (hour >= 12 && hour < 17) {
        greeting = "Good afternoon"
    } else if (hour >= 17) {
        greeting = "Good evening"
    }

    const updateMutation = useUpdateAssignment()

    const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
        const message = "[Frontend] assignment " + assignment.ID + " remote_id " + assignment.RemoteID + " " + column + " changed to " + value
        LogInfo(format(new Date(), "yyyy/MM/dd HH:mm:ssxxx") + " " + message)

        // Use the optimistic update mutation
        updateMutation.mutate({
            assignment,
            column,
            value
        }, {
            onSuccess: () => {
                toast.success("Assignment updated successfully")
            },
            onError: () => {
                toast.error("Assignment update failed")
            }
        })
    }

    const onMoveAssignment = (assignment: assignment.LocalAssignment, status: string) => {
        if (assignment.StatusName === status) {
            return
        }
        handleEditAssignment(assignment, "status_name", status)
    }


    // Track current slide
    const onSelect = useCallback(() => {
        if (!emblaApi) return
        setSelectedIndex(emblaApi.selectedScrollSnap())
    }, [emblaApi])


    useEffect(() => {
        if (!emblaApi) return
        onSelect()
        emblaApi.on('select', onSelect)
        return () => {
            emblaApi.off('select', onSelect)
        }
    }, [emblaApi, onSelect])



    const StatusCard = ({ status, className, color, overColor }: { status: string, className?: string, color?: string, overColor?: string }) => {

        const [{ isOver }, drop] = useDrop({
            accept: "assignment",
            drop: (item: { assignment: assignment.LocalAssignment }) => {
                if (item.assignment) {
                    onMoveAssignment(item.assignment, status)
                }
            },
            collect: (monitor) => ({
                isOver: monitor.isOver(),
            }),
        })

        const assignments = useMemo(() => {
            return (week_assignments || [])
                .filter((assignment) => assignment.StatusName === status)
                .sort((a, b) => {
                    return new Date(a.Deadline).getTime() - new Date(b.Deadline).getTime()
                })
        }, [week_assignments, status])

        return (
            <div
                key={status}
                ref={drop}

                className={`
                    flex flex-col 
                    items-center 
                    p-3 rounded-lg 
                    border ${className} ${isOver ? overColor : color}
                    space-y-2
                `}
            >

                <div className={`text-[10px] font-medium text-white bg-transparent border ${color} p-1.5 rounded-full`}>
                    {status} ({assignments.length})
                </div>


                <div className="w-full h-full">
                    {assignments.length > 0 ? (
                        assignments
                            .slice(0, 1)
                            .map((assignment) => (
                                <div key={assignment.ID}>
                                    <CalendarItem assignment={assignment} onEdit={handleEditAssignment} onAssignmentClick={() => { }} />
                                </div>
                            ))
                    ) : (
                        <div className="flex flex-col items-center justify-center h-full">
                            <CheckCircle2 className="h-8 w-8 text-white/20 mx-auto mb-4" />
                        </div>
                    )}

                </div>

            </div>
        )
    }

    return (
        <DndProvider backend={HTML5Backend}>
            <div className="flex-1 space-y-5 mb-4">
                <div className="">
                    <p className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                        {greeting}, {user?.Username} 👋
                    </p>
                </div>

                <StatsCards />


                <div className="flex-1 grid grid-cols-4 grid-rows-3 gap-5">


                    <div className="grid grid-cols-2 gap-5 col-span-4 row-span-2">
                        <Card className="flex flex-col glass" >
                            <CardHeader className="flex flex-row items-center justify-between">
                                <div className="flex flex-row items-center space-x-2">
                                    <div className={`p-1 rounded-lg bg-white/20`}>
                                        <ClipboardList className={`h-4 w-4 text-white`} strokeWidth={1.5} />
                                    </div>
                                    <p className="text-sm font-medium text-white">
                                        Upcoming Assignments
                                    </p>
                                </div>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="w-24 bg-transparent border-gray-600 text-xs"
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        router.push(`/assignments?view=week`)
                                    }}
                                >
                                    <ClipboardList className="mr-1 w-2 h-2" />
                                    View All
                                </Button>
                            </CardHeader>
                            <CardContent className="flex-1 grid grid-cols-3 gap-4">

                                <StatusCard status="Not started" className="bg-gray-800/40" color="border-gray-600" overColor="border-gray-300" />
                                <StatusCard status="In progress" className="bg-blue-500/20" color="border-blue-500" overColor="border-blue-300" />
                                <StatusCard status="Done" className="bg-green-500/20" color="border-green-500" overColor="border-green-300" />

                            </CardContent>
                        </Card>

                        <div className="grid grid-cols-2 gap-5">

                            <Card
                                onClick={(e) => {
                                    e.stopPropagation()
                                    router.push(`/courses?view=schedule&course=${course?.Code}`)
                                }}

                                className="h-full flex cursor-pointer flex-col glass hover:scale-105 transition-all duration-300 " >
                                <CardHeader className="flex flex-row items-center justify-between">
                                    <div className="flex flex-row items-center space-x-2">
                                        <div className={`p-1 rounded-lg bg-white/20`}>
                                            <BookOpen className={`h-4 w-4 text-white`} strokeWidth={1.5} />
                                        </div>
                                        <p className="text-sm font-medium text-white">
                                            Upcoming Class
                                        </p>
                                    </div>
                                    {isOn && (
                                        <Badge className="text-xs text-blue-500 bg-blue-500/10 border-blue-500">
                                            On Going
                                        </Badge>
                                    )}
                                </CardHeader>
                                {course ? (
                                    <CardContent className="flex-1 space-y-4">
                                        <div className="flex flex-row items-center gap-2">
                                            <div className={`h-2 w-2  rounded-full ${course?.Color}`} />
                                            <div className="font-semibold truncate">
                                                {course?.Code}
                                            </div>
                                        </div>

                                        {until && (
                                            <div className="flex flex-row space-x-2 items-center">
                                                <Clock className="w-4 h-4 text-white" />

                                                <div className="flex flex-row space-x-1 text-xs text-white">
                                                    {daysUntil > 0 && (
                                                        <div className="flex flex-row space-x-1">
                                                            <span className="font-semibold">{daysUntil}</span>
                                                            <span>day{daysUntil > 1 ? "s," : ","}</span>
                                                        </div>

                                                    )}

                                                    {hoursUntil > 0 && (
                                                        <div className="flex flex-row space-x-1">
                                                            <span className="font-semibold">{hoursUntil}</span>
                                                            <span>hour{hoursUntil > 1 ? "s," : ","}</span>
                                                        </div>
                                                    )}

                                                    <span className="font-semibold">{minutesUntil}</span>
                                                    <span>minute{minutesUntil > 1 ? "s" : ""} left{isOn ? " in class" : ""}</span>

                                                </div>


                                            </div>
                                        )}

                                        <div className="flex space-x-2 items-center">
                                            <Calendar className="w-4 h-4 text-white" />
                                            <div className="text-xs text-white">
                                                {course?.Schedule}
                                            </div>
                                        </div>
                                        <div className="flex space-x-2 items-center">
                                            <Users className="w-4 h-4 text-white" />
                                            <div className="text-xs text-white">
                                                {course?.Instructor}
                                            </div>
                                        </div>


                                        <div className="flex space-x-2 items-center">
                                            <MapPin className="w-4 h-4 text-white" />
                                            <div className="text-xs text-white">
                                                {course?.Location || "Online"}
                                            </div>
                                        </div>
                                    </CardContent>
                                ) : (
                                    <div className="flex flex-col gap-4 items-center justify-center h-full">
                                        <CheckCircle2 className="h-12 w-12 text-white/20 mx-auto" />
                                        <p className="text-xs text-gray-400">No course found</p>
                                    </div>
                                )}
                                <CardFooter >
                                    {course ? (
                                        <div className="grid grid-cols-5 w-full gap-2">
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className="flex-1 col-span-3 bg-transparent border-gray-600 text-xs"
                                                onClick={(e) => {
                                                    e.stopPropagation()
                                                    router.push(`/assignments?view=list&course=${course?.Code}`)
                                                }}
                                            >
                                                <ClipboardList className="w-2 h-2" />
                                                Assignments
                                            </Button>

                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className="flex-1 col-span-2 bg-transparent border-gray-600 text-xs"
                                                onClick={(e) => {
                                                    e.stopPropagation()
                                                    router.push(`/notes?course=${course?.Code}`)
                                                }}
                                            >
                                                <FileText className="w-2 h-2" />
                                                Notes
                                            </Button>
                                        </div>
                                    ) : (
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            className="flex-1 bg-transparent border-gray-600 text-xs"
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                router.push(`/courses?view=list`)
                                            }}
                                        >
                                            <BookOpen className="mr-1 w-2 h-2" />
                                            See Courses
                                        </Button>
                                    )}

                                </CardFooter>
                            </Card>

                            <Card className="flex flex-col glass" >
                                <CardHeader className="flex flex-row items-center space-x-2">
                                    <div className={`p-1 rounded-lg bg-white/20`}>
                                        <Calendar className={`h-4 w-4 text-white`} strokeWidth={1.5} />
                                    </div>
                                    <p className="text-sm font-medium text-white">
                                        Upcoming Exams
                                    </p>
                                </CardHeader>
                                <CardContent className="flex-1 space-y-4 overflow-scroll">
                                    {UpcomingExams.length > 0 ? (
                                        UpcomingExams
                                            .map((exam) => (
                                                <div key={exam.ID} className="flex flex-row items-center">
                                                    <Dot className="h-6 w-6 text-white" />
                                                    <div className="flex flex-col">
                                                        <div className="text-xs text-gray-400">
                                                            {exam.CourseCode}
                                                        </div>
                                                        <div className="text-sm font-medium text-white">
                                                            {exam.Title}
                                                        </div>
                                                    </div>
                                                </div>
                                            ))
                                    ) : (
                                        <div className="flex flex-col gap-4 items-center justify-center h-full">
                                            <Calendar className="h-12 w-12 text-white/20 mx-auto" />
                                            <p className="text-xs text-gray-400">No exams found</p>
                                        </div>
                                    )}
                                </CardContent>
                                <CardFooter className="flex space-x-2">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="flex-1 bg-transparent border-gray-600 text-xs"
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            router.push(`/assignments?view=exam`)
                                        }}
                                    >
                                        <ClipboardList className="mr-1 w-2 h-2" />
                                        View All
                                    </Button>
                                </CardFooter>
                            </Card>
                        </div>
                    </div>
                    <div className="grid grid-cols-2 gap-5 col-span-4 ">
                        <div className="flex glass" />

                        <Card className="flex flex-col glass" >
                            <CardHeader className="pb-4 flex flex-row items-center justify-between">
                                <div className="flex flex-row items-center space-x-2">
                                    <div className={`p-1 rounded-lg bg-white/20`}>
                                        <FileText className={`h-4 w-4 text-white`} strokeWidth={1.5} />
                                    </div>
                                    <p className="text-sm font-medium text-white">
                                        Latest Notes
                                    </p>
                                </div>
                                <div className="flex flex-row items-center space-x-4">
                                    <div>
                                        {notesPages.length > 1 && (
                                            <div className="flex flex-row items-center space-x-2">
                                                <Button
                                                    variant="outline"
                                                    size="icon"
                                                    className="rounded-full z-10 h-6 w-6 bg-gray-800/50 border border-gray-600"
                                                    onClick={scrollPrev}
                                                >
                                                    <ChevronLeft className="h-4 w-4" />
                                                </Button>

                                                <Button
                                                    variant="outline"
                                                    size="icon"
                                                    className="rounded-full z-10 h-6 w-6 bg-gray-800/50 border border-gray-600"
                                                    onClick={scrollNext}
                                                >
                                                    <ChevronRight className="h-4 w-4" />
                                                </Button>
                                            </div>
                                        )}
                                    </div>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="w-24 bg-transparent border-gray-600 text-xs"
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            router.push(`/notes`)
                                        }}
                                    >
                                        <FileText className="mr-1 w-2 h-2" />
                                        View All
                                    </Button>



                                </div>
                            </CardHeader>
                            <CardContent className="flex flex-row gap-4 ">

                                {notes && notes.length > 0 ? (
                                    <div className="relative w-full">
                                        <div className="overflow-hidden" ref={emblaRef}>
                                            <div className="flex">
                                                {notesPages.map((page, pageIndex) => (
                                                    <div
                                                        key={pageIndex}
                                                        className="flex-none w-full min-w-0"
                                                    >
                                                        <div className="grid grid-cols-2 gap-4 w-full">
                                                            {page?.map((note: note.LocalNote) => {

                                                                return (
                                                                    <div key={note.ID} className="flex-1 w-full p-2 rounded-lg bg-gray-800/50 border border-gray-600">
                                                                        <div className="flex flex-col space-y-2">
                                                                            <div className="text-sm font-medium text-white line-clamp-1">
                                                                                {note.Title}
                                                                            </div>
                                                                            <div className="flex flex-row items-center justify-between">

                                                                                <div className="flex flex-row items-center space-x-2">
                                                                                    <div className={`w-2 h-2 rounded-full ${note.Course.Color}`} />
                                                                                    <div className="text-xs text-white">
                                                                                        {note.CourseCode}
                                                                                    </div>
                                                                                </div>
                                                                                <div className="text-xs text-gray-400">
                                                                                    {note.Subject}
                                                                                </div>
                                                                            </div>

                                                                        </div>
                                                                    </div>
                                                                )
                                                            })}

                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>



                                    </div>

                                ) : (
                                    <div className="flex flex-row gap-4 items-center justify-center h-full">
                                        <FileText strokeWidth={1.5} className="h-10 w-10 text-white/20 mx-auto" />
                                        <p className="text-xs text-gray-400">No notes found</p>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>


                </div >
            </div >
        </DndProvider >
    )
}