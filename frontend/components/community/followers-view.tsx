"use client"

import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Search, UserMinus, MessageCircle, MapPin, Calendar, Users } from "lucide-react"
import { useState } from "react"
import { UserDetailsModal } from "./user-details-modal"

const followers = [
  {
    id: 13,
    name: "Alice Cooper",
    username: "@alice_c",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Sophomore",
    courses: ["CS 101", "MATH 201", "ENG 102"],
    followers: 89,
    following: 134,
    posts: 23,
    location: "Austin, TX",
    joinedDate: "2023-03-15",
    followedBack: true,
  },
  {
    id: 14,
    name: "Bob Wilson",
    username: "@bob_w",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Junior",
    courses: ["PHY 201", "MATH 301", "CS 201"],
    followers: 156,
    following: 178,
    posts: 45,
    location: "Austin, TX",
    joinedDate: "2022-09-20",
    followedBack: false,
  },
  {
    id: 15,
    name: "Carol Davis",
    username: "@carol_d",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Senior",
    courses: ["BIO 301", "CHEM 201", "MATH 201"],
    followers: 234,
    following: 189,
    posts: 67,
    location: "Austin, TX",
    joinedDate: "2021-08-10",
    followedBack: true,
  },
]

const following = [
  {
    id: 16,
    name: "Daniel Kim",
    username: "@daniel_k",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Junior",
    courses: ["AI 401", "CS 301", "MATH 401"],
    followers: 298,
    following: 156,
    posts: 78,
    location: "Austin, TX",
    joinedDate: "2022-08-25",
    followsBack: true,
  },
  {
    id: 17,
    name: "Eva Martinez",
    username: "@eva_m",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Senior",
    courses: ["ENG 301", "HIST 201", "PHIL 101"],
    followers: 189,
    following: 234,
    posts: 52,
    location: "Austin, TX",
    joinedDate: "2021-09-05",
    followsBack: false,
  },
  {
    id: 18,
    name: "Frank Thompson",
    username: "@frank_t",
    avatar: "/placeholder.svg?height=40&width=40",
    university: "Austin Community College",
    year: "Sophomore",
    courses: ["CHEM 151", "BIO 201", "MATH 201"],
    followers: 145,
    following: 167,
    posts: 34,
    location: "Austin, TX",
    joinedDate: "2023-01-12",
    followsBack: true,
  },
]

export function FollowersView() {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedUser, setSelectedUser] = useState(null)
  const [showUserModal, setShowUserModal] = useState(false)
  const [followingList, setFollowingList] = useState<number[]>([16, 17, 18])

  const filteredFollowers = followers.filter(
    (user) =>
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.username.toLowerCase().includes(searchQuery.toLowerCase()),
  )

  const filteredFollowing = following.filter(
    (user) =>
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.username.toLowerCase().includes(searchQuery.toLowerCase()),
  )

  const handleUnfollow = (userId: number) => {
    setFollowingList((prev) => prev.filter((id) => id !== userId))
  }

  const UserCard = ({ user, isFollowing = false }: { user: any; isFollowing?: boolean }) => (
    <Card
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
                .map((n: string) => n[0])
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
              {isFollowing && user.followsBack && (
                <Badge className="bg-green-500/20 text-green-400 border-green-500/30 text-xs">Follows back</Badge>
              )}
              {!isFollowing && user.followedBack && (
                <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30 text-xs">Following you</Badge>
              )}
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
          {user.courses.slice(0, 3).map((course: string) => (
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
          {isFollowing ? (
            <Button
              variant="outline"
              size="sm"
              className="flex-1 border-red-600 text-red-400 hover:bg-red-600/10 bg-transparent"
              onClick={(e) => {
                e.stopPropagation()
                handleUnfollow(user.id)
              }}
            >
              <UserMinus className="h-3 w-3 mr-1" />
              Unfollow
            </Button>
          ) : (
            <Button variant="outline" size="sm" className="flex-1 border-gray-600 bg-transparent">
              Follow Back
            </Button>
          )}
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
  )

  return (
    <div className="space-y-6">
      {/* Search */}
      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search followers and following..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 bg-gray-800/50 border-gray-600 focus:border-blue-400"
            />
          </div>
        </CardContent>
      </Card>

      {/* Tabs */}
      <Tabs defaultValue="followers" className="w-full">
        <TabsList className="glass border-0 mb-6">
          <TabsTrigger value="followers" className="flex items-center space-x-2">
            <Users className="h-4 w-4" />
            <span>Followers ({followers.length})</span>
          </TabsTrigger>
          <TabsTrigger value="following" className="flex items-center space-x-2">
            <Users className="h-4 w-4" />
            <span>Following ({following.length})</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="followers">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredFollowers.map((user) => (
              <UserCard key={user.id} user={user} />
            ))}
          </div>
          {filteredFollowers.length === 0 && (
            <div className="text-center py-12">
              <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-white mb-2">No followers found</h3>
              <p className="text-gray-400">No followers match your search criteria.</p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="following">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredFollowing.map((user) => (
              <UserCard key={user.id} user={user} isFollowing={true} />
            ))}
          </div>
          {filteredFollowing.length === 0 && (
            <div className="text-center py-12">
              <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-white mb-2">No following found</h3>
              <p className="text-gray-400">No following match your search criteria.</p>
            </div>
          )}
        </TabsContent>
      </Tabs>

      <UserDetailsModal isOpen={showUserModal} onClose={() => setShowUserModal(false)} user={selectedUser} />
    </div>
  )
}
