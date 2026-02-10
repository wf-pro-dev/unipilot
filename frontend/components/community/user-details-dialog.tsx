"use client"

import { Dialog, DialogContent } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Separator } from "@/components/ui/separator"
import {
  MessageCircle,
  UserPlus,
  UserMinus,
  Mail,
  Calendar,
  GraduationCap,
  BookOpen,
  Users,
  Clock,
  Globe
} from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { format } from "date-fns"
import {
  useAcceptFriendRequest,
  useCancelFriendRequest,
  useFriendShipStatus,
  useRemoveFriend,
  useSendFriendRequest
} from "@/hooks/use-friends"
import { cn } from "@/lib/utils"
import { useMemo } from "react"
import { useRCourses } from "@/hooks/use-courses"
import { CourseItem } from "../courses/course-item"
import { FriendList } from "./friend-list"
import { UserItem } from "./user-item"

interface UserDetailsDialogProps {
  isOpen: boolean
  onClose: () => void
  user: models.User
}

export function UserDetailsDialog({ isOpen, onClose, user }: UserDetailsDialogProps) {

  if (!isOpen || !user) return null

  const { data: friendShipStatus, isLoading: friendShipStatusLoading } = useFriendShipStatus(user?.ID)
  const { mutate: sendFriendRequest } = useSendFriendRequest(user.ID)
  const { mutate: cancelFriendRequest } = useCancelFriendRequest(user.ID)
  const { mutate: acceptFriendRequest } = useAcceptFriendRequest(user.ID)
  const { mutate: removeFriend } = useRemoveFriend(user.ID)
  const { data: courses, isLoading: coursesLoading } = useRCourses(user.ID)


  // Get friendship action - matching AssignmentDetailsDialog button logic
  const friendshipAction = useMemo(() => {
    if (friendShipStatus?.status === "accepted") {
      return {
        label: "Remove Friend",
        icon: UserMinus,
        onClick: () => removeFriend(),
        variant: "danger" as const
      }
    }

    if (friendShipStatus?.status === "pending") {
      if (friendShipStatus?.is_pending_for_you) {
        return {
          label: "Accept Request",
          icon: UserPlus,
          onClick: () => acceptFriendRequest(),
          variant: "outline" as const
        }
      }

      return {
        label: "Cancel Request",
        icon: UserMinus,
        onClick: () => cancelFriendRequest(),
        variant: "outline" as const
      }
    }

    return {
      label: "Add Friend",
      icon: UserPlus,
      onClick: () => sendFriendRequest(),
      variant: "outline" as const
    }
  }, [friendShipStatus])

  

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-4xl p-0 gap-0 max-h-[90vh] overflow-hidden">

        <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl pointer-events-none" />



        {/* Two-column layout */}
        <div className="flex flex-col md:flex-row h-full max-h-[90vh]">

          {/* Left Sidebar - Profile Card */}
          <div className="md:w-80 bg-white/5 border-r border-white/5 relative overflow-hidden">

            <div className="relative p-6 flex flex-col h-full">

              {/* Avatar */}
              <div className="flex flex-col items-center mb-6 mt-4">
                <Avatar className="w-32 h-32 ring-2 ring-white/10 shadow-2xl shadow-black/50 mb-4">
                  <AvatarImage src={user.Avatar} alt={user.Username} />
                  <AvatarFallback className="bg-white/10 border border-white/15 text-white text-4xl font-bold">
                    {user.Username
                      .split(" ")
                      .map((n: string) => n[0])
                      .join("")
                      .slice(0, 2)
                      .toUpperCase()}
                  </AvatarFallback>
                </Avatar>

                <h2 className="text-h3 text-text-title text-center mb-1">
                  {user.Username}
                </h2>

                <p className="text-caption text-text-caption text-center mb-4">
                  {user.Email}
                </p>

                {/* Quick stats */}
                <div className="flex gap-3 mb-6">
                  <div className="flex flex-col items-center px-4 py-2 rounded-lg bg-white/5 border border-white/5">
                    <span className="text-h5 text-white font-bold">
                      {friendShipStatus?.friends_count || 0}
                    </span>
                    <span className="text-caption text-text-caption uppercase tracking-wider">
                      Friends
                    </span>
                  </div>
                  <div className="flex flex-col items-center px-4 py-2 rounded-lg bg-white/5 border border-white/5">
                    <span className="text-h5 text-white font-bold">
                      {friendShipStatus?.courses_count || 0}
                    </span>
                    <span className="text-caption text-text-caption uppercase tracking-wider">
                      Courses
                    </span>
                  </div>
                </div>
              </div>

              <Separator className="bg-white/10 mb-6" />

              {/* Key info - matching AssignmentDetailsDialog info section styling */}
              <div className="space-y-4 flex-1">

                <div className="flex items-start gap-3">
                  <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                    <GraduationCap className="w-4 h-4 text-text-caption" />
                  </div>
                  <div className="flex-1">
                    <p className="text-caption text-text-caption uppercase tracking-wider mb-1">
                      University
                    </p>
                    <p className="text-body text-white font-medium">
                      {user.University}
                    </p>
                  </div>
                </div>

                {user.Semester && user.Year && (
                  <div className="flex items-start gap-3">
                    <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                      <Calendar className="w-4 h-4 text-text-caption" />
                    </div>
                    <div className="flex-1">
                      <p className="text-caption text-text-caption uppercase tracking-wider mb-1">
                        Academic Info
                      </p>
                      <p className="text-body text-white font-medium">
                        {user.Semester} • {user.Year}
                      </p>
                    </div>
                  </div>
                )}

                <div className="flex items-start gap-3">
                  <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                    <Clock className="w-4 h-4 text-text-caption" />
                  </div>
                  <div className="flex-1">
                    <p className="text-caption text-text-caption uppercase tracking-wider mb-1">
                      Joined
                    </p>
                    <p className="text-body text-white font-medium">
                      {format(new Date(user.CreatedAt), "MMMM d, yyyy")}
                    </p>
                  </div>
                </div>

                {user.Language && (
                  <div className="flex items-start gap-3">
                    <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                      <Globe className="w-4 h-4 text-text-caption" />
                    </div>
                    <div className="flex-1">
                      <p className="text-caption text-text-caption uppercase tracking-wider mb-1">
                        Language
                      </p>
                      <p className="text-body text-white font-medium uppercase">
                        {user.Language}
                      </p>
                    </div>
                  </div>
                )}
              </div>

              {/* Actions at bottom - matching AssignmentDetailsDialog action buttons */}
              <div className="grid grid-cols-3 gap-2 mt-6">
                <Button
                  variant={friendshipAction.variant}
                  size="sm"
                  className="rounded-full col-span-2"
                  onClick={friendshipAction.onClick}
                >
                  <friendshipAction.icon className="w-4 h-4" />
                  <span>{friendshipAction.label}</span>
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-full"
                >
                  <MessageCircle className="w-4 h-4" />
                </Button>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="rounded-full w-full mt-2"
              >
                <Mail className="w-4 h-4" />
                <span>Email</span>
              </Button>
            </div>
          </div>

          {/* Right Content Area */}
          <div className="flex-1 overflow-y-auto">

            <div className="p-6 space-y-6">

              {/* Activity Section */}
              <div>
                <div className="flex items-center space-x-2 text-body font-medium text-text-caption uppercase tracking-wider mb-3">
                  <Users className="w-4 h-4" />
                  <span>Social Activity</span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">

                  {/* Friendship status */}
                  <div className="p-4 rounded-xl bg-white/5 border border-white/5">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 rounded-lg bg-white/5">
                        <Users className="w-4 h-4 text-text-caption" />
                      </div>
                      <div>
                        <p className="text-caption text-text-caption uppercase tracking-wider">
                          Status
                        </p>
                        <p className="text-body font-bold text-white">
                          {friendShipStatus?.status === "accepted"
                            ? "Friends"
                            : friendShipStatus?.status === "pending"
                              ? "Pending"
                              : "Not Connected"}
                        </p>
                      </div>
                    </div>
                    {friendShipStatus?.status === "accepted" && (
                      <p className="text-caption text-text-caption">
                        Connected since {format(new Date(user.CreatedAt), "MMM d, yyyy")}
                      </p>
                    )}
                  </div>

                  {/* Total friends */}
                  <div className="p-4 rounded-xl bg-white/5 border border-white/5">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 rounded-lg bg-white/5">
                        <Users className="w-4 h-4 text-text-caption" />
                      </div>
                      <div>
                        <p className="text-caption text-text-caption uppercase tracking-wider">
                          Friends
                        </p>
                        <p className="text-body font-bold text-white">
                          {friendShipStatus?.friends_count || 0}
                        </p>
                      </div>
                    </div>
                    <p className="text-caption text-text-caption">
                      Network connections
                    </p>
                  </div>

                  {/* Mutual friends (placeholder) */}
                  <div className="p-4 rounded-xl bg-white/5 border border-white/5">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 rounded-lg bg-white/5">
                        <Users className="w-4 h-4 text-text-caption" />
                      </div>
                      <div>
                        <p className="text-caption text-text-caption uppercase tracking-wider">
                          Mutual
                        </p>
                        <p className="text-body font-bold text-white">
                          0
                        </p>
                      </div>
                    </div>
                    <p className="text-caption text-text-caption">
                      Friends in common
                    </p>
                  </div>

                  {/* Last activity */}
                  <div className="p-4 rounded-xl bg-white/5 border border-white/5">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 rounded-lg bg-white/5">
                        <Clock className="w-4 h-4 text-text-caption" />
                      </div>
                      <div>
                        <p className="text-caption text-text-caption uppercase tracking-wider">
                          Last Sync
                        </p>
                        <p className="text-body font-bold text-white">
                          {user.LastSync
                            ? format(new Date(user.LastSync), "MMM d, h:mm a")
                            : "Never"}
                        </p>
                      </div>
                    </div>
                    <p className="text-caption text-text-caption">
                      Last data update
                    </p>
                  </div>

                </div>
              </div>

              <Separator className="bg-white/10" />


              {/* Courses Section */}
              <div>
                <div className="flex items-center space-x-2 text-body font-medium text-text-caption uppercase tracking-wider mb-3">
                  <BookOpen className="w-4 h-4" />
                  <span>Current Courses</span>
                </div>

                {courses && courses.length > 0 ? (
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    {courses.map((course: models.Course) => (
                      <CourseItem key={course.ID} courseId={course.ID} courseRO={course} size="compact" mode="readonly" />
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 px-4 rounded-xl bg-white/5 border border-white/5">
                    <BookOpen className="w-12 h-12 text-text-muted mx-auto mb-3" />
                    <p className="text-body text-text-caption">
                      No courses enrolled yet
                    </p>
                  </div>
                )}
              </div>

              <Separator className="bg-white/10" />

              {/* Friends Section */}
              <div>
                <div className="flex items-center space-x-2 text-body font-medium text-text-caption uppercase tracking-wider mb-3">
                  <Users className="w-4 h-4" />
                  <span>Friends</span>
                </div>


                <FriendList
                  userID={user.ID}
                  numColumns={1}
                  itemsPerPage={6}
                  containerClassName="gap-4"
                  renderItem={(user: models.User) => <UserItem user={user} size="compact" />}
                />
              </div>


            </div>
          </div>

        </div>
      </DialogContent>
    </Dialog>
  )
}