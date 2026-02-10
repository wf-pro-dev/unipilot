"use client"

import { useSearchParams } from 'next/navigation';
import AIChatInterface from '@/components/ai-chat/ai-chat-interface';
import { AiChatSidebar } from '@/components/ai-chat/ai-chat-sidebar';
import { SidebarProvider } from '@/components/ui/sidebar';
import { useAssignment } from '@/hooks/use-assignments';

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
  const assignmentId = searchParams.get('assignment');
  
  if (!assignmentId) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500">Assignment not found</div>
        </div>
      </div>
    );
  }
  // Fetch all assignments to leverage cached data instead of individual assignment fetch
  const { data: assignment, isLoading } = useAssignment(assignmentId);



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
    <div className="" >
      {/* SidebarProvider enables collapsible sidebar state management */}

      <SidebarProvider>
        <AiChatSidebar assignment={assignment} />
        <div className="relative w-full">
          <AIChatInterface assignment={assignment} />
        </div>
      </SidebarProvider>



    </div>
  );
}