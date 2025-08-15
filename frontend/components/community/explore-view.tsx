"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Search, Filter } from "lucide-react"
import { useState } from "react"
import { UserDetailsModal } from "./user-details-modal"
import { useUsers } from "@/hooks/use-users"
import { UserItem } from "./user-item"
import { user } from "@/wailsjs/go/models"
import { Input } from "../ui/input"



export function ExploreView() {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedUniversity, setSelectedUniversity] = useState("All Universities")
  const [followingUsers, setFollowingUsers] = useState<number[]>([2, 5])
  const [selectedUser, setSelectedUser] = useState<user.User | null>(null)
  const [showUserModal, setShowUserModal] = useState(false)

  const { data: users } = useUsers()

  const handleFollow = (userId: number) => {
    // This is now handled by the UserItem component
    console.log("Follow handled by UserItem")
  }


  const universities = Array.from(new Set(users?.map((user) => user.University) || [])).filter((university) => university !== "")

  const filteredUsers = (users || []).filter((user) => {
    const matchesSearch =
      user.Username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.Email.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesUniversity = selectedUniversity === "All Universities" || user.University === selectedUniversity

    return matchesSearch && matchesUniversity
  })

  const handleFollowToggle = (userId: number) => {
    setFollowingUsers((prev) => (prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId]))
  }

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
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

      {/* Users Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredUsers.map((user) => (
          <UserItem
            key={user.ID}
            user={user as any}
            setSelectedUser={setSelectedUser}
            setShowUserModal={setShowUserModal}
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
      <UserDetailsModal isOpen={showUserModal} onClose={() => setShowUserModal(false)} user={selectedUser} />
    </div >
  )
}
