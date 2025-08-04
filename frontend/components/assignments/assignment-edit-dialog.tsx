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
import { BookOpen, CalendarIcon, Flag } from "lucide-react"
import { cn } from "@/lib/utils"
import { assignment } from "@/wailsjs/go/models"
import { useCourses } from "@/hooks/use-courses"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { Textarea } from "../ui/textarea"


const priorities = [
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" },
]

const types = [
  { value: "HW", label: "HW", color: "text-blue-400 border-blue-400" },
  { value: "Group Project", label: "Group Project", color: "text-yellow-400 border-yellow-400" },
  { value: "Exam", label: "Exam", color: "text-red-400 border-red-400" },
  { value: "Quiz", label: "Quiz", color: "text-orange-400 border-orange-400" },
  { value: "Lab", label: "Lab", color: "text-green-400 border-green-400" },
]

const statuses = [
  { value: "Not started", label: "Not started" },
  { value: "In progress", label: "In progress" },
  { value: "Done", label: "Done" },
]

interface AssignmentEditDialogProps {
  open: boolean
  setOpen: (open: boolean) => void
  assignment: assignment.LocalAssignment
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}

export function AssignmentEditDialog({ open, setOpen, assignment, onEdit }: AssignmentEditDialogProps) {
  const [deadline, setDeadline] = useState<Date>(new Date(assignment.Deadline) || new Date())
  const [formData, setFormData] = useState({
    title: assignment.Title || "",
    course_code: assignment.CourseCode || "",
    course_name: assignment.Course?.Name || "",
    type_name: assignment.TypeName || "",
    status_name: assignment.StatusName || "",
    priority: assignment.Priority || "",
    todo: assignment.Todo || "",
  })

  const key_to_column = {
    title: "Title",
    course_code: "CourseCode",
    course_name: "CourseName",
    type_name: "TypeName",
    status_name: "StatusName",
    priority: "Priority",
    todo: "Todo",
  }



  const { data: courses } = useCourses()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setOpen(false)

    for (const [key, value] of Object.entries(formData)) {
      if (key === "course_name") { continue }
        
      const column = key_to_column[key as keyof typeof key_to_column] as keyof assignment.LocalAssignment
      if (value !== assignment[column]) {

        const message = `Changes to ${column} value: ${value} assignment: ${assignment[column]}`
        LogInfo(message)
        onEdit(assignment, key, value)
      }
      else {
        const message = `No changes to ${column} value: ${value} assignment: ${assignment[column]}`
        LogInfo(message)
      }
    }

    if (!isSameDay(deadline, new Date(assignment.Deadline))) {
      onEdit(assignment, "deadline", format(deadline, "yyyy-MM-dd HH:mm:ssxxx"))
    }
    setDeadline(new Date())

  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>

      <DialogContent className="glass border-0 text-white max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Assignment</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="title" className="text-gray-300">
              Assignment Title
            </Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="Homework #3"
              className="bg-gray-800/50 border-gray-600"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="course" className="text-gray-300">
                Course
              </Label>
              <Select
                value={formData.course_code}
                onValueChange={(value) => {
                  const course = courses?.find((course) => course.Code === value)
                  setFormData({ ...formData, course_code: value, course_name: course?.Name || "" })
                }}
              >
                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                  <div className="flex items-center">
                    <div className={` h-2 w-2 rounded-full ${assignment.Course?.Color}`} />
                    <p className="line-clamp-1">
                      {formData.course_code} - {formData.course_name}
                    </p>
                  </div>
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  {courses?.map((course) => (
                    <SelectItem key={course.Code} value={course.Code}>
                      <div className="flex items-center gap-2">
                        <div className={` h-2 w-2 rounded-full ${course.Color}`} />
                        {course.Code} - {course.Name}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="type" className="text-gray-300">
                Type
              </Label>
              <Select value={formData.type_name} onValueChange={(value) => setFormData({ ...formData, type_name: value })}>
                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                  <SelectValue placeholder="Select type" />
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  {types.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      <div className="flex items-center gap-2">
                        <BookOpen className={` h-4 w-4 ${type.color}`} />
                        {type.label}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="text-gray-300">Due Date</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal bg-gray-800/50 border-gray-600",
                      !deadline && "text-muted-foreground",
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {deadline ? format(deadline, "PPP") : <span>Pick a date</span>}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0 glass border-gray-600">
                  <Calendar mode="single" selected={deadline} onSelect={setDeadline} required />
                </PopoverContent>
              </Popover>
            </div>

            <div>
              <Label htmlFor="priority" className="text-gray-300">
                Status
              </Label>
              <Select
                value={formData.status_name}
                onValueChange={(value) => setFormData({ ...formData, status_name: value })}
              >
                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  {statuses.map((status) => (
                    <SelectItem key={status.value} value={status.value}>
                      {status.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <Label htmlFor="todo" className="text-gray-300">
              Todo
            </Label>
            <Textarea
              id="todo"
              value={formData.todo}
              onChange={(e) => setFormData({ ...formData, todo: e.target.value })}
              placeholder="What do you need to do?"
              className="bg-gray-800/50 border-gray-600"
            />
          </div>


          <div className="flex items-center justify-between">
            <div>
              <div>
                <Label htmlFor="priority" className="text-gray-300">
                  Priority
                </Label>
                <Select
                  value={formData.priority}
                  onValueChange={(value) => setFormData({ ...formData, priority: value })}
                >
                  <SelectTrigger className="bg-gray-800/50 border-gray-600">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-gray-600">
                    {priorities.map((priority) => (
                      <SelectItem key={priority.value} value={priority.value}>
                        <div className="flex items-center gap-2">
                          <Flag className=" h-4 w-4" />
                          {priority.label}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
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
              <Button type="submit">Edit Assignment</Button>

            </div>
          </div>

        </form>
      </DialogContent>
    </Dialog>
  )
}
