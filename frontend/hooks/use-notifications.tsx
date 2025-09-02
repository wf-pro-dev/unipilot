
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { notifications } from '@/wailsjs/go/models'

// Query keys for consistent cache management
export const notificationKeys = {
    all: ['notifications'] as const,
    lists: () => [...notificationKeys.all, 'list'] as const,
    list: (filters: string) => [...notificationKeys.lists(), { filters }] as const,
    details: () => [...notificationKeys.all, 'detail'] as const,
    detail: (id: number) => [...notificationKeys.details(), id] as const,
    search: (query: string) => [...notificationKeys.all, 'search', query] as const,
  }
  
  export function useNotifications() {
    return useQuery({
      queryKey: notificationKeys.lists(),
      queryFn: async (): Promise<notifications.LocalNotification[]> => {
        try {
          return await window.go.main.App.GetNotifications()
        } catch (error) {
          LogError("Failed to fetch notifications: " + error)
          throw new Error(error instanceof Error ? error.message : "Failed to fetch notifications")
        }
      },
      staleTime: 2 * 60 * 1000, // Consider fresh for 2 minutes
      gcTime: 10 * 60 * 1000,   // Keep in cache for 10 minutes
      refetchInterval: 10 * 1000, // Refetch every 10 seconds
    })
  }

  // Hook for deleting assignments@
export function useDeleteNotification() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (notification: notifications.LocalNotification) => {
      return await window.go.main.App.DeleteNotification(notification)
    },
    
    // Optimistically remove the assignment
    onMutate: async (notification) => {
      await queryClient.cancelQueries({ queryKey: notificationKeys.lists() })
      
      const previousNotifications = queryClient.getQueryData<notifications.LocalNotification[]>(notificationKeys.lists())
      
      queryClient.setQueryData<notifications.LocalNotification[]>(notificationKeys.lists(), (old) => {
        if (!old) return []
        return old.filter(n => n.ID !== notification.ID)
      })
      
      return { previousNotifications }
    },
    
    onError: (err, variables, context) => {
      if (context?.previousNotifications) {
        queryClient.setQueryData(notificationKeys.lists(), context.previousNotifications)
      }
      LogError("Failed to delete notification: " + err)
    },
    
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.lists() })
    },
  })
}