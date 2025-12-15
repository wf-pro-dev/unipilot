"use client"

import { useMemo, useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Plus, Play, X, ExternalLink, AlertCircle } from "lucide-react"
import { BrowserOpenURL } from "@/wailsjs/runtime/runtime"
import { note } from "@/wailsjs/go/models"


interface NoteVideoProps {
  videos: string[]
  note: note.LocalNote
  onRemoveVideo: (note: note.LocalNote, videoId: string) => void
  setIsAddDialogOpen: (isOpen: boolean) => void
}

export function NoteVideo({ videos, note, onRemoveVideo, setIsAddDialogOpen }: NoteVideoProps) {
  const [embedErrors, setEmbedErrors] = useState<Set<string>>(new Set())

  const getYouTubeEmbedUrl = (videoId: string) => {
    // Add enablejsapi=1 to allow postMessage communication for error detection
    return `https://www.youtube.com/embed/${videoId}?enablejsapi=1&origin=${window.location.origin}`
  }

  const getYouTubeWatchUrl = (videoId: string) => {
    return `https://www.youtube.com/watch?v=${videoId}`
  }

  // Listen for YouTube iframe API error messages
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      // YouTube sends error messages via postMessage
      if (event.origin !== 'https://www.youtube.com') return

      if (event.data && typeof event.data === 'string') {
        try {
          const data = JSON.parse(event.data)
          // YouTube error 153 is reported in various ways
          if (data.error === 153 || data.errorCode === 153 ||
            (data.info && data.info.includes && data.info.includes('153'))) {
            // Extract video ID from iframe source if possible
            // This is a fallback - we'll also check via iframe load timeout
          }
        } catch {
          // Not JSON, ignore
        }
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [])

  const handleVideoError = (videoId: string) => {
    setEmbedErrors(prev => new Set(prev).add(videoId))
  }

  // Manual error reporting - users can click to report embedding issues
  const handleReportError = (videoId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    handleVideoError(videoId)
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
                <div className="relative aspect-video bg-gray-900 rounded-lg overflow-hidden group">
                  {embedErrors.has(videoId) ? (
                    <div className="w-full h-full flex flex-col items-center justify-center p-4 text-center">
                      <AlertCircle className="w-8 h-8 text-yellow-400 mb-2" />
                      <p className="text-sm text-gray-300 mb-2">
                        This video cannot be embedded
                      </p>
                      <p className="text-xs text-gray-400 mb-4">
                        Error 153: Video player configuration error. The video may have embedding disabled or domain restrictions.
                      </p>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => BrowserOpenURL(getYouTubeWatchUrl(videoId))}
                        className="text-blue-400 border-blue-400/30 hover:bg-blue-400/10"
                      >
                        <ExternalLink className="w-3 h-3 mr-2" />
                        Watch on YouTube
                      </Button>
                    </div>
                  ) : (
                    <>
                      {/* <iframe
                        src={getYouTubeEmbedUrl(videoId)}
                        title={`Video ${index + 1}`}
                        className="w-full h-full"
                        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                        allowFullScreen
                      /> */}
                      <iframe
                        width="100%"
                        height="100%"
                        src="https://www.youtube-nocookie.com/embed/Ilk7UXzV_Qc?si=6cOLhgdzVosGwCbw"
                        title="YouTube video player"
                        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                        referrerpolicy="strict-origin-when-cross-origin"
                        allowfullscreen
                      />
                      {/* Manual error reporting button - appears on hover */}
                      <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => handleReportError(videoId, e)}
                          className="bg-black/50 hover:bg-black/70 text-yellow-400 text-xs h-6 px-2"
                          title="Report embedding error (Error 153)"
                        >
                          <AlertCircle className="w-3 h-3 mr-1" />
                          Report Error
                        </Button>
                      </div>
                    </>
                  )}
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
