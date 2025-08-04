"use client"

import type React from "react"
import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { AlertCircle, CheckCircle } from "lucide-react"
import { AudioRecorder } from "./audio-recorder"

interface RecordingMetadata {
  courseId: string
  duration: number
  timestamp: Date
}

export function TranscriptGeneration() {
  const [recordingComplete, setRecordingComplete] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleRecordingComplete = async (metadata: RecordingMetadata) => {
    try {
      setError(null)
      setRecordingComplete(true)
      
      // The backend automatically handles transcription when recording stops
      // No additional processing needed here
      console.log("✅ Recording completed, transcription handled by backend")
      
    } catch (err) {
      console.error("Recording completion error:", err)
      setError("Failed to complete recording. Please try again.")
    }
  }

  const handleNewRecording = () => {
    setRecordingComplete(false)
    setError(null)
  }

  return (
    <div className="space-y-6">
      <AudioRecorder 
        onRecordingComplete={handleRecordingComplete}
        isProcessing={false}
      />

      {/* Success Message */}
      {recordingComplete && (
        <Alert className="border-green-500 bg-green-500/10">
          <CheckCircle className="w-4 h-4 text-green-500" />
          <AlertDescription className="text-green-300">
            Recording completed! Check the terminal for transcription results and look for transcript files in the transcripts/ folder.
          </AlertDescription>
        </Alert>
      )}

      {/* Error Display */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="w-4 h-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Instructions */}
      <Card className="border-0 glass">
        <CardHeader>
          <CardTitle className="text-white">How it works</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-gray-300">
          <p>1. Select a course and start recording</p>
          <p>2. Speak for 10-30 seconds</p>
          <p>3. Stop recording</p>
          <p>4. Backend automatically transcribes using Hugging Face API</p>
          <p>5. Check terminal logs and transcripts/ folder for results</p>
        </CardContent>
      </Card>
    </div>
  )
}
