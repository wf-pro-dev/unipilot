
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { user } from '@/wailsjs/go/models'
import { LogError } from '@/wailsjs/runtime/runtime'

export const followingKeys = {
  all: ['following'] as const,
  lists: () => [...followingKeys.all, 'list'] as const,
  list: (userID: number) => [...followingKeys.lists(), userID] as const,
  details: () => [...followingKeys.all, 'detail'] as const,
  detail: (id: number) => [...followingKeys.details(), id] as const,
}

export const followersKeys = {
  all: ['followers'] as const,
  lists: () => [...followersKeys.all, 'list'] as const,
  list: (userID: number) => [...followersKeys.lists(), userID] as const,
  details: () => [...followersKeys.all, 'detail'] as const,
  detail: (id: number) => [...followersKeys.details(), id] as const,
}

export function useFollowing(userID: number) {
  return useQuery({
    queryKey: followingKeys.list(userID),
    queryFn: async (): Promise<user.User[]> => {
      const response = await window.go.main.App.GetFollowing(userID)
      return response.users
    },
    staleTime: 30 * 60 * 1000, // Consider fresh for 30 minutes
    gcTime: 60 * 60 * 1000,   // Keep in cache for 1 hour
    enabled: !!userID, // Only run query if userID is provided
  })
}

export function useFollowers(userID: number) {
  return useQuery({
    queryKey: followersKeys.list(userID),
    queryFn: async (): Promise<user.User[]> => {    
      const response = await window.go.main.App.GetFollowers(userID)
      return response.users
    },
    staleTime: 30 * 60 * 1000, // Consider fresh for 30 minutes
    gcTime: 60 * 60 * 1000,   // Keep in cache for 1 hour
    enabled: !!userID, // Only run query if userID is provided
  })
}

export function useFollow() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (followed: user.User) => {
      return await window.go.main.App.Follow(followed.ID)
    },
   
    onMutate: async (followed: user.User) => {
      // Cancel any outgoing refetches (so they don't overwrite our optimistic update)
      await queryClient.cancelQueries({ queryKey: followingKeys.lists() })
      
      // Snapshot the previous value
      const previousFollowing = queryClient.getQueryData<user.User[]>(followingKeys.list(followed.ID))
      
      // Optimistically update to the new value
      queryClient.setQueryData<user.User[]>(followingKeys.list(followed.ID), (old) => {
        if (!old) return [followed]
        return [...old, followed]
      })
      
      return { previousFollowing }
    },

    onError: (err, variables, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      if (context?.previousFollowing) {
        queryClient.setQueryData(followingKeys.list(variables.ID), context.previousFollowing)
      }
      LogError("Failed to follow user: " + err)
    },

    onSettled: (data, error, variables) => {
      // Always refetch after error or success to ensure cache consistency
      queryClient.invalidateQueries({ queryKey: followingKeys.list(variables.ID) })
      queryClient.invalidateQueries({ queryKey: followersKeys.list(variables.ID) })
    },
  })
}

// Add a hook to get follow counts
export function useFollowCounts(userID: number) {
  return useQuery({
    queryKey: ['follow-counts', userID],
    queryFn: async (): Promise<{ followers: number; following: number }> => {
      const [followersResponse, followingResponse] = await Promise.all([
        window.go.main.App.GetFollowers(userID),
        window.go.main.App.GetFollowing(userID)
      ])
      return { 
        followers: followersResponse.Count, 
        following: followingResponse.Count 
      }
    },
    staleTime: 5 * 60 * 1000, // Consider fresh for 5 minutes
    gcTime: 30 * 60 * 1000,   // Keep in cache for 30 minutes
    enabled: !!userID,
  })
}
