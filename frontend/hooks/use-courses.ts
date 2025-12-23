"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { course, user, note, assignment } from "@/wailsjs/go/models"
import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { assignmentKeys } from './use-assignments'
import { documentKeys } from './use-documents'
import { GetCourses, CreateCourse, UpdateCourse, DeleteCourse, RequestLinkCourse, GetCoursesLinked } from '@/wailsjs/go/main/App'


// Query keys for consistent cache management
export const courseKeys = {
  all: ['courses'] as const,
  lists: () => [...courseKeys.all, 'list'] as const,
  list: (filters: string) => [...courseKeys.lists(), { filters }] as const,
  details: () => [...courseKeys.all, 'detail'] as const,
  detail: (id: number) => [...courseKeys.details(), id] as const,
  linked: () => [...courseKeys.all, 'linked'] as const,
}

// Main hook for fetching courses with caching
export function useCourses() {
  return useQuery({
    queryKey: courseKeys.lists(),
    queryFn: async (): Promise<course.LocalCourse[]> => {
      try {
        return await GetCourses()
      } catch (error) {
        LogError("Failed to fetch courses: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch courses")
      }
    },
    staleTime: 5 * 60 * 1000, // Courses change less frequently - 5 minutes
    gcTime: 15 * 60 * 1000,   // Keep in cache for 15 minutes
  })
}

// Type matching the API response (lowercase keys)
export type CoursesLinkedData = Record<string, { 
  Users: user.User[], 
  Assignments: assignment.LocalAssignment[], 
  Notes: note.LocalNote[] 
}>

export function useCoursesLinked() {
  return useQuery({
    queryKey: courseKeys.linked(),
    queryFn: async (): Promise<CoursesLinkedData> => { 
      try {
        const coursesLinked = await GetCoursesLinked()
        return coursesLinked as CoursesLinkedData
      } catch (error) {
        LogError("Failed to fetch courses linked: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch courses linked")
      }
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
  })
}


// Hook for creating new courses
export function useCreateCourse() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (newCourse: course.LocalCourse) => {
      return await CreateCourse(newCourse)
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
      return await UpdateCourse(course, column, value)
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
      return await DeleteCourse(course)
    },
    
    // Optimistically remove the course, assignments, and documents
    onMutate: async (course) => {
      
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

// Hook for requesting to link a course to a list of users
export function useRequestLinkCourse() {
  return useMutation({
    mutationFn: async ({ c, usersID }: { c: course.LocalCourse, usersID: number[] }) => {
      return await RequestLinkCourse(c, usersID)
    }
  })
}

// Hook for accepting a link request
export function useAcceptLink() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ courseData }: { courseData: string }) => {
      return await window.go.main.App.AcceptLink(courseData)
    },

    onMutate: async () => {
      // Force refetch for courses, assignments, and documents
      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      await queryClient.cancelQueries({ queryKey: assignmentKeys.lists() })
      await queryClient.cancelQueries({ queryKey: documentKeys.lists() })
        
    }

  })
}

// Derived data hooks for specific views

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