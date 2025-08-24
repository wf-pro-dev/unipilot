import { Card, CardHeader, CardContent } from "@/components/ui/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { User as UserIcon, Calendar, MessageCircle, UserPlus, UserMinus } from "lucide-react"
import { useFollow, useFollowers, useFollowing } from "@/hooks/use-follows"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { useUsers } from "@/hooks/use-users"
import { useAuthContext } from "../provider/auth-provider"
import { Skeleton } from "../ui/skeleton"

interface UserItemProps {
  userID: number
}

export function UserItem({ userID }: UserItemProps) {


  const { data: followers, isLoading: followersLoading } = useFollowers(userID)
  const { data: following, isLoading: followingLoading } = useFollowing(userID)
 
  const { data: users, isLoading: userLoading } = useUsers()
  const user = users?.find((user) => user.ID === userID)
  const { user: currentUser } = useAuthContext()
  
  const { data: currentUserFollowers } = useFollowers(currentUser?.ID as number)
  const { data: currentUserFollowing } = useFollowing(currentUser?.ID as number)

  // Check if current user is following this user by checking if current user is in the followers list
  const isFollowed = currentUserFollowing?.some((following) => following.ID === userID) 
  // Check if this user is following the current user by checking if the current user is in the following list
  const isFollowing = currentUserFollowers?.some((follower) => follower.ID === userID) 

  const followMutation = useFollow(currentUser!, isFollowed!)


  const handleFollow = () => {
    const message = "following " + user?.Username
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    followMutation.mutate(user!)
  }

  return (
    <Card
      key={userID}
      className="border-0 transition-all duration-300 cursor-pointer glass group"
    >
      <CardHeader className="pb-5">
        <div className="flex items-start space-x-4">
          <Avatar className="w-12 h-12">
            <AvatarImage src={user?.Avatar || "/placeholder.svg"} alt={user?.Username} />
            <AvatarFallback>
              {user?.Username
                .split(" ")
                .map((n: string) => n[0])
                .join("")}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-white truncate">{user?.Username}</h3>
            <p className="text-sm text-blue-400">{user?.Email}</p>
            <div className="flex items-center mt-1 space-x-2">
              <Badge variant="outline" className="text-xs border-gray-600">
                {user?.University}
              </Badge>

            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-5">

        <div className="flex items-center gap-2.5">
          <div className="flex items-center space-x-1 text-xs text-gray-400">
            <UserIcon className="w-3 h-3" />
            {followersLoading ? (
              <Skeleton className="w-16 h-4" />
            ) : (
              <span>{followers?.length} followers</span>
            )}
          </div>

          <div className="flex items-center space-x-1 text-xs text-gray-400">
            <UserIcon className="w-3 h-3" />
            {followingLoading ? (
              <Skeleton className="w-16 h-4" />
            ) : (
              <span>{following?.length} following</span>
            )}
          </div>

          <div className="flex items-center space-x-1 text-xs text-gray-400">
            <Calendar className="w-3 h-3" />
            <span>Joined {new Date(user?.CreatedAt).toLocaleDateString()}</span>
          </div>
        </div>


            
        <div className="flex gap-1 items-center text-xs text-gray-400">

          {user?.CoursesCode && user?.CoursesCode?.length > 0 && (
            user?.CoursesCode?.slice(0, 2).map((course: string) => (
              <Badge key={course} variant="outline" className="text-xs border-gray-600">
                {course}
              </Badge>
            ))

          )}

          {user?.CoursesCode && user?.CoursesCode?.length > 2 && (
            <Badge variant="outline" className="text-xs border-gray-600">
              +{user?.CoursesCode?.length - 2} more
            </Badge>
          )}

          {(!user?.CoursesCode || user?.CoursesCode?.length === 0) && (
            <Badge variant="outline" className="text-xs border-gray-600">
              No courses
            </Badge>
          )}


        </div>


        <div className="flex pt-2 space-x-2">
          {isFollowed! ? (
            <Button
              variant="outline"
              size="sm"
              className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10"
              onClick={(e) => {
                e.stopPropagation()
                handleFollow()
              }}
            >
              <UserMinus className="mr-1 w-3 h-3" />
              Unfollow
            </Button>
          ) : isFollowing! ? (
            <Button
              variant="outline"
              size="sm"
              className="flex-1 text-blue-400 bg-transparent border-blue-600 hover:bg-blue-600/10"
              onClick={(e) => {
                e.stopPropagation()
                handleFollow()
              }}>
              Follow Back
            </Button>

          ) : (

            <Button
              variant="outline"
              size="sm"
              className="flex-1 bg-transparent border-gray-600"
              onClick={(e) => {
                e.stopPropagation()
                handleFollow()
              }}
            >
              <UserPlus className="mr-1 w-3 h-3" />
              Follow
            </Button>
          )}



          <Button
            variant="outline"
            size="sm"
            className="bg-transparent border-gray-600"
            onClick={(e) => e.stopPropagation()}
          >
            <MessageCircle className="w-3 h-3" />
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

