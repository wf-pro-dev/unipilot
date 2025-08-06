"use client"

import { useMemo } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Plus, Play, X, ExternalLink } from "lucide-react"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { note } from "@/wailsjs/go/models"


interface NoteVideoProps {
  videos: string[]
  note: note.LocalNote
  onRemoveVideo: (note: note.LocalNote, videoId: string) => void
  setIsAddDialogOpen: (isOpen: boolean) => void
}

export function NoteVideo({ videos, note, onRemoveVideo, setIsAddDialogOpen }: NoteVideoProps) {

  const getYouTubeEmbedUrl = (videoId: string) => {
    return `https://www.youtube.com/embed/${videoId}`
  }

  const getYouTubeWatchUrl = (videoId: string) => {
    return `https://www.youtube.com/watch?v=${videoId}`
  }

  if (videos.length === 0) {
    return (
      <div className="text-center py-8">
        <div className="mb-4">
          <Play className="w-12 h-12 text-gray-400 mx-auto mb-2" />
          <h3 className="text-lg font-medium text-gray-300 mb-2">No videos yet</h3>
          <p className="text-gray-400 text-sm mb-4">
            Add YouTube videos to enhance your notes
          </p>
        </div>
        <Button
          onClick={() => setIsAddDialogOpen(true)}
          className="bg-blue-600 hover:bg-blue-700"
        >
          <Plus className="w-4 h-4 mr-2" />
          Add Video
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header with add button */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Play className="w-5 h-5 text-gray-400" />
          <h3 className="text-lg font-medium text-white">
            Videos ({videos.length})
          </h3>
        </div>
        <Button
          onClick={() => setIsAddDialogOpen(true)}
          size="sm"
          className="bg-blue-600 hover:bg-blue-700"
        >
          <Plus className="w-4 h-4 mr-2" />
          Add Video
        </Button>
      </div>

      {/* Video grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {videos.map((videoId: string, index: number) => (
          <Card key={videoId} className="bg-gray-800/50 border-gray-700">
            <CardContent className="p-4">
              <div className="space-y-3">
                {/* Video thumbnail/embed */}
                <div className="relative aspect-video bg-gray-900 rounded-lg overflow-hidden">
                  <iframe
                    src={getYouTubeEmbedUrl(videoId)}
                    title={`Video ${index + 1}`}
                    className="w-full h-full"
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                    allowFullScreen
                  />
                </div>

                {/* Video actions */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <Badge variant="outline" className="text-xs">
                      Video {index + 1}
                    </Badge>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => BrowserOpenURL(getYouTubeWatchUrl(videoId))}
                      className="text-gray-400 hover:text-white p-1"
                    >
                      <ExternalLink className="w-3 h-3" />
                    </Button>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onRemoveVideo(note, videoId)}
                    className="text-red-400 hover:text-red-300 p-1"
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

       
    </div>
  )
}
