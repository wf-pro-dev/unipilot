"use client"

import { useEffect, useMemo, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { NoteItem } from "./note-item"
import { CalendarDays, CheckCircle2, Filter, Loader2, Search, X } from "lucide-react"
import { note } from "@/wailsjs/go/models"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { useRouter } from "next/navigation"
import { useCoursesBySemester } from "@/hooks/use-courses"
import { useAuthContext } from "../provider/auth-provider"

interface Filter {
  course: string | null
}


interface NoteViewProps {
  title: string
  notes: note.LocalNote[]
  onNoteClick: (noteID: number) => void
  onEdit: (note: note.LocalNote, column: string, value: string) => void
  onDelete: (note: note.LocalNote) => void
  filter: Filter
  isLoading?: boolean
}

export function NoteView({ title, notes, onNoteClick, onDelete, onEdit, filter, isLoading }: NoteViewProps) {
  const router = useRouter()
  const [selectedCourse, setSelectedCourse] = useState(filter.course || "all")
  const [searchTerm, setSearchTerm] = useState("")

  const { user } = useAuthContext()
  const { data: courses } = useCoursesBySemester(user?.Semester || "FALL 2025")

  const courseCodes = Array.from(new Set((courses || []).map((course) => course.Code)))

  const filteredNotes = useMemo(() => {
    return notes
    .filter((note) => {
      const matchesSearch =
        note.Title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        note.Keywords.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesCourse = selectedCourse === "all" || note.CourseCode === selectedCourse
      return matchesSearch && matchesCourse
    })
    .sort((a, b) => {
      return new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime()
    })
  }, [notes, searchTerm, selectedCourse])

  const hasActiveFilters = searchTerm !== "" || selectedCourse !== "all"

  const clearFilters = () => {
    setSearchTerm("")
    router.push("/notes")
  }

  const onCourseChange = (course: string) => {
    router.push(`/notes?course=${course}`)
  }

  useEffect(() => {
    if (filter.course) {
      setSelectedCourse(filter.course)
    }
  }, [filter])

  return (
    <div className="space-y-4">
      {/* Search and Filters */}
      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <Input
                    placeholder="Search notes by title or keywords..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10 bg-gray-800/50 border-gray-600"
                  />
                </div>
              </div>


              <Select value={selectedCourse} onValueChange={onCourseChange}>
                <SelectTrigger className="w-48 bg-gray-800/50 border-gray-600">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  <SelectItem value="all">All Courses</SelectItem>
                  {courseCodes.map((courseCode) => (
                    <SelectItem key={courseCode} value={courseCode}>
                      {courseCode}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {hasActiveFilters && (
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Filter className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-400">Active filters:</span>
                  {searchTerm && (
                    <Badge variant="secondary" className="bg-blue-500/20 text-blue-400">
                      Search: {searchTerm}
                    </Badge>
                  )}
                  {selectedCourse !== "all" && (
                    <Badge variant="secondary" className="bg-green-500/20 text-green-400">
                      {selectedCourse}
                    </Badge>
                  )}

                </div>
                <Button variant="ghost" size="sm" onClick={clearFilters} className="text-gray-400 hover:text-white">
                  <X className="h-4 w-4 mr-1" />
                  Clear
                </Button>
              </div>
            )}

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
