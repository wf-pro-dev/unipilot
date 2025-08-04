"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Progress } from "@/components/ui/progress"
import {
  User,
  MapPin,
  Calendar,
  BookOpen,
  GraduationCap,
  Mail,
  Phone,
  Edit,
  Save,
  Camera,
  TrendingUp,
  FileText,
  Clock,
  Users,
} from "lucide-react"
import { useState } from "react"
import { useAuth } from "@/hooks/use-auth"
import { useAssignments, useCompletedAssignments } from "@/hooks/use-assignments"
import { useCourses } from "@/hooks/use-courses"
import { CourseItem } from "@/components/courses/course-item"
import { CourseDetailsModal } from "@/components/courses/course-details-modal"
import { Course } from "@/types/models"

// Mock user data
const userData = {
  id: 1,
  name: "John Doe",
  username: "@john_doe",
  email: "john.doe@student.acc.edu",
  phone: "+1 (512) 555-0123",
  avatar: "/placeholder.svg?height=120&width=120",
  university: "Austin Community College",
  year: "Junior",
  major: "Computer Science",
  courses: [
    {
      id: 1,
      code: "CS 201",
      name: "Data Structures",
      instructor: "Dr. Smith",
      credits: 3,
      color: "bg-blue-500",
      schedule: "MWF 10:00-11:00 AM",
      assignments: 8,
      completed: 6,
      semester: "Spring 2024",
    },
    {
      id: 2,
      code: "MATH 301",
      name: "Calculus III",
      instructor: "Prof. Johnson",
      credits: 4,
      color: "bg-green-500",
      schedule: "TTh 2:00-3:30 PM",
      assignments: 12,
      completed: 10,
      semester: "Spring 2024",
    },
    {
      id: 3,
      code: "AI 401",
      name: "Artificial Intelligence",
      instructor: "Dr. Williams",
      credits: 3,
      color: "bg-purple-500",
      schedule: "MWF 1:00-2:00 PM",
      assignments: 6,
      completed: 4,
      semester: "Spring 2024",
    },
    {
      id: 4,
      code: "ENG 102",
      name: "English Composition",
      instructor: "Prof. Davis",
      credits: 3,
      color: "bg-orange-500",
      schedule: "TTh 11:00-12:30 PM",
      assignments: 5,
      completed: 5,
      semester: "Spring 2024",
    },
  ],
  followers: 156,
  following: 203,
  posts: 42,
  location: "Austin, TX",
  joinedDate: "2022-08-15",
  stats: {
    assignmentsCompleted: 87,
    totalAssignments: 95,
    notesCreated: 23,
    studyHours: 156,
  },
}

