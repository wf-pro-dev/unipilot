import { format } from "date-fns"
import { CalendarItem } from "./calendar-item"
import { assignment } from "@/wailsjs/go/models"
import { useDrop } from "react-dnd"

interface CalendarContainerProps {
    day: Date
    dayAssignments: assignment.LocalAssignment[]
    isCurrentMonth: boolean
    isToday: boolean
    onMoveAssignment: (assignment: assignment.LocalAssignment, date: Date) => void
    index: number
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
    onAssignmentClick: (assignment: assignment.LocalAssignment) => void
    onDateClick: (date: Date) => void
}

function CalendarContainer({ day, dayAssignments, isCurrentMonth, isToday, onMoveAssignment, index, onEdit, onAssignmentClick, onDateClick }: CalendarContainerProps) {
    const [{ isOver }, drop] = useDrop({
        accept: "assignment",
        drop: (item: { assignment: assignment.LocalAssignment }) => {
            if (item.assignment) {
                onMoveAssignment(item.assignment, day)
            }
        },
        collect: (monitor) => ({
            isOver: monitor.isOver(),
        }),
    })

    return (
        <div
            ref={drop as any}
            key={index}
            className={`min-h-[120px] p-2 transition-colors duration-200 border border-white/10 ${
                isOver ? "bg-blue-500/20" :
                isCurrentMonth
                ? ""
                : "bg-black/40 opacity-50"
                } ${isToday ? "bg-blue-500/5 relative overflow-hidden" : ""}`}
        >
            {isToday && (
                <div className="absolute top-0 left-0 w-full h-0.5 bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]" />
            )}

            <div className="group hover:bg-white/5 flex items-center justify-between mb-2 cursor-pointer" onClick={() => onDateClick(day)}>
                <span
                    className={`text-xs font-medium ${isCurrentMonth
                        ? isToday
                            ? "text-blue-400"
                            : "text-gray-400 group-hover:text-white transition-colors"
                        : "text-gray-600"
                        }`}
                >
                    {format(day, "d")}
                </span>
                {dayAssignments.length > 0 && (
                    <span className={`text-[10px] px-1.5 py-0.5 rounded-full font-medium ${isToday ? "bg-blue-500 text-white" : "bg-white/10 text-gray-300"}`}>
                        {dayAssignments.length}
                    </span>
                )}
            </div>

            <div className="space-y-1.5">
                {dayAssignments.slice(0, 2).map((assignment) => (
                    <CalendarItem
                        key={assignment.ID}
                        assignment={assignment}
                        onEdit={onEdit}
                        onAssignmentClick={onAssignmentClick}
                    />
                ))}
                {dayAssignments.length > 2 && (
                    <div className="text-[10px] text-gray-500 text-center font-medium hover:text-gray-300 transition-colors cursor-pointer pt-1" onClick={() => onDateClick(day)}>
                        +{dayAssignments.length - 2} more
                    </div>
                )}
            </div>
            
        </div>
    )
}

export { CalendarContainer }
