"use client"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Plus, Play, X, AlertCircle } from "lucide-react"
import { useToast } from "@/hooks/use-toast"
import { YouTubeVideo } from "@/types/models"

interface YouTubeVideoInputProps {
  videos: YouTubeVideo[]
  onVideosChange: (videos: YouTubeVideo[]) => void
  disabled?: boolean
}

export function YouTubeVideoInput({ videos, onVideosChange, disabled }: YouTubeVideoInputProps) {
  const [videoUrl, setVideoUrl] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const { toast } = useToast()

  // Extract YouTube video ID from various URL formats
  const extractVideoId = (url: string): string | null => {
    const patterns = [
      /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([^&\n?#]+)/,
      /youtube\.com\/v\/([^&\n?#]+)/,
      /youtube\.com\/watch\?.*v=([^&\n?#]+)/
    ]

    for (const pattern of patterns) {
      const match = url.match(pattern)
      if (match) {
        return match[1]
      }
    }

    return null
  }

  // Mock function to get video info from YouTube API
  const getVideoInfo = async (videoId: string): Promise<{ title: string; thumbnail?: string; duration?: string }> => {
    // TODO: Replace with actual YouTube Data API v3 call
    // For now, return mock data
    await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API call
    
    return {
      title: `Video ${videoId}`,
      thumbnail: `https://img.youtube.com/vi/${videoId}/default.jpg`,
      duration: "10:30"
    }
  }

  const handleAddVideo = async () => {
    if (!videoUrl.trim()) {
      toast({
        title: "Invalid URL",
        description: "Please enter a valid YouTube URL.",
        variant: "destructive"
      })
      return
    }

    const videoId = extractVideoId(videoUrl)
    if (!videoId) {
      toast({
        title: "Invalid YouTube URL",
        description: "Please enter a valid YouTube video URL.",
        variant: "destructive"
      })
      return
    }

    // Check if video already exists
    if (videos.some(video => video.ID === videoId)) {
      toast({
        title: "Video Already Added",
        description: "This video has already been added to the note.",
        variant: "destructive"
      })
      return
    }

    setIsLoading(true)

    try {
      const videoInfo = await getVideoInfo(videoId)
      
      const newVideo: YouTubeVideo = {
        ID: videoId,
        Title: videoInfo.title,
        ThumbnailURL: videoInfo.thumbnail,
        Duration: videoInfo.duration
      }

      onVideosChange([...videos, newVideo])
      setVideoUrl("")

      toast({
        title: "Video Added",
        description: "YouTube video has been added successfully.",
      })
    } catch (error) {
      toast({
        title: "Failed to Add Video",
        description: "Could not retrieve video information. Please try again.",
        variant: "destructive"
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleRemoveVideo = (videoId: string) => {
    onVideosChange(videos.filter(video => video.ID !== videoId))
    toast({
      title: "Video Removed",
      description: "Video has been removed from the note.",
    })
  }

  return (
    <Card className="border-0 glass">
      <CardHeader>
        <CardTitle className="flex items-center space-x-2 text-white">
          <Play className="w-5 h-5 text-red-400" />
          <span>YouTube Videos</span>
          {videos.length > 0 && (
            <Badge variant="secondary" className="text-red-300 bg-red-500/20">
              {videos.length}
            </Badge>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Add Video Input */}
        <div className="flex space-x-2">
          <div className="flex-1">
            <Label className="text-sm font-medium text-gray-300">Video URL</Label>
            <Input
              value={videoUrl}
              onChange={(e) => setVideoUrl(e.target.value)}
              placeholder="https://www.youtube.com/watch?v=..."
              className="mt-1 placeholder-gray-400 text-white border-gray-600 bg-gray-800/50"
              disabled={disabled || isLoading}
            />
          </div>
          <Button
            onClick={handleAddVideo}
            disabled={!videoUrl.trim() || disabled || isLoading}
            className="mt-6 bg-red-600 hover:bg-red-700"
          >
            {isLoading ? (
              <div className="w-4 h-4 rounded-full border-b-2 border-white animate-spin"></div>
            ) : (
              <>
                <Plus className="mr-2 w-4 h-4" />
                Add
              </>
            )}
          </Button>
        </div>

        {/* Video List */}
        {videos.length > 0 && (
          <div className="space-y-2">
            <Label className="text-sm font-medium text-gray-300">Added Videos</Label>
            <div className="space-y-2">
              {videos.map((video) => (
                <div
                  key={video.ID}
                  className="flex items-center p-3 space-x-3 rounded-lg bg-gray-800/30"
                >
                  {video.ThumbnailURL && (
                    <img
                      src={video.ThumbnailURL}
                      alt={video.Title}
                      className="object-cover w-16 h-12 rounded"
                    />
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-white truncate">
                      {video.Title}
                    </p>
                    <p className="text-xs text-gray-400">
                      ID: {video.ID}
                      {video.Duration && ` • ${video.Duration}`}
                    </p>
                  </div>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => handleRemoveVideo(video.ID)}
                    className="text-red-400 hover:text-red-300"
                    disabled={disabled}
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Help Text */}
        <div className="flex items-start p-3 space-x-2 rounded-lg bg-blue-500/10">
          <AlertCircle className="h-4 w-4 text-blue-400 mt-0.5 flex-shrink-0" />
          <div className="text-xs text-blue-300">
            <p className="mb-1 font-medium">Supported YouTube URL formats:</p>
            <ul className="space-y-1 text-blue-200">
              <li>• https://www.youtube.com/watch?v=VIDEO_ID</li>
              <li>• https://youtu.be/VIDEO_ID</li>
              <li>• https://www.youtube.com/embed/VIDEO_ID</li>
            </ul>
          </div>
        </div>
      </CardContent>
    </Card>
  )
} 