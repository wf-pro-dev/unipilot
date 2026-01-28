import { models } from "@/wailsjs/go/models"
import { useState, useMemo } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { EmptyState } from "@/components/ui/empty-state"
import { differenceInMinutes, format, isAfter, isBefore } from "date-fns"
import { CourseItem } from "./course-item"
import { useAuthContext } from "../provider/auth-provider"
import { ParsedSchedule, parseSchedule } from "@/lib/date-utils"
import { BookOpen, FileText } from "lucide-react"

interface CoursesScheduleProps {
    courses: models.LocalCourse[]
    onCourseClick: (course: models.LocalCourse) => void
    selectedSemester: string
}



const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const HOURS = Array.from({ length: 24 }, (_, i) => i)



interface CourseWithSchedule extends models.LocalCourse {
    parsedSchedule: ParsedSchedule | null
}


function formatTime(hour: number): string {
    if (hour === 0) return '12:00 AM'
    if (hour < 12) return `${hour}:00 AM`
    if (hour === 12) return '12:00 PM'
    return `${hour - 12}:00 PM`
}

function CoursesSchedule({ courses, selectedSemester, onCourseClick }: CoursesScheduleProps) {
    const { user } = useAuthContext()
    if (!user) return null

    // Filter courses by semester and parse schedules
    const scheduledCourses = useMemo((): CourseWithSchedule[] => {
        if (!courses) return []

        return courses
            .filter(course => course.Semester === selectedSemester)
            .map(course => ({
                ...course,
                parsedSchedule: parseSchedule(course.Schedule)
            }))
            .filter((course): course is CourseWithSchedule => course.parsedSchedule !== null)
    }, [courses, selectedSemester])

    const asyncCourses = useMemo(() => {
        return courses.filter(course => course.Semester === selectedSemester && (course.Schedule === "Async" || course.Schedule === "Asynchronous"))
    }, [courses, selectedSemester])


    const timeSlots = Array.from({ length: 12 }, (_, i) => i + 8) // 8 AM to 8 PM

    const now = new Date()
    const currentHour = now.getHours()
    const currentMinute = now.getMinutes()
    const currentTop = (currentHour - 8) * 60 + currentMinute

    const hasScheduledCourses = scheduledCourses.length > 0
    const hasAsyncCourses = asyncCourses.length > 0
    const isEmpty = !hasScheduledCourses && !hasAsyncCourses

    return (
        <div className="flex flex-1 space-y-6">

            {hasScheduledCourses && (
                <GlassCard variant="board" className=" overflow-hidden">
                    <CardContent className="p-0">
                        <div className="overflow-auto custom-scrollbar">
                            <div className="flex min-w-full">
                                {/* Time column */}
                                <div className="flex-shrink-0 w-20 bg-white/[0.02] border-r border-white/5">
                                    <div className="h-12 p-3 text-xs font-medium text-gray-400 uppercase tracking-wider text-center border-b border-white/5 flex items-center justify-center">
                                        Time
                                    </div>
                                    {timeSlots.map(hour => (
                                        <div
                                            key={hour}
                                            className="h-[60px] p-2 text-xs font-medium text-gray-500 text-center border-b border-white/5 flex items-center justify-center"
                                        >
                                            {formatTime(hour)}
                                        </div>
                                    ))}
                                </div>

                                {/* Days columns */}
                                {DAYS.map((day, dayIndex) => {
                                    return (
                                        <div key={day} className="flex-1 min-w-[140px] relative border-r border-white/5 last:border-r-0">
                                            <div className="h-12 p-3 text-xs font-medium text-gray-400 uppercase tracking-wider text-center border-b border-white/5 flex items-center justify-center bg-white/[0.02]">
                                                {day}
                                            </div>
                                            <div className="relative bg-transparent">
                                                {/* Hour grid lines */}
                                                {timeSlots.map(hour => (
                                                    <div
                                                        key={hour}
                                                        className="h-[60px] border-b border-white/5"
                                                    />
                                                ))}

                                                {/* Time indicator */}
                                                {currentTop > 0 && now.getDay() == dayIndex && currentHour < 20 && (
                                                    <div className="absolute -left-[1px] -right-[1px] h-[2px] bg-blue-500 z-50 shadow-[0_0_8px_rgba(59,130,246,0.8)]" style={{
                                                        top: `${currentTop}px`
                                                    }}>
                                                        <div className="absolute -left-1.5 -top-1.5 w-3 h-3 rounded-full bg-blue-500 shadow-md" />
                                                    </div>
                                                )}

                                                {/* Course blocks */}
                                                {scheduledCourses
                                                    .filter(course => course.parsedSchedule?.days.includes(dayIndex))
                                                    .map((course, index) => {
                                                        return (
                                                            <CourseItem key={course.ID} mode="schedule" courseId={course.ID} timeSlots={timeSlots} day={day} />
                                                        )
                                                    })
                                                }
                                            </div>
                                        </div>
                                    )
                                })}
                            </div>
                        </div>
                    </CardContent>
                </GlassCard>
            )}

            {hasAsyncCourses && (
                <div className="space-y-4">
                    <h3 className="text-lg font-semibold text-white px-1">Asynchronous Courses</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {asyncCourses.map((course) => (
                            <CourseItem key={course.ID} courseId={course.ID} />
                        ))}
                    </div>
                </div>
            )}

            {isEmpty && (
                <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">

                    <EmptyState
                        icon={BookOpen}
                        title="No courses found"
                        description={`No courses found for ${selectedSemester} semester`}
                        className="flex-1 items-center"
                    />
                </div>
            )}
        </div>
    )
}

export default CoursesSchedule
