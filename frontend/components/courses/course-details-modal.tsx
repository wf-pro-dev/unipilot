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
  Clock,
  MapPin,
  Edit,
  Trash2,
  Plus,
  TrendingUp,
  FileText,
} from "lucide-react"
import Link from "next/link"
import { Course } from "@/types/models"
import { useAssignments } from "@/hooks/use-assignments"
import { formatDeadline, parseDeadline } from "@/lib/date-utils"

interface CourseDetailsModalProps {
  isOpen: boolean
  onClose: () => void
  course: Course | null
}

export function CourseDetailsModal({ isOpen, onClose, course }: CourseDetailsModalProps) {
  if (!course) return null
  const { assignments } = useAssignments()

  var  course_assignments = assignments.filter((assignment) => assignment.Course?.Code === course.Code)
  var completed_assignments_count = course_assignments.filter((assignment) => assignment.StatusName === "Done").length
  var completionPercentage = (completed_assignments_count / course_assignments.length) * 100

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
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-0 text-white max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <BookOpen className="h-5 w-5 text-blue-400" />
            <span>Course Details</span>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6 mt-6">
          {/* Course Header */}
          <div className="flex items-start justify-between">
            <div className="flex items-center space-x-4">

              <div className={`w-6 h-6 rounded-full ${courseData.Color}`} />
              <div className="space-y-3">
                <h2 className="text-2xl font-bold text-white">{courseData.Code}</h2>
                <p className="text-lg text-gray-300">{courseData.Name}</p>
              </div>
            </div>
            <div className="flex flex-col items-end space-y-4">
              <div className="flex items-center space-x-2">
                <Button variant="outline" size="sm" className="border-gray-600 bg-transparent">
                  <Edit className="h-3 w-3 mr-1" />
                  Edit
                </Button>
                <Button variant="destructive" size="sm">
                  <Trash2 className="h-3 w-3 mr-1" />
                  Delete
                </Button>
              </div>

              <Badge variant="outline" className="border-gray-600">
                <GraduationCap className="h-3 w-3 mr-1" />
                {courseData.Semester}
              </Badge>

            </div>
          </div>

          <Separator className="bg-gray-700" />

          {/* Course Info Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Left Column */}
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Instructor</label>
                <div className="flex items-center space-x-2 text-white">
                  <Users className="h-4 w-4 text-blue-400" />
                  <span>{courseData.Instructor}</span>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Schedule</label>
                <div className="flex items-center space-x-2 text-white">
                  <Calendar className="h-4 w-4 text-purple-400" />
                  <span>{courseData.Schedule}</span>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Location</label>
                <div className="flex items-center space-x-2 text-white">
                  <MapPin className="h-4 w-4 text-green-400" />
                  <span>{courseData.location}</span>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Credits</label>
                <Badge variant="outline" className="border-gray-600">
                  {courseData.Credits} credits
                </Badge>
              </div>
            </div>

            {/* Right Column */}
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Email</label>
                <div className="text-blue-400 hover:text-blue-300 cursor-pointer">{courseData.InstructorEmail}</div>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Office</label>
                <div className="flex items-center space-x-2 text-white">
                  <MapPin className="h-4 w-4 text-orange-400" />
                  <span>{courseData.RoomNumber}</span>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-400 block mb-2">Office Hours</label>
                <div className="flex items-center space-x-2 text-white">
                  <Clock className="h-4 w-4 text-yellow-400" />
                  <span>{courseData.OfficeHours}</span>
                </div>
              </div>
            </div>
          </div>

          <Separator className="bg-gray-700" />

          {/* Assignment Progress */}
          <div>
            <label className="text-sm font-medium text-gray-400 block mb-4">Assignment Progress</label>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <TrendingUp className="h-4 w-4 text-green-400" />
                  <span className="text-white">
                    {completed_assignments_count} of {course_assignments.length} assignments completed
                  </span>
                </div>
                <span className="text-sm text-gray-400">{Math.round(completionPercentage)}%</span>
              </div>
              <Progress value={completionPercentage} className="h-2" />
            </div>
          </div>

          {/* Recent Assignmen  ts */}
          <div>
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-gray-400 block mb-3">Recent Assignments</label>
                <Link href={`/assignments?view=list&course=${courseData.Code}`}>
                <Button variant="ghost" size="sm" className="text-blue-400 hover:text-blue-300">
                  View All
                </Button>
              </Link>
            </div>
            <div className="space-y-2">
              {course_assignments.slice(0, 3).map((assignment, index) => (
                <div key={index} className="flex items-center justify-between p-3 rounded-lg glass-dark">
                  
                  <span className="text-white">{assignment.Title}</span>
                  <div className="flex items-center space-x-2">
                    <span className="text-xs text-gray-400">{formatDeadline(assignment.Deadline)}</span>

                    <Badge
                      variant={
                        assignment.StatusName === "Done"
                          ? "default"
                          : assignment.StatusName === "In Progress"
                            ? "secondary"
                            : "outline"
                      }
                      className="text-xs"
                    >
                      {assignment.StatusName}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <Separator className="bg-gray-700" />

          {/* Actions */}
          <div className="flex space-x-2">
            <Button variant="outline" className="border-gray-600 bg-transparent">
              <Plus className="h-4 w-4 mr-2" />
              Add Assignment
            </Button>
            <Button variant="outline" className="border-gray-600 bg-transparent" onClick={handleCreateNote}>
              <FileText className="h-4 w-4 mr-2" />
              Create Note
            </Button>
            <Button variant="outline" className="border-gray-600 bg-transparent">
              <Users className="h-4 w-4" />
              Add Student
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