export default function ProfilePage() {
  const [isEditing, setIsEditing] = useState(false)
  const [editedData, setEditedData] = useState(userData)
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  
  const { user } = useAuth()
  const { data: assignments } = useAssignments()
  const { data: completedAssignments } =  useCompletedAssignments()
  const { data: courses } = useCourses()

  const completionPercentage = ( (completedAssignments || []).length / (assignments || []).length) * 100

  const handleSave = () => {
    // Here you would typically save to a backend
    setIsEditing(false)
    console.log("Saving profile data:", editedData)
  }

  const handleCancel = () => {
    setEditedData(userData)
    setIsEditing(false)
  }

  const handleCourseClick = (courseId: number) => {
    setSelectedCourseId(courseId)
  }

  const handleEditCourse = (course: Course, column: string, value: string) => {
    // Handle course editing if needed
    console.log("Editing course:", course, column, value)
  }

  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="max-w-6xl mx-auto relative z-10">
        <div className="mb-8">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            My Profile
          </h1>
          <p className="text-gray-400 mt-2">Manage your account information and academic progress</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Profile Info */}
          <div className="lg:col-span-1 space-y-6">
            {/* Profile Card */}
            <Card className="glass border-0">
              <CardContent className="p-6">
                <div className="flex flex-col items-center space-y-4">
                  <div className="relative">
                    <Avatar className="h-24 w-24">
                      <AvatarImage src={userData.avatar || "/placeholder.svg"} alt={userData.name} />
                      <AvatarFallback className="text-lg">
                        {user?.username
                          .split(" ")
                          .map((n) => n[0])
                          .join("")}
                      </AvatarFallback>
                    </Avatar>
                    <Button
                      size="sm"
                      className="absolute -bottom-2 -right-2 h-8 w-8 rounded-full p-0"
                      disabled={!isEditing}
                    >
                      <Camera className="h-4 w-4" />
                    </Button>
                  </div>

                  {isEditing ? (
                    <div className="w-full space-y-3">
                      <Input
                        value={editedData.name}
                        onChange={(e) => setEditedData({ ...editedData, name: e.target.value })}
                        className="bg-gray-800/50 border-gray-600"
                      />
                      <Input
                        value={editedData.username}
                        onChange={(e) => setEditedData({ ...editedData, username: e.target.value })}
                        className="bg-gray-800/50 border-gray-600"
                      />
                    </div>
                  ) : (
                    <div className="text-center">
                      <h2 className="text-xl font-bold text-white">{user?.username}</h2>
                      <p className="text-blue-400">{userData?.email}</p>
                    </div>
                  )}

                  <div className="flex items-center space-x-2">
                    <Badge variant="outline" className="border-gray-600">
                      <GraduationCap className="h-3 w-3 mr-1" />
                      {userData.year}
                    </Badge>
                  </div>

                  <div className="flex justify-center space-x-2 w-full">
                    {isEditing ? (
                      <>
                        <Button onClick={handleSave} size="sm">
                          <Save className="h-4 w-4 mr-1" />
                          Save
                        </Button>
                        <Button onClick={handleCancel} variant="outline" size="sm">
                          Cancel
                        </Button>
                      </>
                    ) : (
                      <Button onClick={() => setIsEditing(true)} size="sm">
                        <Edit className="h-4 w-4 mr-1" />
                        Edit Profile
                      </Button>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Stats Card */}
            <Card className="glass border-0">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2 text-white">
                  <TrendingUp className="h-5 w-5 text-green-400" />
                  <span>Academic Stats</span>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">Assignments Completed</span>
                    <span className="text-sm text-white">{userData.stats.assignmentsCompleted}/{userData.stats.totalAssignments}</span>
                  </div>
                  <Progress value={(userData.stats.assignmentsCompleted / userData.stats.totalAssignments) * 100} className="h-2" />
                </div>

                <div className="grid grid-cols-2 gap-4 text-center">
                  <div>
                    <div className="text-2xl font-bold text-white">{userData.stats.notesCreated}</div>
                    <div className="text-sm text-gray-400">Notes Created</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-white">{userData.stats.studyHours}</div>
                    <div className="text-sm text-gray-400">Study Hours</div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Social Stats */}
            <Card className="glass border-0">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2 text-white">
                  <Users className="h-5 w-5 text-blue-400" />
                  <span>Social</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-3 gap-4 text-center">
                  <div>
                    <div className="text-2xl font-bold text-white">{userData.followers}</div>
                    <div className="text-sm text-gray-400">Followers</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-white">{userData.following}</div>
                    <div className="text-sm text-gray-400">Following</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-white">{userData.posts}</div>
                    <div className="text-sm text-gray-400">Posts</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Right Column - Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Personal Information */}
            <Card className="glass border-0">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2 text-white">
                  <User className="h-5 w-5 text-blue-400" />
                  <span>Personal Information</span>
                  {!isEditing && (
                    <Button onClick={() => setIsEditing(true)} size="sm" className="ml-auto">
                      <Edit className="h-4 w-4 mr-1" />
                      Edit
                    </Button>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-4">
                    <div>
                      <label className="text-sm font-medium text-gray-400 block mb-2">Email</label>
                      {isEditing ? (
                        <Input
                          value={editedData.email}
                          onChange={(e) => setEditedData({ ...editedData, email: e.target.value })}
                          className="bg-gray-800/50 border-gray-600"
                        />
                      ) : (
                        <div className="flex items-center space-x-2 text-white">
                          <Mail className="h-4 w-4 text-blue-400" />
                          <span>{userData.email}</span>
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="text-sm font-medium text-gray-400 block mb-2">Phone</label>
                      {isEditing ? (
                        <Input
                          value={editedData.phone}
                          onChange={(e) => setEditedData({ ...editedData, phone: e.target.value })}
                          className="bg-gray-800/50 border-gray-600"
                        />
                      ) : (
                        <div className="flex items-center space-x-2 text-white">
                          <Phone className="h-4 w-4 text-green-400" />
                          <span>{userData.phone}</span>
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="text-sm font-medium text-gray-400 block mb-2">Location</label>
                      {isEditing ? (
                        <Input
                          value={editedData.location}
                          onChange={(e) => setEditedData({ ...editedData, location: e.target.value })}
                          className="bg-gray-800/50 border-gray-600"
                        />
                      ) : (
                        <div className="flex items-center space-x-2 text-white">
                          <MapPin className="h-4 w-4 text-red-400" />
                          <span>{userData.location}</span>
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div>
                      <label className="text-sm font-medium text-gray-400 block mb-2">University</label>
                      <div className="text-white">{userData.university}</div>
                    </div>

                    <div>
                      <label className="text-sm font-medium text-gray-400 block mb-2">Major</label>
                      {isEditing ? (
                        <Input
                          value={editedData.major}
                          onChange={(e) => setEditedData({ ...editedData, major: e.target.value })}
                          className="bg-gray-800/50 border-gray-600"
                        />
                      ) : (
                        <div className="text-white">{userData.major}</div>
                      )}
                    </div>

                    <div>
                      <label className="text-sm font-medium text-gray-400 block mb-2">Joined</label>
                      <div className="flex items-center space-x-2 text-white">
                        <Calendar className="h-4 w-4 text-purple-400" />
                        <span>{new Date(userData.joinedDate).toLocaleDateString()}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Current Courses */}
            <Card className="glass border-0">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2 text-white">
                  <BookOpen className="h-5 w-5 text-green-400" />
                  <span>Current Courses</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {courses?.slice(0, 4).map((course) => {
                    return (
                     <CourseItem 
                        course={course} 
                        onEdit={() => {}} 
                        onDelete={() => {}} 
                        onToggleComplete={() => {}} 
                        onCourseClick={handleCourseClick} 
                     />
                    )
                  })}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
      <CourseDetailsModal
          isOpen={!!selectedCourseId}
          courseId={selectedCourseId}
          onClose={() => setSelectedCourseId(null)}
          onEdit={handleEditCourse}
        />
    </div>
  )
}
