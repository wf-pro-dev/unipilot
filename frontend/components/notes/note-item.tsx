"use client"

import { Button } from "@/components/ui/button"
import { CardContent, CardFooter } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { CopyPlus, X } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useCourses } from "@/hooks/use-courses"
import { cn } from "@/lib/utils"
import { parseDeadline } from "@/lib/date-utils"
import { format } from "date-fns"
import { Avatar, AvatarImage } from "@radix-ui/react-avatar"
import { AvatarFallback } from "../ui/avatar"
import { Badge } from "../ui/badge"
import { NoteDetailModal } from "./note-detail-modal"
import { useState } from "react"

interface NoteItemProps<T extends models.LocalNote | models.Note> {
  note: T
  onNoteClick?: (noteID: number) => void
  onDelete?: (note: T) => void
  disabled?: boolean
  mode: "default" | "user"
  user?: models.User
  onCopy?: (note: models.Note) => void
}



export function NoteItem({
  note,
  mode = "default",
  onDelete,
  onNoteClick,
  disabled = false,
  onCopy,
  user,
}: NoteItemProps<models.LocalNote | models.Note>) {

  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false)

  const handleCardClick = () => {
    setIsDetailModalOpen(true)
  }

  const handleCloseDetailModal = () => {
    setIsDetailModalOpen(false)
  }


  function DefaultNoteItem({
    note,
    onDelete,
    disabled = false,
    mode = "default",
  }: NoteItemProps<models.LocalNote>) {
  
    if (!note || !onDelete) return null
  
    const handleDelete = (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation()
      onDelete(note)
    }
  
  
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
                    <h5 className="text-h5 font-medium line-clamp-1 tracking-tight ">{note.Title}</h5>
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
  
  function UserNoteItem({
    note,
    onDelete,
    disabled = false,
    onCopy,
    mode = "user",
    user,
  }: NoteItemProps<models.Note>) {
  
    if (!note || !onCopy || !user) return null
  
    const handleCopy = (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation()
      onCopy(note)
    }
  
  
    return (
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
                  <p className="text-caption text-text-caption flex items-center line-clamp-1 leading-relaxed">{note.Subject}</p>
                </div>
                <div className="flex items-center flex-shrink-0">
                  <Button onClick={handleCopy} variant={"outline"} size="icon" className="rounded-full w-7 h-7">
                    <CopyPlus className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
  
              <h5 className="text-h5 font-medium line-clamp-2 tracking-tight ">{note.Title}</h5>
  
              <div className="w-full h-px bg-gray-600">
  
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex-row-reverse p-4 pt-0 gap-2">
          <Badge variant="outline" className="gap-2">
            <span className="text-caption text-text-body">
              {user?.Username || user?.Email}
            </span>
            <Avatar className="h-5 w-5 rounded-full overflow-hidden border border-white/10">
              <AvatarImage src={user?.Avatar || "/placeholder-user.jpg"} />
              <AvatarFallback className="text-[10px]">
                {user?.Username?.split(" ").map((n: string) => n[0]).join("")}
              </AvatarFallback>
            </Avatar>
          </Badge>
        </CardFooter>
        <NoteDetailModal
          key={note.ID} // Force re-render when note changes
          note={note}
          isOpen={isDetailModalOpen}
          onClose={handleCloseDetailModal}
        />
      </GlassCard>
    )
  }

  switch (mode) {
    case "user":
      return <UserNoteItem note={note as models.Note} onDelete={onDelete} onNoteClick={onNoteClick} disabled={disabled} onCopy={onCopy} mode={mode} user={user} />
    default:
      return <DefaultNoteItem note={note as models.LocalNote} onDelete={onDelete} onNoteClick={onNoteClick} disabled={disabled} onCopy={onCopy} mode={mode} />
  }
}
