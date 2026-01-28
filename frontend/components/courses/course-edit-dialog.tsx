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
import { models} from "@/wailsjs/go/models"
import { toast } from "sonner"
import { useCourse, useUpdateCourse } from "@/hooks/use-courses"

const colors = [
    { name: "Blue", value: "bg-blue-500" },
    { name: "Green", value: "bg-green-500" },
    { name: "Purple", value: "bg-purple-500" },
    { name: "Red", value: "bg-red-500" },
    { name: "Orange", value: "bg-orange-500" },
    { name: "Pink", value: "bg-pink-500" },
]

const semesters = [
    { name: "SUMMER 2028", value: "SUMMER 2028" },
    { name: "SPRING 2028", value: "SPRING 2028" },
    { name: "FALL 2027", value: "FALL 2027" },
    { name: "SUMMER 2027", value: "SUMMER 2027" },
    { name: "SPRING 2027", value: "SPRING 2027" },
    { name: "FALL 2026", value: "FALL 2026" },
    { name: "SUMMER 2026", value: "SUMMER 2026" },  
    { name: "SPRING 2026", value: "SPRING 2026" },
    { name: "FALL 2025", value: "FALL 2025" },
    { name: "SUMMER 2025", value: "SUMMER 2025" },
    { name: "SPRING 2025", value: "SPRING 2025" },
    { name: "FALL 2024", value: "FALL 2024" },
]

interface CourseEditDialogProps {
    courseId: number
    isOpen: boolean
    onClose: () => void
}

