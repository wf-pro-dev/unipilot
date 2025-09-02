"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Search, Users, Filter  } from "lucide-react"
import { useState } from "react"
import { UserItem } from "./user-item"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useNetworkStatus } from "@/hooks/use-network-status"
import { OfflineBanner } from "../ui/offline-banner"
import { user } from "@/wailsjs/go/models"

interface FollowingViewProps {
  following: user.User[] | undefined
}

export function FollowingView({ following }: FollowingViewProps) {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedUniversity, setSelectedUniversity] = useState("All Universities")

  const {isOnline} = useNetworkStatus()

  const universities = Array.from(new Set(following?.map((user) => user.University) || [])).filter((university) => university !== "")

  const filteredFollowing = following?.filter((user) => {
    const matchesSearch =
      user.Username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.Email.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesUniversity = selectedUniversity === "All Universities" || user.University === selectedUniversity

    return matchesSearch && matchesUniversity
  })

  if (!isOnline) {
    return <OfflineBanner />
  }

  return (
    <div className="space-y-6">
      {/* Search */}
      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search students by name or username..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-gray-800/50 border-gray-600"
                />
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-4">
              <div className="flex items-center space-x-2">
                <Filter className="h-4 w-4 text-gray-400" />
                <span className="text-sm text-gray-400">University:</span>
                <Select value={selectedUniversity} onValueChange={setSelectedUniversity}>
                  <SelectTrigger className="w-48 bg-gray-800/50 border-gray-600">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-gray-800 border-gray-600">
                    <SelectItem value="All Universities">All Universities</SelectItem>
                    {universities?.map((university) => (
                      <SelectItem key={university} value={university}>
                        {university}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>


      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredFollowing?.map((user) => (
          <UserItem key={user.ID} userID={user.ID}/>
        ))}
      </div>
      {filteredFollowing?.length === 0 && (
        <div className="text-center py-12">
          <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-white mb-2">No following found</h3>
          <p className="text-gray-400">No following match your search criteria.</p>
        </div>
      )}
    </div>
  )
}
