import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MessageCircle, UserPlus, UserMinus } from "lucide-react"
import { useFollow, useFollowers, useFollowing } from "@/hooks/use-follows"
import { useUsers } from "@/hooks/use-users"
import { useAuthContext } from "../provider/auth-provider"
import { Skeleton } from "../ui/skeleton"
import { toast } from "sonner"

import { models} from "@/wailsjs/go/models"

interface UserItemProps {
  userID: number
  user?: models.User
}

export function UserItem({ userID, user: userProp }: UserItemProps) {


  const { data: followers, isLoading: followersLoading } = useFollowers(userID)
  const { data: following, isLoading: followingLoading } = useFollowing(userID)

  const { data: users, isLoading: userLoading } = useUsers()
  const user = userProp || users?.find((user) => user.ID === userID)
  const { user: currentUser, followers: currentUserFollowers, following: currentUserFollowing } = useAuthContext()


  // Check if current user is following this user by checking if current user is in the followers list
  const isFollowed = currentUserFollowing?.some((following) => following.ID === userID)
  // Check if this user is following the current user by checking if the current user is in the following list
  const isFollowing = currentUserFollowers?.some((follower) => follower.ID === userID)

  const followMutation = useFollow(currentUser!, isFollowed!)


  const handleFollow = () => {
    followMutation.mutate(user!, {
      onSuccess: () => {
        if (isFollowed) {
          toast.success("You just unfollowed " + user?.Username)
        } else {
          toast.success("You are now following " + user?.Username)
        }
      },
      onError: () => {
        if (isFollowed) {
          toast.error("Failed to unfollow " + user?.Username)
        } else {
          toast.error("Failed to follow " + user?.Username)
        }
      }
    })
  }

  return (
    <GlassCard
      key={userID}
      variant="outline"
      className="cursor-pointer group"
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

          <div className="grid grid-cols-3 gap-3">
            <div className="flex flex-col items-center justify-center p-2 rounded-xl bg-white/5 border border-white/5">
              {followersLoading ? (
                <Skeleton className="w-8 h-5 bg-white/10" />
              ) : (
                <span className="text-white font-bold text-lg">{followers?.length}</span>
              )}
              <span className="text-[10px] uppercase tracking-wider text-gray-500 font-medium mt-0.5">Followers</span>
            </div>

            <div className="flex flex-col items-center justify-center p-2 rounded-xl bg-white/5 border border-white/5">
              {followingLoading ? (
                <Skeleton className="w-8 h-5 bg-white/10" />
              ) : (
                <span className="text-white font-bold text-lg">{following?.length}</span>
              )}
              <span className="text-[10px] uppercase tracking-wider text-gray-500 font-medium mt-0.5">Following</span>
            </div>

            <div className="flex flex-col items-center justify-center p-2 rounded-xl bg-white/5 border border-white/5">
              <span className="text-white font-bold text-lg">{users?.find(u => u.ID === userID)?.CoursesCode?.length || 0}</span>
              <span className="text-[10px] uppercase tracking-wider text-gray-500 font-medium mt-0.5">Courses</span>
            </div>
          </div>

          <div className="flex flex-wrap gap-1.5 items-center min-h-[24px]">
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
              <span className="text-[10px] text-gray-500 italic pl-1">No courses enrolled</span>
            )}
          </div>

          <div className="flex gap-2 pt-1">
            {isFollowed! ? (
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-red-400 bg-red-500/5 border-red-500/20 hover:bg-red-500/10 hover:text-red-300 hover:border-red-500/30 transition-all h-9 text-xs font-medium"
                onClick={(e) => {
                  e.stopPropagation()
                  handleFollow()
                }}
              >
                <UserMinus className="mr-1.5 w-3.5 h-3.5" />
                Unfollow
              </Button>
            ) : isFollowing! ? (
              <Button
                variant="default"
                size="sm"
                className="flex-1 bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 transition-all h-9 text-xs font-medium"
                onClick={(e) => {
                  e.stopPropagation()
                  handleFollow()
                }}>
                Follow Back
              </Button>

            ) : (

              <Button
                variant="default"
                size="sm"
                className="flex-1 bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 transition-all h-9 text-xs font-medium"
                onClick={(e) => {
                  e.stopPropagation()
                  handleFollow()
                }}
              >
                <UserPlus className="mr-1.5 w-3.5 h-3.5" />
                Follow
              </Button>
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
    </GlassCard>
  )
}