export function CourseEditDialog({ courseId, isOpen, onClose }: CourseEditDialogProps) {
    const { data: course } = useCourse(courseId)
    if (!course) return null

    const updateMutation = useUpdateCourse()
   

    const [startDate, setStartDate] = useState<Date>(new Date(course.StartDate) || new Date())
    const [endDate, setEndDate] = useState<Date>(new Date(course.EndDate) || new Date())
    const [formData, setFormData] = useState({
        name: course.Name || "",
        code: course.Code || "",
        color: course.Color || "bg-blue-500",
        semester: course.Semester || "",
        schedule: course.Schedule || "",
        credits: course.Credits.toString() || "3",

        instructor: course.Instructor || "",
        instructor_email: course.InstructorEmail || "",
        location: course.Location || "",
    })

    const key_to_column = {
        name: "Name",
        code: "Code",
        color: "Color",
        semester: "Semester",
        schedule: "Schedule",
        credits: "Credits",
        instructor: "Instructor",
        instructor_email: "InstructorEmail",
        location: "Location",
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

        // Validate format: "<day>, <day> <hour>:<minutes> <period> - <hour>:<minutes> <period>"
        // Days: M, T, W, Th, F, Sa, Su (separated by ", ") - at least one day required
        // Hours: HH:MM AM/PM (1-2 digits for hour, 00-59 for minutes, AM/PM)
        // All sections separated by spaces
        const schedulePattern = /^((?:M|T|W|Th|F|Sa|Su)(?:,\s(?:M|T|W|Th|F|Sa|Su))*)\s+(\d{1,2}:[0-5]\d\s(?:AM|PM))\s-\s(\d{1,2}:[0-5]\d\s(?:AM|PM))$/

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
        const validDays = ['M', 'T', 'W', 'Th', 'F', 'Sa', 'Su']

        for (const day of days) {
            if (!validDays.includes(day)) {
                toast.error(`Invalid day '${day}'. Valid days: M, T, W, Th, F, Sa, Su`)
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

        if (startHour < 0 || startHour > 12 || endHour < 0 || endHour > 12) {
            toast.error("Hour must be between 0 and 12")
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

        onClose()

        for (const [key, value] of Object.entries(formData)) {

            const column = key_to_column[key as keyof typeof key_to_column] as keyof models.LocalCourse
            if (value != course[column]) {
                updateMutation.mutate({
                    course,
                    column,
                    value
                })
                isChanged = true
            }
        }

        const formattedStartDate = format(startDate, "yyyy-MM-dd HH:mm:ssxxx")
        const formattedEndDate = format(endDate, "yyyy-MM-dd HH:mm:ssxxx")

        if (!isSameDay(startDate, new Date(course.StartDate))) {
            updateMutation.mutate({
                course,
                column: "start_date",
                value: formattedStartDate
            })
            isChanged = true
        }
        if (!isSameDay(endDate, new Date(course.EndDate))) {
            updateMutation.mutate({
                course,
                column: "end_date",
                value: formattedEndDate
            })
            isChanged = true
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
        <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-lg p-0 overflow-hidden gap-0">
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <DialogTitle className="text-xl font-semibold">Edit Course</DialogTitle>
        </DialogHeader>
        
        <div className="p-6 max-h-[80vh] overflow-y-auto custom-scrollbar">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <Label htmlFor="name" className="text-gray-400 text-xs font-medium uppercase tracking-wider block mb-2">
                Course Name
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Data Structures"
                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="location" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Location
                </Label>
                <Input
                  id="location"
                  value={formData.location}
                  onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                  placeholder="Building 1, Room 101 / Online"
                  className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="credits" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Credits
                </Label>
                <Input
                  id="credits"
                  type="number"
                  value={formData.credits}
                  onChange={(e) => setFormData({ ...formData, credits: e.target.value })}
                  placeholder="3"
                  className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                  min={1}
                  max={4}
                  required
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="start_date" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Start Date
                </Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10",
                        !startDate && "text-muted-foreground",
                      )}
                    >
                      <CalendarIcon className="mr-2 w-4 h-4 text-blue-400" />
                      {startDate ? format(startDate, "MMM do, yyyy") : <span>Pick a start date</span>}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="p-0 w-auto border-white/10 glass">
                    <Calendar mode="single" className="glass text-white" selected={startDate} onSelect={setStartDate} required />
                  </PopoverContent>
                </Popover>
              </div>
              <div className="space-y-2">
                <Label htmlFor="end_date" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  End Date
                </Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10",
                        !endDate && "text-muted-foreground",
                      )}
                    >
                      <CalendarIcon className="mr-2 w-4 h-4 text-blue-400" />
                      {endDate ? format(endDate, "MMM do, yyyy") : <span>Pick a end date</span>}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="p-0 w-auto border-white/10 glass">
                    <Calendar mode="single" className="glass text-white" selected={endDate} onSelect={setEndDate} required />
                  </PopoverContent>
                </Popover>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="instructor" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Instructor
                </Label>
                <Input
                  id="instructor"
                  value={formData.instructor}
                  onChange={(e) => setFormData({ ...formData, instructor: e.target.value })}
                  placeholder="Dr. Smith"
                  className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="semester" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Semester
                </Label>
                <Select value={formData.semester} onValueChange={(value) => setFormData({ ...formData, semester: value })}>
                  <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-white/10">
                    {semesters.map((semester) => (
                      <SelectItem key={semester.value} value={semester.value}>
                        <span className="text-sm">{semester.name}</span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="instructor_email" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                Instructor Email
              </Label>
              <Input
                id="instructor_email"
                value={formData.instructor_email}
                onChange={(e) => setFormData({ ...formData, instructor_email: e.target.value })}
                placeholder="smith@example.com"
                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="schedule" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                Schedule
              </Label>
              <Input
                id="schedule"
                value={formData.schedule}
                onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                placeholder="M, T, W 9:00 AM - 10:30 AM / Async / Asynchronous"
                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                required
              />
            </div>

            <div className="flex justify-between items-center pt-4 border-t border-white/5 mt-6">
              <div>
                <Select value={formData.color} onValueChange={(value) => setFormData({ ...formData, color: value })}>
                  <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10 w-[140px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-white/10">
                    {colors.map((color) => (
                      <SelectItem key={color.value} value={color.value}>
                        <div className="flex items-center space-x-2">
                          <div className={`w-4 h-4 rounded-full ${color.value}`} />
                          <span className="text-sm">{color.name}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex space-x-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={onClose}
                  className="border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white"
                >
                  Cancel
                </Button>
                <Button 
                  type="submit"
                  className="bg-blue-600 hover:bg-blue-500 text-white px-6 shadow-[0_0_15px_rgba(37,99,235,0.2)]"
                >
                  Edit Course
                </Button>
              </div>
            </div>
          </form>
        </div>
      </DialogContent>
    </Dialog>
    )
}
