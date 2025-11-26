import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
} from "@/components/ui/sidebar"
import { useAssignmentDocuments } from "@/hooks/use-documents";
import { assignment } from "@/wailsjs/go/models";
import { Separator } from "@radix-ui/react-separator";
import { useState } from "react";
import { AiDocumentCard } from "./ai-chat-documents";

interface AiChatSidebarProps {
  assignment: assignment.LocalAssignment;
}



export function AiChatSidebar({ assignment }: AiChatSidebarProps) {
  const { data: documents } = useAssignmentDocuments(assignment.ID)

  return (
    <Sidebar defaultChecked side="left" className="glass mt-16">
      <SidebarHeader >
        <p className="text-sm font-medium">{assignment.Title}</p>
      </SidebarHeader>
      <Separator className="my-2 h-[1px] bg-white/10" />
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel className="text-xs font-normal text-gray-400">Documents</SidebarGroupLabel>
          <div className="flex flex-col gap-2 mt-2">
            {
              documents?.map((document) => (
                <AiDocumentCard document={document} added={false} />
              ))
            }
          </div>
          <SidebarGroupContent></SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter />
    </Sidebar>
  )
}