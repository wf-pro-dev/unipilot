"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
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
  const course = courses?.find(c => c.Code === note.CourseCode)

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
  const keywords = note.Keywords ?
    (note.Keywords.startsWith('[') ? JSON.parse(note.Keywords) : note.Keywords.split(',')) :
    []

  // Parse videos if they're stored as JSON string
  const videos = note.Videos ?
    (note.Videos.startsWith('[') ? JSON.parse(note.Videos) : []) :
    []

  // Truncate content for preview
  const contentPreview = note.Content ?
    note.Content.length > 150 ?
      note.Content.substring(0, 150) + "..." :
      note.Content :
    "No content"

  return (
    <div>
      <Card
        className={`glass border-0 hover:bg-white/5 transition-colors ${!disabled && onNoteClick ? 'cursor-pointer' : ''
          } ${disabled ? 'opacity-50' : ''}`}
        onClick={handleCardClick}
      >
        <CardContent className="p-4">
          <div className="flex items-start space-x-3">
            <div className="flex-1 space-y-3">
              <div className="flex gap-2 justify-between items-start">
                <div className="space-y-1">
                  <h3 className="font-medium text-white line-clamp-1">{note.Title}</h3>
                  <p className="text-sm text-gray-400 line-clamp-1">{note.Subject}</p>
                </div>
                <div className="flex items-center">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="p-0 w-8 h-8 text-gray-400 hover:text-white"
                        disabled={disabled}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="border-gray-700 glass">
                      <DropdownMenuItem
                        onClick={handleEditOpen}
                        disabled={disabled}
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
                        className="text-red-400"
                      >
                        <Trash2 className="mr-2 w-4 h-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>

              <div className="space-y-2">


                <div className="flex items-center space-x-2">
                  <BookOpen className="w-3 h-3 text-gray-400" />
                  <Badge
                    variant="outline"
                    className="text-xs"
                    style={{
                      borderColor: course?.Color || '#6b7280',
                      color: course?.Color || '#6b7280'
                    }}
                  >
                    {course?.Code} - {course?.Name}
                  </Badge>
                </div>


                {/* Keywords */}

                <div className="flex items-center space-x-2">
                  <Tag className="w-3 h-3 text-gray-400" />
                  {keywords.length > 0 ? (
                    <div className="flex flex-wrap gap-1">
                      {keywords.slice(0, 3).map((keyword: string, index: number) => (
                        <Badge key={index} variant="secondary" className="text-xs">
                          {keyword}
                        </Badge>
                      ))}
                      {keywords.length > 3 && (
                        <Badge variant="secondary" className="text-xs">
                          +{keywords.length - 3} more
                        </Badge>
                      )}
                    </div>
                  ) :
                    <Skeleton className="h-4 w-3/4" />
                  }

                </div>



                <div className="flex items-center space-x-2">
                  <Video className="w-3 h-3 text-gray-400" />
                  <span className="text-xs text-gray-400">
                    {videos.length < 1 ? 'No videos' : videos.length + ' video' + (videos.length !== 1 ? 's' : '') + ' '} available
                  </span>
                </div>


                {/* Timestamp */}
                <div className="flex justify-between items-center text-xs text-gray-400">
                  <span>
                    Created: {new Date(note.CreatedAt).toLocaleDateString()}
                  </span>
                  {note.UpdatedAt && note.UpdatedAt !== note.CreatedAt && (
                    <span>
                      Updated: {new Date(note.UpdatedAt).toLocaleDateString()}
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
