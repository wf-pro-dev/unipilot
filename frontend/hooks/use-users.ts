"use client"

import { InfiniteData, keepPreviousData, useQuery } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { models } from '@/wailsjs/go/models'
import { GetClusterUsers, GetFriends, GetRemoteUsers } from '@/wailsjs/go/main/App'
import { useInfiniteQuery } from '@tanstack/react-query'
import { PageResponse } from '@/types/models'
import { FilterState } from '@/components/core/search-filter/types'

// Query keys for consistent cache management
export const userKeys = {
  all: ['users'] as const,
  lists: () => [...userKeys.all, 'list'] as const,
  search: (search: string, filters: string) => [...userKeys.all, 'search', search, filters] as const,
  cluster: (courseID: string, search: string) => [...userKeys.all, 'cluster', courseID, search] as const,
  details: () => [...userKeys.all, 'detail'] as const,
  detail: (id: number) => [...userKeys.details(), id] as const,
}

// Main hook for fetching users 
export function useUsersScroll({limit = 20, search = "", filters = {}}: {entityID?: string, limit?: number, search?: string, filters?: FilterState}) {

  return useInfiniteQuery({
      queryKey: userKeys.search(search, JSON.stringify(filters)),
      queryFn: async ({ pageParam }): Promise<PageResponse<models.User>> => {
          try {
              // pageParam will be undefined for first page, then the cursor for subsequent pages
              return await GetRemoteUsers(pageParam!, limit, search, filters)
          } catch (error) {
              LogError("Failed to fetch friends: " + error)
              throw new Error(error instanceof Error ? error.message : "Failed to fetch friends")
          }
      },
      initialPageParam: undefined as models.Cursor | undefined,
      getNextPageParam: (lastPage) => {
          // Return the cursor for the next page, or undefined if no more pages
          return lastPage.HasMore ? lastPage.Cursor : undefined
      },
      staleTime: 5 * 60 * 1000, // Consider fresh for 5 minutes
      gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
      placeholderData: keepPreviousData,
  })
}


export function useClusterUsersScroll({entityID: courseID, limit = 20, search = ""}: {entityID: string, limit?: number, search?: string}) {

  return useInfiniteQuery({
      queryKey: userKeys.cluster(courseID, search),
      queryFn: async ({ pageParam }): Promise<PageResponse<models.User>> => {
          try {
              // pageParam will be undefined for first page, then the cursor for subsequent pages
              return await GetClusterUsers(courseID, pageParam!, limit, search)
          } catch (error) {
              LogError("Failed to fetch cluster users: " + error)
              throw new Error(error instanceof Error ? error.message : "Failed to fetch cluster users")
          }
      },
      initialPageParam: undefined as models.Cursor | undefined,
      getNextPageParam: (lastPage) => {
          // Return the cursor for the next page, or undefined if no more pages
          return lastPage.HasMore ? lastPage.Cursor : undefined
      },
      staleTime: 5 * 60 * 1000, // Consider fresh for 5 minutes
      gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
      placeholderData: keepPreviousData,
  })
}


export function useUser(id: string) { 
  
  // Directly subscribe to assignment cache changes
  const { data: users } = useQuery({
    queryKey: userKeys.lists(),
    enabled: false,
  })
  
  const user = (users as InfiniteData<PageResponse<models.User>, unknown>)?.pages.flatMap(page => page.Data)?.find(u => u.ID === id)
  return {
    data: user,
    isLoading: false,
    isError: false,
  }
}
// Legacy support - keep the same interface for existing components
export { useUsersScroll as useUsersLegacy } 