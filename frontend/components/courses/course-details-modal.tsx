"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Progress } from "@/components/ui/progress"
import {
  BookOpen,
  Users,
  Calendar,
  GraduationCap,
  MapPin,
  Edit,
  Trash2,
  Plus,
  TrendingUp,
  FileText,
} from "lucide-react"
import Link from "next/link"
import { course as Course } from "@/wailsjs/go/models"
import { useAssignments, useUpdateAssignment } from "@/hooks/use-assignments"
import { formatDeadline } from "@/lib/date-utils"
import { assignment } from "@/wailsjs/go/models"
import { StatusTag } from "@/components/assignments/utils/status-tag"
import { CourseEditDialog } from "./course-edit-dialog"
import { useState } from "react"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"

interface CourseDetailsModalProps {
  isOpen: boolean
  onClose: () => void
  courseId: number | null
  courses: Course.LocalCourse[]
  onEdit: (course: Course.LocalCourse, column: string, value: string) => void
  onDelete: (course: Course.LocalCourse) => void
}

export function CourseDetailsModal({ isOpen, onClose, courseId, courses, onEdit, onDelete }: CourseDetailsModalProps) {
  const course = courses.find(c => c.ID === courseId) || null

  if (!course) return null

  const { data: assignments, isLoading } = useAssignments()
  const updateMutation = useUpdateAssignment()

  var course_assignments = (assignments || []).filter((assignment: assignment.LocalAssignment) => assignment.Course?.Code === course.Code) || []
  var completed_assignments_count = course_assignments.filter((assignment: assignment.LocalAssignment) => assignment.StatusName === "Done").length
  var completionPercentage = (completed_assignments_count / course_assignments.length) * 100
  var isCompleted = completionPercentage === 100
  const [open, setOpen] = useState(false)

  const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
    console.log("Editing assignment:", assignment)
    const message = "assignment " + assignment.ID + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      assignment,
      column,
      value
    })
  }
  // Mock additional data
  const courseData = {
    ...course,
    location: "Science Building, Room 204",
    email: "instructor@university.edu",
    office: "Faculty Building, Room 301",
    officeHours: "MW 2:00-4:00 PM",
  }

  const handleCreateNote = () => {
    // Navigate to notes creation page with course pre-selected
    window.open(`/notes?course=${encodeURIComponent(courseData.Code)}`, "_blank")
  }

  return (
    <div>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="glass border-0 text-white max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center space-x-2">

            </DialogTitle>
          </DialogHeader>

          <div className="mt-6 space-y-6">
            {/* Course Header */}
            <div className="flex justify-between items-start">
              <div className="flex items-center space-x-4">

                <div className={`w-6 h-6 rounded-full ${courseData.Color}`} />
                <div className="space-y-3">
                  <h2 className="text-2xl font-bold text-white">{courseData.Code}</h2>
                  <p className="text-lg text-gray-300">{courseData.Name}</p>
                </div>
              </div>
              <div className="flex flex-col items-end space-y-4">
                <div className="flex items-center space-x-2">
                  <Button variant="outline" size="sm" className="bg-transparent border-gray-600" onClick={() => setOpen(true)}>
                    <Edit className="mr-1 w-3 h-3" />
                    Edit
                  </Button>
                  <Button variant="destructive" size="sm" onClick={() => onDelete(course)}>
                    <Trash2 className="mr-1 w-3 h-3" />
                    Delete
                  </Button>
                </div>

                <Badge variant="outline" className="border-gray-600">
                  <GraduationCap className="mr-1 w-3 h-3" />
                  {courseData.Semester}
                </Badge>

              </div>
            </div>

            <Separator className="bg-gray-700" />

            {/* Course Info Grid */}
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              {/* Left Column */}
              <div className="space-y-4">
                <div>
                  <label className="block mb-2 text-sm font-medium text-gray-400">Instructor</label>
                  <div className="flex items-center space-x-2 text-white">
                    <Users className="w-4 h-4 text-blue-400" />
                    <span>{courseData.Instructor}</span>
                  </div>
                </div>

                <div>
                  <label className="block mb-2 text-sm font-medium text-gray-400">Schedule</label>
                  <div className="flex items-center space-x-2 text-white">
                    <Calendar className="w-4 h-4 text-purple-400" />
                    <span>{courseData.Schedule}</span>
                  </div>
                </div>


              </div>

              {/* Right Column */}
              <div className="space-y-4">
                <div>
                  <label className="block mb-2 text-sm font-medium text-gray-400">Email</label>
                  <div className="text-blue-400 cursor-pointer hover:text-blue-300">{courseData.InstructorEmail}</div>
                </div>

                <div>
                  <label className="block mb-2 text-sm font-medium text-gray-400">Location</label>
                  <div className="flex items-center space-x-2 text-white">
                    <MapPin className="w-4 h-4 text-orange-400" />
                    <span>{courseData.RoomNumber}</span>
                  </div>
                </div>

              </div>
            </div>

            {/* Credits & Course Dates */}

            <div className="grid grid-cols-1 gap-6 md:grid-cols-3">

              <div>
                <label className="block mb-2 text-sm font-medium text-gray-400">Credits</label>
                <Badge variant="outline" className="border-gray-600">
                  {courseData.Credits} credits
                </Badge>
              </div>

              <div>
                <label className="block mb-2 text-sm font-medium text-gray-400">Start Date</label>
                <p className="text-white">{format(courseData.StartDate, "MMMM d, yyyy")}</p>
              </div>

              <div>
                <label className="block mb-2 text-sm font-medium text-gray-400">End Date</label>
                <p className="text-white">{format(courseData.EndDate, "MMMM d, yyyy")}</p>
              </div>

            </div>

            <Separator className="bg-gray-700" />

            {/* Assignment Progress */}
            <div>
              <label className="block mb-4 text-sm font-medium text-gray-400">Assignment Progress</label>
              <div className="space-y-3">
                <div className="flex justify-between items-center">
                  <div className="flex items-center space-x-2">
                    <TrendingUp className="w-4 h-4 text-green-400" />
                    <span className={`${isCompleted ? "text-green-400" : "text-white"}`}>
                      {completed_assignments_count} of {course_assignments.length} assignments completed
                    </span>
                  </div>
                  <span className="text-sm text-gray-400">{Math.round(completionPercentage)}%</span>
                </div>
                <Progress color={isCompleted ? "green" : "white"} value={completionPercentage} className="h-2" />
              </div>
            </div>

            {/* Recent Assignments */}
            <div>
              <div className="flex justify-between items-center">
                <label className="block mb-3 text-sm font-medium text-gray-400">Recent Assignments</label>
                <Link href={`/assignments?view=list&course=${courseData.Code}`}>
                  <Button variant="ghost" size="sm" className="text-blue-400 hover:text-blue-300">
                    View All
                  </Button>
                </Link>
              </div>
              <div className="grid gap-2 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-2">
                {course_assignments.slice(0, 4).map((assignment, index) => (
                  <div key={index} className="flex items-center p-3 rounded-lg border border-gray-600 bg-gray-800/50">

                    <span className="w-2/3 text-sm text-white line-clamp-2">{assignment.Title}</span>
                    <div className="flex flex-col items-end space-y-2 grow">
                      <span className="text-xs text-gray-400">{formatDeadline(assignment.Deadline)}</span>
                      <StatusTag assignment={assignment} onEdit={handleEditAssignment} />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <Separator className="bg-gray-700" />

            {/* Actions */}
            <div className="flex space-x-2">
              <Button variant="outline" className="bg-transparent border-gray-600">
                <Plus className="mr-2 w-4 h-4" />
                Add Assignment
              </Button>
              <Button variant="outline" className="bg-transparent border-gray-600" onClick={handleCreateNote}>
                <FileText className="mr-2 w-4 h-4" />
                Create Note
              </Button>
              <Button variant="outline" className="bg-transparent border-gray-600">
                <Users className="w-4 h-4" />
                Add Student
              </Button>
            </div>
          </div>
        </DialogContent>

      </Dialog>
      <CourseEditDialog
        open={open}
        setOpen={setOpen}
        course={course}
        onEdit={onEdit}
      />
    </div>
  )
}
