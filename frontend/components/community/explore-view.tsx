"use client"

import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Search, Filter, UserPlus, MessageCircle, MapPin, Calendar } from "lucide-react"
import { useState } from "react"
import { UserDetailsModal } from "./user-details-modal"

const users = [
  {
    id: 1,
    name: "Sarah Johnson",
    username: "@sarah_j",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Junior",
    courses: ["CS 201", "MATH 301", "AI 401"],
    followers: 234,
    following: 189,
    posts: 45,
    isFollowing: false,
    location: "Austin, TX",
    joinedDate: "2023-08-15",
  },
  {
    id: 2,
    name: "Michael Chen",
    username: "@mike_chen",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Senior",
    courses: ["ENG 301", "PHY 201", "ROBO 401"],
    followers: 156,
    following: 203,
    posts: 32,
    isFollowing: true,
    location: "Austin, TX",
    joinedDate: "2022-09-01",
  },
  {
    id: 3,
    name: "Emily Rodriguez",
    username: "@emily_r",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Sophomore",
    courses: ["CHEM 201", "BIO 301", "MATH 201"],
    followers: 89,
    following: 145,
    posts: 28,
    isFollowing: false,
    location: "Austin, TX",
    joinedDate: "2023-01-20",
  },
  {
    id: 4,
    name: "David Thompson",
    username: "@david_t",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Junior",
    courses: ["MATH 401", "STAT 301", "CS 201"],
    followers: 178,
    following: 167,
    posts: 52,
    isFollowing: false,
    location: "Austin, TX",
    joinedDate: "2022-08-30",
  },
  {
    id: 5,
    name: "Jessica Park",
    username: "@jess_park",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Senior",
    courses: ["BUS 401", "ECON 301", "MKT 201"],
    followers: 312,
    following: 234,
    posts: 67,
    isFollowing: true,
    location: "Austin, TX",
    joinedDate: "2021-09-15",
  },
  {
    id: 6,
    name: "Alex Rivera",
    username: "@alex_rivera",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Sophomore",
    courses: ["ART 201", "DES 301", "CS 101"],
    followers: 145,
    following: 198,
    posts: 89,
    isFollowing: false,
    location: "Austin, TX",
    joinedDate: "2023-02-10",
  },
]

const universities = ["All Universities", "Austin Community College", "UT Austin", "Texas State", "ACC Highland"]
const years = ["All Years", "Freshman", "Sophomore", "Junior", "Senior"]
const courses = ["All Courses", "CS 201", "MATH 201", "ENG 102", "PHY 151", "HIST 105", "CHEM 151"]

export function ExploreView() {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedUniversity, setSelectedUniversity] = useState("All Universities")
  const [selectedYear, setSelectedYear] = useState("All Years")
  const [selectedCourse, setSelectedCourse] = useState("All Courses")
  const [followingUsers, setFollowingUsers] = useState<number[]>([2, 5])
  const [selectedUser, setSelectedUser] = useState(null)
  const [showUserModal, setShowUserModal] = useState(false)

  const filteredUsers = users.filter((user) => {
    const matchesSearch =
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.username.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesUniversity = selectedUniversity === "All Universities" || user.university === selectedUniversity
    const matchesYear = selectedYear === "All Years" || user.year === selectedYear
    const matchesCourse = selectedCourse === "All Courses" || user.courses.includes(selectedCourse)

    return matchesSearch && matchesUniversity && matchesYear && matchesCourse
  })

  const handleFollowToggle = (userId: number) => {
    setFollowingUsers((prev) => (prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId]))
  }

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search students by name or username..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-gray-800/50 border-gray-600"
                />
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-4">
              <div className="flex items-center space-x-2">
                <Filter className="h-4 w-4 text-gray-400" />
                <span className="text-sm text-gray-400">University:</span>
                <Select value={selectedUniversity} onValueChange={setSelectedUniversity}>
                  <SelectTrigger className="w-48 bg-gray-800/50 border-gray-600">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-gray-800 border-gray-600">
                    {universities.map((university) => (
                      <SelectItem key={university} value={university}>
                        {university}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-400">Year:</span>
                <Select value={selectedYear} onValueChange={setSelectedYear}>
                  <SelectTrigger className="w-32 bg-gray-800/50 border-gray-600">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-gray-800 border-gray-600">
                    {years.map((year) => (
                      <SelectItem key={year} value={year}>
                        {year}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-400">Course:</span>
                <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                  <SelectTrigger className="w-32 bg-gray-800/50 border-gray-600">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-gray-800 border-gray-600">
                    {courses.map((course) => (
                      <SelectItem key={course} value={course}>
                        {course}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Users Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredUsers.map((user) => (
          <Card
            key={user.id}
            className="glass border-0 hover:scale-105 transition-all duration-300 group cursor-pointer"
            onClick={() => {
              setSelectedUser(user)
              setShowUserModal(true)
            }}
          >
            <CardHeader className="pb-3">
              <div className="flex items-start space-x-4">
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
                      {user.year}
                    </Badge>
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center space-x-4 text-xs text-gray-500">
                <div className="flex items-center space-x-1">
                  <MapPin className="h-3 w-3" />
                  <span>{user.location}</span>
                </div>
                <div className="flex items-center space-x-1">
                  <Calendar className="h-3 w-3" />
                  <span>Joined {new Date(user.joinedDate).getFullYear()}</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs text-gray-400">
                <span>{user.followers} followers</span>
                <span>{user.following} following</span>
                <span>{user.posts} posts</span>
              </div>

              <div className="flex flex-wrap gap-1">
                {user.courses.slice(0, 3).map((course) => (
                  <Badge key={course} variant="secondary" className="text-xs">
                    {course}
                  </Badge>
                ))}
                {user.courses.length > 3 && (
                  <Badge variant="secondary" className="text-xs">
                    +{user.courses.length - 3} more
                  </Badge>
                )}
              </div>

              <div className="flex space-x-2 pt-2">
                <Button
                  variant={followingUsers.includes(user.id) ? "secondary" : "default"}
                  size="sm"
                  className="flex-1"
                  onClick={(e) => {
                    e.stopPropagation()
                    handleFollowToggle(user.id)
                  }}
                >
                  <UserPlus className="h-3 w-3 mr-1" />
                  {followingUsers.includes(user.id) ? "Following" : "Follow"}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="border-gray-600 bg-transparent"
                  onClick={(e) => e.stopPropagation()}
                >
                  <MessageCircle className="h-3 w-3" />
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredUsers.length === 0 && (
        <div className="text-center py-12">
          <Search className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-white mb-2">No users found</h3>
          <p className="text-gray-400">Try adjusting your search or filter criteria.</p>
        </div>
      )}
      <UserDetailsModal isOpen={showUserModal} onClose={() => setShowUserModal(false)} user={selectedUser} />
    </div>
  )
}
