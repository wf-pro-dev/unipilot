"use client"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { CreateNotesView } from "@/components/notes/create-notes-view"
import { SavedNotesList } from "@/components/notes/saved-notes-list"
import { AllTranscriptsView } from "@/components/notes/all-transcripts-view"
import { AllNotesView } from "@/components/notes/all-notes-view"
import { Plus, FileText, FileAudio, Library } from "lucide-react"

export default function NotesPage() {
  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="relative z-10">
        <div className="mb-8">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            Study Notes
          </h1>
          <p className="text-gray-400 mt-2">Create, organize, and manage your study materials</p>
        </div>

        <Tabs defaultValue="create" className="w-full">
          <TabsList className="grid w-full grid-cols-4 glass border-0 mb-8">
            <TabsTrigger value="create" className="flex items-center space-x-2">
              <Plus className="h-4 w-4" />
              <span>Create</span>
            </TabsTrigger>
            <TabsTrigger value="notes" className="flex items-center space-x-2">
              <FileText className="h-4 w-4" />
              <span>All Notes</span>
            </TabsTrigger>
            <TabsTrigger value="transcripts" className="flex items-center space-x-2">
              <FileAudio className="h-4 w-4" />
              <span>Transcripts</span>
            </TabsTrigger>
            <TabsTrigger value="library" className="flex items-center space-x-2">
              <Library className="h-4 w-4" />
              <span>Library</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="create">
            <CreateNotesView />
          </TabsContent>

          <TabsContent value="notes">
            <AllNotesView />
          </TabsContent>

          <TabsContent value="transcripts">
            <AllTranscriptsView />
          </TabsContent>

          <TabsContent value="library">
            <SavedNotesList />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
