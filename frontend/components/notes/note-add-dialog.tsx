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
import { CoursesSelect } from "../courses/courses-select"

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
        })
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button className="text-white bg-green-600 hover:bg-green-500 shadow-[0_0_15px_rgba(22,163,74,0.3)] transition-all duration-300 border-0">
                    <Plus className="h-4 w-4 mr-2" />
                    Add Note
                </Button>
            </DialogTrigger>
            <DialogContent className="glass border-white/10 text-white max-w-md p-0 overflow-hidden gap-0">
                <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <DialogTitle className="text-xl font-semibold">Add Note</DialogTitle>
                </DialogHeader>
                
                <div className="p-6">
                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="subject" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                                    Subject
                                </Label>
                                <Select
                                    value={formData.subject}
                                    onValueChange={(value) => setFormData({ ...formData, subject: value })}
                                >
                                    <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10">
                                        <SelectValue placeholder="Select subject" />
                                    </SelectTrigger>
                                    <SelectContent className="glass border-white/10">
                                        {subjects.map((subject) => (
                                            <SelectItem key={subject.value} value={subject.value}>
                                                <span className="text-sm">{subject.label}</span>
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="course" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                                    Course
                                </Label>
                                <CoursesSelect
                                    value={formData.course_code}
                                    onValueChange={(value) => setFormData({ 
                                        ...formData, 
                                        course_code: value,
                                    })}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="title" className="text-gray-400 text-xs font-medium uppercase tracking-wider">
                                Note Title
                            </Label>
                            <Input
                                id="title"
                                value={formData.title}
                                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                                placeholder="Lecture 5: Introduction to React"
                                className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10"
                                required
                            />
                        </div>

                        <div className="flex justify-end space-x-3 pt-4 border-t border-white/5 mt-6">
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
                                className="bg-green-600 hover:bg-green-500 text-white px-6 shadow-[0_0_15px_rgba(22,163,74,0.2)]"
                            >
                                Add Note
                            </Button>
                        </div>
                    </form>
                </div>
            </DialogContent>
        </Dialog>
    )
}
