import { CardContent } from "@/components/ui/card"
import { GlassCard, GlassCardVariants } from "@/components/ui/glass-card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MessageCircle, UserPlus, UserMinus, Users } from "lucide-react"
import { Skeleton } from "../ui/skeleton"

import { models } from "@/wailsjs/go/models"
import { memo, useCallback, useMemo } from "react"
import {
  useAcceptFriendRequest,
  useCancelFriendRequest,
  useFriendShipStatus,
  useRemoveFriend,
  useSendFriendRequest
} from "@/hooks/use-friends"
import { cn } from "@/lib/utils"
import { LucideIcon } from "lucide-react"
import { useDialogContext } from "../provider/dialog-provider"

interface UserItemProps {
  user: models.User
  size?: "default" | "compact"
  variant?: GlassCardVariants
}

interface FriendshipActionProps {
  label: string
  variant: "default" | "outline"
  Icon: LucideIcon
  onClick: (e: React.MouseEvent) => void
  className: string
}

const BaseUserItem = ({
  user,
  size = "default",
  variant = "outline"
}: UserItemProps) => {

  const { SetDialogState } = useDialogContext()
  const { data: friendShipStatus, isLoading: friendShipStatusLoading } = useFriendShipStatus(user.ID)
  const { mutate: sendFriendRequest } = useSendFriendRequest(user.ID)
  const { mutate: cancelFriendRequest } = useCancelFriendRequest(user.ID)
  const { mutate: acceptFriendRequest } = useAcceptFriendRequest(user.ID)
  const { mutate: removeFriend } = useRemoveFriend(user.ID)

  const handleOpenDetails = useCallback(() => {
    SetDialogState({ 
      modelType: "user", 
      dialogType: "details", 
      id: user.ID, 
      open: true, 
      item: user,
      viewMode: "readonly"
    })
  }, [])

  const handleMessage = useCallback((e: React.MouseEvent) => {
    e.stopPropagation()
    // Message functionality
  }, [])

  // Get friendship action state
  const getFriendshipAction: FriendshipActionProps = useMemo(() => {
    if (friendShipStatus?.status === "accepted") {
      return {
        label: "Friends",
        variant: "outline" as const,
        Icon: Users,
        onClick: (e: React.MouseEvent) => {
          e.stopPropagation()
          removeFriend()
        },
        className: "text-emerald-400 border-emerald-500/20 bg-emerald-500/5 hover:bg-emerald-500/10"
      }
    }

    if (friendShipStatus?.status === "pending") {
      if (friendShipStatus?.is_pending_for_you) {
        return {
          label: "Accept",
          variant: "default" as const,
          Icon: UserPlus,
          onClick: (e: React.MouseEvent) => {
            e.stopPropagation()
            acceptFriendRequest()
          },
          className: "bg-blue-600 hover:bg-blue-500 text-white"
        }
      }

      return {
        label: "Pending",
        variant: "outline" as const,
        Icon: UserMinus,
        onClick: (e: React.MouseEvent) => {
          e.stopPropagation()
          cancelFriendRequest()
        },
        className: "text-gray-400 border-gray-500/20 bg-gray-500/5"
      }
    }

    return {
      label: "Add Friend",
      variant: "default" as const,
      Icon: UserPlus,
      onClick: (e: React.MouseEvent) => {
        e.stopPropagation()
        sendFriendRequest()
      },
      className: "bg-blue-600 hover:bg-blue-500 text-white"
    }
  }, [friendShipStatus])

  const FriendshipAction = memo(({
    variant,
    Icon,
    onClick,
    className,
    label
  }: FriendshipActionProps) => {
    return (
      <Button
        variant={variant}
        size="sm"
        className={cn("flex-1 text-xs font-medium h-9", className)}
        onClick={onClick}
      >
        <Icon className="w-3.5 h-3.5 mr-1.5" />
        {label}
      </Button>
    )
  })



  // Default size - Modern card with focus on avatar and key info
  const DefaultUserItem = () => {
    return (
      <GlassCard
        variant={variant}
        className="group cursor-pointer"
        onClick={handleOpenDetails}
      >
        <CardContent className="p-0">
          {/* Header Section with gradient background */}
          <div className="relative h-24 bg-gradient-to-br from-blue-600/20 via-purple-600/20 to-pink-600/20 overflow-hidden">
            {/* Decorative background pattern */}
            <div className="absolute inset-0 opacity-10">
              <div className="absolute top-0 left-0 w-32 h-32 bg-white rounded-full blur-3xl" />
              <div className="absolute bottom-0 right-0 w-32 h-32 bg-blue-500 rounded-full blur-3xl" />
            </div>

            {/* Stats overlay */}
            <div className="absolute top-3 right-3 flex gap-2">
              <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-black/40 backdrop-blur-md border border-white/10">
                <Users className="w-3 h-3 text-white/80" />
                {friendShipStatusLoading ? (
                  <Skeleton className="w-6 h-3 bg-white/20" />
                ) : (
                  <span className="text-xs font-semibold text-white">
                    {friendShipStatus?.friends_count || 0}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-black/40 backdrop-blur-md border border-white/10">
                <div className="w-2 h-2 rounded-full bg-blue-400" />
                <span className="text-xs font-semibold text-white">
                  {user.CoursesCode?.length || 0}
                </span>
              </div>
            </div>
          </div>

          {/* Avatar - overlapping header */}
          <div className="absolute px-5 -translate-y-1/2">
            <Avatar className="w-20 h-20 border-4 border-background shadow-2xl shadow-black/50 ring-2 ring-white/10">
              <AvatarImage src={user.Avatar} alt={user.Username} />
              <AvatarFallback className="bg-white/10 border border-white/15 text-black text-2xl font-bold">
                {user.Username
                  .split(" ")
                  .map((n: string) => n[0])
                  .join("")
                  .slice(0, 2)
                  .toUpperCase()}
              </AvatarFallback>
            </Avatar>
          </div>

          {/* Main content */}
          <div className="px-5 pb-5 pt-16 space-y-4">
            {/* User info */}
            <div className="space-y-1">
              <h3 className="text-lg font-bold text-white tracking-tight">
                {user.Username}
              </h3>
              <p className="text-sm text-gray-400">
                {user.Email}
              </p>
            </div>

            {/* University badge */}
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className="text-xs border-white/10 bg-white/5 text-gray-300 font-normal px-3 py-1"
              >
                {user.University}
              </Badge>
              {user.Semester && user.Year && (
                <span className="text-xs text-gray-500">
                  {user.Semester} • {user.Year}
                </span>
              )}
            </div>

            {/* Courses - Only show if exists */}
            {user.CoursesCode && user.CoursesCode.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {user.CoursesCode.slice(0, 3).map((course: string) => (
                  <span
                    key={course}
                    className="text-xs px-2 py-0.5 rounded-md bg-blue-500/10 text-blue-300 font-medium"
                  >
                    {course}
                  </span>
                ))}
                {user.CoursesCode.length > 3 && (
                  <span className="text-xs px-2 py-0.5 rounded-md bg-white/5 text-gray-400">
                    +{user.CoursesCode.length - 3}
                  </span>
                )}
              </div>
            )}

            {/* Actions */}
            <div className="flex gap-2 pt-2">

              <FriendshipAction {...getFriendshipAction} />

              <Button
                variant="outline"
                size="icon"
                className="h-9 w-9 bg-white/5 border-white/10 hover:bg-white/10"
                onClick={handleMessage}
              >
                <MessageCircle className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </CardContent>

       
      </GlassCard>
    )
  }

  // Compact size - Minimal horizontal layout
  const CompactUserItem = () => {
    return (
      <GlassCard
        variant={variant}
        className="group cursor-pointer hover:bg-white/10 transition-all"
        onClick={handleOpenDetails}
      >
        <CardContent className="p-4">
          <div className="flex items-center gap-4">
            {/* Avatar with status indicator */}
            <div className="relative flex-shrink-0">
              <Avatar className="w-12 h-12 ring-2 ring-offset-2 ring-offset-background ring-blue-500/20">
                <AvatarImage src={user.Avatar || "/placeholder.svg"} alt={user.Username} />
                <AvatarFallback className="bg-white/10 border border-white/15 text-black text-2xl font-bold">
                  {user.Username
                    .split(" ")
                    .map((n: string) => n[0])
                    .join("")
                    .slice(0, 2)
                    .toUpperCase()}
                </AvatarFallback>
              </Avatar>
              {/* Online/status indicator */}
              <div className="absolute -bottom-0.5 -right-0.5 w-4 h-4 bg-emerald-500 rounded-full border-2 border-background" />
            </div>

            {/* User info */}
            <div className="flex-1 min-w-0">
              <h4 className="text-sm font-semibold text-white truncate">
                {user.Username}
              </h4>
              <div className="flex items-center gap-2 mt-0.5">
                <span className="text-xs text-gray-400 truncate">
                  {user.University}
                </span>
                {user.CoursesCode && user.CoursesCode.length > 0 && (
                  <>
                    <span className="text-xs text-gray-600">•</span>
                    <span className="text-xs text-gray-500">
                      {user.CoursesCode.length} {user.CoursesCode.length === 1 ? 'course' : 'courses'}
                    </span>
                  </>
                )}
              </div>
            </div>

            {/* Action button */}
            <FriendshipAction {...getFriendshipAction} />
          </div>
        </CardContent>

       
      </GlassCard>
    )
  }

  return size === "compact" ? <CompactUserItem /> : <DefaultUserItem />
}

export const UserItem = memo(BaseUserItem, (prevProps, nextProps) => {
  return (
    prevProps.user.ID === nextProps.user.ID &&
    prevProps.size === nextProps.size &&
    prevProps.variant === nextProps.variant
  )
})