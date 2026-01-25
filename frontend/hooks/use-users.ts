"use client"

import { useQuery } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { models } from '@/wailsjs/go/models'
import { GetRemoteUsers } from '@/wailsjs/go/main/App'

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
    queryFn: async (): Promise<models.User[]> => {
      try {
        var users = await GetRemoteUsers()
        return users
      } catch (error) {
        LogError(error as string)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch users")
      }
    },


    staleTime: 2 * 60 * 60 * 1000, // Consider fresh for 2 hours
    gcTime: 10 * 60 * 60 * 1000,   // Keep in cache for 10 hours
  })
}

// Legacy support - keep the same interface for existing components
export { useUsers as useUsersLegacy } 