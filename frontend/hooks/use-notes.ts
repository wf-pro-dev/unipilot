"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { note } from '@/wailsjs/go/models'
import { useToast } from './use-toast'

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
    queryFn: async (): Promise<note.LocalNote[]> => {
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
      note: note.LocalNote
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
      const previousNotes = queryClient.getQueryData<note.LocalNote[]>(noteKeys.lists())
      
      // Optimistically update the cache
      queryClient.setQueryData<note.LocalNote[]>(noteKeys.lists(), (old) => {
        if (!old) return []
        return old.map(n => 
          n.ID === note.ID 
            ? { ...n, [column]: value, UpdatedAt: new Date() } as note.LocalNote
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
    mutationFn: async (newNote: note.LocalNote) => {
      return await window.go.main.App.CreateNote(newNote)
    },
    // Optimistically add the new note
   // Optimistically add the new note
   onMutate: async (newNote) => {
    await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
    
    const previousNotes = queryClient.getQueryData<note.LocalNote[]>(noteKeys.lists())
    
    queryClient.setQueryData<note.LocalNote[]>(noteKeys.lists(), (old) => {
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
    mutationFn: async (noteToDelete: note.LocalNote) => {
      return await window.go.main.App.DeleteNote(noteToDelete)
    },
    
    // Optimistically remove the note
    onMutate: async (noteToDelete) => {
      await queryClient.cancelQueries({ queryKey: noteKeys.lists() })
      
      const previousNotes = queryClient.getQueryData<note.LocalNote[]>(noteKeys.lists())
      
      queryClient.setQueryData<note.LocalNote[]>(noteKeys.lists(), (old) => {
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


