import { assignment, course as Course } from "@/wailsjs/go/models"
import { useState, useMemo } from "react"
import { Card, CardContent } from "@/components/ui/card"
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
        <div className="space-y-4">
            <Card className="flex items-center justify-between glass border-0 p-4">
                <h2 className="text-lg font-medium">Weekly Schedule</h2>
                <div className="w-64">
                    <Select value={selectedSemester} onValueChange={setSelectedSemester}>
                        <SelectTrigger className="bg-gray-800/50 border border-gray-600">
                            <SelectValue placeholder="Filter by semester" />
                        </SelectTrigger>
                        <SelectContent className="glass border-0">
                            {semesters.map(semester => (
                                <SelectItem key={semester} value={semester}>
                                    {semester}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            </Card>

            <Card className="glass border-0">
                <CardContent>
                    <div className="overflow-auto">
                        <div className="flex min-w-full">
                            {/* Time column */}
                            <div className="flex-shrink-0 w-20">
                                <div className="h-12 p-2 font-medium text-center border-b">Time</div>
                                {timeSlots.map(hour => (
                                    <div
                                        key={hour}
                                        className="h-[60px] p-2 text-sm text-muted-foreground text-center border-b border-r flex items-center justify-center"
                                    >
                                        {formatTime(hour)}
                                    </div>
                                ))}
                            </div>

                            {/* Days columns */}
                            {DAYS.map((day, dayIndex) => {
                                console.log(currentTop, now.getDay(), dayIndex, currentHour, currentTop > 0 && now.getDay() == dayIndex && currentHour < 20)
                                return (
                                    <div key={day} className="flex-1 min-w-[120px]">
                                        <div className="h-12 p-2 font-medium text-center border-b">
                                            {day}
                                        </div>
                                        <div className="relative">
                                            {/* Hour grid lines */}
                                            {timeSlots.map(hour => (
                                                <div
                                                    key={hour}
                                                    className="h-[60px] border-b border-r border-border"
                                                />
                                            ))}


                                            {/* Time indicator */}
                                            {currentTop > 0 && now.getDay() == dayIndex && currentHour < 20 && (
                                                <div className="absolute -left-1 -right-1 h-1 bg-blue-500/70 z-50" style={{
                                                    top: `${currentTop}px`
                                                }} />
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

                                                    const duration = differenceInMinutes(endDate, startDate) + 1

                                                    // Calculate position: each hour slot is 60px (h-15)
                                                    var topPosition = ((startHour - timeSlots[0]) * 60)
                                                    if (startMinute != 0) {
                                                        topPosition = topPosition + (60 / (60 / startMinute))
                                                    }
                                                    const height = duration - 2 // 1px/min / Subtract 2px for border spacing

                                                    return (

                                                        <Card
                                                            key={`${course.ID}-${index}`}
                                                            className={`absolute left-1 right-1 text-xs text-white font-medium shadow-sm ${isOn ? 'bg-blue-500/50 hover:bg-blue-500/70' : 'glass hover:bg-white/5'} border-0  transition-all duration-300 `}
                                                            style={{
                                                                backgroundColor: course.Color || '#3b82f6',
                                                                top: `${topPosition}px`,
                                                                height: `${height}px`
                                                            }}
                                                            onClick={() => onCourseClick(course)}
                                                        >
                                                            <CardContent className="p-2">
                                                                <div className="flex flex-col space-y-2">
                                                                    <div className="flex flex-row items-center gap-2">
                                                                        <div className={`h-2 w-2  rounded-full ${course.Color}`} />
                                                                        <div className="font-semibold truncate">
                                                                            {course.Code}
                                                                        </div>
                                                                    </div>

                                                                    <div className="text-xs opacity-90">
                                                                        {course.parsedSchedule?.startTimeString} - {course.parsedSchedule?.endTimeString}
                                                                    </div>

                                                                </div>
                                                            </CardContent>
                                                        </Card>

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
                        <div className="
                            absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 
                            bg-gray-800/50 border border-gray-600 
                            p-4 ml-10 rounded-lg 
                            text-center text-muted-foreground
                            flex flex-col items-center justify-center
                            space-y-1
                            ">
                            <BookOpen strokeWidth={1.5} className="h-10 w-10 text-white/20 mx-auto mb-2" />
                            <p className="text-sm text-white/50">No courses found</p>
                            <p className="text-sm text-white/50"> for <span className="font-semibold">{selectedSemester}</span> semester</p>
                        </div>
                    )}
                </CardContent>
            </Card>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {asyncCourses.map((course) => (
                    <CourseItem key={course.ID} course={course} onCourseClick={onCourseClick} onEdit={onEdit} onDelete={onDelete} />
                ))}
            </div>


        </div>
    )
}

export default CoursesSchedule
