"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Search, Filter, X, Users } from "lucide-react"
import { useState } from "react"
import { UserItem } from "./user-item"
import { Input } from "../ui/input"
import { useNetworkStatus } from "@/hooks/use-network-status"
import { OfflineBanner } from "../ui/offline-banner"
import { models } from "@/wailsjs/go/models"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { useAuthContext } from "../provider/auth-provider"
import { GlassCard } from "../ui/glass-card"
import { EmptyState } from "../ui/empty-state"

interface ExploreViewProps {
  users: models.User[] | undefined
}

export function ExploreView({ users }: ExploreViewProps) {
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedUniversity, setSelectedUniversity] = useState("All Universities")
  const { user: currentUser } = useAuthContext()

  const { isOnline } = useNetworkStatus()


  const universities = Array.from(new Set(users?.map((user) => user.University) || [])).filter((university) => university !== "")

  const filteredUsers = (users || []).filter((user) => {

    const matchesSearch =
      user.Username.toLowerCase().includes(searchTerm.toLowerCase()) ||
      user.Email.toLowerCase().includes(searchTerm.toLowerCase())

    const matchesUniversity = selectedUniversity === "All Universities" || user.University === selectedUniversity

    return matchesSearch && matchesUniversity
  })

  

  const hasActiveFilters = searchTerm !== "" || selectedUniversity !== "All Universities"

  const clearFilters = () => {
    setSearchTerm("")
    setSelectedUniversity("All Universities")
  }

  if (!isOnline) {
    return <OfflineBanner />
  }

  // if there is only one user (current user), show a message that there are no users to explore
  if (users?.length === 0) {
    return (
      <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
        <EmptyState
          icon={Users}
          title="No users found"
          description="Wait for other users to join the platform."
          className="flex-1 items-center"
        />
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 space-y-6">
      {/* Search and Filters */}
      <GlassCard variant="board" className="flex-grow-0 flex-row">
        <CardContent className="flex-1 p-5">
          <div className="space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-4">
              
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search students by name or username..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-white/5 border-white/10 transition-all duration-300 h-10"
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
            {hasActiveFilters && (
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Filter className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-400">Active filters:</span>
                  {searchTerm && (
                    <Badge variant="secondary" className="bg-blue-500/20 text-blue-400">
                      Search: {searchTerm}
                    </Badge>
                  )}
                  {selectedUniversity !== "All Universities" && (
                    <Badge variant="secondary" className="bg-green-500/20 text-green-400">
                      {selectedUniversity}
                    </Badge>
                  )}

                </div>
                <Button variant="ghost" size="sm" onClick={clearFilters} className="text-gray-400 hover:text-white">
                  <X className="h-4 w-4 mr-1" />
                  Clear
                </Button>
              </div>
            )}
          </div>


        </CardContent>
      </GlassCard>



      {
        filteredUsers.length === 0 ? (
          <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
            <EmptyState
              icon={Search}
              title="No users found"
              description="Try adjusting your search or filter criteria."
              className="flex-1 items-center"
              onClick={clearFilters}
              buttonText="Clear Filters"
            />
          </div>
        ) : (
          // Users Grid
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredUsers.map((user) => (
              <UserItem
                key={user.ID}
                userID={user.ID}
              />
            ))}
          </div>
        )
      }
    </div >
  )
}
