"use client"

import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query"
import { SaveUIMessage, GetConversationHistory } from "@/wailsjs/go/main/App"
import { UIMessage } from "@ai-sdk/react"
import { LogError } from "@/wailsjs/runtime/runtime"
import { aimessage } from "@/wailsjs/go/models"


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
        queryFn: async (): Promise<aimessage.LocalAiMessage[]> => {
            try {
                return await GetConversationHistory(assignmentID)
            } catch (error) {
                LogError("Failed to fetch conversation history: " + error)
                throw new Error(error instanceof Error ? error.message : "Failed to fetch conversation history")
            }
        },
        staleTime: 1000 * 60 * 30, // 30 minutes
        gcTime: 1000 * 60 * 60, // 1 hour
        refetchOnWindowFocus: false,
    });
};