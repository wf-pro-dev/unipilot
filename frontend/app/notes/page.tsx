"use client"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { SavedNotesList } from "@/components/notes/saved-notes-list"
import { AllNotesView } from "@/components/notes/all-notes-view"
import { NoteGenerationForm } from "@/components/notes/note-generation-form"
import { Plus, FileText, Library, Sparkles } from "lucide-react"

export default function NotesPage() {
  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute left-10 top-20 w-72 h-72 rounded-full blur-3xl bg-blue-500/10 animate-float"></div>
      <div className="absolute right-10 bottom-20 w-96 h-96 rounded-full blur-3xl bg-purple-500/10 animate-float-delayed"></div>

      <div className="relative z-10">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
            Study Notes
          </h1>
          <p className="mt-2 text-gray-400">Generate AI-powered study notes for your courses</p>
        </div>

        <Tabs defaultValue="generate" className="w-full">
          <TabsList className="grid grid-cols-3 mb-8 w-full border-0 glass">
            <TabsTrigger value="generate" className="flex items-center space-x-2">
              <Sparkles className="w-4 h-4" />
              <span>Generate</span>
            </TabsTrigger>
            <TabsTrigger value="notes" className="flex items-center space-x-2">
              <FileText className="w-4 h-4" />
              <span>All Notes</span>
            </TabsTrigger>
            <TabsTrigger value="library" className="flex items-center space-x-2">
              <Library className="w-4 h-4" />
              <span>Library</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="generate">
            <NoteGenerationForm />
          </TabsContent>

          <TabsContent value="notes">
            <AllNotesView />
          </TabsContent>

          <TabsContent value="library">
            <SavedNotesList />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
