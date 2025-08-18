"use client"

import { useQuery } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { user } from '@/wailsjs/go/models'
import { useAuth } from './use-auth'
import { useFollowing } from './use-follows'

// Query keys for consistent cache management
export const userKeys = {
  all: ['users'] as const,
  lists: () => [...userKeys.all, 'list'] as const,
  list: (filters: string) => [...userKeys.lists(), { filters }] as const,
  details: () => [...userKeys.all, 'detail'] as const,
  detail: (id: number) => [...userKeys.details(), id] as const,
}

// Main hook for fetching assignments with caching
export function useUsers() {
  return useQuery({
    queryKey: userKeys.lists(),
    queryFn: async (): Promise<user.User[]> => {
      try {
        var users = await window.go.main.App.GetRemoteUsers()
        console.log("Users fetched", users)
        return users
      } catch (error) {
        LogError("Failed to fetch users: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch users")
      }
    },


    staleTime: 2 * 60 * 1000, // Consider fresh for 2 minutes
    gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
  })
}

export function useFollowRecommendations() {
  const {data : users, ...rest} = useUsers()
  const {user} = useAuth()
  const currentUser = users?.find(u => u.ID === user?.ID)
  const {data: following} = useFollowing(currentUser?.ID as number)

  const followRecommendations = users?.filter(user => {
    return user.ID !== currentUser?.ID && 
    !following?.some(followingUser => followingUser.ID === user.ID) &&
    user.University === currentUser?.University &&
    currentUser?.CoursesCode.some(courseCode => user.CoursesCode.includes(courseCode))
  })

  return {
    data: followRecommendations,
    ...rest
  }
}
// Legacy support - keep the same interface for existing components
export { useUsers as useUsersLegacy } 