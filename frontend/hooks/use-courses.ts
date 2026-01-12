"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { models } from "@/wailsjs/go/models"
import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { assignmentKeys } from './use-assignments'
import { documentKeys } from './use-documents'
import { GetCourses, CreateCourse, UpdateCourse, DeleteCourse, CourseShare, GetCoursesLinked, AcceptCourseInvitation, DeclineCourseInvitation } from '@/wailsjs/go/main/App'
import { authKeys } from './use-auth'
import { toast } from 'sonner'

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
    queryFn: async (): Promise<models.LocalCourse[]> => {
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

export function useCoursesLinked() {
  return useQuery({
    queryKey: courseKeys.linked(),
    queryFn: async (): Promise<models.Course[]> => { 
      try {
        const coursesLinked = await GetCoursesLinked()
        return coursesLinked as models.Course[]
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
    mutationFn: async (newCourse: models.LocalCourse) => {
      return await CreateCourse(newCourse)
    },
    
    // Optimistically add the new course
    onMutate: async (newCourse) => {
      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      
      const previousCourses = queryClient.getQueryData<models.LocalCourse[]>(courseKeys.lists())
      
      queryClient.setQueryData<models.LocalCourse[]>(courseKeys.lists(), (old) => {
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
    mutationFn: async ({ course, column, value }: { course: models.LocalCourse, column: string, value: string }) => {
      return await UpdateCourse(course, column, value)
    },
    
    // Optimistic update for instant UI feedback
    onMutate: async ({ course, column, value }) => {
      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      
      const previousCourses = queryClient.getQueryData<models.LocalCourse[]>(courseKeys.lists())
      
      queryClient.setQueryData<models.LocalCourse[]>(courseKeys.lists(), (old) => {
        if (!old) return []
        return old.map(c => 
          c.ID === course.ID 
            ? { ...course, [column]: value, UpdatedAt: new Date() } as models.LocalCourse
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
    mutationFn: async (course: models.LocalCourse) => {
      return await DeleteCourse(course)
    },
    
    // Optimistically remove the course, assignments, and documents
    onMutate: async (course) => {
      
      const previousCourses = queryClient.getQueryData<models.LocalCourse[]>(courseKeys.lists())
      const previousAssignments = queryClient.getQueryData<models.LocalAssignment[]>(assignmentKeys.lists())
      
      // Remove course from cache
      queryClient.setQueryData<models.LocalCourse[]>(courseKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(c => c.ID !== course.ID)
      })
      
      // Remove assignments that belong to this course from cache
      queryClient.setQueryData<models.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
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
export function useCourseShare() {
  return useMutation({
    mutationFn: async ({ c, usersID }: { c: models.LocalCourse, usersID: number[] }) => {
      return await CourseShare(c, usersID)
    },
    onSuccess: () => {
      toast.success("Course shared successfully")
    },
    onError: (err) => {
      LogError("Failed to share course: " + err)
      toast.error("Failed to share course")
    },
  })
}

export function useAcceptCourseInvitation() {
  const queryClient = useQueryClient()
  const createCourse = useCreateCourse()
  const updateCourse = useUpdateCourse()
  return useMutation({
    
    mutationFn: async ({ invitation }: { invitation: models.CourseInvitation }) => {
     
      return await AcceptCourseInvitation(invitation)
    },
    onMutate: async ({ invitation }) => {


      await queryClient.cancelQueries({ queryKey: courseKeys.lists() })
      const previousCourses = queryClient.getQueryData<models.LocalCourse[]>(courseKeys.lists())

      const targetiD = invitation.Course?.ParentID || invitation.Course?.ID || 0

      var existingCourse = previousCourses?.find(c => c.Code === invitation.CourseCode)
      if (existingCourse) {
         await updateCourse.mutate({ course: existingCourse, column: "parent_id", value: targetiD.toString() })
      } else {
        const newCourse = models.LocalCourse.createFrom(invitation.Course)
        newCourse.ParentID = targetiD
        console.log("newCourse", newCourse)
        await createCourse.mutate(newCourse)
      }
     
      await queryClient.cancelQueries({ queryKey: authKeys.coursesInvitations })
      const previousInvitations = queryClient.getQueryData<models.CourseInvitation[]>(authKeys.coursesInvitations)
      queryClient.setQueryData<models.CourseInvitation[]>(authKeys.coursesInvitations, (old) => {
        if (!old) return []
        return old.filter(i => i.ID !== invitation.ID)
      })

      await queryClient.invalidateQueries({ queryKey: courseKeys.linked() })

      return { previousCourses, previousInvitations }
    },
    onSuccess: () => {
      toast.success("Course invitation accepted successfully")
    },
    onError: (err, variables, context) => {
      if (context?.previousCourses) {
        queryClient.setQueryData(courseKeys.lists(), context.previousCourses)
      }
      if (context?.previousInvitations) {
        queryClient.setQueryData(authKeys.coursesInvitations, context.previousInvitations)
      }
      LogError("Failed to accept course invitation: " + err)
      toast.error("Failed to accept course invitation")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: courseKeys.lists() })
      queryClient.invalidateQueries({ queryKey: authKeys.coursesInvitations })
      queryClient.invalidateQueries({ queryKey: courseKeys.linked() })
    },
  })
}

export function useDeclineCourseInvitation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (invitation: models.CourseInvitation) => {
      return await DeclineCourseInvitation(invitation)
    },
    onMutate: async (invitation) => {
      await queryClient.cancelQueries({ queryKey: authKeys.coursesInvitations })
      const previousInvitations = queryClient.getQueryData<models.CourseInvitation[]>(authKeys.coursesInvitations)
      queryClient.setQueryData<models.CourseInvitation[]>(authKeys.coursesInvitations, (old) => {
        if (!old) return []
        return old.filter(i => i.ID !== invitation.ID)
      })
      return { previousInvitations }
    },
    onSuccess: () => {
      toast.success("Course invitation declined successfully")
    },
    onError: (err, variables, context) => {
      if (context?.previousInvitations) {
        queryClient.setQueryData(authKeys.coursesInvitations, context.previousInvitations)
      }
      LogError("Failed to decline course invitation: " + err)
      toast.error("Failed to decline course invitation")
    },
  
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