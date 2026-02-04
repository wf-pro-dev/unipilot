import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { useCourse, useCourseAssignments, useDeleteCourse } from "@/hooks/use-courses"
import { models } from "@/wailsjs/go/models"

interface CourseDeleteDialogProps {
    courseId: string
    isOpen: boolean
    onClose: () => void
}
/**
 * Confirmation dialog component for course deletion.
 * 
 * Displays a warning dialog before deleting a course, showing the number of
 * assignments that will be deleted along with the course. Provides visual
 * feedback about the destructive nature of the action.
 * 
 * Features:
 * - Shows assignment count that will be deleted
 * - Warns about permanent data loss
 * - Fetches course assignments to display accurate count
 * - Toast notification on successful deletion
 * 
 * @param {CourseDeleteDialogProps} props - Component props
 * @param {boolean} props.isOpen - Controls dialog visibility
 * @param {() => void} props.onClose - Callback to close the dialog
 * @param {number | null} props.courseId - ID of the course to delete
 * @param {course.LocalCourse[]} props.courses - Array of all courses to find the target course
 * @param {(course: course.LocalCourse) => void} props.onDelete - Callback to execute deletion
 * @returns {JSX.Element | null} The delete confirmation dialog or null if course not found
 */
export function CourseDeleteDialog({ isOpen, onClose, courseId }: CourseDeleteDialogProps) {
    // Locate the course to delete from the courses array
    const { data: course } = useCourse(courseId)
    if (!course) return null
    const { data: course_assignments } = useCourseAssignments(courseId)
    const course_assignments_count = course_assignments?.length || 0

    const deleteCourse = useDeleteCourse()

    /**
     * Handles the delete confirmation action.
     * 
     * Executes the deletion callback, shows success toast, and closes the dialog.
     * This is called when user confirms the deletion.
     */
    const handleDelete = () => {
    if (course) {
        deleteCourse.mutate(course)
        onClose()
    }
}

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-md p-0 overflow-hidden gap-0">
                <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <DialogTitle className="text-xl font-semibold text-white">Delete Course</DialogTitle>
                </DialogHeader>
                <div className="p-6">
                    <DialogDescription className="text-gray-300 text-base leading-relaxed">
                        Are you sure you want to delete this course?
                        <br />
                        <span className="block mt-2">
                            This will delete <b className="text-red-400">{course_assignments_count}</b> assignments and <b className="text-red-400">all</b> related documents.
                        </span>
                        <span className="block mt-2 text-red-400 font-medium bg-red-900/20 p-3 rounded-lg border border-red-500/20">
                            This action cannot be undone.
                        </span>
                    </DialogDescription>
                    <DialogFooter className="mt-6 flex gap-3">
                        <Button 
                            variant="outline" 
                            onClick={onClose}
                            className="border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white"
                        >
                            Cancel
                        </Button>
                        <Button 
                            variant="destructive" 
                            onClick={handleDelete}
                            className="bg-red-600 hover:bg-red-500 text-white shadow-[0_0_15px_rgba(220,38,38,0.3)] border-0"
                        >
                            Delete Course
                        </Button>
                    </DialogFooter>
                </div>
            </DialogContent>
        </Dialog>
    )
}