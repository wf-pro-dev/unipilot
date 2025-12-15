"use client"

import { memo } from "react";
import { assignment } from "@/wailsjs/go/models";
import { ArrowRight } from "lucide-react";
import { formatDeadline } from "@/lib/date-utils";
import { useRouter } from "next/navigation";

interface AssignmentCardProps {
  assignment: assignment.LocalAssignment,
}

function BaseAssignmentCard(props: AssignmentCardProps) {
  const router = useRouter();

  return (
    <div
      key={props.assignment.ID}
      className={`bg-white/5 flex items-center gap-4 p-4 border-white/5 shadow-lg shadow-black/40 hover:border-white/10 hover:translate-y-1 rounded-xl transition-all duration-300 cursor-pointer group w-full text-left border relative overflow-hidden group/document-card`}
      onClick={() => router.push(`/chat?assignment=${props.assignment.ID}`)}>

      {/* Shine effect on hover - inspired by courses-schedule.tsx */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent opacity-0 group-hover/document-card:opacity-100 transition-opacity duration-300" />

      <div className="flex-1 min-w-0 flex flex-col relative z-10">
        <span className={`text-body-small font-semibold text-white truncate transition-colors duration-300 `}>
          {props.assignment.Title}
        </span>
        <span className="text-caption text-white/50 truncate transition-colors">
          {formatDeadline(props.assignment.Deadline)}
        </span>
      </div>

      <div className=" flex items-center justify-center p-2 rounded-full bg-white/10 border border-white/10 shadow-lg shadow-black/40">
        <ArrowRight className="w-4 h-4 text-white" strokeWidth={1.5} />
      </div>
      
    </div>);
}

export const AiAssignmentCard = memo(BaseAssignmentCard, (prevProps, nextProps) => {
  return prevProps.assignment.ID === nextProps.assignment.ID
})  