"use client"

import { useMutation, useQuery, useQueryClient, useInfiniteQuery, UseInfiniteQueryResult, InfiniteData } from '@tanstack/react-query'
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
import { PageResponse } from '@/types/models'  

export const friendKeys = {
    all: ['friends'] as const,
    lists: () => [...friendKeys.all, 'list'] as const,
    list: (userID: string) => [...friendKeys.lists(), userID] as const,
    allStatus: ['friends', 'status'] as const,
    status: (userID: string) => [...friendKeys.allStatus, userID] as const,
}

export function useFriendsScroll({limit = 20, userID}: {limit?: number, userID: string}) {
    console.log("[Hook] Getting friends for user:", userID)
    return useInfiniteQuery({
        queryKey: friendKeys.lists(),
        queryFn: async ({ pageParam }): Promise<PageResponse<models.User>> => {
            try {
                // pageParam will be undefined for first page, then the cursor for subsequent pages
                var friends =  await GetFriends(userID, pageParam!, limit)
                console.log("[Hook] Friends:", friends)
                return friends
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
    })
}

// Main hook for fetching assignments with caching
export function useFriends(userID: string, limit: number, cursor?: models.Cursor) {
    return useQuery({
        queryKey: friendKeys.lists(),
        queryFn: async (): Promise<PageResponse<models.User>> => {
            try {
                return await GetFriends(userID, cursor!, limit)
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
            const previousFriends = queryClient.getQueryData<InfiniteData<PageResponse<models.User>, unknown>>(friendKeys.lists())
            
            queryClient.setQueryData(friendKeys.lists(), (old: InfiniteData<PageResponse<models.User>, unknown>) => {
                if (!old) return { pages: [{ Data: [userID], HasMore: false, Cursor: undefined }], pageParams: undefined }
                return { ...old, pages: [...old.pages, { Data: [userID], HasMore: false, Cursor: undefined }] }
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
            const previousFriends = queryClient.getQueryData<InfiniteData<PageResponse<models.User>, unknown>>(friendKeys.lists())
            queryClient.setQueryData(friendKeys.lists(), (old: InfiniteData<PageResponse<models.User>, unknown>) => {
                if (!old) return { Data: [], HasMore: false, Cursor: undefined }
                return { ...old, Data: old.pages.flatMap(page => page.Data).filter((f: models.User) => f.ID !== userID) }
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