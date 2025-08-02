"use client"

import type React from "react"

import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { cn } from "@/lib/utils"
import { Calendar } from "@/components/ui/calendar"
import { CalendarIcon, Plus } from "lucide-react"
import { format } from "date-fns"
import { course } from "@/wailsjs/go/models"
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

interface AddCourseDialogProps {
  onAdd: (course: course.LocalCourse) => void
}

export function AddCourseDialog({ onAdd }: AddCourseDialogProps) {
  const [open, setOpen] = useState(false)
  const [startDate, setStartDate] = useState<Date | undefined>(undefined)
  const [endDate, setEndDate] = useState<Date | undefined>(undefined)

  const [formData, setFormData] = useState({
    name: "",
    code: "",
    color: "bg-blue-500",
    semester: "",
    schedule: "",
    credits: "3",
    room_number: "",
    instructor: "",
    instructor_email: "",
    location: "",
  })
  const validateDates = (startDate: Date, endDate: Date) => {
    if (startDate > endDate) {
        toast.error("Start date must be before end date")
        return false
    }

    return true
}

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (startDate && endDate && !validateDates(startDate,endDate)) return 
    
   
    onAdd({
      Name: formData.name,
      Code: formData.code,
      Color: formData.color,
      Semester: formData.semester,
      Schedule: formData.schedule,
      Credits: parseInt(formData.credits),
      RoomNumber: formData.room_number,
      Instructor: formData.instructor,
      InstructorEmail: formData.instructor_email,
      StartDate: startDate,
      EndDate: endDate,
    } as course.LocalCourse)
    setStartDate(undefined)
    setEndDate(undefined)

    toast.success("Course added successfully")

    setOpen(false)
    setFormData({
      name: "",
      code: "",
      color: "bg-blue-500",
      semester: "",
      schedule: "",
      credits: "3",
      room_number: "",
      instructor: "",
      instructor_email: "",
      location: "",
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700">
          <Plus className="mr-2 w-4 h-4" />
          Add Course
        </Button>
      </DialogTrigger>
      <DialogContent className="glass border-0 text-white max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Course</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="name" className="text-gray-300">
                Course Name
              </Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Data Structures"
                className="border-gray-600 bg-gray-800/50"
                required
              />
            </div>
            <div>
              <Label htmlFor="code" className="text-gray-300">
                Course Code
              </Label>
              <Input
                id="code"
                value={formData.code}
                onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                placeholder="CS 101"
                className="border-gray-600 bg-gray-800/50"
                required
              />
            </div>
          </div>


          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="code" className="text-gray-300">
                Location
              </Label>
              <Input
                id="location"
                value={formData.location}
                onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                placeholder="Building 1, Room 101 / Online"
                className="border-gray-600 bg-gray-800/50"
                required
              />
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
                className="border-gray-600 bg-gray-800/50"
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
                    <CalendarIcon className="mr-2 w-4 h-4" />
                    {startDate ? format(startDate, "PPP") : <span>Pick a start date</span>}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="p-0 w-auto border-gray-600 glass">
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
                    <CalendarIcon className="mr-2 w-4 h-4" />
                    {endDate ? format(endDate, "PPP") : <span>Pick a end date</span>}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="p-0 w-auto border-gray-600 glass">
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
                className="border-gray-600 bg-gray-800/50"
                required
              />
            </div>

            <div>
              <Label htmlFor="semester" className="text-gray-300">
                Semester
              </Label>
              <Select value={formData.semester} onValueChange={(value) => setFormData({ ...formData, semester: value })}>
                <SelectTrigger className="border-gray-600 bg-gray-800/50">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-gray-600 glass">
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
              className="border-gray-600 bg-gray-800/50"
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
              placeholder="M, T, W 9:00 AM - 10:30 AM / Async / Asynchronous"
              className="border-gray-600 bg-gray-800/50"
              required
            />
          </div>





          <div className="flex justify-between items-center pt-4">
            <div>
              <Select value={formData.color} onValueChange={(value) => setFormData({ ...formData, color: value })}>
                <SelectTrigger className="border-gray-600 bg-gray-800/50">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-gray-600 glass">
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



            <div className="flex space-x-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
                className="bg-transparent border-gray-600"
              >
                Cancel
              </Button>
              <Button type="submit">Add Course</Button>
            </div>

          </div>

        </form>
      </DialogContent>
    </Dialog >
  )
}
