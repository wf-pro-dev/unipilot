"use client"

import { memo } from "react";
import { models } from "@/wailsjs/go/models";
import { ArrowRight } from "lucide-react";
import { formatDeadline } from "@/lib/date-utils";
import { useRouter } from "next/navigation";

interface AssignmentCardProps {
  assignment: models.LocalAssignment,
}

function BaseAssignmentCard(props: AssignmentCardProps) {
  const router = useRouter();

  return (
    <div
      key={props.assignment.ID}
      className={` flex items-center gap-4 px-4 py-2 transition-all duration-300 cursor-pointer group w-full relative  group/document-card`}
      onClick={() => router.push(`/chat?assignment=${props.assignment.ID}`)}>

      {/* Shine effect on hover - inspired by courses-schedule.tsx */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent opacity-0 group-hover/document-card:opacity-100 transition-opacity duration-300" />

      <div className="flex-1 min-w-0 flex flex-col relative z-10">
        <p className="text-body font-medium text-white truncate transition-colors duration-300">
          {props.assignment.Title}
        </p>
        <p className="text-caption text-gray-400 truncate transition-colors">
          {formatDeadline(props.assignment.Deadline)}
        </p>
      </div>

      <div className=" flex items-center justify-center p-1 rounded-full bg-white/10 border border-white/10 shadow-lg shadow-black/40">
        <ArrowRight className="w-4 h-4 text-white" strokeWidth={1.5} />
      </div>
      
    </div>);
}

export const AiAssignmentCard = memo(BaseAssignmentCard, (prevProps, nextProps) => {
  return prevProps.assignment.ID === nextProps.assignment.ID
})  