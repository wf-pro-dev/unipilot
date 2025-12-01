import { assignment, course as Course } from "@/wailsjs/go/models"
import { useState, useMemo } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { EmptyState } from "@/components/ui/empty-state"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { differenceInMinutes, format, isAfter, isBefore, isSameDay } from "date-fns"
import { CourseItem } from "./course-item"
import { useAuthContext } from "../provider/auth-provider"
import { ParsedSchedule, parseSchedule } from "@/lib/date-utils"
import { BookOpen, FileText } from "lucide-react"

interface CoursesScheduleProps {
    courses: Course.LocalCourse[]
    onCourseClick: (course: Course.LocalCourse) => void
    onEdit: (course: Course.LocalCourse, column: string, value: string) => void
    onDelete: (course: Course.LocalCourse) => void
}



const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const HOURS = Array.from({ length: 24 }, (_, i) => i)



interface CourseWithSchedule extends Course.LocalCourse {
    parsedSchedule: ParsedSchedule | null
}


function formatTime(hour: number): string {
    if (hour === 0) return '12:00 AM'
    if (hour < 12) return `${hour}:00 AM`
    if (hour === 12) return '12:00 PM'
    return `${hour - 12}:00 PM`
}

function CoursesSchedule({ courses, onCourseClick, onEdit, onDelete }: CoursesScheduleProps) {
    const { user } = useAuthContext()
    if (!user) return null
    const [selectedSemester, setSelectedSemester] = useState<string>(user.Semester)

    // Get unique semesters
    const semesters = useMemo(() => {
        if (!courses) return []
        const uniqueSemesters = [...new Set([...courses.map(course => course.Semester), user.Semester])]

        // Custom sorting function for semester format: "<season> <year>"
        const sortSemesters = (a: string, b: string) => {
            const parseSemester = (semester: string) => {
                const parts = semester.split(' ')
                if (parts.length !== 2) return { season: '', year: 0, seasonPriority: 0 }

                const season = parts[0].toUpperCase()
                const year = parseInt(parts[1])

                // Season priority: SPRING = 1, SUMMER = 2, FALL = 3
                const seasonPriority: Record<string, number> = {
                    'FALL': 1,
                    'SUMMER': 2,
                    'SPRING': 3
                }

                return {
                    season,
                    year,
                    seasonPriority: seasonPriority[season] || 0
                }
            }

            const semesterA = parseSemester(a)
            const semesterB = parseSemester(b)

            // First sort by year (descending)
            if (semesterA.year !== semesterB.year) {
                return semesterB.year - semesterA.year
            }

            // Then sort by season (SPRING -> SUMMER -> FALL)
            return semesterA.seasonPriority - semesterB.seasonPriority
        }

        return uniqueSemesters.sort(sortSemesters)
    }, [courses])

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

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between px-1">
                <h2 className="text-2xl font-bold text-white tracking-tight">Weekly Schedule</h2>
                <div className="w-64">
                    <Select value={selectedSemester} onValueChange={setSelectedSemester}>
                        <SelectTrigger className="bg-white/5 border-white/10 text-white hover:bg-white/10 transition-all duration-200 backdrop-blur-sm">
                            <SelectValue placeholder="Filter by semester" />
                        </SelectTrigger>
                        <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl text-gray-200">
                            {semesters.map(semester => (
                                <SelectItem 
                                    key={semester} 
                                    value={semester}
                                    className="focus:bg-white/10 focus:text-white cursor-pointer"
                                >
                                    {semester}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            </div>

            <GlassCard className="border-white/5 bg-white/5 overflow-hidden shadow-xl shadow-black/20">
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
                                                    if (!course.parsedSchedule) return null
                                                    
                                                    var isOn = false
                                                    const startHour = course.parsedSchedule.startHour
                                                    const startMinute = course.parsedSchedule.startMinute
                                                    const endHour = course.parsedSchedule.endHour
                                                    const endMinute = course.parsedSchedule.endMinute

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
                                                            key={`${course.ID}-${index}`}
                                                            className={` absolute left-1 right-1 border rounded-lg hover:translate-y-0.5 backdrop-blur-lg transition-all duration-300 overflow-hidden  cursor-pointer group ${
                                                                isOn 
                                                                    ? 'border-blue-400/50 ring-2 ring-blue-500/30 shadow-[0_0_15px_rgba(59,130,246,0.3)]' 
                                                                    : 'bg-white/5 border-white/5 shadow-lg shadow-black/60'
                                                            }`}
                                                            style={{
                                                                top: `${topPosition}px`,
                                                                height: `${height}px`,
                                                            }}
                                                            onClick={() => onCourseClick(course)}
                                                        >
                                                            <div className="h-full w-full p-2.5 flex flex-col relative overflow-hidden">
                                                                {/* Shine effect on hover */}
                                                                <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                                                                
                                                                <div className="flex items-center gap-2 mb-0.5 relative z-10">
                                                                    <div className="font-semibold text-sm text-white drop-shadow-md truncate">
                                                                        {course.Code}
                                                                    </div>
                                                                </div>

                                                                <div className="text-[10px] font-medium text-white/90 truncate relative z-10">
                                                                    {course.parsedSchedule?.startTimeString} - {course.parsedSchedule?.endTimeString}
                                                                </div>
                                                                
                                                                {course.Name && height > 50 && (
                                                                    <div className="text-[10px] text-white/80 truncate mt-1 relative z-10">
                                                                        {course.Name}
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </div>
                                                    )
                                                })
                                            }
                                        </div>
                                    </div>
                                )
                            })}
                        </div>
                    </div>

                    {scheduledCourses.length === 0 && (
                        <div className="py-20 flex justify-center items-center">
                             <EmptyState
                                icon={BookOpen}
                                title="No courses found"
                                description={`No courses found for ${selectedSemester} semester`}
                                className="bg-transparent border-0"
                            />
                        </div>
                    )}
                </CardContent>
            </GlassCard>

            {asyncCourses.length > 0 && (
                 <div className="space-y-4">
                    <h3 className="text-lg font-semibold text-white px-1">Asynchronous Courses</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {asyncCourses.map((course) => (
                            <CourseItem key={course.ID} course={course} onCourseClick={onCourseClick} onEdit={onEdit} onDelete={onDelete} />
                        ))}
                    </div>
                 </div>
            )}
        </div>
    )
}

export default CoursesSchedule
