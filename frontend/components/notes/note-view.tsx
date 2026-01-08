"use client"

import { useEffect, useMemo, useState } from "react"
import { CardContent } from "@/components/ui/card"
import { NoteItem } from "./note-item"
import { FileText, Filter, Search, StickyNote, X } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { useRouter } from "next/navigation"
import { useCoursesBySemester } from "@/hooks/use-courses"
import { useAuthContext } from "../provider/auth-provider"
import { GlassCard } from "../ui/glass-card"
import { EmptyState } from "../ui/empty-state"

interface Filter {
  course: string | null
}


interface NoteViewProps {
  title: string
  notes: models.LocalNote[]
  onNoteClick: (noteID: number) => void
  onEdit: (note: models.LocalNote, column: string, value: string) => void
  onDelete: (note: models.LocalNote) => void
  setAddOpen: (open: boolean) => void
  filter: Filter
  isLoading?: boolean
}

export function NoteView({ title, notes, onNoteClick, onDelete, onEdit, setAddOpen, filter, isLoading }: NoteViewProps) {
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
          note.Title.toLowerCase().includes(searchTerm.toLowerCase())
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

  if (notes.length === 0) {
    return (
      <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
        <EmptyState
          icon={StickyNote}
          title="No notes found"
          description="Create a note to get started"
          className="flex-1 items-center"
          onClick={() => setAddOpen(true)}
          buttonText="Create Note"
        />
      </div>
    )
  }


  return (
    <div className="flex flex-col flex-1 space-y-6">
      {/* Search and Filters */}

      <GlassCard variant="board" className="flex-grow-0 flex-row">
        <CardContent className="p-5 flex-1">
            <div className="flex flex-1  lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-4">
              
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search notes by title or keywords..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-white/5 border-white/10    transition-all duration-300 h-10"
                />
              </div>


              <Select value={selectedCourse} onValueChange={onCourseChange}>
                <SelectTrigger className="w-60 bg-white/5 border-white/10 h-10">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
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


        </CardContent>
      </GlassCard>

      {(filteredNotes || []).length === 0 ? (
        <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
          <EmptyState
            icon={StickyNote}
            title="No notes found"
            description="Clear your filters or search terms to see all notes"
            className="flex-1 items-center"
            onClick={clearFilters}
            buttonText="Clear Filters"
          />
        </div>


      ) : (



        <div>
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
        </div>

      )}

    </div >
  )
}
