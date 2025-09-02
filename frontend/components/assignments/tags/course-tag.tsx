import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { useCourses } from "@/hooks/use-courses"

interface CourseTagProps {
    assignment: assignment.LocalAssignment
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}


function CourseTag({ assignment, onEdit }: CourseTagProps) {
    const { data: courses } = useCourses()

    const handleEdit = (code: string) => {

        onEdit(assignment, "course_code", code)
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="p-0">
                    <Badge variant="secondary" className="text-xs flex flex-row gap-2">
                        <div className={`h-2 w-2  rounded-full ${assignment.Course?.Color}`} />
                        {assignment.Course?.Code || 'No Course'}
                    </Badge>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass border-gray-600 grid grid-cols-2">
                {courses?.map((course) => (
                    <DropdownMenuItem key={course.Code} onClick={() => handleEdit(course.Code)}>
                        <Badge variant="secondary" className={`text-xs flex flex-row gap-2`}>
                            <div className={`h-2 w-2  rounded-full ${course.Color}`} />
                            {course.Code}
                        </Badge>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export { CourseTag }