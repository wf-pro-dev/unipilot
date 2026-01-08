"use client"

import { Button } from "@/components/ui/button"
import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { X } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useCourses } from "@/hooks/use-courses"
import { cn } from "@/lib/utils"
import { parseDeadline } from "@/lib/date-utils"
import { format } from "date-fns"
import { toast } from "sonner"

interface NoteItemProps {
  note: models.LocalNote
  onEdit: (note: models.LocalNote, column: string, value: string) => void
  onNoteClick?: (noteID: number) => void
  onDelete: (note: models.LocalNote) => void
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
  const handleDelete = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    onDelete(note)
  }


  // Parse keywords if they're stored as JSON string
  // Parse videos if they're stored as JSON string
  const videos = note.Videos ?
    (note.Videos.startsWith('[') ? JSON.parse(note.Videos) : []) :
    []




  return (
    <div>
      <GlassCard
        variant="outline"
        className={`${disabled ? 'opacity-50' : ''}`}
        onClick={handleCardClick}
      >
        <CardContent className="p-5">
          <div className="flex items-start space-x-4">
            <div className="flex-1 space-y-3">
              <div className="flex gap-3 justify-between items-start">
                <div className="space-y-1 flex-1 min-w-0">
                  <h5 className="text-h5 line-clamp-1 tracking-tight ">{note.Title}</h5>
                  <p className="text-caption flex items-center gap-1 line-clamp-1 leading-relaxed">{note.Subject}</p>
                </div>
                <div className="flex items-center flex-shrink-0">
                  <Button onClick={handleDelete} variant={"outline"} size="icon" className="rounded-full w-7 h-7">
                    <X className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>

              <div className="space-y-3">


                <p className="text-caption flex items-center gap-2">
                  <div className={cn("w-2 h-2 mx-2.5 rounded-full shrink-0", note.Course?.Color)} />
                  <span className="text-caption">{note.Course?.Code}</span>
                </p>


                {/* Timestamp */}
                <div className="flex justify-between items-center text-[10px] uppercase tracking-wider text-gray-500 border-t border-white/5 pt-3 mt-3 font-medium">
                  <span>
                    Created {format(parseDeadline(note.CreatedAt), "EEEE MMMM d, yyyy")}
                  </span>

                </div>

              </div>
            </div>
          </div>
        </CardContent>
      </GlassCard>
    </div>
  )
}
