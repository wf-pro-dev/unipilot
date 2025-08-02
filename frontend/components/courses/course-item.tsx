import { useState } from "react"
import { Card, CardContent } from "../ui/card"
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
    const completed_assignments_count = course_assignments.filter((assignment) => assignment.StatusName === "Done").length
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

    return (
        <div>

            <Card
                className={`glass border-0 hover:bg-white/5 transition-colors ${!disabled && onCourseClick ? 'cursor-pointer' : ''
                    } ${disabled ? 'opacity-50' : ''}`}
                onClick={handleCardClick}
                key={course.ID}
            >
                <CardContent className="p-4">
                    <div className="flex justify-between items-start mb-3">
                        <div className="flex items-center space-x-3">
                            <div className={`w-4 h-4 rounded-full ${course.Color}`} />
                            <div>
                                <h3 className="font-semibold text-white">{course.Code}</h3>
                                <p className="text-sm text-gray-400 line-clamp-1">{course.Name}</p>
                            </div>
                        </div>
                        <div className="flex items-center">
                            <Badge variant="outline" className="text-xs border-gray-600">
                                {course.Credits} credits
                            </Badge>

                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        className="p-0 w-8 h-8 text-gray-400 hover:text-white"
                                        disabled={disabled}
                                        onClick={(e) => e.stopPropagation()}
                                    >
                                        <MoreVertical className="w-4 h-4" />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end" className="border-gray-700 glass">
                                    <DropdownMenuItem
                                        onClick={handleEditOpen}
                                        disabled={disabled}
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
                                        className="text-red-400"
                                    >
                                        <Trash2 className="mr-2 w-4 h-4" />
                                        Delete
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>



                        </div>

                    </div>

                    <div className="mb-3 space-y-2">
                        <div className="flex items-center space-x-2 text-sm">
                            <Users className="w-3 h-3 text-blue-400" />
                            <span className="text-gray-300">{course.Instructor}</span>
                        </div>
                        <div className="flex items-center space-x-2 text-sm">
                            <Clock className="w-3 h-3 text-purple-400" />
                            <span className="text-gray-300">{course.Schedule}</span>
                        </div>
                    </div>

                    <div className="flex gap-5 items-center">
                        <div className="flex flex-col space-y-2 grow">
                            <div className="flex justify-between text-sm">
                                <span className="text-gray-400">Progress</span>
                                <span className="text-gray-400">
                                    {completed_assignments_count}/{course_assignments.length}
                                </span>
                            </div>
                            <Progress value={completionPercentage} className="h-1.5" />
                        </div>
                        <Badge variant="secondary" className="text-xs border-gray-600">
                            {course.Semester}
                        </Badge>
                    </div>
                </CardContent>
            </Card>
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