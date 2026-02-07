"use client"

import { Button } from "@/components/ui/button"
import { CardContent, CardFooter } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { CopyPlus, X } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { cn } from "@/lib/utils"
import { parseDeadline } from "@/lib/date-utils"
import { format } from "date-fns"
import { Avatar, AvatarImage } from "@radix-ui/react-avatar"
import { AvatarFallback } from "../ui/avatar"
import { Badge } from "../ui/badge"
import { useNote, useDeleteNote } from "@/hooks/use-notes"
import { useDialogContext } from "../provider/dialog-provider"

interface NoteItemProps{
  noteID: string
  noteRO?: models.LocalNote | models.Note
  disabled?: boolean
  mode?: "default" | "user"
  user?: models.User
  onCopy?: (note: models.Note) => void
}



export function NoteItem({
  noteID,
  noteRO,
  mode = "default",
  disabled = false,
  onCopy,
  user,
}: NoteItemProps) {

  const { SetDialogState } = useDialogContext()


  function DefaultNoteItem({
    noteID,
    disabled = false,
  }: NoteItemProps) {
  
    const { data: note } = useNote(noteID)
    const { mutate: deleteNote } = useDeleteNote()


    if (!note) return null
  
    const handleDelete = (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation()
      deleteNote(note as models.LocalNote)
    }
  
  
    return (
      <div>
        <GlassCard
          variant="outline"
          className={`${disabled ? 'opacity-50' : ''}`}
          onClick={()=>{SetDialogState({ modelType: "note", dialogType: "details", id: noteID })}}
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
    noteRO,
    disabled = false,
    onCopy,
    mode = "user",
    user,
  }: NoteItemProps) {
  
    if (!noteRO || !onCopy || !user) return null
  
    const handleCopy = (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation()
      onCopy(noteRO as models.Note)
    }
  
  
    return (
      <GlassCard
        variant="outline"
        className={`${disabled ? 'opacity-50' : ''}`}
        onClick={()=>{SetDialogState({ modelType: "note", dialogType: "details", id: noteRO.ID })}}
      >
        <CardContent className="p-5">
          <div className="flex items-start space-x-4">
            <div className="flex-1 space-y-3">
              <div className="flex gap-3 justify-between items-start">
                <div className="space-y-1 flex-1 min-w-0">
                  <p className="text-caption text-text-caption flex items-center line-clamp-1 leading-relaxed">{noteRO.Subject}</p>
                </div>
                <div className="flex items-center flex-shrink-0">
                  <Button onClick={handleCopy} variant={"outline"} size="icon" className="rounded-full w-7 h-7">
                    <CopyPlus className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
  
              <h5 className="text-h5 font-medium line-clamp-2 tracking-tight ">{noteRO.Title}</h5>
  
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
       
      </GlassCard>
    )
  }

  switch (mode) {
    case "user":
      return <UserNoteItem noteID={noteID} noteRO={noteRO} disabled={disabled} onCopy={onCopy} mode={mode} user={user} />
    default:
      return <DefaultNoteItem noteID={noteID} disabled={disabled}  />
  }
}
