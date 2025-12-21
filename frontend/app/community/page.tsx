"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ExploreView } from "@/components/community/explore-view"
import { FollowersView } from "@/components/community/followers-view"
import { FollowingView } from "@/components/community/following-view"
import { Search, Users } from "lucide-react"
import { useAuthContext } from "@/components/provider/auth-provider"

/**
 * Community page component for user discovery and social connections.
 * 
 * Provides a tabbed interface for exploring users, viewing followers, and managing
 * following relationships. The page displays user cards in a grid layout with
 * options to follow/unfollow users and discover study partners.
 * 
 * Features:
 * - Tab-based navigation (Explore, Followers, Following)
 * - User discovery and search functionality
 * - Follow/unfollow user management
 * - Responsive grid layout for user cards
 * 
 * Data Source:
 * - Uses `useAuthContext` to access user relationship data (followers, following, users)
 * - Data is fetched and managed at the provider level for global state access
 * 
 * @returns {JSX.Element} The community page with tab navigation and user views
 */
export default function CommunityPage() {
  // Extract user relationship data from auth context (followers, following, all users)
  // Context provides pre-fetched data avoiding prop drilling through component tree
  const { followers, following, users } = useAuthContext()

  return (
    <div className="">

      {/* Main content container with max-width constraint and z-index above background */}
      <div className="max-w-7xl mx-auto relative z-10">
        {/* Page header with gradient text effect for visual appeal */}
        <div className="mb-8">
          <h1 className="text-h1 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            Community
          </h1>
          <p className="text-body-small text-gray-400 mt-3">Connect with fellow students and discover study partners</p>
        </div>

        {/* Tab navigation with default "explore" view for user discovery */}
        <Tabs defaultValue="explore" className="w-full">
          {/* Tab list with glass morphism styling matching design system */}
          <TabsList className="h-full flex w-fit bg-white/5 p-1 rounded-xl mb-6 border border-white/5">
            <TabsTrigger 
              value="explore" 
              className="flex w-60 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Search className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Explore</span>
            </TabsTrigger>
            <TabsTrigger 
              value="followers" 
              className="flex w-60 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Followers</span>
            </TabsTrigger>
            <TabsTrigger 
              value="following" 
              className="flex w-60 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Following</span>
            </TabsTrigger>
          </TabsList>

          {/* Explore tab: displays all users for discovery and following */}
          <TabsContent value="explore">
            <ExploreView users={users} />
          </TabsContent>

          {/* Followers tab: shows users who follow the current user */}
          <TabsContent value="followers">
            <FollowersView followers={followers} />
          </TabsContent>

          {/* Following tab: displays users the current user is following */}
          <TabsContent value="following">
            <FollowingView following={following} />
          </TabsContent>

        </Tabs>
      </div>
    </div>
  )
}
