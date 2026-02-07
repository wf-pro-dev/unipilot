"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { ExploreView } from "@/components/community/explore-view"
import { Search, Users } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"
import { SocialTab } from "@/components/community/social-tab"
import { FriendList } from "@/components/community/friend-list"
import { UserList } from "@/components/community/user-list"
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
  const router = useRouter()
  const searchParams = useSearchParams()
  // Get the current view from URL parameters, default to "today"
  const currentView = searchParams.get("view") || "explore"

  // Valid view values
  const validViews = ["explore", "friends", "social"]

  // Ensure the current view is valid, otherwise default to "today"
  const activeView = validViews.includes(currentView) ? currentView : "explore"

  const { user } = useAuthContext()


  /**
  * Handles tab change and synchronizes the active view with URL query parameters.
  * 
  * Updates the URL to reflect the selected tab view while preserving other
  * query parameters (filters, assignment ID, etc.).
  * 
  * @param {string} value - The tab value to switch to ("explore" | "followers" | "following")
  */
  const handleTabChange = (value: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("view", value)
    router.push(`/community?${params.toString()}`)
  }

  return (
    <div className="flex flex-col flex-1">

      {/* Main content container with max-width constraint and z-index above background */}
      <div className="flex flex-col flex-1 relative z-10">
        {/* Page header with gradient text effect for visual appeal */}
        <div className="mb-8">
          <h1 className="text-h1 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            Community
          </h1>
          <p className="text-body-small text-gray-400 mt-3">Connect with fellow students and discover study partners</p>
        </div>

        {/* Tab navigation with default "explore" view for user discovery */}
        <Tabs value={activeView} onValueChange={handleTabChange} className="flex flex-col flex-1 w-full">
          {/* Tab list with glass morphism styling matching design system */}
          <TabsList className="flex w-fit bg-white/5 p-1 rounded-xl mb-6 border border-white/5">
            <TabsTrigger
              value="explore"
              className="flex w-48 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Search className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Explore</span>
            </TabsTrigger>
            <TabsTrigger
              value="friends"
              className="flex w-48 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Friends</span>
            </TabsTrigger>
            
            <TabsTrigger
              value="social"
              className="flex w-48 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Social</span>
            </TabsTrigger>
          </TabsList>

          {/* Explore tab: displays all users for discovery and following */}
          <TabsContent value="explore" className="flex flex-col data-[state=active]:flex-1 m-0">
           <ExploreView ListComponent={<UserList userID={user?.ID!} />} />
          </TabsContent>


          {/* Following tab: displays users the current user is following */}
          <TabsContent value="friends" className="flex flex-col data-[state=active]:flex-1 m-0">
            <ExploreView ListComponent={<FriendList userID={user?.ID!} />} />
          </TabsContent>

          <TabsContent value="social" className="flex flex-col data-[state=active]:flex-1 m-0">
            <SocialTab />
          </TabsContent>

        </Tabs>
      </div>
    </div >
  )
}
