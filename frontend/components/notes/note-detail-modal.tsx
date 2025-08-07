"use client"

import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { BookOpen, Tag, Video, Calendar, User, X, List } from "lucide-react"
import { note } from "@/wailsjs/go/models"
import { NoteVideo } from "./note-video"
import { useMemo, useState } from "react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCourses } from "@/hooks/use-courses"
import { useNotes } from "@/hooks/use-notes"
import { AddVideosDialog } from "./add-videos-dialog"
import { toast } from "sonner"
import { StyledMarkdownRenderer } from "./markdown-renderer"

interface NoteDetailModalProps {
  noteID: number | null
  isOpen: boolean
  onClose: () => void
  onEdit: (note: note.LocalNote, column: string, value: string) => void
  onDelete: (note: note.LocalNote) => void
}

export function NoteDetailModal({
  noteID,
  isOpen,
  onClose,
  onEdit,
  onDelete
}: NoteDetailModalProps) {

  const [activeView, setActiveView] = useState("videos")
  const { data: courses } = useCourses()
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)

  const { data: notes } = useNotes()
  const note = notes?.find(n => n.ID === noteID)

  const course = courses?.find(c => c.Code === note?.CourseCode)

  const handleDelete = () => {
    if (confirm("Are you sure you want to delete this note?")) {
      onDelete(note!)
      onClose()
    }
  }
  
  const videos = useMemo(() => {
    if (!note?.Videos) return []
    try {
      return note.Videos.startsWith('[') ? JSON.parse(note.Videos) : []
    } catch (error) {
      console.error('Error parsing videos:', error)
      return []
    }
  }, [note?.Videos])

  // Extract YouTube video ID from various URL formats
  const extractVideoId = (url: string): string | null => {
    const patterns = [
      /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([^&\n?#]+)/,
      /youtube\.com\/v\/([^&\n?#]+)/,
      /youtube\.com\/watch\?.*v=([^&\n?#]+)/
    ]

    for (const pattern of patterns) {
      const match = url.match(pattern)
      if (match) return match[1]
    }
    return null
  }

  const handleAddVideo = (note: note.LocalNote, video: string) => {
    const videoId = extractVideoId(video)
    if (videoId) {
      // Use the current videos state instead of parsing from note
      if (!videos.includes(videoId)) {
        const newVideos = [...videos, videoId]
        onEdit(note, "videos", JSON.stringify(newVideos))
        toast.success("Video added successfully")
      } else {
        toast.error("This video is already in the list")
      }
    } else {
      toast.error("Please enter a valid YouTube URL")
    }
  }

  const handleRemoveVideo = (note: note.LocalNote, videoId: string) => {
    const newVideos = videos.filter((id: string) => id !== videoId)
    onEdit(note, "videos", JSON.stringify(newVideos))
    toast.success("Video removed successfully")
  }

  if (!note) return null

  // Parse keywords if they're stored as JSON string
  const keywords = note.Keywords ?
    (note.Keywords.startsWith('[') ? JSON.parse(note.Keywords) : note.Keywords.split(',')) :
    []


  return (
    <div>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto border-0 glass">
          <DialogHeader className="space-y-4">

            <div className="space-y-2">
              <DialogTitle className="text-2xl font-bold text-white note-dialog-title">
                {note.Title}
              </DialogTitle>
              <p className="text-gray-300 note-dialog-subtitle">{note.Subject}</p>
            </div>


            {/* Metadata */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              {/* Course Info */}
              {course && (
                <div className="flex items-center space-x-2">
                  <BookOpen className="w-4 h-4 text-gray-400" />
                  <Badge variant="outline" className="text-sm flex flex-row gap-2">
                    <div className={`h-2 w-2  rounded-full ${course?.Color}`} />
                    {course?.Code}
                  </Badge>
                </div>
              )}

              {/* Keywords */}
              {keywords.length > 0 && (
                <div className="flex items-center space-x-2">
                  <Tag className="w-4 h-4 text-gray-400" />
                  <div className="flex flex-wrap gap-1">
                    {keywords.map((keyword: string, index: number) => (
                      <Badge key={index} variant="secondary" className="text-xs">
                        {keyword}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              {/* Videos */}
              { videos.length > 0 && (
                <div className="flex items-center space-x-2">
                  <Video className="w-4 h-4 text-gray-400" />
                  <span className="text-gray-300">
                    {videos.length} video{videos.length !== 1 ? 's' : ''} available
                  </span>
                </div>
              )}

              {/* Timestamps */}
              <div className="flex items-center space-x-2">
                <Calendar className="w-4 h-4 text-gray-400" />
                <span className="text-gray-300">
                  Created: {new Date(note.CreatedAt).toLocaleDateString()}
                </span>
              </div>

              {note.UpdatedAt && note.UpdatedAt !== note.CreatedAt && (
                <div className="flex items-center space-x-2">
                  <Calendar className="w-4 h-4 text-gray-400" />
                  <span className="text-gray-300">
                    Updated: {new Date(note.UpdatedAt).toLocaleDateString()}
                  </span>
                </div>
              )}
            </div>

            <Separator className="bg-gray-700" />
          </DialogHeader>

          <Tabs defaultValue="note" value={activeView} onValueChange={setActiveView} className="w-full">
            <TabsList className="flex flex-row gap-2 mb-4">
              <TabsTrigger value="note" className="glass border-gray-600  flex items-center space-x-2 py-2 px-4 rounded-full">
                <Calendar className="w-4 h-4" />
                <span>Note</span>
              </TabsTrigger>
              <TabsTrigger value="videos" className="glass border-gray-600 flex items-center space-x-2 py-2 px-4 rounded-full">
                <List className="w-4 h-4" />
                <span>Videos</span>
              </TabsTrigger>
            </TabsList>

            {/* Content */}

            <TabsContent value="note">
              <div className="space-y-4">
                  <div className="space-y-2">
                  {/* Check if we have HTML content from the server */}
                  {note.Content ? (
                    <StyledMarkdownRenderer
                      content={note.Content}
                      className="bg-gray-800/50 rounded-lg p-4"
                    />
                  ) : (
                    <div className="text-gray-400 text-center py-8">
                      No content available
                    </div>
                  )}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="videos">
              <NoteVideo
                videos={videos}
                note={note}
                onRemoveVideo={handleRemoveVideo}
                setIsAddDialogOpen={setIsAddDialogOpen}
              />
            </TabsContent>

          </Tabs>

          {/* Actions */}
          <div className="flex justify-end space-x-2 pt-4 border-t border-gray-700">
           
            <Button
              variant="destructive"
              onClick={handleDelete}
              className="bg-red-600 hover:bg-red-700"
            >
              Delete Note
            </Button>
          </div>
        </DialogContent>
      </Dialog>
      {/* Add Videos Dialog */}
      <AddVideosDialog
        isOpen={isAddDialogOpen}
        onClose={() => setIsAddDialogOpen(false)}
        note={note}
        onAddVideo={handleAddVideo}
      />
    </div>
  )
} 