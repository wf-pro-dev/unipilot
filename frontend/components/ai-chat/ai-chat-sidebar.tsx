"use client"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
} from "@/components/ui/sidebar"
import { useAssignmentDocumentIDsRAG } from "@/hooks/use-documents";
import { models } from "@/wailsjs/go/models";
import { AiDocumentCard } from "./ai-chat-documents";
import { GlassCard } from "../ui/glass-card";
import { useNextAssignments } from "@/hooks/use-assignments";
import { AiAssignmentCard } from "./ai-chat-assignments";
import { useCallback } from "react";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";

interface AiChatSidebarProps {
  assignment: models.LocalAssignment;
}

export function AiChatSidebar({ assignment }: AiChatSidebarProps) {
  const router = useRouter();

  const { data: nextAssignments } = useNextAssignments()
  const { data: documentRagIDs } = useAssignmentDocumentIDsRAG(assignment.ID)

  const IsDocumentAdded = useCallback((document: models.LocalDocument) => {
    return documentRagIDs?.includes(document.ID) || false;
  }, [documentRagIDs]);

  const getNextAssignments = useCallback(() => {
    return nextAssignments?.
      filter((assign) => assign.ID != assignment.ID).
      slice(0, 3) || [];
  }, [nextAssignments, assignment.ID]);

  const handleBack = () => {
    router.back();
  };

  return (
    <Sidebar collapsible="icon" className="h-screen shadow-2xl border-none bg-transparent" variant="sidebar">
      <GlassCard variant="board" className="flex flex-col flex-1 border-white/5 bg-white/5 rounded-none h-full p-0 overflow-hidden">
        <SidebarHeader className="p-6 pb-4 border-b border-white/5 bg-white/5 backdrop-blur-3xl">
          <div className="flex flex-col gap-3 group-data-[collapsible=icon]:hidden">
            <div className="flex items-center justify-between">
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 rounded-full hover:bg-white/10 -ml-2"
                onClick={handleBack}
              >
                <ArrowLeft className="h-4 w-4 text-white/70" />
              </Button>
              <div className="flex items-center gap-2 px-1 py-0.5 rounded-full bg-white/5 border border-white/5 w-fit">
                <div className={`w-1.5 h-1.5 rounded-full ${assignment.Course?.Color} shadow-[0_0_8px] shadow-${assignment.Course?.Color}/80 ml-1.5`} />
                <span className="text-[10px] font-medium pr-2 opacity-80 uppercase tracking-wider">{assignment.Course?.Code}</span>
              </div>
            </div>
            <h1 className="text-h4 font-semibold leading-tight text-white drop-shadow-sm truncate">{assignment.Title}</h1>
          </div>
          <div className="hidden group-data-[collapsible=icon]:flex flex-col gap-4 items-center justify-center">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 rounded-full hover:bg-white/10"
              onClick={handleBack}
            >
              <ArrowLeft className="h-4 w-4 text-white/70" />
            </Button>
            <div className={`w-4 h-4 rounded-full ${assignment.Course?.Color} shadow-[0_0_12px] shadow-${assignment.Course?.Color}`} />
          </div>
        </SidebarHeader>

        <SidebarContent className="gap-6 pt-4 flex justify-self-center scrollbar-none">
          <SidebarGroup className="p-0 group-data-[collapsible=icon]:hidden">
            <div >
              <SidebarGroupLabel className="text-xs uppercase tracking-widest text-white/40 font-semibold">
                Context Sources
              </SidebarGroupLabel>
            </div>
            <div className="flex flex-col">
              {assignment.Documents?.map((document) => {
                const added = IsDocumentAdded(document)
                return (
                  <AiDocumentCard key={document.ID} document={document} added={added} />
                )
              })}
              {assignment.Documents?.length === 0 && (
                <div className="px-4 py-8 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                  <p className="text-[10px] text-muted-foreground/60">No documents linked</p>
                </div>
              )}
            </div>
          </SidebarGroup>
          <SidebarGroup className="p-0 group-data-[collapsible=icon]:hidden">
            <div>
              <SidebarGroupLabel className="text-xs uppercase tracking-widest text-white/40 font-semibold">
                Upcoming Assignments
              </SidebarGroupLabel>
            </div>
            <div className="flex flex-col">
              {getNextAssignments().map((assignment) => (
                <AiAssignmentCard key={assignment.ID} assignment={assignment} />
              ))}
              {getNextAssignments().length === 0 && (
                <div className="px-4 py-8 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                  <p className="text-[10px] text-muted-foreground/60">No upcoming assignments</p>
                </div>
              )}
            </div>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter className="p-0 border-t border-white/5 bg-gradient-to-t from-black/20 to-transparent group-data-[collapsible=icon]:hidden">
          <div className="p-5 flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <p className="text-[10px] text-muted-foreground/60 uppercase tracking-wider font-semibold">Due Date</p>
              <div className={`w-1.5 h-1.5 rounded-full ${new Date(assignment.Deadline) < new Date() ? 'bg-red-500 shadow-red-500/50' : 'bg-emerald-500 shadow-emerald-500/50'} shadow-[0_0_8px]`} />
            </div>
            <div className="p-3 rounded-xl bg-white/5 border border-white/5 backdrop-blur-md shadow-inner">
              <p className="text-body-small font-medium text-center text-white/90">
                {new Date(assignment.Deadline).toLocaleDateString(undefined, { weekday: 'long', year: 'numeric', month: 'short', day: 'numeric' })}
              </p>
            </div>
          </div>
        </SidebarFooter>
      </GlassCard>
    </Sidebar>
  )
}