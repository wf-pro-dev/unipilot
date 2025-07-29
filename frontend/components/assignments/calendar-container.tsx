import { format } from "date-fns"
import { CalendarItem } from "./calendar-item"
import { Assignment } from "@/types/models"
import { useDrop } from "react-dnd"
import { useState } from "react"

interface CalendarContainerProps {
    day: Date
    dayAssignments: Assignment[]
    isCurrentMonth: boolean
    isToday: boolean
    onMoveAssignment: (assignment: Assignment, date: Date) => void
    index: number
    onEdit: (assignment: Assignment, column: string, value: string) => void
    onAssignmentClick: (assignment: Assignment) => void
    onDateClick: (date: Date) => void
}

function CalendarContainer({ day, dayAssignments, isCurrentMonth, isToday, onMoveAssignment, index, onEdit, onAssignmentClick, onDateClick }: CalendarContainerProps) {
    const [{ isOver }, drop] = useDrop({
        accept: "assignment",
        drop: (item: { assignment: Assignment }) => {
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
            ref={drop}
            key={index}
            className={`min-h-[120px] p-2 rounded-lg border ${
                isOver ? "bg-blue-500/50" :
                isCurrentMonth
                ? "bg-gray-800/50 border-gray-600"
                : "bg-gray-900/30 border-gray-700"
                } ${isToday ? "ring-2 ring-blue-500" : ""}`}
        >

            <div className="flex items-center justify-between mb-2" onClick={() => onDateClick(day)}>
                <span
                    className={`text-sm font-medium ${isCurrentMonth
                        ? isToday
                            ? "text-blue-400"
                            : "text-white"
                        : "text-gray-500"
                        }`}
                >
                    {format(day, "d")}
                </span>
                {dayAssignments.length > 0 && (
                    <span className="text-xs bg-blue-500 text-white px-1.5 py-0.5 rounded-full">
                        {dayAssignments.length}
                    </span>
                )}
            </div>

            <div className="space-y-1">
                {dayAssignments.slice(0, 1).map((assignment) => (
                    <CalendarItem
                        key={assignment.ID}
                        assignment={assignment}
                        onEdit={onEdit}
                        onAssignmentClick={onAssignmentClick}
                    />
                ))}
                {dayAssignments.length > 1 && (
                    <div className="text-xs text-gray-400 text-center">
                        +{dayAssignments.length - 1} more
                    </div>
                )}
            </div>
            
        </div>
    )
}

export { CalendarContainer }