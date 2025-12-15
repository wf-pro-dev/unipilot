"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { MoreVertical, Edit, Trash2, BookOpen, Tag, Video, FileText } from "lucide-react"
import { note } from "@/wailsjs/go/models"
import { useCourses } from "@/hooks/use-courses"
import { Skeleton } from "../ui/skeleton"

interface NoteItemProps {
  note: note.LocalNote
  onEdit: (note: note.LocalNote, column: string, value: string) => void
  onNoteClick?: (noteID: number) => void
  onDelete: (note: note.LocalNote) => void
  disabled?: boolean
}

export function NoteItem({
  note,
  onEdit,
  onDelete,
  onNoteClick,
  disabled = false
}: NoteItemProps) {

  const { data: courses } = useCourses()
  const course = courses?.find(c => c.Code === note.course_code)

  const handleCardClick = () => {
    if (onNoteClick && !disabled) {
      onNoteClick(note.ID)
    }
  }

  const handleEditOpen = (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation()
    // TODO: Implement edit dialog similar to AssignmentEditDialog
    console.log("Edit note:", note)
  }

  // Parse keywords if they're stored as JSON string
  // Parse videos if they're stored as JSON string
  const videos = note.videos ?
    (note.videos.startsWith('[') ? JSON.parse(note.videos) : []) :
    []


    return (
    <div>
      <GlassCard
        variant={!disabled && onNoteClick ? "interactive" : "default"}
        className={`border-white/5 bg-white/5 hover:bg-white/10 transition-all duration-300 ${disabled ? 'opacity-50' : ''}`}
        onClick={handleCardClick}
      >
        <CardContent className="p-5">
          <div className="flex items-start space-x-4">
            <div className="flex-1 space-y-3">
              <div className="flex gap-3 justify-between items-start">
                <div className="space-y-1 flex-1 min-w-0">
                  <h3 className="text-base font-semibold text-white line-clamp-1 tracking-tight">{note.title}</h3>
                  <p className="text-xs text-gray-400 line-clamp-1 font-medium">{note.subject}</p>
                </div>
                <div className="flex items-center flex-shrink-0">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="p-0 w-8 h-8 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                        disabled={disabled}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="border-white/10 bg-black/90 backdrop-blur-xl glass">
                      <DropdownMenuItem
                        onClick={handleEditOpen}
                        disabled={disabled}
                        className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
                      >
                        <Edit className="mr-2 w-4 h-4" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={(e) => {
                          e.stopPropagation()
                          onDelete(note)
                        }}
                        disabled={disabled}
                        className="text-red-400 focus:text-red-300 focus:bg-red-500/10 cursor-pointer"
                      >
                        <Trash2 className="mr-2 w-4 h-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>

              <div className="space-y-3">


                <div className="flex items-center space-x-2">
                  <div className="p-1 bg-blue-500/10 rounded-md border border-blue-500/10">
                    <BookOpen className="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
                  </div>
                  <Badge
                    variant="outline"
                    className="text-[10px] border-white/10 bg-white/5 px-2 py-0.5 font-medium text-gray-300"
                    style={{
                      borderColor: course?.Color ? undefined : 'rgba(255,255,255,0.1)',
                    }}
                  >
                    <span className="mr-1" style={{ color: course?.Color || '#9ca3af' }}>●</span>
                    {course?.Code || 'No Course'} 
                  </Badge>
                </div>



                <div className="flex items-center space-x-2">
                  <div className="p-1 bg-purple-500/10 rounded-md border border-purple-500/10">
                    <Video className="w-3.5 h-3.5 text-purple-400 flex-shrink-0" />
                  </div>
                  <span className="text-xs text-gray-400 font-medium">
                    {videos.length < 1 ? 'No videos' : videos.length + ' video' + (videos.length !== 1 ? 's' : '')} available
                  </span>
                </div>


                {/* Timestamp */}
                <div className="flex justify-between items-center text-[10px] uppercase tracking-wider text-gray-500 pt-3 border-t border-white/5 mt-3 font-medium">
                  <span>
                    Created {new Date(note.CreatedAt).toLocaleDateString()}
                  </span>
                  {note.UpdatedAt && note.UpdatedAt !== note.CreatedAt && (
                    <span>
                      Updated {new Date(note.UpdatedAt).toLocaleDateString()}
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </GlassCard>
    </div>
  )
}
