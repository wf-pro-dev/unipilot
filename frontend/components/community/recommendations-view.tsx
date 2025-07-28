"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { UserDetailsModal } from "./user-details-modal"
import { BookOpen, MapPin, Calendar, UserPlus, MessageCircle, GraduationCap } from "lucide-react"
import { useState } from "react"

// Mock current user's courses
const currentUserCourses = ["CS 201", "MATH 301", "AI 401", "ENG 102"]

// Mock recommended users data
const recommendedUsers = [
  {
    id: 1,
    name: "Sarah Johnson",
    username: "@sarah_j",
    avatar: "/placeholder.svg?height=40&width=40",
    year: "Junior",
    location: "Austin, TX",
    joinedDate: "2022-09-01",
    courses: ["CS 201", "MATH 301", "PHY 151", "ENG 102"],
    sharedCourses: ["CS 201", "MATH 301", "ENG 102"],
    matchReason: "3 shared courses",
  },
  {
    id: 2,
    name: "Michael Chen",
    username: "@mike_chen",
    avatar: "/placeholder.svg?height=40&width=40",
    year: "Senior",
    location: "Round Rock, TX",
    joinedDate: "2021-08-15",
    courses: ["AI 401", "CS 301", "MATH 301"],
    sharedCourses: ["AI 401", "MATH 301"],
    matchReason: "2 shared courses",
  },
  {
    id: 3,
    name: "Emily Rodriguez",
    username: "@emily_r",
    avatar: "/placeholder.svg?height=40&width=40",
    year: "Sophomore",
    location: "Cedar Park, TX",
    joinedDate: "2023-01-10",
    courses: ["CS 201", "ENG 102", "HIST 101"],
    sharedCourses: ["CS 201", "ENG 102"],
    matchReason: "2 shared courses",
  },
  {
    id: 4,
    name: "David Kim",
    username: "@david_k",
    avatar: "/placeholder.svg?height=40&width=40",
    year: "Junior",
    location: "Austin, TX",
    joinedDate: "2022-08-20",
    courses: ["AI 401", "CS 301", "PHIL 201"],
    sharedCourses: ["AI 401"],
    matchReason: "1 shared course",
  },
  {
    id: 5,
    name: "Jessica Taylor",
    username: "@jess_t",
    avatar: "/placeholder.svg?height=40&width=40",
    year: "Senior",
    location: "Pflugerville, TX",
    joinedDate: "2021-08-25",
    courses: ["MATH 301", "STAT 301", "ECON 201"],
    sharedCourses: ["MATH 301"],
    matchReason: "1 shared course",
  },
]

export function RecommendationsView() {
  const [selectedUser, setSelectedUser] = useState<(typeof recommendedUsers)[0] | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)

  const handleUserClick = (user: (typeof recommendedUsers)[0]) => {
    setSelectedUser(user)
    setIsModalOpen(true)
  }

  const handleFollow = (userId: number) => {
    console.log("Following user:", userId)
  }

  const handleMessage = (userId: number) => {
    console.log("Messaging user:", userId)
  }

  return (
    <div className="space-y-6">
      {/* Current User's Courses */}
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <BookOpen className="h-5 w-5 text-blue-400" />
            <span>Your Current Courses</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {currentUserCourses.map((course) => (
              <Badge key={course} className="bg-blue-500/20 text-blue-400 border-blue-500/30">
                {course}
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Recommended Users */}
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <UserPlus className="h-5 w-5 text-green-400" />
            <span>Recommended Classmates</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {recommendedUsers.map((user) => (
              <Card
                key={user.id}
                className="glass-dark border-0 hover:glass-hover transition-all duration-300 cursor-pointer"
                onClick={() => handleUserClick(user)}
              >
                <CardContent className="p-4">
                  <div className="flex items-start space-x-3 mb-3">
                    <Avatar className="h-12 w-12">
                      <AvatarImage src={user.avatar || "/placeholder.svg"} alt={user.name} />
                      <AvatarFallback>
                        {user.name
                          .split(" ")
                          .map((n) => n[0])
                          .join("")}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-white truncate">{user.name}</h3>
                      <p className="text-sm text-blue-400">{user.username}</p>
                      <div className="flex items-center space-x-2 mt-1">
                        <Badge variant="outline" className="border-gray-600 text-xs">
                          <GraduationCap className="h-3 w-3 mr-1" />
                          {user.year}
                        </Badge>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-2 mb-3">
                    <div className="flex items-center space-x-2 text-sm">
                      <MapPin className="h-3 w-3 text-red-400" />
                      <span className="text-gray-300">{user.location}</span>
                    </div>
                    <div className="flex items-center space-x-2 text-sm">
                      <Calendar className="h-3 w-3 text-purple-400" />
                      <span className="text-gray-300">Joined {new Date(user.joinedDate).toLocaleDateString()}</span>
                    </div>
                  </div>

                  {/* Match Reason */}
                  <div className="mb-3">
                    <Badge className="bg-green-500/20 text-green-400 border-green-500/30 text-xs">
                      {user.matchReason}
                    </Badge>
                  </div>

                  {/* Shared Courses */}
                  <div className="mb-4">
                    <p className="text-xs text-gray-400 mb-2">Shared Courses:</p>
                    <div className="flex flex-wrap gap-1">
                      {user.sharedCourses.map((course) => (
                        <Badge key={course} variant="outline" className="border-green-500/30 text-green-400 text-xs">
                          {course}
                        </Badge>
                      ))}
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex space-x-2">
                    <Button
                      size="sm"
                      className="flex-1"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleFollow(user.id)
                      }}
                    >
                      <UserPlus className="h-3 w-3 mr-1" />
                      Follow
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      className="flex-1 border-gray-600 bg-transparent"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleMessage(user.id)
                      }}
                    >
                      <MessageCircle className="h-3 w-3 mr-1" />
                      Message
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>

      <UserDetailsModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} user={selectedUser} />
    </div>
  )
}
