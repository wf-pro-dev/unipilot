
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { user } from '@/wailsjs/go/models'
import { LogError } from '@/wailsjs/runtime/runtime'
import { toast } from 'sonner'

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

export function useFollow(currentUser: user.User, isFollow: boolean) {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (followed: user.User) => {
      return await window.go.main.App.Follow(followed.ID)
    },
   
    onMutate: async (followed: user.User) => {
      // Cancel any outgoing refetches (so they don't overwrite our optimistic update)
      await queryClient.cancelQueries({ queryKey: followersKeys.lists() })
      
      // Snapshot the previous value
      const previousCurrentUserFollowing = queryClient.getQueryData<user.User[]>(followingKeys.list(currentUser.ID))
      const previousFollowers = queryClient.getQueryData<user.User[]>(followersKeys.list(currentUser.ID))
      
      // Optimistically update to the new value
      if (isFollow) {
       
        queryClient.setQueryData<user.User[]>(followingKeys.list(currentUser.ID), (old) => {
          if (!old) return [followed]
          return old.filter((user) => user.ID !== followed.ID)
        })

        queryClient.setQueryData<user.User[]>(followersKeys.list(followed.ID), (old) => {
          if (!old) return [currentUser]
          return old.filter((user) => user.ID !== currentUser.ID)
        })
      } else {

        queryClient.setQueryData<user.User[]>(followingKeys.list(currentUser.ID), (old) => {
          if (!old) return [followed]
          return [...old, followed]
        })

        queryClient.setQueryData<user.User[]>(followersKeys.list(followed.ID), (old) => {
          if (!old) return [currentUser]
          return [...old, currentUser]
        })
      }

  
      return { previousCurrentUserFollowing, previousFollowers }
    },

    onError: (err, variables, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      if (context?.previousCurrentUserFollowing) {
        queryClient.setQueryData(followingKeys.list(variables.ID), context.previousCurrentUserFollowing)
      }
      if (context?.previousFollowers) {
        queryClient.setQueryData(followersKeys.list(variables.ID), context.previousFollowers)
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
