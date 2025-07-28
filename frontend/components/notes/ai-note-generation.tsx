"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Sparkles, Download, Copy, RefreshCw } from "lucide-react"

const courses = [
  { value: "cs-201", label: "CS 201 - Data Structures" },
  { value: "math-301", label: "MATH 301 - Calculus III" },
  { value: "ai-401", label: "AI 401 - Artificial Intelligence" },
  { value: "eng-102", label: "ENG 102 - English Composition" },
]

const noteTypes = [
  { value: "summary", label: "Summary Notes" },
  { value: "outline", label: "Topic Outline" },
  { value: "flashcards", label: "Flashcards" },
  { value: "study-guide", label: "Study Guide" },
]

export function AINotesGeneration() {
  const [topic, setTopic] = useState("")
  const [selectedCourse, setSelectedCourse] = useState("")
  const [noteType, setNoteType] = useState("")
  const [generatedNotes, setGeneratedNotes] = useState("")
  const [isGenerating, setIsGenerating] = useState(false)

  const handleGenerate = async () => {
    if (!topic || !selectedCourse || !noteType) return

    setIsGenerating(true)

    // Simulate AI generation
    setTimeout(() => {
      const mockNotes = `# ${topic} - Study Notes

## Key Concepts
- Important concept 1: Detailed explanation of the first key concept
- Important concept 2: Detailed explanation of the second key concept
- Important concept 3: Detailed explanation of the third key concept

## Examples
1. **Example 1**: Practical application showing how the concept works
2. **Example 2**: Another example demonstrating different use cases

## Summary
This topic covers the fundamental aspects of ${topic} and its applications in ${courses.find((c) => c.value === selectedCourse)?.label}.

## Study Tips
- Review the key concepts regularly
- Practice with the provided examples
- Connect this topic to previous lessons`

      setGeneratedNotes(mockNotes)
      setIsGenerating(false)
    }, 2000)
  }

  const copyToClipboard = () => {
    navigator.clipboard.writeText(generatedNotes)
  }

  const downloadNotes = () => {
    const blob = new Blob([generatedNotes], { type: "text/markdown" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `${topic.replace(/\s+/g, "-").toLowerCase()}-notes.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <Sparkles className="h-5 w-5 text-purple-400" />
            <span>AI Notes Generation</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-300 block mb-2">Course</label>
              <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                  <SelectValue placeholder="Select a course" />
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  {courses.map((course) => (
                    <SelectItem key={course.value} value={course.value}>
                      {course.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium text-gray-300 block mb-2">Note Type</label>
              <Select value={noteType} onValueChange={setNoteType}>
                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                  <SelectValue placeholder="Select note type" />
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  {noteTypes.map((type) => (
                    <SelectItem key={type.value} value={type.value}>
                      {type.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <label className="text-sm font-medium text-gray-300 block mb-2">Topic or Content</label>
            <Textarea
              placeholder="Enter the topic you want to generate notes for, or paste content to summarize..."
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              className="bg-gray-800/50 border-gray-600 min-h-[100px]"
            />
          </div>

          <Button
            onClick={handleGenerate}
            disabled={!topic || !selectedCourse || !noteType || isGenerating}
            className="w-full bg-gradient-to-r from-purple-500 to-pink-600 hover:from-purple-600 hover:to-pink-700"
          >
            {isGenerating ? (
              <>
                <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                Generating Notes...
              </>
            ) : (
              <>
                <Sparkles className="h-4 w-4 mr-2" />
                Generate Notes
              </>
            )}
          </Button>
        </CardContent>
      </Card>

      {generatedNotes && (
        <Card className="glass border-0">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-white">Generated Notes</CardTitle>
            <div className="flex space-x-2">
              <Button variant="outline" size="sm" onClick={copyToClipboard} className="border-gray-600 bg-transparent">
                <Copy className="h-4 w-4 mr-1" />
                Copy
              </Button>
              <Button variant="outline" size="sm" onClick={downloadNotes} className="border-gray-600 bg-transparent">
                <Download className="h-4 w-4 mr-1" />
                Download
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-600">
              <pre className="text-gray-300 whitespace-pre-wrap text-sm">{generatedNotes}</pre>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
