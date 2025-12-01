"use client"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
} from "@/components/ui/sidebar"
import { useAssignmentDocuments } from "@/hooks/use-documents";
import { assignment } from "@/wailsjs/go/models";
import { AiDocumentCard } from "./ai-chat-documents";
import { GlassCard } from "../ui/glass-card";
import { useNextAssignments } from "@/hooks/use-assignments";
import { AiAssignmentCard } from "./ai-chat-assignments";
import { useMemo } from "react";

interface AiChatSidebarProps {
  assignment: assignment.LocalAssignment;
}



export function AiChatSidebar({ assignment }: AiChatSidebarProps) {
  const { data: documents } = useAssignmentDocuments(assignment.ID)
  const { data: nextAssignments } =  useNextAssignments()

  const getNextAssignments = useMemo(() => {
    return nextAssignments?.
    filter((assign)=> assign.ID != assignment.ID).
    slice(0, 3) || [];
  }, [nextAssignments, assignment.ID]);

  return (
    <Sidebar defaultChecked side="left" className="h-screen pt-16 shadow-2xl w-80" collapsible="icon">
      <GlassCard
        variant={"default"}
        className={`flex flex-col flex-1 border-white/5 bg-white/5`}
      >
        <SidebarHeader className="p-6 pb-4 border-b border-white/5 bg-white/5 backdrop-blur-3xl">
          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 px-1 py-0.5 rounded-full bg-white/5 border border-white/5 w-fit">
                <div className={`w-1.5 h-1.5 rounded-full ${assignment.Course.Color} shadow-[0_0_8px] shadow-${assignment.Course.Color}/80 ml-1.5`} />
                <span className="text-[10px] font-medium pr-2 opacity-80 uppercase tracking-wider">{assignment.Course.Code}</span>
              </div>
            </div>
            <h1 className="text-h4 font-semibold leading-tight text-white drop-shadow-sm">{assignment.Title}</h1>
          </div>
        </SidebarHeader>

        <SidebarContent className="px-4 py-4 gap-6 flex justify-self-center">
          <SidebarGroup className="p-0">
            <div className="flex items-center justify-between mb-3 px-1">
              <SidebarGroupLabel className="text-h4 uppercase tracking-widest text-white/60 font-bold">
                Context Sources
              </SidebarGroupLabel>
              <span className="text-caption font-semibold text-muted-foreground/40 bg-white/5 rounded">{documents?.length || 0}</span>
            </div>
            <div className="flex flex-col gap-2.5">
              {documents?.map((document) => (
                <AiDocumentCard key={document.ID} document={document} added={false} />
              ))}
              {documents?.length === 0 && (
                <div className="px-4 py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                  <p className="text-caption text-muted-foreground/60">No documents linked</p>
                </div>
              )}
            </div>
          </SidebarGroup>
          <SidebarGroup className="p-0">
            <div className="flex items-center justify-between mb-3 px-1">
              <SidebarGroupLabel className="text-h4 uppercase tracking-widest text-white/60 font-bold">
                Upcoming Assignments
              </SidebarGroupLabel>
              <span className="text-[10px] text-muted-foreground/40 bg-white/5 px-1.5 rounded">{documents?.length || 0}</span>
            </div>
            <div className="flex flex-col gap-2.5">
              {getNextAssignments.map((assignment) => (
                <AiAssignmentCard key={assignment.ID} assignment={assignment} />
              ))}
              {getNextAssignments.length === 0 && (
                <div className="px-4 py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                  <p className="text-caption text-muted-foreground/60">No upcoming assignments</p>
                </div>
              )}
            </div>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter className="p-0 border-t border-white/5 bg-gradient-to-t from-black/20 to-transparent">
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