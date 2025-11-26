"use client"

import { useQuery } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { user } from '@/wailsjs/go/models'

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


    staleTime: 2 * 60 * 60 * 1000, // Consider fresh for 2 hours
    gcTime: 10 * 60 * 60 * 1000,   // Keep in cache for 10 hours
  })
}

// Legacy support - keep the same interface for existing components
export { useUsers as useUsersLegacy } 