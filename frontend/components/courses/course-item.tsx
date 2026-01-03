import { useState } from "react"
import { CardContent } from "../ui/card"
import { GlassCard } from "../ui/glass-card"
import { Progress } from "../ui/progress"
import { Badge } from "../ui/badge"
import { Users, Clock, Edit, MoreVertical, Trash2 } from "lucide-react"
import { course } from "@/wailsjs/go/models"
import { useAssignments } from "@/hooks/use-assignments"
import { LogPrint } from "@/wailsjs/runtime/runtime"
import { CourseEditDialog } from "./course-edit-dialog"
import { Button } from "../ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "../ui/dropdown-menu"

interface CourseItemProps {
    course: course.LocalCourse
    onCourseClick?: (course: course.LocalCourse) => void
    onEdit: (course: course.LocalCourse, column: string, value: string) => void
    onDelete: (course: course.LocalCourse) => void
    disabled?: boolean
}

function CourseItem({ course,
    onCourseClick,
    onEdit,
    onDelete,
    disabled = false
}: CourseItemProps) {

    const { data: assignments } = useAssignments()

    const course_assignments = (assignments || []).filter((assignment) => assignment.Course?.Code === course.Code)
    const completed_assignments_count = course_assignments.filter((assignment) => assignment.Status === "Done").length
    const completionPercentage = course_assignments.length > 0 ? (completed_assignments_count / course_assignments.length) * 100 : 0

    const [open, setOpen] = useState(false)

    const handleCardClick = () => {
        LogPrint("handleCardClick")
        if (onCourseClick && !disabled) {
            onCourseClick(course)
        }
    }

    const handleEditOpen = (e: React.MouseEvent<HTMLDivElement>) => {
        e.stopPropagation()
        setOpen(true)
    }

    // Helper function to get gradient classes for course colors
    // This ensures Tailwind can detect all possible class combinations
    const getCourseGradientClasses = (color: string | undefined, isOn: boolean | null) => {
        if (!color || !isOn) {
            return {
                bg: "bg-white/5",
                hover: "hover:bg-white/10 hover:border-white/10"
            }
        }

        // Map color values to gradient classes that Tailwind can detect
        const colorMap: Record<string, { bg: string; hover: string }> = {
            "bg-blue-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-blue-500/20 before:via-blue-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-blue-500/40 after:via-blue-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-green-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-green-500/20 before:via-green-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-green-500/40 after:via-green-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-purple-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-purple-500/20 before:via-purple-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-purple-500/40 after:via-purple-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-red-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-red-500/20 before:via-red-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-red-500/40 after:via-red-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-orange-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-orange-500/20 before:via-orange-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-orange-500/40 after:via-orange-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
            },
            "bg-pink-500": {
                bg: "before:absolute before:inset-0 before:bg-gradient-to-br before:from-pink-500/20 before:via-pink-500/5 before:to-transparent before:transition-opacity before:duration-500",
                hover: "hover:before:opacity-0 after:absolute after:inset-0 after:bg-gradient-to-br after:from-pink-500/40 after:via-pink-500/15 after:to-transparent after:opacity-0 hover:after:opacity-100 after:transition-opacity after:duration-500",
            }
        }

        return colorMap[color] || {
            bg: "bg-white/5",
            hover: "hover:bg-white/10 hover:border-white/10"
        }
    }

    return (
        <div>

            <GlassCard
                variant={"outline"}
                className={`${disabled ? 'opacity-50' : ''}`}
                onClick={handleCardClick}
                key={course.ID}
            >
                <CardContent className="p-5">
                    <div className="flex justify-between items-start mb-5 ">
                        <div className={`flex items-center border border-white/5  shadow-lg shadow-black/60 rounded-xl p-3 overflow-hidden  space-x-4 w-full  relative `}>

                            <div className={`absolute inset-0 z-0 ${getCourseGradientClasses(course.Color, true).bg} ${getCourseGradientClasses(course.Color, true).hover} transition-colors duration-300`} />
                            
                            <div className="flex flex-col w-full min-w-0">
                                <div className="flex items-center justify-between gap-2">
                                    <h3 className="text-base font-bold text-white line-clamp-1 tracking-tight">{course.Code}</h3>
                                    <div className="flex items-center gap-2 flex-shrink-0">
                                        <Badge variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-300 uppercase tracking-wider px-2 py-0.5">
                                            {course.Credits} credits
                                        </Badge>

                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="p-0 w-8 h-8 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                                                    disabled={disabled}
                                                    onClick={(e) => e.stopPropagation()}
                                                >
                                                    <MoreVertical className="w-4 h-4" />
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end" className="border-white/10 bg-black/90 backdrop-blur-xl glass">
                                                <DropdownMenuItem
                                                    onClick={handleEditOpen}
                                                    disabled={disabled}
                                                    className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
                                                >
                                                    <Edit className="mr-2 w-4 h-4" />
                                                    Edit
                                                </DropdownMenuItem>
                                                <DropdownMenuItem
                                                    onClick={(e) => {
                                                        e.stopPropagation()
                                                        onDelete(course)
                                                    }}
                                                    disabled={disabled}
                                                    className="text-red-400 focus:text-red-300 focus:bg-red-500/10 cursor-pointer"
                                                >
                                                    <Trash2 className="mr-2 w-4 h-4" />
                                                    Delete
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    </div>
                                </div>

                                <p className="text-xs text-gray-400 line-clamp-1 mt-0.5 font-medium">{course.Name}</p>
                            </div>
                        </div>

                    </div>

                    <div className="mb-5 space-y-2">
                        <div className="flex items-center space-x-3 text-xs">
                            <div className="p-1 bg-blue-500/10 rounded-md">
                                <Users className="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
                            </div>
                            <span className="text-gray-300 line-clamp-1 font-medium">{course.Instructor}</span>
                        </div>
                        <div className="flex items-center space-x-3 text-xs">
                            <div className="p-1 bg-purple-500/10 rounded-md">
                                <Clock className="w-3.5 h-3.5 text-purple-400 flex-shrink-0" />
                            </div>
                            <span className="text-gray-300 line-clamp-1 font-medium">{course.Schedule}</span>
                        </div>
                    </div>

                    <div className="flex gap-4 items-end pt-1">
                        <div className="flex flex-col space-y-2 grow">
                            <div className="flex justify-between text-[10px] uppercase tracking-wider font-medium">
                                <span className="text-gray-500">Progress</span>
                                <span className="text-gray-400">
                                    {Math.round(completionPercentage)}%
                                </span>
                            </div>
                            <Progress value={completionPercentage} className="h-1.5 bg-white/10" />
                        </div>
                        <Badge variant="secondary" className="text-[10px] border-white/10 bg-white/5 text-gray-400 flex-shrink-0 font-medium px-2">
                            {course.Semester}
                        </Badge>
                    </div>
                </CardContent>
            </GlassCard>
            <CourseEditDialog
                open={open}
                setOpen={setOpen}
                course={course}
                onEdit={onEdit}
            />
        </div>


    )
}

export { CourseItem }