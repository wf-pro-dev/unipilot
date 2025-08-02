"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { course } from "@/wailsjs/go/models"
import { LogError } from "@/wailsjs/runtime/runtime"
import { assignmentKeys } from './use-assignments'
import { assignment } from '@/wailsjs/go/models'
import { documentKeys } from './use-documents'

// Query keys for consistent cache management
export const courseKeys = {
  all: ['courses'] as const,
  lists: () => [...courseKeys.all, 'list'] as const,
  list: (filters: string) => [...courseKeys.lists(), { filters }] as const,
  details: () => [...courseKeys.all, 'detail'] as const,
  detail: (id: number) => [...courseKeys.details(), id] as const,
}

// Main hook for fetching courses with caching
export function useCourses() {
  return useQuery({
    queryKey: courseKeys.lists(),
    queryFn: async (): Promise<course.LocalCourse[]> => {
      try {
        return await window.go.main.App.GetCourses()
      } catch (error) {
        LogError("Failed to fetch courses: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch courses")
      }
    },
    staleTime: 5 * 60 * 1000, // Courses change less frequently - 5 minutes
    gcTime: 15 * 60 * 1000,   // Keep in cache for 15 minutes
  })
}

// Hook for creating new courses
export function useCreateCourse() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (newCourse: course.LocalCourse) => {
      return await window.go.main.App.CreateCourse(newCourse)
    },
    
    // Optimistically add the new course
    onMutate: async (newCourse) => {
      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      
      const previousCourses = queryClient.getQueryData<course.LocalCourse[]>(courseKeys.lists())
      
      queryClient.setQueryData<course.LocalCourse[]>(courseKeys.lists(), (old) => {
        if (!old) return [newCourse]
        return [newCourse, ...old]
      })
      
      return { previousCourses }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousCourses) {
        queryClient.setQueryData(courseKeys.lists(), context.previousCourses)
      }
      LogError("Failed to create course: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: courseKeys.lists() })
    },
  })
}

// Hook for updating courses
export function useUpdateCourse() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ course, column, value }: { course: course.LocalCourse, column: string, value: string }) => {
      return await window.go.main.App.UpdateCourse(course, column, value)
    },
    
    // Optimistic update for instant UI feedback
    onMutate: async ({ course, column, value }) => {
      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      
      const previousCourses = queryClient.getQueryData<course.LocalCourse[]>(courseKeys.lists())
      
      queryClient.setQueryData<course.LocalCourse[]>(courseKeys.lists(), (old) => {
        if (!old) return []
        return old.map(c => 
          c.ID === course.ID 
            ? { ...course, [column]: value, UpdatedAt: new Date() } as course.LocalCourse
            : c
        )
      })
      
      return { previousCourses }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousCourses) {
        queryClient.setQueryData(courseKeys.lists(), context.previousCourses)
      }
      LogError("Failed to update course: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: courseKeys.lists() })
    },
  })
}

// Hook for deleting courses
export function useDeleteCourse() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (course: course.LocalCourse) => {
      return await window.go.main.App.DeleteCourse(course)
    },
    
    // Optimistically remove the course, assignments, and documents
    onMutate: async (course) => {
      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      await queryClient.cancelQueries({ queryKey: assignmentKeys.all })
      await queryClient.cancelQueries({ queryKey: documentKeys.all })
      
      const previousCourses = queryClient.getQueryData<course.LocalCourse[]>(courseKeys.lists())
      const previousAssignments = queryClient.getQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists())
      
      // Remove course from cache
      queryClient.setQueryData<course.LocalCourse[]>(courseKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(c => c.ID !== course.ID)
      })
      
      // Remove assignments that belong to this course from cache
      queryClient.setQueryData<assignment.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(a => a.Course?.ID !== course.ID)
      })
      
      // Remove all document caches for assignments that belong to this course
      const assignmentsToRemove = previousAssignments?.filter(a => a.Course?.ID === course.ID) || []
      
      assignmentsToRemove.forEach(assignment => {
        // Remove assignment documents
        queryClient.removeQueries({ queryKey: documentKeys.list(assignment.ID) })
        queryClient.removeQueries({ queryKey: documentKeys.support(assignment.ID) })
        queryClient.removeQueries({ queryKey: documentKeys.submissions(assignment.ID) })
      })
      
      return { previousCourses, previousAssignments }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousCourses) {
        queryClient.setQueryData(courseKeys.lists(), context.previousCourses)
      }
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to delete course: " + err)
    },
    
    onSettled: () => {
      // Invalidate all related caches to ensure consistency
      queryClient.invalidateQueries({ queryKey: courseKeys.all })
      queryClient.invalidateQueries({ queryKey: assignmentKeys.all })
      queryClient.invalidateQueries({ queryKey: documentKeys.all })
      
      // Also invalidate storage info since documents were deleted
      queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
    },
  })
}

// Derived data hooks for specific views

export function useActiveCourses() {
  const { data: courses, ...rest } = useCourses()
  
  const activeCourses = courses?.filter(course => {
    const now = new Date()
    const startDate = new Date(course.StartDate)
    const endDate = new Date(course.EndDate)
    return startDate <= now && endDate >= now
  }) || []
  
  return {
    data: activeCourses,
    ...rest
  }
}

export function useUpcomingCourses() {
  const { data: courses, ...rest } = useCourses()
  
  const upcomingCourses = courses?.filter(course => {
    const now = new Date()
    const startDate = new Date(course.StartDate)
    return startDate > now
  }) || []
  
  return {
    data: upcomingCourses,
    ...rest
  }
}

export function useCoursesBySemester(semester: string) {
  const { data: courses, ...rest } = useCourses()
  
  const semesterCourses = courses?.filter(course => 
    course.Semester === semester
  ) || []
  
  return {
    data: semesterCourses,
    ...rest
  }
} 