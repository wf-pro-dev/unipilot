import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { useCourseAssignments } from "@/hooks/use-assignments"
import { course } from "@/wailsjs/go/models"
import { toast } from "sonner"

interface CourseDeleteDialogProps {
    isOpen: boolean
    onClose: () => void
    courseId: number | null
    courses: course.LocalCourse[]
    onDelete: (course: course.LocalCourse) => void
}

export function CourseDeleteDialog({ isOpen, onClose, courseId, courses, onDelete }: CourseDeleteDialogProps) {

    const course = courses.find(c => c.ID === courseId) || null
    if (!course) {
        return null
    }

    const { data: course_assignments } = useCourseAssignments(course as course.LocalCourse)
    const course_assignments_count = course_assignments?.length || 0


const handleDelete = () => {
    if (course) {
        onDelete(course)
        toast.success("Course deleted successfully")
        onClose()
    }
}

return (
    <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="border-gray-700 glass">
            <DialogHeader>
                <DialogTitle>Are you sure you want to delete this course?</DialogTitle>
            </DialogHeader>
            <DialogDescription>
            
                <span >This will delete <b className="text-red-500">{course_assignments_count}</b> assignments and <b className="text-red-500">all</b> related documents.</span>
                <br />
                <span className="text-red-500">This action cannot be undone.</span>
            </DialogDescription>
            <DialogFooter>
                <Button variant="outline" onClick={onClose}>Cancel</Button>
                <Button variant="destructive" onClick={handleDelete}>Delete</Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
)
}