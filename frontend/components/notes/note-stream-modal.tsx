"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Loader2, X, Sparkles, BookOpen } from "lucide-react"
import { useCourses } from "@/hooks/use-courses"
import { StyledMarkdownRenderer } from "../markdown/markdown-renderer"
import { useEffect, useRef } from "react"
import { models } from "@/wailsjs/go/models"

interface NoteStreamModalProps {
  isOpen: boolean
  onClose: () => void
  onStop?: () => void
  // Streaming state from useStreamNote
  note: models.LocalNote
  content: string
  isStreaming: boolean
  error: string | null
}

export function NoteStreamModal({
  isOpen,
  onClose,
  onStop,
  note,
  content,
  isStreaming,
  error
}: NoteStreamModalProps) {
  const { data: courses } = useCourses()
  const { CourseID ,Subject, Title } = note
  const course = courses?.find(c => c.ID === CourseID)
  const contentEndRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom as content streams
  useEffect(() => {
    if (isStreaming && content && contentEndRef.current) {
      contentEndRef.current.scrollIntoView({ behavior: 'smooth', block: 'end' })
    }
  }, [content, isStreaming])

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col glass border-white/10 p-0 overflow-hidden gap-0">
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5 flex-shrink-0">
          <div className="flex items-center justify-between">
            <div className="space-y-1 flex-1">
              <div className="flex items-center gap-2">
                <BookOpen className="w-4 h-4 text-blue-400" />
                <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">
                  {Subject}
                </span>
                {isStreaming && (
                  <Badge variant="secondary" className="bg-blue-500/10 text-blue-400 border-blue-500/20">
                    <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                    Generating...
                  </Badge>
                )}
              </div>
              <DialogTitle className="text-xl font-semibold text-white">
                {Title}
              </DialogTitle>
              {course && (
                <div className="flex items-center gap-2 mt-2">
                  <div className={`h-2 w-2 rounded-full ${course.Color}`} />
                  <span className="text-xs text-gray-400">{course.Code}</span>
                </div>
              )}
            </div>
            {isStreaming && onStop && (
              <Button
                variant="outline"
                size="sm"
                onClick={onStop}
                className="border-red-500/20 bg-red-500/10 hover:bg-red-500/20 text-red-400"
              >
                <X className="w-4 h-4 mr-1" />
                Stop
              </Button>
            )}
          </div>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-4">
            {error ? (
              <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400">
                {error}
              </div>
            ) : content ? (
              <div className="space-y-2">
                <StyledMarkdownRenderer
                  content={content}
                  className="bg-white/5 rounded-xl p-6 border border-white/5 min-h-[300px]"
                />
                {isStreaming && (
                  <div className="flex items-center gap-2 text-sm text-blue-400">
                    <Sparkles className="w-4 h-4 animate-pulse" />
                    <span>AI is generating content...</span>
                    <span className="inline-block w-2 h-4 bg-blue-400 animate-pulse ml-1" ref={contentEndRef} />
                  </div>
                )}
                {!isStreaming && <div ref={contentEndRef} />}
              </div>
            ) : (
              <div className="text-center py-12 text-gray-400">
                <Loader2 className="w-8 h-8 mx-auto mb-3 animate-spin" />
                <p>Waiting for content to stream...</p>
              </div>
            )}
          </div>
        </div>

      </DialogContent>
    </Dialog>
  )
}

