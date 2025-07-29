import { Card, CardContent } from "../ui/card"
import { Progress } from "../ui/progress"
import { Badge } from "../ui/badge"
import { Users, Clock } from "lucide-react"
import { Course } from "@/types/models"
import { useAssignments } from "@/hooks/use-assignments"
import { LogPrint } from "@/wailsjs/runtime/runtime"

interface CourseItemProps {
    course: Course
    onEdit: (course: Course, column: string, value: string) => void
    onDelete: (id: number) => void
    onToggleComplete: (course: Course) => void
    onCourseClick?: (course: Course) => void
    disabled?: boolean
}

function CourseItem({ course,
    onEdit,
    onDelete,
    onToggleComplete,
    onCourseClick,
    disabled = false
}: CourseItemProps) {

    const { data: assignments } = useAssignments()

    const course_assignments = (assignments || []).filter((assignment) => assignment.Course?.Code === course.Code)
    const completed_assignments_count = course_assignments.filter((assignment) => assignment.StatusName === "Done").length
    const completionPercentage = course_assignments.length > 0 ? (completed_assignments_count / course_assignments.length) * 100 : 0

    const handleCardClick = () => {
        LogPrint("handleCardClick")
        if (onCourseClick && !disabled) {
            onCourseClick(course)
        }
    }

    return (

        <Card
            className={`glass border-0 hover:bg-white/5 transition-colors ${!disabled && onCourseClick  ? 'cursor-pointer' : ''
                } ${disabled ? 'opacity-50' : ''}`}
            onClick={handleCardClick}
            key={course.ID}
        >
            <CardContent className="p-4">
                <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center space-x-3">
                        <div className={`w-4 h-4 rounded-full ${course.Color}`} />
                        <div>
                            <h3 className="font-semibold text-white">{course.Code}</h3>
                            <p className="text-sm text-gray-400">{course.Name}</p>
                        </div>
                    </div>
                    <Badge variant="outline" className="border-gray-600 text-xs">
                        {course.Credits} credits
                    </Badge>
                </div>

                <div className="space-y-2 mb-3">
                    <div className="flex items-center space-x-2 text-sm">
                        <Users className="h-3 w-3 text-blue-400" />
                        <span className="text-gray-300">{course.Instructor}</span>
                    </div>
                    <div className="flex items-center space-x-2 text-sm">
                        <Clock className="h-3 w-3 text-purple-400" />
                        <span className="text-gray-300">{course.Schedule}</span>
                    </div>
                </div>

                <div className="flex gap-5 items-center">
                    <div className="space-y-2 flex flex-col grow">
                        <div className="flex justify-between text-sm">
                            <span className="text-gray-400">Progress</span>
                            <span className="text-gray-400">
                                {completed_assignments_count}/{course_assignments.length}
                            </span>
                        </div>
                        <Progress value={completionPercentage} className="h-1.5" />
                    </div>
                    <Badge variant="secondary" className="border-gray-600 text-xs">
                        {course.Semester}
                    </Badge>
                </div>
            </CardContent>
        </Card>

    )
}

export { CourseItem }