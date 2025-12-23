"use client"

import { GlassCard } from "@/components/ui/glass-card"
import { Button } from "@/components/ui/button"
import { CopyPlus } from "lucide-react"
import { note } from "@/wailsjs/go/models"
import { toast } from "sonner"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

interface NoteItemCompactProps {
  note: note.LocalNote
  onCopy?: (note: note.LocalNote) => void
  disabled?: boolean
  className?: string
}

export function NoteItemCompact({
  note,
  onCopy,
  disabled = false,
  className
}: NoteItemCompactProps) {

  // Mock user data
  const user = {
    username: "Student",
    avatar: "/placeholder-user.jpg"
  }

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (onCopy) {
      onCopy(note)
    } else {
        toast.success("Note copied to your notebook")
    }
  }

  // Use course color if available, or fallback
  const accentColor = note.Course?.Color || "bg-gray-500"

  return (
    <div className={className}>
      <GlassCard
        className={`p-4 border-white/5 bg-white/5 hover:bg-white/10 transition-all ${disabled ? 'opacity-50' : ''}`}
      >
        <div className="flex items-start gap-3">
          {/* Course Accent Dot */}
          <div className={`w-2 h-2 rounded-full mt-2 shrink-0 ${accentColor.replace('bg-', 'bg-')}`} />

          <div className="flex-1 min-w-0 space-y-4">
            <div className="space-y-2">
              <h3 className="text-sm font-medium text-gray-200 truncate leading-tight">
                {note.title}
              </h3>
              <div className="flex items-center gap-2">
                <p className="text-xs text-gray-400 font-medium truncate">
                  {note.course_code || note.Course?.Code}
                </p>
                <span className="text-gray-600 text-[10px]">•</span>
                <p className="text-xs text-gray-500 truncate">
                  {note.subject}
                </p>
              </div>
            </div>

            <div className="flex items-center justify-between py-4 border-t border-white/20">
              {/* User Info */}
              <div className="flex items-center gap-2">
                <Avatar className="h-5 w-5 border border-white/10">
                  <AvatarImage src={note.User?.Avatar || "/placeholder-user.jpg"} />
                  <AvatarFallback className="text-[10px]">ST</AvatarFallback>
                </Avatar>
                <span className="text-xs text-gray-400 truncate max-w-[150px]">
                  {note.User?.Username || note.User?.Email}
                </span>
              </div>
            </div>
          </div>

          <Button
            size="icon"
            variant="ghost"
            className="h-8 w-8 -mt-1 -mr-1 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg shrink-0"
            onClick={handleCopy}
            disabled={disabled}
            title="Create personal copy"
          >
            <CopyPlus className="h-4 w-4" />
          </Button>
        </div>
      </GlassCard>
    </div>
  )
}

