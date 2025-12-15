"use client"

import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query"
import { SaveUIMessage, GetConversationHistory } from "@/wailsjs/go/main/App"
import { UIMessage } from "@ai-sdk/react"


export const aimessageKeys = {
    all: ['aimessages'] as const,
    lists: () => [...aimessageKeys.all, 'list'] as const,
    list: (assignmentID: number) => [...aimessageKeys.lists(), { assignmentID }] as const,
}
// Custom hook for saving messages
// Hook for updating assignments with optimistic updates
export function useSaveUIMessage() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: async ({
            assignmentID,
            vercelMessage
        }: {
            assignmentID: number
            vercelMessage: UIMessage
        }) => {
            return await SaveUIMessage(assignmentID, vercelMessage)
        },

        onSuccess: (savedMessage, variables) => {
            // Invalidate and refetch the conversation history
            queryClient.invalidateQueries({ 
              queryKey: aimessageKeys.list(variables.assignmentID) 
            });
            
            // Optional: Optimistically update the cache
            queryClient.setQueryData(
              aimessageKeys.list(variables.assignmentID),
              (old: UIMessage[] | undefined) => 
                old ? [...old, savedMessage] : [savedMessage]
            );
        },

        // If the mutation fails, rollback
        onError: (error) => {
            console.error('Failed to save message:', error)
        },
    })
}

// Custom hook for fetching conversation history
export const useConversationHistory = (assignmentID: number) => {
    return useQuery({
        queryKey: ['conversation', assignmentID],
        queryFn: () => GetConversationHistory(assignmentID),
        enabled: !!assignmentID, // Only run if assignmentID is provided
        staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
        gcTime: 10 * 60 * 1000, // Keep in cache for 10 minutes
    });
};