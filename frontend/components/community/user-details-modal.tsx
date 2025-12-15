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
      <DialogContent className="glass border-white/10 text-white max-w-2xl p-0 overflow-hidden gap-0">
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <DialogTitle className="flex items-center space-x-2 text-xl font-semibold">
            <User className="h-5 w-5 text-blue-400" />
            <span>User Profile</span>
          </DialogTitle>
        </DialogHeader>

        <div className="p-6 space-y-6">
          {/* User Header */}
          <div className="flex items-start space-x-6">
            <Avatar className="h-24 w-24 ring-4 ring-white/5 shadow-2xl">
              <AvatarImage src={user.avatar || "/placeholder.svg"} alt={user.name} />
              <AvatarFallback className="text-2xl bg-gradient-to-br from-blue-500 to-purple-600 text-white">
                {user.name
                  .split(" ")
                  .map((n: string) => n[0])
                  .join("")}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-2 pt-1">
              <div>
                <h2 className="text-2xl font-bold text-white tracking-tight">{user.name}</h2>
                <p className="text-base text-blue-400 font-medium">@{user.username}</p>
              </div>
              <div className="flex items-center space-x-2">
                <Badge variant="outline" className="border-white/10 bg-white/5 text-gray-300">
                  <GraduationCap className="h-3 w-3 mr-1.5 text-purple-400" />
                  {user.year}
                </Badge>
                <Badge variant="outline" className="border-white/10 bg-white/5 text-gray-300">
                  {user.university}
                </Badge>
              </div>
            </div>
          </div>

          {/* User Stats */}
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center p-4 rounded-xl bg-white/5 border border-white/5">
              <div className="text-2xl font-bold text-blue-400 mb-1">{user.followers}</div>
              <div className="text-xs font-medium text-gray-400 uppercase tracking-wider">Followers</div>
            </div>
            <div className="text-center p-4 rounded-xl bg-white/5 border border-white/5">
              <div className="text-2xl font-bold text-green-400 mb-1">{user.following}</div>
              <div className="text-xs font-medium text-gray-400 uppercase tracking-wider">Following</div>
            </div>
            <div className="text-center p-4 rounded-xl bg-white/5 border border-white/5">
              <div className="text-2xl font-bold text-purple-400 mb-1">{user.posts}</div>
              <div className="text-xs font-medium text-gray-400 uppercase tracking-wider">Posts</div>
            </div>
          </div>

          {/* User Info */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 p-4 rounded-xl bg-white/5 border border-white/5">
            <div>
              <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-2">Location</label>
              <div className="flex items-center space-x-2 text-white">
                <MapPin className="h-4 w-4 text-green-400" />
                <span className="text-sm">{user.location}</span>
              </div>
            </div>

            <div>
              <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-2">Joined</label>
              <div className="flex items-center space-x-2 text-white">
                <Calendar className="h-4 w-4 text-blue-400" />
                <span className="text-sm">{formatDate(user.joinedDate)}</span>
              </div>
            </div>
          </div>

          {/* Courses */}
          <div>
            <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block mb-3">Current Courses</label>
            <div className="flex flex-wrap gap-2">
              {user.courses.map((course: string) => (
                <Badge key={course} className="bg-blue-500/10 text-blue-300 border-blue-500/20 px-3 py-1.5 hover:bg-blue-500/20 transition-colors">
                  <BookOpen className="h-3.5 w-3.5 mr-1.5" />
                  {course}
                </Badge>
              ))}
            </div>
          </div>

          {/* Mutual Followers (if available) */}
          {user.mutualFollowers && (
            <div className="flex items-center space-x-2 text-sm text-gray-400 bg-white/5 p-3 rounded-lg border border-white/5 w-fit">
              <Users className="h-4 w-4 text-purple-400" />
              <span><span className="text-white font-medium">{user.mutualFollowers}</span> mutual connections</span>
            </div>
          )}

          <div className="h-px bg-white/5 w-full" />

          {/* Actions */}
          <div className="flex justify-between gap-3">
            <div className="flex space-x-3 flex-1">
              <Button className="bg-blue-600 hover:bg-blue-500 text-white shadow-[0_0_15px_rgba(37,99,235,0.2)] border-0 flex-1">
                <UserPlus className="h-4 w-4 mr-2" />
                Follow
              </Button>
              <Button variant="outline" className="border-white/10 bg-white/5 hover:bg-white/10 text-white flex-1">
                <MessageCircle className="h-4 w-4 mr-2" />
                Message
              </Button>
            </div>
            <Button variant="outline" className="border-white/10 bg-white/5 hover:bg-white/10 text-white">
              <Mail className="h-4 w-4 mr-2" />
              Email
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
