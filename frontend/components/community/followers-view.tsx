"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Search, Users, Filter } from "lucide-react"
import { useState } from "react"
import { UserItem } from "./user-item"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useNetworkStatus } from "@/hooks/use-network-status"
import { OfflineBanner } from "../ui/offline-banner"
import { models } from "@/wailsjs/go/models"
import { GlassCard } from "../ui/glass-card"
import { EmptyState } from "../ui/empty-state"
import { useRouter } from "next/navigation"

interface FollowersViewProps {
  followers: models.User[] | undefined
}

export function FollowersView({ followers }: FollowersViewProps) {
  
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedUniversity, setSelectedUniversity] = useState("All Universities")

  const router = useRouter()
  const { isOnline } = useNetworkStatus()

  const universities = Array.from(new Set(followers?.map((user) => user.University) || [])).filter((university) => university !== "")


  const filteredFollowers = followers?.filter((user) => {
    const matchesSearch =
      user.Username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.Email.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesUniversity = selectedUniversity === "All Universities" || user.University === selectedUniversity

    return matchesSearch && matchesUniversity
  })

  const clearFilters = () => {
    setSearchQuery("")
    setSelectedUniversity("All Universities")
  }

  if (!isOnline) {
    return <OfflineBanner />
  }

  if (followers?.length === 0) {
    return (
      <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
        <EmptyState
          icon={Users}
          title="No followers found"
          description="Don't be shy! Follow someone"
          className="flex-1 items-center"
          onClick={() => router.push("/community?view=explore")}
          buttonText="Go to Explore"
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Search */}
      <GlassCard variant="board" className="flex-grow-0 flex-row">
        <CardContent className="flex-1 p-5">
          
          <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-4">
            
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder="Search students by name or username..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 bg-white/5 border-white/10    transition-all duration-300 h-10"
              />
            </div>



            <Select value={selectedUniversity} onValueChange={setSelectedUniversity}>
              <SelectTrigger className="w-60 bg-white/5 border-white/10 h-10">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                <SelectItem value="All Universities">All Universities</SelectItem>
                {universities?.map((university) => (
                  <SelectItem key={university} value={university}>
                    {university}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

          </div>
        </CardContent>
      </GlassCard>


      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredFollowers?.map((user) => (
          <UserItem key={user.ID} userID={user.ID} />
        ))}
      </div>
      {filteredFollowers?.length === 0 && (
        <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
          <EmptyState
            icon={Users}
            title="No followers found"
            description="Try adjusting your search or filter criteria."
            className="flex-1 items-center"
            onClick={clearFilters}
            buttonText="Clear Filters"
          />
        </div>

      )}


    </div>
  )
}
