"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Search, Filter, X } from "lucide-react"
import { useState } from "react"
import { UserItem } from "./user-item"
import { Input } from "../ui/input"
import { useNetworkStatus } from "@/hooks/use-network-status"
import { OfflineBanner } from "../ui/offline-banner"
import { user } from "@/wailsjs/go/models"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { useAuthContext } from "../provider/auth-provider"
import { GlassCard } from "../ui/glass-card"

interface ExploreViewProps {
  users: user.User[] | undefined
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

    return user.ID != currentUser?.ID && matchesSearch && matchesUniversity
  })

  const hasActiveFilters = searchTerm !== "" || selectedUniversity !== "All Universities"

  const clearFilters = () => {
    setSearchTerm("")
    setSelectedUniversity("All Universities")
  }

  if (!isOnline) {
    return <OfflineBanner />
  }

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
      <GlassCard className="border-white/5 bg-white/5 backdrop-blur-xl shadow-lg shadow-black/20">
        <CardContent className="p-5">
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

      {/* Users Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredUsers.map((user) => (
          <UserItem
            key={user.ID}
            userID={user.ID}
          />
        ))}
      </div>

      {
        filteredUsers.length === 0 && (
          <div className="text-center py-12">
            <Search className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-white mb-2">No users found</h3>
            <p className="text-gray-400">Try adjusting your search or filter criteria.</p>
          </div>
        )
      }
    </div >
  )
}
