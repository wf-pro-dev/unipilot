"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { addDays, startOfWeek, endOfWeek, isWithinInterval, isAfter, isSameDay } from 'date-fns'
import { assignment, course } from '@/wailsjs/go/models'
import { useMemo } from 'react'

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
    queryFn: async (): Promise<assignment.LocalAssignment[]> => {
      try {
        return await window.go.main.App.GetAssignments()
      } catch (error) {
        LogError("Failed to fetch assignments: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch assignments")
      }
    },
    staleTime: 5 * 60 * 60 * 1000, // Consider fresh for 5 hours
    gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
  })
}




export function useUpdateAssignment() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ assignment, column, value }: { assignment: assignment.LocalAssignment, column: string, value: string }) => {
      return await window.go.main.App.UpdateAssignment(assignment, column, value)
    },
    
    // Optimistic update for instant UI feedback
    onMutate: async ({ assignment, column, value }) => {
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })
      
      const previousAssignments = queryClient.getQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists())
      
      queryClient.setQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return []
        return old.map(a => 
          a.ID === assignment.ID 
            ? { ...a, [column]: value, UpdatedAt: new Date() } as assignment.LocalAssignment
            : a
        )
      })
      
      return { previousAssignments }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to update assignment: " + err)
      // Invalidate on error to ensure we have correct server state
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
    },
    

  })
}

// Hook for creating new assignments
export function useCreateAssignment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (newAssignment: assignment.LocalAssignment) => {
      return await window.go.main.App.CreateAssignment(newAssignment)
    },

    // Optimistically add the new assignment
    onMutate: async (newAssignment) => {
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })

      const previousAssignments = queryClient.getQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists())

      queryClient.setQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
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
    mutationFn: async (assignment: assignment.LocalAssignment) => {
      return await window.go.main.App.DeleteAssignment(assignment)
    },

    // Optimistically remove the assignment
    onMutate: async (assignment) => {
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })

      const previousAssignments = queryClient.getQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists())

      queryClient.setQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
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

export function useCourseAssignments(course: course.LocalCourse) {
  return useQuery({
    queryKey: assignmentKeys.list(course.ID.toString()),
    queryFn: async (): Promise<assignment.LocalAssignment[]> => {
      try {
        return await window.go.main.App.GetCourseAssignments(course)
      } catch (error) {
        LogError("Failed to fetch course assignments: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch course assignments")
      }
    },
    staleTime: 2 * 60 * 1000, // Consider fresh for 2 minutes
    gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
  })
}

// Derived data hooks for specific views (memoized automatically by React Query)

export function useOverdueAssignments() {
  const { data: assignments, ...rest } = useAssignments()

  const overdueAssignments = assignments?.filter(assignment => {
    if (!assignment.Deadline) return false
    return new Date(assignment.Deadline) < addDays(new Date(), -1) && assignment.Status !== 'Done'
  }) || []

  return {
    data: overdueAssignments,
    ...rest
  }
}

export function useTodayAssignments() {
  const { data: assignments, ...rest } = useAssignments()

  const todayAssignments = useMemo(() => assignments?.filter(assignment => {
    if (!assignment.Deadline) return false
    const today = new Date()
    const deadline = new Date(assignment.Deadline)
    return deadline.toDateString() === today.toDateString()
  }) || [], [assignments]) // Memoize the result to avoid unnecessary re-renders

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

  const weekAssignments = useMemo(() => assignments?.filter(assignment => {
    if (!assignment.Deadline) return false
    const deadline = new Date(assignment.Deadline)
    return isWithinInterval(deadline, { start: startOfWeek(new Date()), end: endOfWeek(new Date()) })
  }) || [], [assignments]) // Memoize the result to avoid unnecessary re-renders

  return {
    data: weekAssignments,
    ...rest
  }
}

// Completed assignments
export function useCompletedAssignments() {
  const { data: assignments, ...rest } = useAssignments()

  const completedAssignments = useMemo(() => assignments?.filter(assignment =>
    assignment.Status === 'Done'
  ) || [], [assignments]) // Memoize the result to avoid unnecessary re-renders

  return {
    data: completedAssignments,
    ...rest
  }
}

// Exam assignments
export function useExamAssignments() {
  const { data: assignments, ...rest } = useAssignments()

  const examAssignments = useMemo(() => assignments?.filter(assignment =>
    assignment.Type === 'Exam'
  ) || [], [assignments]) // Memoize the result to avoid unnecessary re-renders

  return {
    data: examAssignments,
    ...rest
  }
}

// Next assignments
export function useNextAssignments() {
  const { data: assignments, ...rest } = useAssignments()

  const nextAssignments = useMemo(() => assignments?.
    filter(assignment => isAfter(assignment.Deadline, new Date()) || isSameDay(assignment.Deadline, new Date())).
    sort((a, b) => new Date(a.Deadline).getTime() - new Date(b.Deadline).getTime()) || [], [assignments]) // Memoize the result to avoid unnecessary re-renders

  return {
    data: nextAssignments,
    ...rest
  }
}

export function useAcceptAssignment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (assignmentData: string) => {
      return await window.go.main.App.AcceptAssignment(assignmentData)
    },

    onSuccess: async (newAssignment) => {
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })

      const previousAssignments = queryClient.getQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists())

      queryClient.setQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return [newAssignment]
        return [newAssignment, ...old]
      })

      return { previousAssignments }
    },

    onError: (err) => {
      LogError("Failed to accept assignment: " + err)
    },

    onSettled: () => {
      
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
    },
  })
}


// Legacy support - keep the same interface for existing components
export { useAssignments as useAssignmentsLegacy } 