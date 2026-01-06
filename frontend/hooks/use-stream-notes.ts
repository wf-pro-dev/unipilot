// frontend/hooks/use-stream-notes.ts
"use client"

import { fetchEventSource } from '@microsoft/fetch-event-source'
import { useState, useCallback, useRef, useEffect } from 'react'
import { toast } from 'sonner'
import { useQueryClient } from '@tanstack/react-query'
import { noteKeys } from './use-notes'
import { GetAuthToken } from '@/wailsjs/go/main/App'
import { models } from '@/wailsjs/go/models'
import { useCreateNote } from './use-notes'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://wwwill.xyz/unipilot/api/v1'

// Interface for field-specific chunk data
interface ChunkData {
    chunk: string
}

// Interface for error event data
interface ErrorData {
    error: string
}




export function useStreamNote() {
    const [content, setContent] = useState('')
    const [isStreaming, setIsStreaming] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const abortControllerRef = useRef<AbortController | null>(null)
    const accumulatedContentRef = useRef('')
    const noteDataRef = useRef<models.LocalNote | null>(null)
    const queryClient = useQueryClient()
    const { mutate: CreateNote } = useCreateNote()

    const startStream = useCallback(async (noteData: models.LocalNote) => {
        // Store noteData in ref for use in completion handler
        noteDataRef.current = noteData
        // Reset state
        setContent('')
        setError(null)
        setIsStreaming(true)
        accumulatedContentRef.current = ''

        // Create abort controller for cancellation
        const abortController = new AbortController()
        abortControllerRef.current = abortController


        try {
            // Get auth token from Wails backend
            const token = await GetAuthToken()
            if (!token) {
                throw new Error('No auth token found')
            }

            await fetchEventSource(`${API_BASE_URL}/notes/stream`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify(noteData),
                signal: abortController.signal,

                // Handle incoming events - event.event contains the event type
                onmessage(event) {
                    try {
                        // Handle different event types based on event.event
                        const eventType = event.event || '' // Empty string if no event type specified
                        switch (eventType) {

                            case 'error':
                                // Error event from server
                                try {
                                    const errorData: ErrorData = JSON.parse(event.data)
                                    const errorMessage = errorData.error || 'Unknown error occurred'
                                    setError(errorMessage)
                                    setIsStreaming(false)
                                    toast.error(errorMessage)
                                } catch (parseErr) {
                                    console.error('Error parsing error event:', parseErr)
                                    setError('Failed to parse error message')
                                    setIsStreaming(false)
                                    toast.error('Error generating note')
                                }
                                break

                            case 'complete':
                                // Completion event - streaming finished, now create note locally
                                try {
                                    // Finalize the accumulated content
                                    setContent(accumulatedContentRef.current)
                                    setIsStreaming(false)

                                    // Get noteData from ref (safe access)
                                    const currentNoteData = noteDataRef.current
                                    if (!currentNoteData) {
                                        console.error('Note data not available in completion handler')
                                        toast.error('Failed to save note: missing note data')
                                        return
                                    }

                                    // Create note locally using accumulated content
                                    const localNote = {
                                        remote_id: 0, // Will be set by CreateNote when syncing to server
                                        course_code: currentNoteData.CourseCode,
                                        title: currentNoteData.Title,
                                        subject: currentNoteData.Subject,
                                        content: accumulatedContentRef.current,
                                        videos: '',
                                    } as unknown as models.LocalNote

                                    CreateNote(localNote, {
                                        onSuccess: () => {
                                            queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
                                            toast.success('Note generated and saved successfully!')
                                        },
                                        onError: (error) => {
                                            console.error('Error creating note locally:', error)
                                            toast.warning('Note generated but failed to save locally. Please sync.')
                                        }
                                    })

                                } catch (parseErr) {
                                    console.error('Error parsing complete event:', parseErr)
                                    // Even if parsing fails, mark as complete
                                    setIsStreaming(false)
                                    queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
                                    toast.success('Note generation completed')
                                }
                                break

                            default:
                                // Generic data event - extract chunk value from JSON format
                                // All chunks are treated as content
                                if (event.data) {
                                    try {
                                        // Parse JSON to extract chunk value: {"chunk":"..."}
                                        const data: ChunkData = JSON.parse(event.data)
                                        if (data.chunk !== undefined && data.chunk !== null) {
                                            // Accumulate all chunks as content
                                            accumulatedContentRef.current += String(data.chunk)
                                            setContent(accumulatedContentRef.current)
                                        }
                                    } catch (parseErr) {
                                        // JSON parsing failed - likely due to unescaped characters in chunk from server
                                        // The server uses fmt.Fprintf with %s which doesn't escape JSON properly
                                        // Try to extract chunk value manually as fallback
                                        console.warn('Failed to parse chunk JSON (server may have unescaped characters):', parseErr)
                                        
                                        // Try to extract chunk value by finding content between "chunk":" and closing quote
                                        // This is a best-effort fallback for malformed JSON
                                        // Use [\s\S] instead of . with s flag for broader compatibility
                                        const chunkPattern = /"chunk"\s*:\s*"([\s\S]*?)"(?:\s*[,}])/
                                        const match = event.data.match(chunkPattern)
                                        
                                        if (match && match[1] !== undefined) {
                                            // Found chunk value, unescape common escape sequences
                                            let chunkValue = match[1]
                                                .replace(/\\n/g, '\n')
                                                .replace(/\\r/g, '\r')
                                                .replace(/\\t/g, '\t')
                                                .replace(/\\"/g, '"')
                                                .replace(/\\\\/g, '\\')
                                            
                                            accumulatedContentRef.current += chunkValue
                                            setContent(accumulatedContentRef.current)
                                        } else {
                                            // If extraction fails, check if it's plain text (no JSON structure)
                                            // This handles edge cases where server sends plain text instead of JSON
                                            const trimmedData = event.data.trim()
                                            if (!trimmedData.startsWith('{') && !trimmedData.startsWith('"')) {
                                                // Looks like plain text, use it directly
                                                accumulatedContentRef.current += event.data
                                                setContent(accumulatedContentRef.current)
                                            } else {
                                                // Malformed JSON that we can't parse - log error but continue
                                                console.error('Unable to parse chunk data, skipping:', {
                                                    data: event.data.substring(0, 100), // Log first 100 chars
                                                    error: parseErr
                                                })
                                            }
                                        }
                                    }
                                }
                                break
                        }
                    } catch (err) {
                        console.error('Error handling SSE message:', err)
                        setError('Failed to process server response')
                    }
                },

                // Handle connection errors
                onerror(err) {
                    console.error('SSE connection error:', err)
                    setError('Connection error occurred')
                    setIsStreaming(false)
                    toast.error('Connection lost while generating note')
                    throw err // Re-throw to stop retrying
                },

                // Handle connection open
                async onopen(res) {
                    if (res.ok && res.status === 200) {
                        console.log('SSE connection opened successfully')
                    } else if (res.status >= 400 && res.status < 500 && res.status !== 429) {
                        // Client errors - don't retry
                        setError(`HTTP ${res.status}: ${res.statusText}`)
                        setIsStreaming(false)
                        toast.error(`Failed to start stream: ${res.statusText}`)
                    } else {
                        // Server errors - will retry
                        throw new Error(`HTTP ${res.status}: ${res.statusText}`)
                    }
                },
            })
        } catch (err: any) {
            if (err.name === 'AbortError') {
                console.log('Stream aborted by user')
                setIsStreaming(false)
            } else {
                setError(err.message || 'Failed to stream note')
                setIsStreaming(false)
                toast.error('Failed to generate note')
            }
        }
    }, [queryClient, CreateNote])

    const stopStream = useCallback(() => {
        if (abortControllerRef.current) {
            abortControllerRef.current.abort()
            abortControllerRef.current = null
            setIsStreaming(false)
        }
    }, [])

    const reset = useCallback(() => {
        setContent('')
        setError(null)
        setIsStreaming(false)
        accumulatedContentRef.current = ''
        noteDataRef.current = null
        if (abortControllerRef.current) {
            abortControllerRef.current.abort()
            abortControllerRef.current = null
        }
    }, [])

    useEffect(() => {
        return () => {
            if (abortControllerRef.current) {
                abortControllerRef.current.abort()
            }
        }
    }, [])

    return {
        content,
        isStreaming,
        error,
        startStream,
        stopStream,
        reset,
    }
}