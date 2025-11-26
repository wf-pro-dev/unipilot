// frontend/hooks/use-ai-chat.ts
import { useMutation } from "@tanstack/react-query"
import { assignment } from "@/wailsjs/go/models"


interface AIChatRequest {
  assignment: assignment.LocalAssignment
  userMessage: string
  conversationHistory: Array<{
    id: string
    content: string
    role: "user" | "assistant"
    timestamp: Date
  }>
}

interface AIChatResponse {
  answer: string
  sources?: string[]
}

export function useAIChat() {
  return useMutation({
    mutationFn: async (request: AIChatRequest): Promise<AIChatResponse> => {
      // This will call your Go backend
      const response = {
        answer: "Hello, how are you?",
        sources: ["source1", "source2"]
      }
      return response;
    },
  });
}