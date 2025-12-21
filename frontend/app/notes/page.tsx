"use client"
import { AddNoteDialog } from "@/components/notes/note-add-dialog"
import { NoteView } from "@/components/notes/note-view"
import { NoteDetailModal } from "@/components/notes/note-detail-modal"
import { useNotes, useDeleteNote, useUpdateNote, useCreateNote } from "@/hooks/use-notes"
import { note } from "@/wailsjs/go/models"
import { useState } from "react"
import { toast } from "sonner"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { useSearchParams, usePathname } from "next/navigation"

export default function NotesPage() {
  const searchParams = useSearchParams()
  const pathname = usePathname()

  const { data: notes, isLoading } = useNotes()

  const createNote = useCreateNote()
  const deleteNote = useDeleteNote()
  const updateNote = useUpdateNote()
  const [selectedNoteID, setSelectedNoteID] = useState<number | null>(null)
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false)

  const courseFilter = searchParams.get("course") || "all"

  console.log("route", pathname)
  console.log("filter (page)", { courseFilter })

  const handleAddNote = async (note: note.LocalNote) => {
    const message = "note " + note.Title + " added"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    createNote.mutate(note, {
      onSuccess: () => {
        toast.success("Note added successfully")
      },
      onError: () => {
        toast.error("Note addition failed")
      }
    })
  }

  const handleNoteClick = (noteID: number) => {
    setSelectedNoteID(noteID)
    setIsDetailModalOpen(true)
  }

  const handleDeleteNote = async (note: note.LocalNote) => {
    const message = "note " + note.Title + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteNote.mutate(note, {
      onSuccess: () => {
        toast.success("Note deleted successfully")
      },
      onError: () => {
        toast.error("Note deletion failed")
      }
    })
  }

  const handleEditNote = async (note: note.LocalNote, column: string, value: string) => {
    const message = "note " + note.Title + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    updateNote.mutate({ note, column, value }, {
      onError: () => {
        toast.error(`Note ${column} update failed`)
      }
    })
  }

  const handleCloseDetailModal = () => {
    setIsDetailModalOpen(false)
    setSelectedNoteID(null)
  }

  return (
    <div className="">

      <div className="relative z-10">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-h1 text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
              Study Notes
            </h1>
            <p className="mt-3 text-body-small text-gray-400">Generate AI-powered study notes for your courses</p>
          </div>
          <AddNoteDialog onAdd={handleAddNote} />
        </div>



        <NoteView
          title="All Notes"
          notes={notes || []}
          onNoteClick={handleNoteClick}
          onDelete={handleDeleteNote}
          onEdit={handleEditNote}
          isLoading={isLoading}
          filter={{ course: courseFilter }}
        />


        {/* Note Detail Modal */}
        <NoteDetailModal
          key={selectedNoteID} // Force re-render when note changes
          noteID={selectedNoteID}
          isOpen={isDetailModalOpen}
          onClose={handleCloseDetailModal}
          onEdit={handleEditNote}
          onDelete={handleDeleteNote}
        />
      </div>
    </div>
  )
}
