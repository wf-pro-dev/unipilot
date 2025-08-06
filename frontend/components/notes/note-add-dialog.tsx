"use client"

import type React from "react"

import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Plus, BookOpen, FileText, Tag, Video } from "lucide-react"
import { useCourses } from "@/hooks/use-courses"
import { note } from "@/wailsjs/go/models"

const subjects = [
    { value: "Mathematics", label: "Mathematics", color: "text-blue-400 border-blue-400" },
    { value: "Science", label: "Science", color: "text-yellow-400 border-yellow-400" },
    { value: "History", label: "History", color: "text-red-400 border-red-400" },
    { value: "English", label: "English", color: "text-orange-400 border-orange-400" },
    { value: "Computer Science", label: "Computer Science", color: "text-green-400 border-green-400" },
    { value: "Physics", label: "Physics", color: "text-purple-400 border-purple-400" },
    { value: "Chemistry", label: "Chemistry", color: "text-pink-400 border-pink-400" },
    { value: "Biology", label: "Biology", color: "text-gray-400 border-gray-400" },
    { value: "Social Studies", label: "Social Studies", color: "text-brown-400 border-brown-400" },
    { value: "Art", label: "Art", color: "text-indigo-400 border-indigo-400" },
    { value: "Music", label: "Music", color: "text-teal-400 border-teal-400" },
]

interface AddNoteDialogProps {
    onAdd: (note: note.LocalNote) => void
}

export function AddNoteDialog({ onAdd }: AddNoteDialogProps) {
    const [open, setOpen] = useState(false)
    const [formData, setFormData] = useState({
        title: "",
        subject: "",
        course_code: "",
        course_name: "",
        course_color: "",
    })

    const { data: courses } = useCourses()

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        setOpen(false)

        onAdd({
            Title: formData.title,
            Subject: formData.subject,
            CourseCode: formData.course_code,

            ID: 0,
            CreatedAt: new Date(),
            UpdatedAt: new Date(),
            DeletedAt: null,
            Course: null as any,
        } as note.LocalNote)

        setFormData({
            title: "",
            subject: "",
            course_code: "",
            course_name: "",
            course_color: "",
        })
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button className="bg-gradient-to-r from-green-500 to-blue-600 hover:from-green-600 hover:to-blue-700">
                    <Plus className="h-4 w-4 mr-2" />
                    Add Note
                </Button>
            </DialogTrigger>
            <DialogContent className="glass border-0 text-white max-w-md">
                <DialogHeader>
                    <DialogTitle>Add Note</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">


                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="subject" className="text-gray-300">
                                Subject
                            </Label>
                            <Select
                                value={formData.subject}
                                onValueChange={(value) => setFormData({ ...formData, subject: value })}
                            >
                                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                                    <SelectValue placeholder="Select subject" />
                                </SelectTrigger>
                                <SelectContent className="glass border-gray-600">
                                    {subjects.map((subject) => (
                                        <SelectItem key={subject.value} value={subject.value}>
                                            {subject.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>


                        <div>
                            <Label htmlFor="course" className="text-gray-300">
                                Course
                            </Label>
                            <Select
                                value={formData.course_code}
                                onValueChange={(value) => {
                                    const course = courses?.find((course) => course.Code === value)
                                    setFormData({
                                        ...formData,
                                        course_code: value,
                                        course_name: course?.Name || "",
                                        course_color: course?.Color || ""
                                    })
                                }}
                            >
                                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                                    <SelectValue placeholder="Select course" className="text-white" >
                                        <div className="flex items-center gap-2">
                                            <div className={`h-2 w-2 rounded-full ${formData.course_color}`} />
                                            {formData.course_code}
                                        </div>
                                    </SelectValue>
                                </SelectTrigger>
                                <SelectContent className="glass border-gray-600">
                                    {courses?.map((course) => (
                                        <SelectItem key={course.Code} value={course.Code}>
                                            <div className="flex items-center gap-2">
                                                <div className={`h-2 w-2 rounded-full ${course.Color}`} />
                                                {course.Code} - {course.Name}
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    <div>
                        <Label htmlFor="title" className="text-gray-300">
                            Note Title
                        </Label>
                        <Input
                            id="title"
                            value={formData.title}
                            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                            placeholder="Lecture 5: Introduction to React"
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
                        <Button type="submit">Add Note</Button>
                    </div>
                </form>
            </DialogContent>
        </Dialog>
    )
}
