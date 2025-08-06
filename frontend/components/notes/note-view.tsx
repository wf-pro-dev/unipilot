"use client"

import { useEffect, useMemo, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { NoteItem } from "./note-item"
import { CalendarDays, CheckCircle2, Loader2, Search } from "lucide-react"
import { course, note } from "@/wailsjs/go/models"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Input } from "@/components/ui/input"


interface NoteViewProps {
  title: string
  notes: note.LocalNote[]
  onNoteClick: (noteID: number) => void
  onEdit: (note: note.LocalNote, column: string, value: string) => void
  onDelete: (note: note.LocalNote) => void
  isLoading?: boolean
}

export function NoteView({ title, notes, onNoteClick, onDelete, onEdit, isLoading }: NoteViewProps) {

  const [selectedCourse, setSelectedCourse] = useState("All Courses")
  const [searchQuery, setSearchQuery] = useState("")

  const courseCodes = Array.from(new Set((notes || []).map((note) => note.CourseCode || "All Courses")))
  console.log(courseCodes)

  const filteredNotes = useMemo(() => {
    return notes.filter((note) => {
      const matchesSearch =
        note.Title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        note.Keywords.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesCourse = selectedCourse === "All Courses" || note.CourseCode === selectedCourse
      return matchesSearch && matchesCourse
    })
  }, [notes, searchQuery, selectedCourse])

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search notes by title or keywords..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-gray-800/50 border-gray-600"
                />
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-4">
              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-400">Course:</span>
                <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                  <SelectTrigger className="w-48 bg-gray-800/50 border-gray-600">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-gray-600">
                    <SelectItem value="All Courses">All Courses</SelectItem>
                    {courseCodes.map((courseCode) => (
                      <SelectItem key={courseCode} value={courseCode}>
                        {courseCode}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
      <Card className="border-0 glass p-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <CalendarDays className="w-5 h-5" />
            <span>{title}</span>
            {isLoading && <Loader2 className="ml-2 w-4 h-4 animate-spin" />}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {(filteredNotes || []).length === 0 ? (
            <div className="flex justify-center items-center h-48 text-gray-400">
              <div className="text-center">
                <CheckCircle2 className="mx-auto w-12 h-12 opacity-50" />
                <p>No notes</p>
              </div>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-2">
              {(filteredNotes || []).map((note, index) => (
                <NoteItem
                  key={note.ID}
                  note={note}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onNoteClick={onNoteClick}
                  disabled={isLoading || !note.Content}
                />
              ))}
            </div>
          )}

        </CardContent>
      </Card>

    </div >
  )
}
