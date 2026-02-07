"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { UserDetailsModal } from "./user-details-dialog"
import { BookOpen, UserPlus } from "lucide-react"
import { useState } from "react"
import { useNetworkStatus } from "@/hooks/use-network-status"
import { UserItem } from "./user-item"
import { OfflineBanner } from "../ui/offline-banner"
import { models } from "@/wailsjs/go/models"


export function RecommendationsView() {
  const [selectedUser, setSelectedUser] = useState<models.User | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)

  const {isOnline} = useNetworkStatus()
  
  const users: models.User[] = []
  const currentUser = users.find(u => u.ID === user?.ID)
  const currentUserCourses = currentUser?.CoursesCode || []
  const recommendedUsers: models.User[] = []


  if (!isOnline) {
    return <OfflineBanner />
  }

  return (
    <div className="space-y-6">
      {/* Current User's Courses */}
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <BookOpen className="h-5 w-5 text-blue-400" />
            <span>Your Current Courses</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {currentUserCourses.map((course) => (
              <Badge key={course} className="bg-blue-500/20 text-blue-400 border-blue-500/30">
                {course}
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Recommended Users */}
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2 text-white">
            <UserPlus className="h-5 w-5 text-green-400" />
            <span>Recommended Classmates</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {recommendedUsers?.map((user) => (
              <UserItem key={user.ID} userID={user.ID} />
            ))}
          </div>
        </CardContent>
      </Card>

      <UserDetailsModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} user={selectedUser} />
    </div>
  )
}
