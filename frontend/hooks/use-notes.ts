"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { models} from '@/wailsjs/go/models'
import { assignmentKeys } from './use-assignments'

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
    queryFn: async (): Promise<models.LocalNote[]> => {
      try {

        return await window.go.main.App.GetNotes()
      } catch (error) {
        LogError("Failed to fetch notes: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch notes")
      }
    },
    staleTime: 2 * 60 * 1000, // Notes change frequently - 2 minutes
    gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
  })
}

// Hook for updating notes with optimistic updates
export function useUpdateNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ 
      note, 
      column, 
      value 
    }: { 
      note: models.LocalNote
      column: string
      value: string 
    }) => {
      return await window.go.main.App.UpdateNote(note, column, value)
    },
    
    // Optimistic update for instant UI feedback
    onMutate: async ({ note, column, value }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      // Snapshot the previous value
      const previousNotes = queryClient.getQueryData<models.LocalNote[]>(noteKeys.lists())
      
      // Optimistically update the cache
      queryClient.setQueryData<models.LocalNote[]>(noteKeys.lists(), (old) => {
        if (!old) return []
        return old.map(n => 
          n.ID === note.ID 
            ? { ...n, [column]: value, UpdatedAt: new Date() } as models.LocalNote
            : n
        )
      })
      
      return { previousNotes }
    },
    
    // If the mutation fails, rollback
    onError: (err, variables, context) => {
      if (context?.previousNotes) {
        queryClient.setQueryData(noteKeys.lists(), context.previousNotes)
      }
      LogError("Failed to update note: " + err)
    },
    
    // Always refetch after error or success to ensure consistency
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
    },
  })
}

// Hook for creating new notes
export function useCreateNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (newNote: models.LocalNote) => {
      return await window.go.main.App.CreateNote(newNote)
    },
    // Optimistically add the new note
   // Optimistically add the new note
   onMutate: async (newNote) => {
    await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
    
    const previousNotes = queryClient.getQueryData<models.LocalNote[]>(noteKeys.lists())
    
    queryClient.setQueryData<models.LocalNote[]>(noteKeys.lists(), (old) => {
      if (!old) return [newNote]
      return [newNote, ...old]
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



// Hook for deleting notes
export function useDeleteNote() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (noteToDelete: models.LocalNote) => {
      return await window.go.main.App.DeleteNote(noteToDelete)
    },
    
    // Optimistically remove the note
    onMutate: async (noteToDelete) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      const previousNotes = queryClient.getQueryData<models.LocalNote[]>(noteKeys.lists())
      
      queryClient.setQueryData<models.LocalNote[]>(noteKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(n => n.ID !== noteToDelete.ID)
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

// Hook for accepting notes
export function useAcceptNote() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (noteData: string) => {
      return await window.go.main.App.AcceptNote(noteData)
    },

    onSuccess: async (newNote) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })

      const previousNotes = queryClient.getQueryData<models.LocalNote[]>(noteKeys.lists())

      queryClient.setQueryData<models.LocalNote[]>(noteKeys.lists(), (old) => {
        if (!old) return [newNote]
        return [newNote, ...old]
      })

      return { previousNotes }
    },

    onError: (err) => {
      LogError("Failed to accept note: " + err)
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: noteKeys.lists() })
    },
  })
}



export function useCourseNotes(course: models.LocalCourse) {
  const { data: notes } = useNotes()
  const courseNotes = (notes || []).filter(note => note.CourseCode == course.Code)
  return courseNotes
}


