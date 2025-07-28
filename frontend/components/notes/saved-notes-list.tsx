"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { FileText, Search, MoreVertical, Edit, Trash2, Download, Eye, Calendar } from "lucide-react"
import { format } from "date-fns"

const mockNotes = [
  {
    id: 1,
    title: "Data Structures Overview",
    course: "CS 201",
    courseColor: "bg-blue-500",
    type: "summary",
    createdAt: new Date(2024, 0, 15),
    updatedAt: new Date(2024, 0, 16),
    wordCount: 1250,
    tags: ["arrays", "linked-lists", "trees"],
  },
  {
    id: 2,
    title: "Calculus Integration Methods",
    course: "MATH 301",
    courseColor: "bg-green-500",
    type: "study-guide",
    createdAt: new Date(2024, 0, 12),
    updatedAt: new Date(2024, 0, 14),
    wordCount: 890,
    tags: ["integration", "substitution", "parts"],
  },
  {
    id: 3,
    title: "Machine Learning Basics",
    course: "AI 401",
    courseColor: "bg-purple-500",
    type: "outline",
    createdAt: new Date(2024, 0, 10),
    updatedAt: new Date(2024, 0, 10),
    wordCount: 2100,
    tags: ["supervised", "unsupervised", "neural-networks"],
  },
]

const noteTypeLabels = {
  summary: "Summary",
  outline: "Outline",
  flashcards: "Flashcards",
  "study-guide": "Study Guide",
}

export function SavedNotesList() {
  const [searchTerm, setSearchTerm] = useState("")
  const [notes] = useState(mockNotes)

  const filteredNotes = notes.filter(
    (note) =>
      note.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      note.course.toLowerCase().includes(searchTerm.toLowerCase()) ||
      note.tags.some((tag) => tag.toLowerCase().includes(searchTerm.toLowerCase())),
  )

  const handleView = (noteId: number) => {
    console.log("Viewing note:", noteId)
  }

  const handleEdit = (noteId: number) => {
    console.log("Editing note:", noteId)
  }

  const handleDelete = (noteId: number) => {
    console.log("Deleting note:", noteId)
  }

  const handleDownload = (noteId: number) => {
    console.log("Downloading note:", noteId)
  }

  return (
    <div className="space-y-6">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center justify-between text-white">
            <div className="flex items-center space-x-2">
              <FileText className="h-5 w-5 text-blue-400" />
              <span>Saved Notes</span>
              <Badge variant="outline" className="border-gray-600">
                {filteredNotes.length}
              </Badge>
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="relative mb-6">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search notes by title, course, or tags..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 bg-gray-800/50 border-gray-600"
            />
          </div>

          <div className="grid md:grid-cols-1 xl:grid-cols-2 gap-4">
            {filteredNotes.map((note) => (
              <Card key={note.id} className="glass-dark border-0 hover:glass-hover transition-all duration-300">
                <CardContent className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 space-y-2">
                      <div className="flex items-center space-x-3">
                        <div className={`w-3 h-3 rounded-full ${note.courseColor}`} />
                        <h3 className="font-semibold text-white">{note.title}</h3>
                        <Badge variant="outline" className="text-xs border-gray-600">
                          {noteTypeLabels[note.type as keyof typeof noteTypeLabels]}
                        </Badge>
                      </div>

                      <div className="flex items-center space-x-4 text-sm text-gray-400">
                        <span>{note.course}</span>
                        <span>•</span>
                        <span>{note.wordCount} words</span>
                        <span>•</span>
                        <div className="flex items-center space-x-1">
                          <Calendar className="h-3 w-3" />
                          <span>{format(note.updatedAt, "MMM d, yyyy")}</span>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-1">
                        {note.tags.map((tag) => (
                          <Badge key={tag} variant="secondary" className="text-xs bg-gray-700/50 text-gray-300">
                            #{tag}
                          </Badge>
                        ))}
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleView(note.id)}
                        className="text-blue-400 hover:text-blue-300"
                      >
                        <Eye className="h-4 w-4" />
                      </Button>

                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="glass border-gray-600">
                          <DropdownMenuItem onClick={() => handleEdit(note.id)}>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleDownload(note.id)}>
                            <Download className="mr-2 h-4 w-4" />
                            Download
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(note.id)}
                            className="text-red-400 hover:text-red-300"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}

            {filteredNotes.length === 0 && (
              <div className="text-center py-12">
                <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-300 mb-2">No notes found</h3>
                <p className="text-gray-400">
                  {searchTerm ? "Try adjusting your search terms" : "Create your first note to get started"}
                </p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
