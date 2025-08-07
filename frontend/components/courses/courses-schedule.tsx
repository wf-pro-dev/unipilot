import { assignment, course as Course } from "@/wailsjs/go/models"
import { useState, useMemo } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { differenceInMinutes, format, isAfter, isBefore, isSameDay } from "date-fns"
import { CourseItem } from "./course-item"

interface CoursesScheduleProps {
    courses: Course.LocalCourse[]
    onCourseClick: (course: Course.LocalCourse) => void
    onEdit: (course: Course.LocalCourse, column: string, value: string) => void
    onDelete: (course: Course.LocalCourse) => void
}

// Day abbreviations mapping
const DAY_MAPPING: Record<string, number> = {
    'M': 1, 'Mo': 1, 'Mon': 1,
    'T': 2, 'Tu': 2, 'Tue': 2,
    'W': 3, 'We': 3, 'Wed': 3,
    'Th': 4, 'Thu': 4,
    'F': 5, 'Fr': 5, 'Fri': 5,
    'S': 6, 'Sa': 6, 'Sat': 6,
    'Su': 0, 'Sun': 0
}

const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const HOURS = Array.from({ length: 24 }, (_, i) => i)

interface ParsedSchedule {
    days: number[]
    startTime: number // 24-hour format
    endTime: number
    startTimeString: string
    endTimeString: string
}

interface CourseWithSchedule extends Course.LocalCourse {
    parsedSchedule: ParsedSchedule | null
}

function parseSchedule(schedule: string): ParsedSchedule | null {
    if (!schedule) return null

    try {
        // Split by time separator (looking for patterns like "1:00 PM - 2:00 PM")
        const timeMatch = schedule.match(/(\d{1,2}):(\d{2})\s*(AM|PM)\s*-\s*(\d{1,2}):(\d{2})\s*(AM|PM)/i)
        if (!timeMatch) return null

        const [, startHour, startMin, startPeriod, endHour, endMin, endPeriod] = timeMatch

        // Convert to 24-hour format
        let start24 = parseInt(startHour)
        let end24 = parseInt(endHour)

        if (startPeriod.toUpperCase() === 'PM' && start24 !== 12) start24 += 12
        if (startPeriod.toUpperCase() === 'AM' && start24 === 12) start24 = 0
        if (endPeriod.toUpperCase() === 'PM' && end24 !== 12) end24 += 12
        if (endPeriod.toUpperCase() === 'AM' && end24 === 12) end24 = 0

        // Parse days (everything before the time)
        const daysPart = schedule.split(/\d{1,2}:\d{2}/)[0].trim()
        const dayTokens = daysPart.split(/[,\s]+/).filter(token => token.length > 0)

        const days: number[] = []
        for (const token of dayTokens) {
            const day = DAY_MAPPING[token]
            if (day !== undefined) {
                days.push(day)
            }
        }

        return {
            days,
            startTime: start24,
            endTime: end24,
            startTimeString: `${startHour}:${startMin} ${startPeriod}`,
            endTimeString: `${endHour}:${endMin} ${endPeriod}`
        }
    } catch (error) {
        console.error('Error parsing schedule:', schedule, error)
        return null
    }
}

function formatTime(hour: number): string {
    if (hour === 0) return '12:00 AM'
    if (hour < 12) return `${hour}:00 AM`
    if (hour === 12) return '12:00 PM'
    return `${hour - 12}:00 PM`
}

function CoursesSchedule({ courses, onCourseClick, onEdit, onDelete }: CoursesScheduleProps) {
    const [selectedSemester, setSelectedSemester] = useState<string>("SUMMER 2025")

    // Get unique semesters
    const semesters = useMemo(() => {
        if (!courses) return []
        const uniqueSemesters = [...new Set(courses.map(course => course.Semester).filter(Boolean))]
        
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

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold">Weekly Schedule</h2>
                <div className="w-64">
                    <Select value={selectedSemester} onValueChange={setSelectedSemester}>
                        <SelectTrigger className="glass border-0">
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
            </div>

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
                            {DAYS.map((day, dayIndex) => (
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

                                        {/* Course blocks */}
                                        {scheduledCourses
                                            .filter(course => course.parsedSchedule?.days.includes(dayIndex))
                                            .map((course, index) => {
                                                if (!course.parsedSchedule) return null

                                                var isOn = false

                                                const startHour = course.parsedSchedule.startTime
                                                const startMinute = parseInt(course.parsedSchedule.startTimeString.split(":")[1])

                                                const endHour = course.parsedSchedule.endTime
                                                const endMinute = parseInt(course.parsedSchedule.endTimeString.split(":")[1])

                                                const today = new Date()
                                                const startDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), startHour, startMinute)
                                                const endDate = new Date(today.getFullYear(), today.getMonth(), today.getDate(), endHour, endMinute)

                                                isOn = day == format(today, 'EEEE') && isBefore(startDate, today) && isAfter(endDate, today)

                                                const duration = differenceInMinutes(endDate, startDate)

                                                // Calculate position: each hour slot is 60px (h-20)
                                                var topPosition = ((startHour - timeSlots[0]) * 60)
                                                if (startMinute != 0) {
                                                    topPosition = topPosition + (60 / (60 / startMinute))
                                                }
                                                const height = duration - 2 // Subtract 2px for border spacing

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

                                                                <div className="truncate">
                                                                    {course.Name}
                                                                </div>
                                                                <div className="text-xs opacity-90">
                                                                    {course.parsedSchedule?.startTimeString} - {course.parsedSchedule?.endTimeString}
                                                                </div>
                                                                {course.RoomNumber && (
                                                                    <div className="text-xs opacity-90">
                                                                        {course.RoomNumber}
                                                                    </div>
                                                                )}
                                                            </div>
                                                        </CardContent>
                                                    </Card>
                                                )
                                            })
                                        }
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {scheduledCourses.length === 0 && (
                        <div className="text-center py-8 text-muted-foreground">
                            {selectedSemester !== "SUMMER 2025"
                                ? `No courses found for ${selectedSemester} semester`
                                : "No courses with valid schedules found"
                            }
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
