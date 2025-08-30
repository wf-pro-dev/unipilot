"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ExploreView } from "@/components/community/explore-view"
import { FollowersView } from "@/components/community/followers-view"
import { FollowingView } from "@/components/community/following-view"
import { Search, Users } from "lucide-react"
import { useAuthContext } from "@/components/provider/auth-provider"


export default function CommunityPage() {
 
  const { followers, following, users } = useAuthContext()

  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="max-w-7xl mx-auto relative z-10">
        <div className="mb-8">
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            Community
          </h1>
          <p className="text-gray-400 mt-2">Connect with fellow students and discover study partners</p>
        </div>

        <Tabs defaultValue="explore" className="w-full">
          <TabsList className="glass border-0 mb-6">
            <TabsTrigger value="explore" className="flex items-center space-x-2">
              <Search className="h-4 w-4" />
              <span>Explore</span>
            </TabsTrigger>
            <TabsTrigger value="followers" className="flex items-center space-x-2">
              <Users className="h-4 w-4" />
              <span>Followers</span>
            </TabsTrigger>
            <TabsTrigger value="following" className="flex items-center space-x-2">
              <Users className="h-4 w-4" />
              <span>Following</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="explore">
            <ExploreView users={users} />
          </TabsContent>

          <TabsContent value="followers">
            <FollowersView followers={followers} />
          </TabsContent>

          <TabsContent value="following">
            <FollowingView following={following} />
          </TabsContent>

        </Tabs>
      </div>
    </div>
  )
}
