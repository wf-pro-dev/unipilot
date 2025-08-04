"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Search, Filter, Calendar, BookOpen, Tag, Play, Trash2, Edit } from "lucide-react"
import { useNotes, useDeleteNote } from "@/hooks/use-notes"
import { useCourses } from "@/hooks/use-courses"
import { useToast } from "@/hooks/use-toast"
import { Note } from "@/types/models"

export function NotesList() {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedCourse, setSelectedCourse] = useState("")
  const [selectedSubject, setSelectedSubject] = useState("")

  const { data: notes, isLoading } = useNotes()
  const { data: courses } = useCourses()
  const deleteNote = useDeleteNote()
  const { toast } = useToast()

  // Get unique subjects from notes
  const subjects = [...new Set(notes?.map(note => note.Subject) || [])]

  // Filter notes based on search and filters
  const filteredNotes = notes?.filter(note => {
    const matchesSearch = searchQuery === "" || 
      note.Title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      note.Content.toLowerCase().includes(searchQuery.toLowerCase()) ||
      note.Keywords.some(keyword => keyword.toLowerCase().includes(searchQuery.toLowerCase()))

    const matchesCourse = selectedCourse === "" || note.CourseID.toString() === selectedCourse
    const matchesSubject = selectedSubject === "" || note.Subject === selectedSubject

    return matchesSearch && matchesCourse && matchesSubject
  }) || []

  const handleDeleteNote = async (noteId: number) => {
    try {
      await deleteNote.mutateAsync(noteId)
      toast({
        title: "Note Deleted",
        description: "Note has been deleted successfully.",
      })
    } catch (error) {
      toast({
        title: "Delete Failed",
        description: "Failed to delete note. Please try again.",
        variant: "destructive"
      })
    }
  }

  const formatDate = (date: Date) => {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="w-8 h-8 rounded-full border-b-2 border-purple-500 animate-spin"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
      <Card className="border-0 glass">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <Search className="w-5 h-5 text-blue-400" />
            <span>Search & Filter Notes</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div>
              <Label className="text-sm font-medium text-gray-300">Search</Label>
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search notes, keywords..."
                className="mt-1 placeholder-gray-400 text-white border-gray-600 bg-gray-800/50"
              />
            </div>

            <div>
              <Label className="text-sm font-medium text-gray-300">Course</Label>
              <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                <SelectTrigger className="mt-1 border-gray-600 bg-gray-800/50">
                  <SelectValue placeholder="All courses" />
                </SelectTrigger>
                <SelectContent className="border-gray-600 glass">
                  <SelectItem value="">All courses</SelectItem>
                  {courses?.map((course) => (
                    <SelectItem key={course.ID} value={course.ID.toString()}>
                      {course.Code} - {course.Name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label className="text-sm font-medium text-gray-300">Subject</Label>
              <Select value={selectedSubject} onValueChange={setSelectedSubject}>
                <SelectTrigger className="mt-1 border-gray-600 bg-gray-800/50">
                  <SelectValue placeholder="All subjects" />
                </SelectTrigger>
                <SelectContent className="border-gray-600 glass">
                  <SelectItem value="">All subjects</SelectItem>
                  {subjects.map((subject) => (
                    <SelectItem key={subject} value={subject}>
                      {subject}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Notes List */}
      <div className="space-y-4">
        {filteredNotes.length === 0 ? (
          <Card className="border-0 glass">
            <CardContent className="flex flex-col justify-center items-center py-12">
              <BookOpen className="mb-4 w-12 h-12 text-gray-400" />
              <h3 className="mb-2 text-lg font-medium text-gray-300">No notes found</h3>
              <p className="text-center text-gray-400">
                {searchQuery || selectedCourse || selectedSubject 
                  ? "Try adjusting your search or filters."
                  : "Create your first AI-generated note to get started."
                }
              </p>
            </CardContent>
          </Card>
        ) : (
          filteredNotes.map((note) => (
            <Card key={note.ID} className="border-0 transition-colors glass hover:bg-gray-800/20">
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <CardTitle className="mb-2 text-white">{note.Title}</CardTitle>
                    <div className="flex items-center space-x-4 text-sm text-gray-400">
                      <div className="flex items-center space-x-1">
                        <Calendar className="w-4 h-4" />
                        <span>{formatDate(note.CreatedAt)}</span>
                      </div>
                      <div className="flex items-center space-x-1">
                        <BookOpen className="w-4 h-4" />
                        <span>{courses?.find(c => c.ID === note.CourseID)?.Code || 'Unknown'}</span>
                      </div>
                      <Badge variant="secondary" className="text-purple-300 bg-purple-500/20">
                        {note.Subject}
                      </Badge>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {note.YouTubeVideos && note.YouTubeVideos.length > 0 && (
                      <Button size="sm" variant="ghost" className="text-blue-400 hover:text-blue-300">
                        <Play className="mr-1 w-4 h-4" />
                        {note.YouTubeVideos.length}
                      </Button>
                    )}
                    <Button size="sm" variant="ghost" className="text-gray-400 hover:text-gray-300">
                      <Edit className="w-4 h-4" />
                    </Button>
                    <Button 
                      size="sm" 
                      variant="ghost" 
                      className="text-red-400 hover:text-red-300"
                      onClick={() => handleDeleteNote(note.ID)}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {/* Keywords */}
                  {note.Keywords && note.Keywords.length > 0 && (
                    <div>
                      <Label className="flex items-center mb-2 space-x-1 text-sm font-medium text-gray-300">
                        <Tag className="w-4 h-4" />
                        <span>Keywords</span>
                      </Label>
                      <div className="flex flex-wrap gap-2">
                        {note.Keywords.map((keyword, index) => (
                          <Badge key={index} variant="outline" className="text-purple-300 bg-purple-500/10 border-purple-500/30">
                            {keyword}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Content Preview */}
                  <div>
                    <Label className="mb-2 text-sm font-medium text-gray-300">Preview</Label>
                    <div className="overflow-hidden p-3 max-h-32 rounded-lg bg-gray-800/30">
                      <p className="text-sm text-gray-300 line-clamp-3">
                        {note.Content.replace(/[#*`]/g, '').substring(0, 200)}...
                      </p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  )
} 