"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { Sparkles, Loader2, CheckCircle } from "lucide-react"
import { useCourses } from "@/hooks/use-courses"
import { useToast } from "@/hooks/use-toast"
import { useGenerateNote, useCreateNote } from "@/hooks/use-notes"
import { YouTubeVideoInput } from "./youtube-video-input"
import { YouTubeVideo } from "@/types/models"

// Fixed list of subjects as requested
const SUBJECTS = [
  "Mathematics", "Physics", "Chemistry", "Biology",
  "Computer Science", "Engineering",
  "Literature", "History", "Philosophy",
  "Business", "Economics", "Psychology",
  "Languages", "Arts", "Music"
]

interface NoteGenerationFormProps {
  onNoteGenerated?: (note: any) => void
}

export function NoteGenerationForm({ onNoteGenerated }: NoteGenerationFormProps) {
  const [title, setTitle] = useState("")
  const [selectedCourse, setSelectedCourse] = useState("")
  const [selectedSubject, setSelectedSubject] = useState("")
  const [generatedNote, setGeneratedNote] = useState("")
  const [keywords, setKeywords] = useState<string[]>([])
  const [youtubeVideos, setYoutubeVideos] = useState<YouTubeVideo[]>([])

  const { data: courses, isLoading: coursesLoading } = useCourses()
  const { toast } = useToast()
  const generateNote = useGenerateNote()
  const createNote = useCreateNote()

  const handleGenerate = async () => {
    if (!title || !selectedCourse || !selectedSubject) {
      toast({
        title: "Missing Information",
        description: "Please fill in all fields before generating a note.",
        variant: "destructive"
      })
      return
    }

    try {
      const result = await generateNote.mutateAsync({
        title,
        courseId: selectedCourse,
        subject: selectedSubject
      })

      setGeneratedNote(result.Content)
      setKeywords(result.Keywords)

      toast({
        title: "Note Generated",
        description: "Your AI-generated note has been created successfully!",
      })

      if (onNoteGenerated) {
        onNoteGenerated(result)
      }
    } catch (error) {
      toast({
        title: "Generation Failed",
        description: "Failed to generate note. Please try again.",
        variant: "destructive"
      })
    }
  }

  const handleSave = async () => {
    if (!generatedNote) return

    try {
      await createNote.mutateAsync({
        UserID: 1, // TODO: Get from auth
        CourseID: parseInt(selectedCourse),
        Title: title,
        Subject: selectedSubject,
        Content: generatedNote,
        Keywords: keywords,
        YouTubeVideos: youtubeVideos
      })

      toast({
        title: "Note Saved",
        description: "Your note has been saved successfully!",
      })

      // Reset form
      setTitle("")
      setSelectedCourse("")
      setSelectedSubject("")
      setGeneratedNote("")
      setKeywords([])
      setYoutubeVideos([])
    } catch (error) {
      toast({
        title: "Save Failed",
        description: "Failed to save note. Please try again.",
        variant: "destructive"
      })
    }
  }

  return (
    <div className="space-y-6">
      <Card className="border-0 glass">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <Sparkles className="w-5 h-5 text-purple-400" />
            <span>Generate AI Study Note</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <Label htmlFor="title" className="text-sm font-medium text-gray-300">
                Note Title
              </Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Enter note title..."
                className="mt-1 placeholder-gray-400 text-white border-gray-600 bg-gray-800/50"
              />
            </div>

            <div>
              <Label className="text-sm font-medium text-gray-300">Course</Label>
              <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                <SelectTrigger className="mt-1 border-gray-600 bg-gray-800/50">
                  <SelectValue placeholder="Select a course" />
                </SelectTrigger>
                <SelectContent className="border-gray-600 glass">
                  {coursesLoading ? (
                    <SelectItem value="" disabled>Loading courses...</SelectItem>
                  ) : (
                    courses?.map((course) => (
                      <SelectItem key={course.ID} value={course.ID.toString()}>
                        {course.Code} - {course.Name}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <Label className="text-sm font-medium text-gray-300">Subject</Label>
            <Select value={selectedSubject} onValueChange={setSelectedSubject}>
              <SelectTrigger className="mt-1 border-gray-600 bg-gray-800/50">
                <SelectValue placeholder="Select a subject" />
              </SelectTrigger>
              <SelectContent className="border-gray-600 glass">
                {SUBJECTS.map((subject) => (
                  <SelectItem key={subject} value={subject}>
                    {subject}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <Button
            onClick={handleGenerate}
            disabled={!title || !selectedCourse || !selectedSubject || generateNote.isPending}
            className="w-full bg-gradient-to-r from-purple-500 to-blue-500 hover:from-purple-600 hover:to-blue-600"
          >
            {generateNote.isPending ? (
              <>
                <Loader2 className="mr-2 w-4 h-4 animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <Sparkles className="mr-2 w-4 h-4" />
                Generate Note
              </>
            )}
          </Button>
        </CardContent>
      </Card>

      {generatedNote && (
        <Card className="border-0 glass">
          <CardHeader>
            <CardTitle className="flex justify-between items-center text-white">
              <span>Generated Note</span>
              <Button
                onClick={handleSave}
                size="sm"
                className="bg-green-600 hover:bg-green-700"
              >
                <CheckCircle className="mr-2 w-4 h-4" />
                Save Note
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {keywords.length > 0 && (
              <div>
                <Label className="block mb-2 text-sm font-medium text-gray-300">
                  Keywords
                </Label>
                <div className="flex flex-wrap gap-2">
                  {keywords.map((keyword, index) => (
                    <span
                      key={index}
                      className="px-2 py-1 text-sm text-purple-300 rounded-md bg-purple-500/20"
                    >
                      {keyword}
                    </span>
                  ))}
                </div>
              </div>
            )}
            
            <div>
              <Label className="block mb-2 text-sm font-medium text-gray-300">
                Content
              </Label>
              <div className="overflow-y-auto p-4 max-h-96 rounded-lg bg-gray-800/30">
                <pre className="font-sans text-sm text-gray-300 whitespace-pre-wrap">
                  {generatedNote}
                </pre>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* YouTube Videos Section */}
      {generatedNote && (
        <YouTubeVideoInput
          videos={youtubeVideos}
          onVideosChange={setYoutubeVideos}
          disabled={false}
        />
      )}
    </div>
  )
} 