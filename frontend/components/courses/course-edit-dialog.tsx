"use client"

import type React from "react"
import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { format, isSameDay } from "date-fns"
import { CalendarIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { course as Course } from "@/wailsjs/go/models"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { toast } from "sonner"

const colors = [
    { name: "Blue", value: "bg-blue-500" },
    { name: "Green", value: "bg-green-500" },
    { name: "Purple", value: "bg-purple-500" },
    { name: "Red", value: "bg-red-500" },
    { name: "Orange", value: "bg-orange-500" },
    { name: "Pink", value: "bg-pink-500" },
]

const semesters = [
    { name: "FALL 2024", value: "FALL 2024" },
    { name: "SPRING 2025", value: "SPRING 2025" },
    { name: "SUMMER 2025", value: "SUMMER 2025" },
    { name: "FALL 2025", value: "FALL 2025" },
    { name: "SPRING 2026", value: "SPRING 2026" },
    { name: "SUMMER 2026", value: "SUMMER 2026" },
    { name: "SPRING 2027", value: "SPRING 2027" },
    { name: "SUMMER 2027", value: "SUMMER 2027" },
    { name: "FALL 2027", value: "FALL 2027" },
    { name: "SPRING 2028", value: "SPRING 2028" },
    { name: "SUMMER 2028", value: "SUMMER 2028" },
]

interface CourseEditDialogProps {
    open: boolean
    setOpen: (open: boolean) => void
    course: Course.LocalCourse | null
    onEdit: (course: Course.LocalCourse, column: string, value: string) => void
}

export function CourseEditDialog({ open, setOpen, course, onEdit }: CourseEditDialogProps) {
    if (!course) return null
    const [startDate, setStartDate] = useState<Date>(new Date(course.StartDate) || new Date())
    const [endDate, setEndDate] = useState<Date>(new Date(course.EndDate) || new Date())
    const [formData, setFormData] = useState({
        name: course.Name || "",
        code: course.Code || "",
        color: course.Color || "bg-blue-500",
        semester: course.Semester || "",
        schedule: course.Schedule || "",
        credits: course.Credits.toString() || "3",
        room_number: course.RoomNumber || "",
        duration: course.Duration || "",
        instructor: course.Instructor || "",
        instructor_email: course.InstructorEmail || "",
    })

    const key_to_column = {
        name: "Name",
        code: "Code",
        color: "Color",
        semester: "Semester",
        schedule: "Schedule",
        credits: "Credits",
        room_number: "RoomNumber",
        duration: "Duration",
        instructor: "Instructor",
        instructor_email: "InstructorEmail",
    }

    const validateDates = (startDate: Date, endDate: Date) => {
        if (startDate > endDate) {
            toast.error("Start date must be before end date")
            return false
        }

        return true
    }

    const validateSchedule = (schedule: string) => {
        if (schedule.length === 0) {
            toast.error("Schedule is required")
            return false
        }

        if (schedule == "Async" || schedule == "Asynchronous") {
            return true
        }

        // Validate format: "<day>, <day> <hour> - <hour>"
        // Days: M, T, W, Th, F, S, Su (separated by ", ")
        // Hours: HH:MM AM/PM (1-2 digits for hour, 00-59 for minutes)
        const schedulePattern = /^((?:(?:M|T|W|Th|F|S|Su)(?:,\s(?:M|T|W|Th|F|S|Su))*)?)\s+(\d{1,2}:[0-5]\d\s(?:AM|PM))\s*-\s*(\d{1,2}:[0-5]\d\s(?:AM|PM))$/

        const match = schedule.match(schedulePattern)

        if (!match) {
            toast.error("Invalid schedule format. Expected: 'M, T, W 9:00 AM - 10:30 AM'")
            return false
        }

        const [, daysStr, startTime, endTime] = match

        // Validate that at least one day is specified
        if (!daysStr || daysStr.trim().length === 0) {
            toast.error("At least one day must be specified")
            return false
        }

        // Validate individual days
        const days = daysStr.trim().split(', ')
        const validDays = ['M', 'T', 'W', 'Th', 'F', 'S', 'Su']

        for (const day of days) {
            if (!validDays.includes(day)) {
                toast.error(`Invalid day '${day}'. Valid days: M, T, W, Th, F, S, Su`)
                return false
            }
        }

        // Validate time format more strictly
        const timePattern = /^(\d{1,2}):([0-5]\d)\s(AM|PM)$/

        const startMatch = startTime.match(timePattern)
        const endMatch = endTime.match(timePattern)

        if (!startMatch || !endMatch) {
            toast.error("Invalid time format. Use format like '9:00 AM' or '12:30 PM'")
            return false
        }

        // Validate hour ranges (1-12 for 12-hour format)
        const startHour = parseInt(startMatch[1])
        const endHour = parseInt(endMatch[1])

        if (startHour < 1 || startHour > 12 || endHour < 1 || endHour > 12) {
            toast.error("Hour must be between 1 and 12")
            return false
        }

        // Convert to 24-hour format for comparison
        const convertTo24Hour = (hour: number, minute: number, period: string): number => {
            if (period === 'AM') {
                return hour === 12 ? 0 * 60 + minute : hour * 60 + minute
            } else {
                return hour === 12 ? 12 * 60 + minute : (hour + 12) * 60 + minute
            }
        }

        const startMinutes = convertTo24Hour(startHour, parseInt(startMatch[2]), startMatch[3])
        const endMinutes = convertTo24Hour(endHour, parseInt(endMatch[2]), endMatch[3])

        if (startMinutes >= endMinutes) {
            toast.error("Start time must be before end time")
            return false
        }

        return true
    }

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()

        var isChanged = false

        if (!validateDates(startDate, endDate) || !validateSchedule(formData.schedule)) {
            return
        }

        setOpen(false)

        for (const [key, value] of Object.entries(formData)) {

            const column = key_to_column[key as keyof typeof key_to_column] as keyof Course.LocalCourse
            if (value != course[column]) {
                onEdit(course, key, value)
                isChanged = true
            }
            else {
                const message = `No changes to ${column} value: ${value} course: ${course[column]}`
                LogInfo(message)
            }
        }

        const formattedStartDate = format(startDate, "yyyy-MM-dd HH:mm:ssxxx")
        const formattedEndDate = format(endDate, "yyyy-MM-dd HH:mm:ssxxx")

        if (!isSameDay(startDate, new Date(course.StartDate))) {
            onEdit(course, "start_date", formattedStartDate)
            isChanged = true
        } else {
            const message = `No changes to start date value: ${formattedStartDate} course: ${course.StartDate}`
            LogInfo(message)
        }
        if (!isSameDay(endDate, new Date(course.EndDate))) {
            onEdit(course, "end_date", formattedEndDate)
            isChanged = true
        } else {
            const message = `No changes to end date value: ${formattedEndDate} course: ${course.EndDate}`
            LogInfo(message)
        }

        if (isChanged) {
            toast.success("Course updated successfully")
        } else {
            toast.info("No changes to course")
        }

        setStartDate(new Date())
        setEndDate(new Date())
    }



    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent className="glass border-0 text-white max-w-md max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Edit Course</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <Label htmlFor="name" className="text-gray-300">
                            Course Name
                        </Label>
                        <Input
                            id="name"
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            placeholder="Data Structures"
                            className="bg-gray-800/50 border-gray-600"
                            required
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">

                        <div>
                            <Label htmlFor="color" className="text-gray-300">
                                Color
                            </Label>
                            <Select value={formData.color} onValueChange={(value) => setFormData({ ...formData, color: value })}>
                                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent className="glass border-gray-600">
                                    {colors.map((color) => (
                                        <SelectItem key={color.value} value={color.value}>
                                            <div className="flex items-center space-x-2">
                                                <div className={`w-4 h-4 rounded-full ${color.value}`} />
                                                <span>{color.name}</span>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="credits" className="text-gray-300">
                                Credits
                            </Label>
                            <Input
                                id="credits"
                                type="number"
                                value={formData.credits}
                                onChange={(e) => setFormData({ ...formData, credits: e.target.value })}
                                placeholder="3"
                                className="bg-gray-800/50 border-gray-600"
                                min={1}
                                max={4}
                                required
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="start_date" className="text-gray-300">
                                Start Date
                            </Label>
                            <Popover>
                                <PopoverTrigger asChild>
                                    <Button
                                        variant="outline"
                                        className={cn(
                                            "w-full justify-start text-left font-normal bg-gray-800/50 border-gray-600",
                                            !startDate && "text-muted-foreground",
                                        )}
                                    >
                                        <CalendarIcon className="mr-2 h-4 w-4" />
                                        {startDate ? format(startDate, "PPP") : <span>Pick a start date</span>}
                                    </Button>
                                </PopoverTrigger>
                                <PopoverContent className="w-auto p-0 glass border-gray-600">
                                    <Calendar mode="single" selected={startDate} onSelect={setStartDate} required />
                                </PopoverContent>
                            </Popover>
                        </div>
                        <div>
                            <Label htmlFor="end_date" className="text-gray-300">
                                End Date
                            </Label>
                            <Popover>
                                <PopoverTrigger asChild>
                                    <Button
                                        variant="outline"
                                        className={cn(
                                            "w-full justify-start text-left font-normal bg-gray-800/50 border-gray-600",
                                            !endDate && "text-muted-foreground",
                                        )}
                                    >
                                        <CalendarIcon className="mr-2 h-4 w-4" />
                                        {endDate ? format(endDate, "PPP") : <span>Pick a end date</span>}
                                    </Button>
                                </PopoverTrigger>
                                <PopoverContent className="w-auto p-0 glass border-gray-600">
                                    <Calendar mode="single" selected={endDate} onSelect={setEndDate} required />
                                </PopoverContent>
                            </Popover>
                        </div>
                    </div>


                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="instructor" className="text-gray-300">
                                Instructor
                            </Label>
                            <Input
                                id="instructor"
                                value={formData.instructor}
                                onChange={(e) => setFormData({ ...formData, instructor: e.target.value })}
                                placeholder="Dr. Smith"
                                className="bg-gray-800/50 border-gray-600"
                                required
                            />
                        </div>

                        <div>
                            <Label htmlFor="semester" className="text-gray-300">
                                Semester
                            </Label>
                            <Select value={formData.semester} onValueChange={(value) => setFormData({ ...formData, semester: value })}>
                                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent className="glass border-gray-600">
                                    {semesters.map((semester) => (
                                        <SelectItem key={semester.value} value={semester.value}>
                                            <span>{semester.name}</span>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>


                    </div>

                    <div>
                        <Label htmlFor="name" className="text-gray-300">
                            Instructor Email
                        </Label>
                        <Input
                            id="instructor_email"
                            value={formData.instructor_email}
                            onChange={(e) => setFormData({ ...formData, instructor_email: e.target.value })}
                            placeholder="smith@example.com"
                            className="bg-gray-800/50 border-gray-600"
                            required
                        />
                    </div>



                    <div>
                        <Label htmlFor="schedule" className="text-gray-300">
                            Schedule
                        </Label>
                        <Input
                            id="schedule"
                            value={formData.schedule}
                            onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                            placeholder="MWF 10:00-11:00 AM"
                            className="bg-gray-800/50 border-gray-600"
                            required
                        />
                    </div>







                    <div className="flex justify-end space-x-2 pt-4">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => setOpen(false)}
                            className="border-gray-600 bg-transparent"
                        >
                            Cancel
                        </Button>
                        <Button type="submit">Edit Course</Button>
                    </div>

                </form>
            </DialogContent>
        </Dialog >
    )
}
