"use client"

import { useSearchParams } from 'next/navigation';
import { useAssignments } from '@/hooks/use-assignments';
import AIChatInterface from '@/components/ai-chat/ai-chat-interface';
import { Loader2 } from 'lucide-react';
import { AiChatSidebar } from '@/components/ai-chat/ai-chat-sidebar';
import { SidebarProvider } from '@/components/ui/sidebar';

export default function AIChatPage() {


  const searchParams = useSearchParams();
  console.log(searchParams.get('assignment'));
  const assignmentId = parseInt(searchParams.get('assignment') as string);
  const { data: assignments, isLoading } = useAssignments();

  // Find the specific assignment from global state
  const assignment = assignments?.find(a => a.ID === assignmentId);

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
      <SidebarProvider>

        {/* Your existing background elements */}
        <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
        <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

        <AiChatSidebar assignment={assignment} />

        <div className="relative w-full z-10">
          <AIChatInterface assignment={assignment} />
        </div>

      </SidebarProvider>
    </div>
  );
}