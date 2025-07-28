"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { FileAudio, Search, MoreVertical, Play, Download, Edit, Trash2, Eye, Calendar, Clock } from "lucide-react"
import { format } from "date-fns"

const mockTranscripts = [
  {
    id: 1,
    title: "Introduction to Data Structures",
    course: "CS 201",
    courseColor: "bg-blue-500",
    duration: "45:30",
    createdAt: new Date(2024, 0, 15),
    fileSize: "12.5 MB",
    audioFormat: "MP3",
    transcriptLength: 2800,
  },
  {
    id: 2,
    title: "Calculus Integration Techniques",
    course: "MATH 301",
    courseColor: "bg-green-500",
    duration: "38:15",
    createdAt: new Date(2024, 0, 12),
    fileSize: "9.8 MB",
    audioFormat: "WAV",
    transcriptLength: 2100,
  },
  {
    id: 3,
    title: "Machine Learning Fundamentals",
    course: "AI 401",
    courseColor: "bg-purple-500",
    duration: "52:45",
    createdAt: new Date(2024, 0, 10),
    fileSize: "15.2 MB",
    audioFormat: "MP3",
    transcriptLength: 3400,
  },
]

export function AllTranscriptsView() {
  const [searchTerm, setSearchTerm] = useState("")
  const [transcripts] = useState(mockTranscripts)

  const filteredTranscripts = transcripts.filter(
    (transcript) =>
      transcript.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      transcript.course.toLowerCase().includes(searchTerm.toLowerCase()),
  )

  const handlePlay = (transcriptId: number) => {
    console.log("Playing transcript:", transcriptId)
  }

  const handleView = (transcriptId: number) => {
    console.log("Viewing transcript:", transcriptId)
  }

  const handleEdit = (transcriptId: number) => {
    console.log("Editing transcript:", transcriptId)
  }

  const handleDelete = (transcriptId: number) => {
    console.log("Deleting transcript:", transcriptId)
  }

  const handleDownload = (transcriptId: number) => {
    console.log("Downloading transcript:", transcriptId)
  }

  return (
    <div className="space-y-6">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center justify-between text-white">
            <div className="flex items-center space-x-2">
              <FileAudio className="h-5 w-5 text-green-400" />
              <span>Audio Transcripts</span>
              <Badge variant="outline" className="border-gray-600">
                {filteredTranscripts.length}
              </Badge>
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="relative mb-6">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search transcripts by title or course..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 bg-gray-800/50 border-gray-600"
            />
          </div>

          <div className="grid md:grid-cols-1 xl:grid-cols-2 gap-4">
            {filteredTranscripts.map((transcript) => (
              <Card key={transcript.id} className="glass-dark border-0 hover:glass-hover transition-all duration-300">
                <CardContent className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 space-y-2">
                      <div className="flex items-center space-x-3">
                        <div className={`w-3 h-3 rounded-full ${transcript.courseColor}`} />
                        <h3 className="font-semibold text-white">{transcript.title}</h3>
                        <Badge variant="outline" className="text-xs border-gray-600">
                          {transcript.audioFormat}
                        </Badge>
                      </div>

                      <div className="flex items-center space-x-4 text-sm text-gray-400">
                        <span>{transcript.course}</span>
                        <span>•</span>
                        <div className="flex items-center space-x-1">
                          <Clock className="h-3 w-3" />
                          <span>{transcript.duration}</span>
                        </div>
                        <span>•</span>
                        <span>{transcript.fileSize}</span>
                        <span>•</span>
                        <div className="flex items-center space-x-1">
                          <Calendar className="h-3 w-3" />
                          <span>{format(transcript.createdAt, "MMM d, yyyy")}</span>
                        </div>
                      </div>

                      <div className="text-sm text-gray-400">
                        <span>{transcript.transcriptLength.toLocaleString()} words transcribed</span>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handlePlay(transcript.id)}
                        className="text-green-400 hover:text-green-300"
                      >
                        <Play className="h-4 w-4" />
                      </Button>

                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleView(transcript.id)}
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
                          <DropdownMenuItem onClick={() => handleEdit(transcript.id)}>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleDownload(transcript.id)}>
                            <Download className="mr-2 h-4 w-4" />
                            Download
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDelete(transcript.id)}
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

            {filteredTranscripts.length === 0 && (
              <div className="text-center py-12">
                <FileAudio className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-300 mb-2">No transcripts found</h3>
                <p className="text-gray-400">
                  {searchTerm ? "Try adjusting your search terms" : "Create your first transcript to get started"}
                </p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
