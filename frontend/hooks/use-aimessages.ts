"use client"

import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query"
import { SaveUIMessage, GetConversationHistory } from "@/wailsjs/go/main/App"
import { UIMessage } from "@ai-sdk/react"
import { LogError } from "@/wailsjs/runtime/runtime"
import { models } from "@/wailsjs/go/models"
import { toast } from "sonner"


export const aimessageKeys = {
    all: ['aimessages'] as const,
    lists: () => [...aimessageKeys.all, 'list'] as const,
    list: (assignmentID: string) => [...aimessageKeys.lists(), { assignmentID }] as const,
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
            assignmentID: string
            vercelMessage: UIMessage
        }) => {
            return await SaveUIMessage(assignmentID, vercelMessage)
        },

        onMutate: async ({ assignmentID, vercelMessage }) => {
            await queryClient.cancelQueries({ queryKey: aimessageKeys.list(assignmentID) })
            const previousMessages = queryClient.getQueryData(aimessageKeys.list(assignmentID))

            queryClient.setQueryData(
                aimessageKeys.list(assignmentID),
                (old: UIMessage[] | undefined) =>
                    old ? [...old, vercelMessage] : [vercelMessage]
            );

            return { previousMessages }
        },

        onSuccess: (savedMessage, variables) => {
            // Invalidate and refetch the conversation history
            queryClient.invalidateQueries({
                queryKey: aimessageKeys.list(variables.assignmentID)
            });

        },

        onError: (error, variables, context) => {
            if (context?.previousMessages) {
                queryClient.setQueryData(aimessageKeys.list(variables.assignmentID), context.previousMessages)
            }
            LogError('Failed to save message: ' + error)
            toast.error('Failed to save message')
        },
    })
}

// Custom hook for fetching conversation history
export const useConversationHistory = (assignmentID: string) => {
    return useQuery({
        queryKey: aimessageKeys.list(assignmentID),
        queryFn: async (): Promise<models.LocalAiMessage[]> => {
            try {
                return await GetConversationHistory(assignmentID)
            } catch (error) {
                LogError("Failed to fetch conversation history: " + error)
                throw new Error(error instanceof Error ? error.message : "Failed to fetch conversation history")
            }
        },
        staleTime: 5 * 60 * 1000, // 5 minutes
        gcTime: 15 * 60 * 1000, // 15 minutes
        refetchOnWindowFocus: false,
    });
};