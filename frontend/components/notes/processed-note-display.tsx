"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Download, Copy, Edit, Save, X } from "lucide-react"
import { useState } from "react"
import { Textarea } from "@/components/ui/textarea"

interface ProcessedNote {
  id: string
  title: string
  keywords: string[]
  content: string
  courseId: string
  lectureTitle: string
  createdAt: Date
  duration: number
}

interface ProcessedNoteDisplayProps {
  note: ProcessedNote
  onSave?: (note: ProcessedNote) => void
  onDiscard?: () => void
}

export function ProcessedNoteDisplay({ note, onSave, onDiscard }: ProcessedNoteDisplayProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editedNote, setEditedNote] = useState(note)

  const copyToClipboard = () => {
    const fullContent = `# ${note.title}\n\n**Keywords:** ${note.keywords.join(", ")}\n\n${note.content}`
    navigator.clipboard.writeText(fullContent)
  }

  const downloadNote = () => {
    const fullContent = `# ${note.title}\n\n**Keywords:** ${note.keywords.join(", ")}\n\n${note.content}`
    const blob = new Blob([fullContent], { type: "text/markdown" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `${note.title.replace(/\s+/g, "-").toLowerCase()}.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleSave = () => {
    if (onSave) {
      onSave(editedNote)
    }
    setIsEditing(false)
  }

  const handleDiscard = () => {
    setEditedNote(note)
    setIsEditing(false)
    if (onDiscard) {
      onDiscard()
    }
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, "0")}`
  }

  return (
    <Card className="border-0 glass">
      <CardHeader className="flex flex-row justify-between items-center">
        <div className="space-y-2">
          <CardTitle className="text-xl text-white">{note.title}</CardTitle>
          <div className="flex items-center space-x-4 text-sm text-gray-400">
            <span>Course: {note.lectureTitle}</span>
            <span>Duration: {formatDuration(note.duration)}</span>
            <span>Created: {note.createdAt.toLocaleDateString()}</span>
          </div>
        </div>
        <div className="flex space-x-2">
          {!isEditing ? (
            <>
              <Button variant="outline" size="sm" onClick={() => setIsEditing(true)} className="bg-transparent border-gray-600">
                <Edit className="mr-1 w-4 h-4" />
                Edit
              </Button>
              <Button variant="outline" size="sm" onClick={copyToClipboard} className="bg-transparent border-gray-600">
                <Copy className="mr-1 w-4 h-4" />
                Copy
              </Button>
              <Button variant="outline" size="sm" onClick={downloadNote} className="bg-transparent border-gray-600">
                <Download className="mr-1 w-4 h-4" />
                Download
              </Button>
            </>
          ) : (
            <>
              <Button size="sm" onClick={handleSave} className="bg-green-600 hover:bg-green-700">
                <Save className="mr-1 w-4 h-4" />
                Save
              </Button>
              <Button variant="outline" size="sm" onClick={handleDiscard} className="bg-transparent border-gray-600">
                <X className="mr-1 w-4 h-4" />
                Cancel
              </Button>
            </>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Keywords */}
        <div>
          <h3 className="mb-2 text-sm font-medium text-gray-300">Keywords</h3>
          <div className="flex flex-wrap gap-2">
            {note.keywords.map((keyword, index) => (
              <Badge key={index} variant="secondary" className="text-blue-300 bg-blue-500/20 border-blue-500/30">
                {keyword}
              </Badge>
            ))}
          </div>
        </div>

        {/* Content */}
        <div>
          <h3 className="mb-2 text-sm font-medium text-gray-300">Content</h3>
          {isEditing ? (
            <Textarea
              value={editedNote.content}
              onChange={(e) => setEditedNote(prev => ({ ...prev, content: e.target.value }))}
              className="bg-gray-800/50 border-gray-600 min-h-[300px] text-gray-300"
              placeholder="Edit your note content..."
            />
          ) : (
            <div className="overflow-y-auto p-4 max-h-96 rounded-lg border border-gray-600 bg-gray-800/50">
              <div className="max-w-none prose prose-invert prose-sm">
                <div 
                  className="text-gray-300 whitespace-pre-wrap"
                  dangerouslySetInnerHTML={{ __html: renderMarkdown(note.content) }}
                />
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

// Simple markdown renderer (you can replace with a proper library like react-markdown)
function renderMarkdown(markdown: string): string {
  return markdown
    .replace(/^### (.*$)/gim, '<h3 class="mb-2 text-lg font-semibold text-white">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 class="mb-3 text-xl font-semibold text-white">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 class="mb-4 text-2xl font-bold text-white">$1</h1>')
    .replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold text-white">$1</strong>')
    .replace(/\*(.*?)\*/g, '<em class="italic text-gray-300">$1</em>')
    .replace(/^- (.*$)/gim, '<li class="ml-4 text-gray-300">$1</li>')
    .replace(/^(\d+)\. (.*$)/gim, '<li class="ml-4 text-gray-300">$2</li>')
    .replace(/\n\n/g, '<br><br>')
} 