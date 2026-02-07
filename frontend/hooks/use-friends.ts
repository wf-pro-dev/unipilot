"use client"

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { models, client } from '@/wailsjs/go/models'
import {
    GetFriends,
    GetFriendShipStatus, 
    SendFriendRequest,
    CancelFriendRequest,
    AcceptFriendRequest,
    RemoveFriend
} from '@/wailsjs/go/main/App'
import { toast } from 'sonner'

export const friendKeys = {
    all: ['friends'] as const,
    lists: () => [...friendKeys.all, 'list'] as const,
    list: (userID: string) => [...friendKeys.lists(), userID] as const,
    allStatus: ['friends', 'status'] as const,
    status: (userID: string) => [...friendKeys.allStatus, userID] as const,
}

// Main hook for fetching assignments with caching
export function useFriends(userID: string, limit: number, offset: number) {
    return useQuery({
        queryKey: friendKeys.lists(),
        queryFn: async (): Promise<models.User[]> => {
            try {
                return await GetFriends(userID, limit, offset)
            } catch (error) {
                LogError("Failed to fetch friends: " + error)
                throw new Error(error instanceof Error ? error.message : "Failed to fetch friends")
            }
        },
        staleTime: 5 * 60 * 60 * 1000, // Consider fresh for 5 hours
        gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
    })
}

export function useFriendShipStatus(userID: string) {
    return useQuery({
        queryKey: friendKeys.status(userID),
        queryFn: async (): Promise<client.FriendStatusResponse> => {
            try {
                return await GetFriendShipStatus(userID)
            } catch (error) {
                LogError("Failed to fetch friend ship status: " + error)
                throw new Error(error instanceof Error ? error.message : "Failed to fetch friend ship status")
            }
        },

        staleTime: 5 * 60 * 60 * 1000, // Consider fresh for 5 hours
        gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
    })
}

export function useSendFriendRequest(userID: string) {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async (): Promise<void> => {
            return await SendFriendRequest(userID)
        },
        onSuccess: () => {
            toast.success("Friend request sent successfully")
        },
        onError: (error) => {
            LogError("Failed to send friend request: " + error)
            toast.error("Failed to send friend request")
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: friendKeys.status(userID) })
        },

    })
}

export function useAcceptFriendRequest(userID: string) {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async (): Promise<void> => {
            return await AcceptFriendRequest(userID)
        },
        onMutate: async () => {
            
            // Optimistically update the friend status
            const previousStatus = queryClient.getQueryData<client.FriendStatusResponse>(friendKeys.status(userID))
            queryClient.setQueryData(friendKeys.status(userID), (old: client.FriendStatusResponse) => {
                if (!old) return { friends_count: 1 }
                return { ...old, friends_count: old.friends_count + 1 }
            })

            // Optimistically update the friends list
            const previousFriends = queryClient.getQueryData<models.User[]>(friendKeys.lists())
            
            queryClient.setQueryData(friendKeys.lists(), (old: models.User[]) => {
                if (!old) return [userID]
                return [...old, userID]
            })


            return { previousFriends, previousStatus }
        },
        onSuccess: () => {
            toast.success("Friend request accepted successfully")
        },
        onError: (error) => {
            
            

            LogError("Failed to accept friend request: " + error)
            toast.error("Failed to accept friend request")
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: friendKeys.lists() })
            queryClient.invalidateQueries({ queryKey: friendKeys.status(userID) })
        },
    })
      
}

export function useCancelFriendRequest(userID: string) {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async (): Promise<void> => {
            return await CancelFriendRequest(userID)
        },
        onSuccess: () => {
            toast.success("Friend request cancelled successfully")
        },
        onError: (error) => {
            LogError("Failed to cancel friend request: " + error)
            toast.error("Failed to cancel friend request")
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: friendKeys.status(userID) })
        },
    })
}

export function useRemoveFriend(userID: string) {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async (): Promise<void> => {
            return await RemoveFriend(userID)
        },
        onMutate: async () => {
            // Optimistically update the friend status
            const previousStatus = queryClient.getQueryData<client.FriendStatusResponse>(friendKeys.status(userID))
            queryClient.setQueryData(friendKeys.status(userID), (old: client.FriendStatusResponse) => {
                if (!old) return { friends_count: 0 }
                return { ...old, friends_count: old.friends_count - 1 }
            })
            // Optimistically update the friends list
            const previousFriends = queryClient.getQueryData<models.User[]>(friendKeys.lists())
            queryClient.setQueryData(friendKeys.lists(), (old: models.User[]) => {
                if (!old) return []
                return old.filter(f => f.ID !== userID)
            })
            return { previousFriends, previousStatus }
        },
        onSuccess: () => {
            toast.success("Friend removed successfully")
        },
        onError: (error) => {
            
            LogError("Failed to remove friend: " + error)
            toast.error("Failed to remove friend")
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: friendKeys.lists() })
            queryClient.invalidateQueries({ queryKey: friendKeys.status(userID) })
        },
    })
}