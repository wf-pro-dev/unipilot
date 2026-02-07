import { models } from "@/wailsjs/go/models"
// YouTube video interface
interface YouTubeVideo {
  ID: string // YouTube video ID
  Title: string // Video title
  ThumbnailURL?: string
  Duration?: string
}

interface PageResponse<T> {
  Data: T[]
  Cursor?: models.Cursor
  HasMore: boolean
}


export type {  
  PageResponse,
  YouTubeVideo
}