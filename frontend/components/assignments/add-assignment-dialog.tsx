"use client"

import type React from "react"

import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { format } from "date-fns"
import { BookOpen, CalendarIcon, Flag, Plus } from "lucide-react"
import { cn } from "@/lib/utils"
import { useCourses } from "@/hooks/use-courses"
import { assignment } from "@/wailsjs/go/models"
import { GlassCard } from "@/components/ui/glass-card"
import { Textarea } from "../ui/textarea"
import { CoursesSelect } from "../courses/courses-select"

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

interface AddAssignmentDialogProps {
  onAdd: (assignment: assignment.LocalAssignment) => void
}

export function AddAssignmentDialog({ onAdd }: AddAssignmentDialogProps) {
  const [open, setOpen] = useState(false)
  const [deadline, setDeadline] = useState<Date>(new Date())
  const intitialFormData = {
    title: "",
    course_color: "",
    course_code: "",
    course_name: "",
    type_name: "",
    status_name: "Not started",
    priority: "low",
    todo: "",
    link: "",
  }
  const [formData, setFormData] = useState(intitialFormData)

  const key_to_column = {
    title: "Title",
    course_code: "CourseCode",
    type_name: "TypeName",
    status_name: "StatusName",
    priority: "Priority",
    todo: "Todo",
    link: "Link",
  }

  const { data: courses } = useCourses()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setOpen(false)

    onAdd({
      Title: formData.title,
      Todo: formData.todo,
      Deadline: deadline,
      CourseCode: formData.course_code,
      TypeName: formData.type_name,
      StatusName: formData.status_name,
      Priority: formData.priority || "low",
      Link: formData.link || "https://acconline.austincc.edu/ultra/stream",
      ID: 0,
      CreatedAt: new Date(),
      UpdatedAt: new Date(),
      DeletedAt: null,
      RemoteID: 0,
      NotionID: "",
      SyncStatus: "pending",
      Course: null as any,
      Type: null as any,
      Status: null as any,
      Documents: [] as any,
    } as assignment.LocalAssignment)
    setFormData({
      title: "",
      course_color: "",
      course_code: "",
      course_name: "",
      type_name: "",
      status_name: "Not started",
      priority: "low",
      todo: "",
      link: "",
    })
  }

  const handleOpenChange = (open: boolean) => {
    setOpen(open)
    if (!open) {
      setFormData(intitialFormData)
      setDeadline(new Date())
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button className="text-white bg-blue-600 hover:bg-blue-500 shadow-[0_0_15px_rgba(37,99,235,0.3)] transition-all duration-300 border-0">
          <Plus className="h-4 w-4 mr-2" />
          Add Assignment
        </Button>
      </DialogTrigger>
      <DialogContent className="glass border-white/10 text-white max-w-lg p-0 overflow-hidden gap-0">
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <DialogTitle className="text-xl font-semibold">Add Assignment</DialogTitle>
        </DialogHeader>
        
        <div className="p-6 max-h-[80vh] overflow-y-auto custom-scrollbar">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="title" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                Assignment Title
              </Label>
              <Input
                id="title"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="e.g. Calculus Midterm, History Essay..."
                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-11"
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="course" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Course
                </Label>
                <CoursesSelect
                  value={formData.course_code}
                  onValueChange={(value) =>
                    setFormData({
                      ...formData,
                      course_code: value,
                    })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="type" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Type
                </Label>
                <Select value={formData.type_name} onValueChange={(value) => setFormData({ ...formData, type_name: value })}>
                  <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent className="glass border-white/10">
                    {types.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        <div className="flex items-center gap-2">
                          <BookOpen className={`h-3.5 w-3.5 ${type.color}`} />
                          <span className="text-sm">{type.label}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="text-gray-400 text-xs font-medium uppercase tracking-wider">Due Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10",
                        !deadline && "text-muted-foreground",
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4 text-blue-400" />
                      {deadline ? format(deadline, "MMM do, yyyy") : <span>Pick a date</span>}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0 glass border-white/10">
                    <Calendar 
                        className="glass text-white" 
                        mode="single" 
                        selected={deadline} 
                        onSelect={(date) => date && setDeadline(date)} 
                        required 
                    />
                  </PopoverContent>
                </Popover>
              </div>

              <div className="space-y-2">
                <Label htmlFor="priority" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                  Priority
                </Label>
                <Select
                  value={formData.priority}
                  onValueChange={(value) => setFormData({ ...formData, priority: value })}
                >
                  <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-white/10">
                    {priorities.map((priority) => (
                      <SelectItem key={priority.value} value={priority.value}>
                        <div className="flex items-center gap-2">
                          <Flag className={`h-3.5 w-3.5 ${
                            priority.value === 'high' ? 'text-red-400 fill-red-400/20' : 
                            priority.value === 'medium' ? 'text-yellow-400' : 'text-green-400'
                          }`} />
                          <span className="text-sm">{priority.label}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="link" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                Link (Optional)
              </Label>
              <Input
                id="link"
                value={formData.link}
                onChange={(e) => setFormData({ ...formData, link: e.target.value })}
                placeholder="https://canvas.university.edu/..."
                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10 font-mono text-xs"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="todo" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                Notes & Todos
              </Label>
              <Textarea
                id="todo"
                value={formData.todo}
                onChange={(e) => setFormData({ ...formData, todo: e.target.value })}
                placeholder="Add specific details, requirements, or sub-tasks..."
                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all min-h-[100px] resize-none"
              />
            </div>

            <div className="flex justify-end gap-3 pt-4 border-t border-white/5 mt-6">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
                className="border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white"
              >
                Cancel
              </Button>
              <Button 
                type="submit"
                className="bg-blue-600 hover:bg-blue-500 text-white px-6 shadow-[0_0_15px_rgba(37,99,235,0.2)]"
              >
                Create Assignment
              </Button>
            </div>
          </form>
        </div>
      </DialogContent>
    </Dialog>
  )
}
