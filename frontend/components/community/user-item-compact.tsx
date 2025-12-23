"use client"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { GlassCard } from "@/components/ui/glass-card"
import { user as User } from "@/wailsjs/go/models"

interface UserItemCompactProps {
  user: User.User
}

export function UserItemCompact({ user }: UserItemCompactProps) {
  
  return (
    <GlassCard className="p-3 flex items-center gap-3 border-white/5 bg-white/5 hover:bg-white/10 transition-all cursor-pointer group">
      <Avatar className="h-10 w-10 border border-white/10">
        <AvatarImage src={user.Avatar || "/placeholder.svg"} />
        <AvatarFallback className="bg-blue-600 text-white text-xs">
          {user.Username?.substring(0, 2).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      
      <div className="flex-1 min-w-0">
        <h4 className="text-sm font-medium text-white truncate leading-none mb-1">
          {user.Username}
        </h4>
        <p className="text-xs text-blue-400 truncate">
          {user.University || "Unknown University"}
        </p>
      </div>
    </GlassCard>
  )
}

