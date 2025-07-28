"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { User, MapPin, Calendar, BookOpen, Users, MessageCircle, UserPlus, GraduationCap, Mail } from "lucide-react"

interface UserDetailsModalProps {
  isOpen: boolean
  onClose: () => void
  user: any | null
}

export function UserDetailsModal({ isOpen, onClose, user }: UserDetailsModalProps) {
  if (!user) return null

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    })
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-0 text-white max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <User className="h-5 w-5 text-blue-400" />
            <span>User Profile</span>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6 mt-6">
          {/* User Header */}
          <div className="flex items-start space-x-6">
            <Avatar className="h-20 w-20">
              <AvatarImage src={user.avatar || "/placeholder.svg"} alt={user.name} />
              <AvatarFallback className="text-lg">
                {user.name
                  .split(" ")
                  .map((n: string) => n[0])
                  .join("")}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1">
              <h2 className="text-2xl font-bold text-white">{user.name}</h2>
              <p className="text-lg text-blue-400 mb-2">{user.username}</p>
              <div className="flex items-center space-x-3">
                <Badge variant="outline" className="border-gray-600">
                  <GraduationCap className="h-3 w-3 mr-1" />
                  {user.year}
                </Badge>
                <Badge variant="outline" className="border-gray-600">
                  {user.university}
                </Badge>
              </div>
            </div>
          </div>

          <Separator className="bg-gray-700" />

          {/* User Stats */}
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <div className="text-2xl font-bold text-blue-400">{user.followers}</div>
              <div className="text-sm text-gray-400">Followers</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-green-400">{user.following}</div>
              <div className="text-sm text-gray-400">Following</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-purple-400">{user.posts}</div>
              <div className="text-sm text-gray-400">Posts</div>
            </div>
          </div>

          <Separator className="bg-gray-700" />

          {/* User Info */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="text-sm font-medium text-gray-400 block mb-2">Location</label>
              <div className="flex items-center space-x-2 text-white">
                <MapPin className="h-4 w-4 text-green-400" />
                <span>{user.location}</span>
              </div>
            </div>

            <div>
              <label className="text-sm font-medium text-gray-400 block mb-2">Joined</label>
              <div className="flex items-center space-x-2 text-white">
                <Calendar className="h-4 w-4 text-blue-400" />
                <span>{formatDate(user.joinedDate)}</span>
              </div>
            </div>
          </div>

          {/* Courses */}
          <div>
            <label className="text-sm font-medium text-gray-400 block mb-3">Current Courses</label>
            <div className="flex flex-wrap gap-2">
              {user.courses.map((course: string) => (
                <Badge key={course} className="bg-blue-500/20 text-blue-400 border-blue-500/30">
                  <BookOpen className="h-3 w-3 mr-1" />
                  {course}
                </Badge>
              ))}
            </div>
          </div>

          {/* Mutual Followers (if available) */}
          {user.mutualFollowers && (
            <div>
              <label className="text-sm font-medium text-gray-400 block mb-2">Mutual Connections</label>
              <div className="flex items-center space-x-2 text-white">
                <Users className="h-4 w-4 text-purple-400" />
                <span>{user.mutualFollowers} mutual followers</span>
              </div>
            </div>
          )}

          <Separator className="bg-gray-700" />

          {/* Actions */}
          <div className="flex justify-between">
            <div className="flex space-x-2">
              <Button className="bg-blue-600 hover:bg-blue-700">
                <UserPlus className="h-4 w-4 mr-2" />
                Follow
              </Button>
              <Button variant="outline" className="border-gray-600 bg-transparent">
                <MessageCircle className="h-4 w-4 mr-2" />
                Message
              </Button>
            </div>
            <Button variant="outline" className="border-gray-600 bg-transparent">
              <Mail className="h-4 w-4 mr-2" />
              Send Email
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
