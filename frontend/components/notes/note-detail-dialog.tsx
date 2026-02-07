"use client"

import { Button } from "@/components/ui/button"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { FileText, Trash2 } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useRef } from "react"
import { useDeleteNote, useNote } from "@/hooks/use-notes"

import { StyledMarkdownRenderer } from "../markdown/markdown-renderer"
import { MarkdownTableOfContents } from "../markdown/markdown-table-of-contents"
import { motion, useScroll, useTransform } from "framer-motion"

interface NoteDetailsDialogProps {
  noteId: string
  noteRO?: models.Note
  isOpen: boolean
  onClose: () => void
  mode?: "default" | "readonly"
}

export function NoteDetailsDialog({
  noteId,
  noteRO,
  isOpen,
  onClose,
  mode = "default",
}: NoteDetailsDialogProps) {



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


  function DefaultNoteDetails({
    noteId,
    isOpen,
    onClose,
    mode = "default"
  }: NoteDetailsDialogProps) {

    const { data: note } = useNote(noteId)
    const { mutate: deleteNote } = useDeleteNote()

    const handleDelete = () => {
      deleteNote(note as models.LocalNote)
      onClose()
    }

    const handleOpenChange = (open: boolean) => {
      if (!open) {
        onClose()
      }
    }

    if (!note) return null
    return (

      <Dialog open={isOpen} onOpenChange={handleOpenChange}>
        <DialogContent className="max-w-7xl max-h-[90vh] flex flex-col glass border-white/10 p-0 overflow-hidden gap-0">

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
                  {note.Course && <div className={`w-1.5 h-1.5 rounded-full ${note.Course?.Color} shadow-[0_0_8px] shadow-${note.Course?.Color}/80 ml-0.5`} />}
                  <span className="text-[10px] font-medium pr-1 opacity-80 uppercase tracking-wider text-gray-300">{note.Subject || "General"}</span>
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
                  {note.Course && (
                    <div className="flex items-center gap-1.5">
                      <span className="font-medium text-gray-300">{note.Course.Code}</span>
                      <span className="w-1 h-1 rounded-full bg-gray-600" />
                      <span>{note.Course.Name}</span>
                    </div>
                  )}
                 
                </div>
              </motion.div>
            </div>
          </motion.div>

          {/* Main Content Area with TOC and Content */}
          <div className="flex flex-1 overflow-hidden">
            {/* Table of Contents - Sticky Sidebar */}
            {note.Content && (
              <aside className="w-64 flex flex-col border-r border-white/5 bg-white/[0.02] flex-shrink-0 overflow-y-auto pb-16">
                <div className="p-4 sticky top-0 shrink-0 ">
                  <MarkdownTableOfContents
                    markdown={note.Content}
                    containerRef={scrollRef}
                    maxLevel={3}
                  />
                </div>
                <div className="absolute bottom-0 w-64 p-4 flex justify-center items-center backdrop-blur-md bg-white/5">
                  <Button
                    variant="danger"
                    size="sm"
                    className="rounded-full"
                    onClick={handleDelete}
                  >
                    <Trash2 className="w-4 h-4" />
                    <span>Delete</span>
                  </Button>
                </div>
              </aside>
            )}

            {/* Scrollable Content Area */}
            <div
              className="flex-1 overflow-y-auto scroll-smooth"
              ref={scrollRef}
              onScroll={(e) => scrollY.set(e.currentTarget.scrollTop)}
            >
              {/* Content */}
              <div className="p-6 pt-4">
                <div className="animate-in fade-in slide-in-from-bottom-4 duration-300 focus-visible:ring-0 focus-visible:outline-none mt-0">

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


                {/* Actions */}

              </div>
            </div>

          </div>
        </DialogContent>
      </Dialog>

    )
  }

  switch (mode) {
    case "default":
      return <DefaultNoteDetails noteId={noteId} noteRO={noteRO} isOpen={isOpen} onClose={onClose} mode={mode} />

  }

}