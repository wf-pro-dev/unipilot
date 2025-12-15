"use client"

import { useSearchParams } from 'next/navigation';
import { useAssignments } from '@/hooks/use-assignments';
import AIChatInterface from '@/components/ai-chat/ai-chat-interface';
import { Loader2 } from 'lucide-react';
import { AiChatSidebar } from '@/components/ai-chat/ai-chat-sidebar';
import { SidebarProvider } from '@/components/ui/sidebar';

/**
 * AI Chat page component for assignment-specific AI assistance.
 * 
 * Provides an AI-powered chat interface for a specific assignment, allowing users
 * to ask questions and get help with their coursework. The page requires an
 * assignment ID in the URL query parameters and displays a sidebar with assignment
 * documents alongside the main chat interface.
 * 
 * Features:
 * - Assignment-specific AI chat interface
 * - Sidebar displaying assignment documents
 * - Conversation history persistence
 * - Loading and error states for assignment lookup
 * 
 * URL Query Parameters:
 * - `assignment`: Required assignment ID (number) to load the chat context
 * 
 * @returns {JSX.Element} The AI chat page with sidebar and chat interface, or loading/error states
 */
export default function AIChatPage() {
  const searchParams = useSearchParams();

  // Parse assignment ID from URL for deep linking support
  const assignmentId = parseInt(searchParams.get('assignment') as string);
  // Fetch all assignments to leverage cached data instead of individual assignment fetch
  const { data: assignments, isLoading } = useAssignments();

  // Locate target assignment using optional chaining to handle undefined state during initial load
  const assignment = assignments?.find(a => a.ID === assignmentId);

  // Prevent rendering until data loads to avoid showing "not found" during fetch
  if (isLoading) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin mr-2" />
          <span>Loading assignment...</span>
        </div>
      </div>
    );
  }

  // Handle invalid assignment ID or missing assignment gracefully
  if (!assignment) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500">Assignment not found</div>
        </div>
      </div>
    );
  }

  return (
    <div className="page-chat" >
      {/* SidebarProvider enables collapsible sidebar state management */}
      <SidebarProvider>

        {/* Decorative background elements */}
        <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
        <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

        <AiChatSidebar assignment={assignment} />

        {/* z-10 ensures chat interface renders above background decorative elements */}
        <div className="relative w-full">
          <AIChatInterface assignment={assignment} />
        </div>

      </SidebarProvider>
    </div>
  );
}