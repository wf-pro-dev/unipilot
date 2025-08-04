"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { Note, YouTubeVideo } from '@/types/models'

// Query keys for consistent cache management
export const noteKeys = {
  all: ['notes'] as const,
  lists: () => [...noteKeys.all, 'list'] as const,
  list: (filters: string) => [...noteKeys.lists(), { filters }] as const,
  details: () => [...noteKeys.all, 'detail'] as const,
  detail: (id: number) => [...noteKeys.details(), id] as const,
  search: (query: string) => [...noteKeys.all, 'search', query] as const,
}

// Main hook for fetching notes with caching
export function useNotes() {
  return useQuery({
    queryKey: noteKeys.lists(),
    queryFn: async (): Promise<Note[]> => {
      try {
        // TODO: Replace with actual API call
        // For now, return empty array
        return []
      } catch (error) {
        LogError("Failed to fetch notes: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch notes")
      }
    },
    staleTime: 2 * 60 * 1000, // Notes change frequently - 2 minutes
    gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
  })
}

// Hook for creating new notes
export function useCreateNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (newNote: Omit<Note, 'ID' | 'CreatedAt' | 'UpdatedAt' | 'DeletedAt'>) => {
      // TODO: Replace with actual API call
      const mockNote: Note = {
        ...newNote,
        ID: Date.now(), // Temporary ID generation
        CreatedAt: new Date(),
        UpdatedAt: new Date(),
      }
      return mockNote
    },
    
    // Optimistically add the new note
    onMutate: async (newNote) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      const previousNotes = queryClient.getQueryData<Note[]>(noteKeys.lists())
      
      const mockNote: Note = {
        ...newNote,
        ID: Date.now(),
        CreatedAt: new Date(),
        UpdatedAt: new Date(),
      }
      
      queryClient.setQueryData<Note[]>(noteKeys.lists(), (old) => {
        if (!old) return [mockNote]
        return [mockNote, ...old]
      })
      
      return { previousNotes }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousNotes) {
        queryClient.setQueryData(noteKeys.lists(), context.previousNotes)
      }
      LogError("Failed to create note: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
    },
  })
}

// Hook for updating notes
export function useUpdateNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ note, updates }: { note: Note, updates: Partial<Note> }) => {
      // TODO: Replace with actual API call
      const updatedNote: Note = {
        ...note,
        ...updates,
        UpdatedAt: new Date(),
      }
      return updatedNote
    },
    
    // Optimistic update for instant UI feedback
    onMutate: async ({ note, updates }) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      const previousNotes = queryClient.getQueryData<Note[]>(noteKeys.lists())
      
      queryClient.setQueryData<Note[]>(noteKeys.lists(), (old) => {
        if (!old) return []
        return old.map(n => 
          n.ID === note.ID 
            ? { ...note, ...updates, UpdatedAt: new Date() } as Note
            : n
        )
      })
      
      return { previousNotes }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousNotes) {
        queryClient.setQueryData(noteKeys.lists(), context.previousNotes)
      }
      LogError("Failed to update note: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
    },
  })
}

// Hook for deleting notes
export function useDeleteNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (noteId: number) => {
      // TODO: Replace with actual API call
      return noteId
    },
    
    // Optimistically remove the note
    onMutate: async (noteId) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      const previousNotes = queryClient.getQueryData<Note[]>(noteKeys.lists())
      
      queryClient.setQueryData<Note[]>(noteKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(n => n.ID !== noteId)
      })
      
      return { previousNotes }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousNotes) {
        queryClient.setQueryData(noteKeys.lists(), context.previousNotes)
      }
      LogError("Failed to delete note: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
    },
  })
}

// Hook for searching notes
export function useSearchNotes(query: string) {
  return useQuery({
    queryKey: noteKeys.search(query),
    queryFn: async (): Promise<Note[]> => {
      try {
        // TODO: Replace with actual search API call
        // For now, return empty array
        return []
      } catch (error) {
        LogError("Failed to search notes: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to search notes")
      }
    },
    enabled: query.length > 0,
    staleTime: 1 * 60 * 1000, // Search results change frequently
    gcTime: 5 * 60 * 1000,
  })
}

// Hook for adding YouTube videos to a note
export function useAddYouTubeVideos() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ noteId, videos }: { noteId: number, videos: YouTubeVideo[] }) => {
      // TODO: Replace with actual API call
      return { noteId, videos }
    },
    
    // Optimistically update the note
    onMutate: async ({ noteId, videos }) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      const previousNotes = queryClient.getQueryData<Note[]>(noteKeys.lists())
      
      queryClient.setQueryData<Note[]>(noteKeys.lists(), (old) => {
        if (!old) return []
        return old.map(n => 
          n.ID === noteId 
            ? { 
                ...n, 
                YouTubeVideos: [...(n.YouTubeVideos || []), ...videos],
                UpdatedAt: new Date() 
              } as Note
            : n
        )
      })
      
      return { previousNotes }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousNotes) {
        queryClient.setQueryData(noteKeys.lists(), context.previousNotes)
      }
      LogError("Failed to add YouTube videos: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
    },
  })
}

// Hook for generating AI notes
export function useGenerateNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ title, courseId, subject }: { title: string, courseId: string, subject: string }) => {
      // TODO: Replace with actual API call to backend/Gemini
      // For now, return mock data
      const mockNote: Note = {
        ID: Date.now(),
        UserID: 1, // TODO: Get from auth
        CourseID: parseInt(courseId),
        Title: title,
        Subject: subject,
        Content: `# ${title}\n\nGenerated content for ${subject}...`,
        Keywords: ["keyword1", "keyword2", "keyword3", "keyword4", "keyword5"],
        CreatedAt: new Date(),
        UpdatedAt: new Date(),
      }
      return mockNote
    },
    
    onSuccess: (newNote) => {
      // Add the generated note to the cache
      queryClient.setQueryData<Note[]>(noteKeys.lists(), (old) => {
        if (!old) return [newNote]
        return [newNote, ...old]
      })
    },
    
    onError: (err) => {
      LogError("Failed to generate note: " + err)
    },
  })
} 