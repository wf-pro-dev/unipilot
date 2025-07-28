"use client"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AINotesGeneration } from "./ai-note-generation"
import { TranscriptGeneration } from "./transcript-generation"
import { Sparkles, FileAudio } from "lucide-react"

export function CreateNotesView() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white mb-2">Create Notes</h2>
        <p className="text-gray-400">Generate study notes using AI or transcribe audio lectures</p>
      </div>

      <Tabs defaultValue="ai-generation" className="w-full">
        <TabsList className="grid w-full grid-cols-2 glass border-0">
          <TabsTrigger value="ai-generation" className="flex items-center space-x-2">
            <Sparkles className="h-4 w-4" />
            <span>AI Generation</span>
          </TabsTrigger>
          <TabsTrigger value="transcription" className="flex items-center space-x-2">
            <FileAudio className="h-4 w-4" />
            <span>Audio Transcription</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="ai-generation" className="mt-6">
          <AINotesGeneration />
        </TabsContent>

        <TabsContent value="transcription" className="mt-6">
          <TranscriptGeneration />
        </TabsContent>
      </Tabs>
    </div>
  )
}
