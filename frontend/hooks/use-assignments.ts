"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Assignment } from "@/types/models"
import { LogError } from "@/wailsjs/runtime/runtime"
import { addDays, startOfWeek, endOfWeek, isWithinInterval } from 'date-fns'

// Query keys for consistent cache management
export const assignmentKeys = {
  all: ['assignments'] as const,
  lists: () => [...assignmentKeys.all, 'list'] as const,
  list: (filters: string) => [...assignmentKeys.lists(), { filters }] as const,
  details: () => [...assignmentKeys.all, 'detail'] as const,
  detail: (id: number) => [...assignmentKeys.details(), id] as const,
}

// Main hook for fetching assignments with caching
export function useAssignments() {
  return useQuery({
    queryKey: assignmentKeys.lists(),
    queryFn: async (): Promise<Assignment[]> => {
      try {
        return await window.go.main.App.GetAssignments()
      } catch (error) {
        LogError("Failed to fetch assignments: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch assignments")
      }
    },
    staleTime: 2 * 60 * 1000, // Consider fresh for 2 minutes
    gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
  })
}

// Hook for updating assignments with optimistic updates
export function useUpdateAssignment() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ 
      assignment, 
      column, 
      value 
    }: { 
      assignment: Assignment
      column: string
      value: string 
    }) => {
      return await window.go.main.App.UpdateAssignment(assignment, column, value)
    },
    
    // Optimistic update for instant UI feedback
    onMutate: async ({ assignment, column, value }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })
      
      // Snapshot the previous value
      const previousAssignments = queryClient.getQueryData<Assignment[]>(assignmentKeys.lists())
      
      // Optimistically update the cache
      queryClient.setQueryData<Assignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return []
        return old.map(a => 
          a.ID === assignment.ID 
            ? { ...a, [column]: value, UpdatedAt: new Date() }
            : a
        )
      })
      
      return { previousAssignments }
    },
    
    // If the mutation fails, rollback
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to update assignment: " + err)
    },
    
    // Always refetch after error or success to ensure consistency
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
    },
  })
}

// Hook for creating new assignments
export function useCreateAssignment() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (newAssignment: Assignment) => {
      return await window.go.main.App.CreateAssignment(newAssignment)
    },
    
    // Optimistically add the new assignment
    onMutate: async (newAssignment) => {
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })
      
      const previousAssignments = queryClient.getQueryData<Assignment[]>(assignmentKeys.lists())
      
      queryClient.setQueryData<Assignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return [newAssignment]
        return [newAssignment, ...old]
      })
      
      return { previousAssignments }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to create assignment: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
    },
  })
}

// Hook for deleting assignments
export function useDeleteAssignment() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (assignment: Assignment) => {
      return await window.go.main.App.DeleteAssignment(assignment)
    },
    
    // Optimistically remove the assignment
    onMutate: async (assignment) => {
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })
      
      const previousAssignments = queryClient.getQueryData<Assignment[]>(assignmentKeys.lists())
      
      queryClient.setQueryData<Assignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(a => a.ID !== assignment.ID)
      })
      
      return { previousAssignments }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to delete assignment: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
    },
  })
}

// Derived data hooks for specific views (memoized automatically by React Query)
export function useOverdueAssignments() {
  const { data: assignments, ...rest } = useAssignments()
  
  const overdueAssignments = assignments?.filter(assignment => {
    if (!assignment.Deadline) return false
    return new Date(assignment.Deadline) < addDays(new Date(), -1) && assignment.StatusName !== 'Done'
  }) || []
  
  return {
    data: overdueAssignments,
    ...rest
  }
}

export function useTodayAssignments() {
  const { data: assignments, ...rest } = useAssignments()
  
  const todayAssignments = assignments?.filter(assignment => {
    if (!assignment.Deadline) return false
    const today = new Date()
    const deadline = new Date(assignment.Deadline)
    return deadline.toDateString() === today.toDateString()
  }) || []
  
  return {
    data: todayAssignments,
    ...rest
  }
}

export function useAssignmentsByCourse(courseId?: number) {
  const { data: assignments, ...rest } = useAssignments()
  
  const courseAssignments = assignments?.filter(assignment => 
    assignment.Course?.ID === courseId
  ) || []
  
  return {
    data: courseAssignments,
    ...rest
  }
}

// Weekly assignments
export function useWeekAssignments() {
  const { data: assignments, ...rest } = useAssignments()
  
  const weekAssignments = assignments?.filter(assignment => {
    if (!assignment.Deadline) return false
    const deadline = new Date(assignment.Deadline)
    return isWithinInterval(deadline, { start: startOfWeek(new Date()), end: endOfWeek(new Date()) })
  }) || []
  
  return {
    data: weekAssignments,
    ...rest
  }
}

// Completed assignments
export function useCompletedAssignments() {
  const { data: assignments, ...rest } = useAssignments()
  
  const completedAssignments = assignments?.filter(assignment => 
    assignment.StatusName === 'Done' || assignment.Completed
  ) || []
  
  return {
    data: completedAssignments,
    ...rest
  }
}

// Completed assignments
export function useExamAssignments() {
  const { data: assignments, ...rest } = useAssignments()
  
  const examAssignments = assignments?.filter(assignment => 
    assignment.TypeName === 'Exam' 
  ) || []
  
  return {
    data: examAssignments,
    ...rest
  }
}

// Legacy support - keep the same interface for existing components
export { useAssignments as useAssignmentsLegacy } 