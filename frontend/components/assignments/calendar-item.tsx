"use client"
import { useDrag } from "react-dnd"
import { StatusTag } from "./tags/status-tag"
import { useAssignment } from "@/hooks/use-assignments"
import { useDialogContext } from "../provider/dialog-provider"

interface CalendarItemProps {
  assignmentId: number
}

export function CalendarItem({ assignmentId }: CalendarItemProps) {

  const { data: assignment } = useAssignment(assignmentId)
  if (!assignment) return null

  const { SetDialogState } = useDialogContext()

  const handleCardClick = (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation()
    SetDialogState({ modelType: "assignment", dialogType: "details", id: assignmentId })
  }

  const [{ isDragging }, drag] = useDrag({
    type: "assignment",
    item: { assignment: assignment },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  })


  const priorityColor = {
    low: "bg-emerald-500",
    medium: "bg-amber-500",
    high: "bg-red-500",
  }[assignment.Priority?.toLowerCase() || "low"] || "bg-gray-500"

  const isDone = assignment.Status === "Done"


  return (
    <div>
      <div
        ref={drag as any}
        className={`cursor-pointer  group/item  ${isDragging ? "opacity-50" : ""}`}
        onClick={handleCardClick}
      >
        <div className={`
        relative backdrop-blur-lg overflow-hidden rounded-xl border transition-all 
        ${isDone
            ? 'bg-black/10 border-white/5 opacity-70 hover:opacity-100 duration-30'
            : 'bg-white/5 border-white/5 shadow-lg shadow-black/60 hover:translate-y-1 hover:border-white/10 duration-100 ease-in-out'
          }
      `}>

          {/* Priority Indicator Bar */}
          <div className={`absolute left-0 top-0 bottom-0 w-1 ${priorityColor}`} />

          {/* Shine effect on hover (only if not done) */}
          {!isDone && (
            <div className="absolute inset-0 bg-gradient-to-br from-white/15 to-transparent opacity-0 group-hover/item:opacity-100 transition-opacity duration-300 pointer-events-none" />
          )}

          <div className="flex flex-col gap-1.5 p-2.5 pl-3.5 relative z-10">
            {/* Header: Course Info & Time */}

            <div className="flex items-center gap-1.5 min-w-0 overflow-hidden">
              <div className={`w-1.5 h-1.5 rounded-full flex-shrink-0 shadow-sm ${assignment.Course?.Color}`} />
              <p className="text-caption text-gray-400 font-medium truncate tracking-wider drop-shadow-md">
                {assignment.Course?.Code}
              </p>
            </div>

            {/* Title */}
            <p className={`text-body font-medium truncate leading-tight drop-shadow-sm ${isDone ? "line-through text-white/30" : "text-white"}`}>
              {assignment.Title}
            </p>

            {/* Footer: Tags */}
            <div className="flex items-center justify-between pt-0.5">
              <div className="scale-[0.85] origin-left -ml-1 transition-opacity">
                <StatusTag variant="outline" assignment={assignment} />
              </div>
            </div>
          </div>
        </div>
      </div>
      
    </div>
  )
}
