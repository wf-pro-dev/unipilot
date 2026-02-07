import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MessageCircle, UserPlus, UserMinus } from "lucide-react"
import { useUser } from "@/hooks/use-users"
import { Skeleton } from "../ui/skeleton"

import { models } from "@/wailsjs/go/models"
import { UserDetailsModal } from "./user-details-modal"
import { useState } from "react"
import { useAcceptFriendRequest, useCancelFriendRequest, useFriendShipStatus, useRemoveFriend, useSendFriendRequest } from "@/hooks/use-friends"

interface UserItemProps {
  userID: string
  user?: models.User
}

export function UserItem({ userID }: UserItemProps) {

  const [isDetailsOpen, setIsDetailsOpen] = useState(false)

  const { data: user } = useUser(userID)
  const { data: friendShipStatus, isLoading: friendShipStatusLoading } = useFriendShipStatus(userID)
  const { mutate: sendFriendRequest } = useSendFriendRequest(userID)
  const { mutate: cancelFriendRequest } = useCancelFriendRequest(userID)
  const { mutate: acceptFriendRequest } = useAcceptFriendRequest(userID)
  const { mutate: removeFriend } = useRemoveFriend(userID)



  const handleOpenDetails = () => {
    setIsDetailsOpen(true)
  }

  return (
    <GlassCard
      key={userID}
      variant="outline"
      className="cursor-pointer group"
      onClick={handleOpenDetails}
    >
      <CardContent className="p-5">
        <div className="flex items-start space-x-4 mb-6">
          <Avatar className="w-14 h-14 shadow-xl flex-shrink-0 shadow-black/80">
            <AvatarImage src={user?.Avatar || "/placeholder.svg"} alt={user?.Username} />
            <AvatarFallback className="bg-gradient-to-br from-blue-600 to-purple-600 text-white font-bold text-lg">
              {user?.Username
                .split(" ")
                .map((n: string) => n[0])
                .join("")}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0 space-y-1">
            <h3 className="text-base font-bold text-white truncate tracking-tight">{user?.Username}</h3>
            <p className="text-xs text-blue-400 font-medium truncate">{user?.Email}</p>
            <Badge variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-400 uppercase tracking-wider px-2 py-0.5">
              {user?.University}
            </Badge>

          </div>
        </div>

        <div className="space-y-5">

          <div className="grid grid-cols-2 gap-4">

            <div className="flex flex-col items-center justify-center p-2 rounded-xl bg-white/5 border border-white/5">
              {friendShipStatusLoading ? (
                <Skeleton className="w-8 h-5 bg-white/10" />
              ) : (
                <span className="text-white font-bold text-lg">{friendShipStatus?.friends_count}</span>
              )}
              <span className="text-[10px] uppercase tracking-wider text-gray-500 font-medium mt-0.5">Following</span>
            </div>

            <div className="flex flex-col items-center justify-center p-2 rounded-xl bg-white/5 border border-white/5">
              <span className="text-white font-bold text-lg">{user?.CoursesCode?.length || 0}</span>
              <span className="text-[10px] uppercase tracking-wider text-gray-500 font-medium mt-0.5">Courses</span>
            </div>
          </div>

          <div className="flex flex-wrap gap-1.5 items-center">
            {user?.CoursesCode && user?.CoursesCode?.length > 0 ? (
              <>
                {user?.CoursesCode?.slice(0, 3).map((course: string) => (
                  <Badge key={course} variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-300 font-medium px-2 py-0.5 hover:bg-white/10 transition-colors cursor-default">
                    {course}
                  </Badge>
                ))}
                {user?.CoursesCode?.length > 3 && (
                  <Badge variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-400 px-2 py-0.5 hover:bg-white/10 transition-colors cursor-default">
                    +{user?.CoursesCode?.length - 3}
                  </Badge>
                )}
              </>
            ) : (
              <span className="text-[10px] text-gray-500 italic">No courses enrolled</span>
            )}
          </div>

          <div className="flex gap-2">
            {friendShipStatus?.status === "accepted" ? (
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-red-400 bg-red-500/5 border-red-500/20 hover:bg-red-500/10 hover:text-red-300 hover:border-red-500/30 transition-all h-9 text-xs font-medium"
                onClick={(e) => {
                  e.stopPropagation()
                  removeFriend()
                }}
              >
                <UserMinus className="mr-1.5 w-3.5 h-3.5" />
                Remove Friend
              </Button>
            ) : (
              friendShipStatus?.status === "pending" ? (

                friendShipStatus?.is_pending_for_you ? (
                  <Button
                    variant="default"
                    size="sm"
                    className="flex-1 bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 transition-all h-9 text-xs font-medium"
                    onClick={(e) => {
                      e.stopPropagation()
                      acceptFriendRequest()
                    }}>
                    Accept Request
                  </Button>

                ) : (

                  <Button
                    variant="danger"
                    size="sm"
                    className="rounded-md flex-1 font-normal"
                    onClick={(e) => {
                      e.stopPropagation()
                      cancelFriendRequest()
                    }}
                  >
                    <UserPlus className="mr-1.5 w-3.5 h-3.5" />
                    Cancel Request
                  </Button>
                )

              ) : (
                <Button
                  variant="default"
                  size="sm"
                  className="flex-1 bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 transition-all h-9 text-xs font-medium"
                  onClick={(e) => {
                    e.stopPropagation()
                    sendFriendRequest()
                  }}

                >
                  <UserPlus className="mr-1.5 w-3.5 h-3.5" />
                  Add Friend
                </Button>
              )
            )}

            <Button
              variant="outline"
              size="icon"
              className="bg-white/5 border-white/10 hover:bg-white/10 hover:text-white w-9 h-9 rounded-lg transition-all flex-shrink-0"
              onClick={(e) => e.stopPropagation()}
            >
              <MessageCircle className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </CardContent>
      <UserDetailsModal
        isOpen={isDetailsOpen}
        onClose={() => setIsDetailsOpen(false)}
        user={user!}
      />
    </GlassCard>
  )
}
