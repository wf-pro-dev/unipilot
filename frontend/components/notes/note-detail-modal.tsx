"use client"

import { Button } from "@/components/ui/button"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { Video, FileText} from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { NoteVideo } from "./note-video"
import { useMemo, useState, useRef } from "react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCourses } from "@/hooks/use-courses"
import { useNotes } from "@/hooks/use-notes"
import { AddVideosDialog } from "./add-videos-dialog"
import { toast } from "sonner"
import { StyledMarkdownRenderer } from "./markdown-renderer"
import { motion, useScroll, useTransform } from "framer-motion"

interface NoteDetailModalProps {
  noteID: number | null
  isOpen: boolean
  onClose: () => void
  onEdit: (note: models.LocalNote, column: string, value: string) => void
  onDelete: (note: models.LocalNote) => void
}

export function NoteDetailModal({
  noteID,
  isOpen,
  onClose,
  onEdit,
  onDelete
}: NoteDetailModalProps) {

  const [activeView, setActiveView] = useState("note")
  const { data: courses } = useCourses()
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false)
  
  const scrollRef = useRef<HTMLDivElement>(null)
  // Using useScroll for cleaner Framer Motion integration
  // We need to ensure the scrollRef is attached to the scrolling container
  const { scrollY } = useScroll({ 
    container: scrollRef
  })

  // Header Animations
  const headerPadding = useTransform(scrollY, [0, 100], ["24px", "16px"])
  const headerBackground = useTransform(
    scrollY, 
    [0, 50], 
    ["rgba(0,0,0,0)", "rgba(0,0,0,0.2)"]
  )
  const headerBackdrop = useTransform(scrollY, [0, 50], ["blur(0px)", "blur(12px)"])
  const headerBorderOpacity = useTransform(scrollY, [0, 50], [0, 0.1])
  
  const metadataOpacity = useTransform(scrollY, [0, 100], [1, 0])
  const metadataHeight = useTransform(scrollY, [0, 100], ["auto", "0px"])
  const metadataMargin = useTransform(scrollY, [0, 100], ["16px", "0px"])
  
  const titleSize = useTransform(scrollY, [0, 150], ["1.5rem", "1.25rem"]) // 2xl to xl

  // Tabs Animations
  // Padding around the tabs list: starts at 24px (p-6), goes to 0px (full width)
  const tabsContainerPadding = useTransform(scrollY, [0, 100], ["24px", "0px"])
  // Radius of the tabs list itself: starts at 12px (rounded-xl), goes to 0px
  const tabsRadius = useTransform(scrollY, [0, 100], [12, 0])
  const tabsBackdrop = useTransform(scrollY, [0, 50], ["blur(0px)", "blur(8px)"])
  const tabsBackground = useTransform(scrollY, [0, 50], ["rgba(0,0,0,0.05)", "rgba(0,0,0,0.4)"])
  
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

  const handleAddVideo = (note: models.LocalNote, video: string) => {
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

  const handleRemoveVideo = (note: models.LocalNote, videoId: string) => {
    const newVideos = videos.filter((id: string) => id !== videoId)
    onEdit(note, "videos", JSON.stringify(newVideos))
    toast.success("Video removed successfully")
  }

  if (!note) return null

  // Parse keywords if they're stored as JSON string
  return (
    <div>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col glass border-white/10 p-0 overflow-hidden gap-0">
          
          {/* Animated Header */}
          <motion.div 
            className="border-b border-white/5 z-30 flex-shrink-0 relative transition-all duration-300 ease-in-out"
            style={{ 
                padding: headerPadding,
                backgroundColor: headerBackground,
                backdropFilter: headerBackdrop,
                borderColor: `rgba(255,255,255,${headerBorderOpacity})`
            }}
          >
            <div className="space-y-4">
               {/* Top Bar with Subject & Actions */}
               <motion.div 
                 className="flex items-start justify-between"
                 style={{ 
                    opacity: metadataOpacity, 
                    height: useTransform(scrollY, [0, 20], ["auto", "0px"]) 
                 }}
               >
                 <div className="flex items-center gap-2 px-2 py-1 rounded-full bg-white/5 border border-white/5 w-fit">
                    {course && <div className={`w-1.5 h-1.5 rounded-full ${course?.Color} shadow-[0_0_8px] shadow-${course?.Color}/80 ml-0.5`} />}
                    <span className="text-[10px] font-medium pr-1 opacity-80 uppercase tracking-wider text-gray-300">{note.Subject || "General"}</span>
                 </div>
                 
                 <div className="flex items-center space-x-2">
                    <div className="text-xs text-gray-500 font-medium px-2">
                        {new Date(note.CreatedAt).toLocaleDateString()}
                    </div>
                 </div>
               </motion.div>
              
              <motion.h2 
                className="font-bold text-white leading-tight tracking-tight"
                style={{ fontSize: titleSize }}
              >
                {note.Title}
              </motion.h2>

              {/* Course & Date Info - Collapsible */}
              <motion.div 
                  style={{ 
                      opacity: metadataOpacity,
                      height: metadataHeight,
                      marginTop: metadataMargin
                  }}
                  className="overflow-hidden"
              >
                  <div className="flex items-center gap-4 text-xs text-gray-400">
                    {course && (
                        <div className="flex items-center gap-1.5">
                            <span className="font-medium text-gray-300">{course.Code}</span>
                            <span className="w-1 h-1 rounded-full bg-gray-600" />
                            <span>{course.Name}</span>
                        </div>
                    )}
                    {note.UpdatedAt && note.UpdatedAt !== note.CreatedAt && (
                        <div className="flex items-center gap-1.5 ml-auto">
                            <span>Updated {new Date(note.UpdatedAt).toLocaleDateString()}</span>
                        </div>
                    )}
                  </div>
              </motion.div>
            </div>
          </motion.div>

          <div 
            className="flex-1 overflow-y-auto scroll-smooth"
            ref={scrollRef}
            onScroll={(e) => scrollY.set(e.currentTarget.scrollTop)}
          >
             <Tabs defaultValue="note" value={activeView} onValueChange={setActiveView} className="w-full flex flex-col min-h-full">
                
                {/* Sticky Tabs List Container */}
                <motion.div 
                    className="sticky top-0 z-20" 
                    style={{ 
                        padding: tabsContainerPadding,
                        backdropFilter: tabsBackdrop,
                        backgroundColor: tabsBackground
                    }}
                >
                    <motion.div style={{ borderRadius: tabsRadius }} className="overflow-hidden">
                        <TabsList className="flex flex-row bg-white/5 p-1 w-full border border-white/5 h-auto">
                            <TabsTrigger 
                              value="note" 
                              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
                            >
                              <FileText className="w-4 h-4" />
                              <span className="text-sm font-medium">Note Content</span>
                            </TabsTrigger>
                            <TabsTrigger 
                              value="videos" 
                              className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
                            >
                              <Video className="w-4 h-4" />
                              <span className="text-sm font-medium">Videos ({videos.length})</span>
                            </TabsTrigger>
                        </TabsList>
                    </motion.div>
                </motion.div>

                {/* Content */}
                <div className="p-6 pt-4 flex-1">
                  <TabsContent value="note" className="animate-in fade-in slide-in-from-bottom-4 duration-300 focus-visible:ring-0 focus-visible:outline-none mt-0">
                    <div className="space-y-4">
                      {/* Check if we have HTML content from the server */}
                      {note.Content ? (
                        <StyledMarkdownRenderer
                          content={note.Content}
                          className="bg-white/5 rounded-xl p-6 border border-white/5 min-h-[300px]"
                        />
                      ) : (
                        <div className="text-gray-400 text-center py-12 border border-dashed border-white/10 rounded-xl bg-white/5">
                          <FileText className="h-8 w-8 mx-auto mb-3 opacity-50" />
                          <p>No content available</p>
                        </div>
                      )}
                    </div>
                  </TabsContent>

                  <TabsContent value="videos" className="animate-in fade-in slide-in-from-bottom-4 duration-300 focus-visible:ring-0 focus-visible:outline-none mt-0">
                    <NoteVideo
                      videos={videos}
                      note={note}
                      onRemoveVideo={handleRemoveVideo}
                      setIsAddDialogOpen={setIsAddDialogOpen}
                    />
                  </TabsContent>

                  {/* Actions */}
                  <div className="flex justify-end space-x-3 pt-6 mt-6 border-t border-white/5">
                    <Button
                      variant="destructive"
                      onClick={handleDelete}
                      className="bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 shadow-none"
                    >
                      Delete Note
                    </Button>
                  </div>
                </div>
             </Tabs>
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
