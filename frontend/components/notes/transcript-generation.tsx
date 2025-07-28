"use client"

import type React from "react"

import { useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Progress } from "@/components/ui/progress"
import { Mic, Upload, FileAudio, Download, Copy, Square } from "lucide-react"

const courses = [
  { value: "cs-201", label: "CS 201 - Data Structures" },
  { value: "math-301", label: "MATH 301 - Calculus III" },
  { value: "ai-401", label: "AI 401 - Artificial Intelligence" },
  { value: "eng-102", label: "ENG 102 - English Composition" },
]

export function TranscriptGeneration() {
  const [selectedCourse, setSelectedCourse] = useState("")
  const [lectureTitle, setLectureTitle] = useState("")
  const [isRecording, setIsRecording] = useState(false)
  const [isProcessing, setIsProcessing] = useState(false)
  const [transcript, setTranscript] = useState("")
  const [uploadedFile, setUploadedFile] = useState<File | null>(null)
  const [recordingTime, setRecordingTime] = useState(0)
  const [processingProgress, setProcessingProgress] = useState(0)

  const startRecording = () => {
    setIsRecording(true)
    setRecordingTime(0)
    // Start recording timer
    const timer = setInterval(() => {
      setRecordingTime((prev) => prev + 1)
    }, 1000)

    // Store timer ID for cleanup
    ;(window as any).recordingTimer = timer
  }

  const stopRecording = () => {
    setIsRecording(false)
    clearInterval((window as any).recordingTimer)
    processRecording()
  }

  const processRecording = () => {
    setIsProcessing(true)
    setProcessingProgress(0)

    // Simulate processing with progress
    const progressTimer = setInterval(() => {
      setProcessingProgress((prev) => {
        if (prev >= 100) {
          clearInterval(progressTimer)
          setIsProcessing(false)
          setTranscript(`# ${lectureTitle || "Lecture Transcript"}

## Introduction
Welcome to today's lecture on advanced topics in ${courses.find((c) => c.value === selectedCourse)?.label || "the selected course"}. 

## Key Points Discussed
1. **First Topic**: Detailed explanation of the first major concept covered in this lecture.
2. **Second Topic**: In-depth analysis of the second important topic.
3. **Third Topic**: Comprehensive overview of the third key area.

## Examples and Applications
- Practical example 1: Real-world application demonstrating the concepts
- Practical example 2: Another example showing different use cases
- Case study: Detailed analysis of a specific scenario

## Summary
Today we covered the fundamental aspects of the topic and explored various applications. The key takeaways include understanding the core principles and their practical implementations.

## Next Steps
- Review the concepts discussed today
- Complete the assigned readings
- Prepare for the upcoming assignment

*Transcript generated on ${new Date().toLocaleDateString()} at ${new Date().toLocaleTimeString()}*`)
          return 100
        }
        return prev + 10
      })
    }, 200)
  }

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      setUploadedFile(file)
      processRecording() // Process the uploaded file
    }
  }

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`
  }

  const copyTranscript = () => {
    navigator.clipboard.writeText(transcript)
  }

  const downloadTranscript = () => {
    const blob = new Blob([transcript], { type: "text/markdown" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `${lectureTitle.replace(/\s+/g, "-").toLowerCase() || "transcript"}.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <FileAudio className="h-5 w-5 text-green-400" />
            <span>Audio Transcription</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-300 block mb-2">Course</label>
              <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                <SelectTrigger className="bg-gray-800/50 border-gray-600">
                  <SelectValue placeholder="Select a course" />
                </SelectTrigger>
                <SelectContent className="glass border-gray-600">
                  {courses.map((course) => (
                    <SelectItem key={course.value} value={course.value}>
                      {course.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium text-gray-300 block mb-2">Lecture Title</label>
              <Input
                placeholder="Enter lecture title..."
                value={lectureTitle}
                onChange={(e) => setLectureTitle(e.target.value)}
                className="bg-gray-800/50 border-gray-600"
              />
            </div>
          </div>

          {/* Recording Controls */}
          <div className="flex items-center justify-center space-x-4 p-6 rounded-lg glass-dark">
            {!isRecording ? (
              <Button
                onClick={startRecording}
                disabled={!selectedCourse || isProcessing}
                className="bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700"
              >
                <Mic className="h-4 w-4 mr-2" />
                Start Recording
              </Button>
            ) : (
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <div className="w-3 h-3 bg-red-500 rounded-full animate-pulse"></div>
                  <span className="text-red-400 font-mono">{formatTime(recordingTime)}</span>
                </div>
                <Button
                  onClick={stopRecording}
                  variant="outline"
                  className="border-red-500 text-red-400 hover:bg-red-500/10 bg-transparent"
                >
                  <Square className="h-4 w-4 mr-2" />
                  Stop Recording
                </Button>
              </div>
            )}
          </div>

          {/* File Upload */}
          <div className="text-center">
            <div className="text-gray-400 mb-2">or</div>
            <div className="relative">
              <input
                type="file"
                accept="audio/*"
                onChange={handleFileUpload}
                className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                disabled={isRecording || isProcessing}
              />
              <Button
                variant="outline"
                disabled={isRecording || isProcessing}
                className="border-gray-600 bg-transparent"
              >
                <Upload className="h-4 w-4 mr-2" />
                Upload Audio File
              </Button>
            </div>
            {uploadedFile && <p className="text-sm text-gray-400 mt-2">Uploaded: {uploadedFile.name}</p>}
          </div>

          {/* Processing Progress */}
          {isProcessing && (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-300">Processing audio...</span>
                <span className="text-gray-400">{processingProgress}%</span>
              </div>
              <Progress value={processingProgress} className="h-2" />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Transcript Output */}
      {transcript && (
        <Card className="glass border-0">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-white">Generated Transcript</CardTitle>
            <div className="flex space-x-2">
              <Button variant="outline" size="sm" onClick={copyTranscript} className="border-gray-600 bg-transparent">
                <Copy className="h-4 w-4 mr-1" />
                Copy
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={downloadTranscript}
                className="border-gray-600 bg-transparent"
              >
                <Download className="h-4 w-4 mr-1" />
                Download
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="bg-gray-800/50 rounded-lg p-4 border border-gray-600 max-h-96 overflow-y-auto">
              <pre className="text-gray-300 whitespace-pre-wrap text-sm">{transcript}</pre>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
