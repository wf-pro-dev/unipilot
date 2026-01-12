"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ExploreView } from "@/components/community/explore-view"
import { FollowersView } from "@/components/community/followers-view"
import { FollowingView } from "@/components/community/following-view"
import { BookOpenIcon, FileText, Search, Users } from "lucide-react"
import { useAuthContext } from "@/components/provider/auth-provider"
import { useCallback, useEffect, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { useCurrentUser, useGetCourseInvitations } from "@/hooks/use-auth"
import { CourseItem } from "@/components/courses/course-item"
import { useAcceptCourseInvitation, useCoursesLinked, useDeclineCourseInvitation } from "@/hooks/use-courses"
import { EmptyState, HorizontalEmptyState } from "@/components/ui/empty-state"
import { GlassCard } from "@/components/ui/glass-card"
import { toast } from "sonner"
import { models } from "@/wailsjs/go/models"
import { useMemo } from "react"
import { AssignmentItem } from "@/components/assignments/assignment-item"
import { useCreateAssignment, useAssignments } from "@/hooks/use-assignments"

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
  const { followers, following, users, courses: userCourses } = useAuthContext()
  const { data: courseInvitations } = useGetCourseInvitations()
  const { data: coursesLinked } = useCoursesLinked()
  const { data: currentUser } = useCurrentUser()
  const { data: assignments } = useAssignments()
  const AcceptCourseInvitation = useAcceptCourseInvitation()
  const DeclineCourseInvitation = useDeclineCourseInvitation()
  const createMutation = useCreateAssignment()

  const [selectedCourse, setSelectedCourse] = useState<models.Course | undefined>(undefined)

  const router = useRouter()
  const searchParams = useSearchParams()
  // Get the current view from URL parameters, default to "today"
  const currentView = searchParams.get("view") || "explore"

  // Valid view values
  const validViews = ["explore", "followers", "following", "social"]

  // Ensure the current view is valid, otherwise default to "today"
  const activeView = validViews.includes(currentView) ? currentView : "explore"



  const handleAcceptCourseInvitation = (invitation: models.CourseInvitation) => {
    AcceptCourseInvitation.mutate({ invitation })
  }

  const handleDeclineCourseInvitation = (invitation: models.CourseInvitation) => {
    DeclineCourseInvitation.mutate(invitation)
  } 

  /**
  * Handles assignment creation with optimistic UI updates.
  * 
  * Creates a new assignment and provides immediate UI feedback. Logs the creation
  * and shows success/error toast notifications.
  * 
  * @param {assignment.LocalAssignment} assignment - The assignment to create
  * @returns {Promise<void>}
  */
  const handleCopyAssignment = async (assignment: models.Assignment) => {
    const message = "[Frontend] assignment " + assignment.Title + " added"
    var correspondingCourse = userCourses?.find((c) => c.Code == assignment.CourseCode)
    var newAssignment = models.LocalAssignment.createFrom(assignment)
    createMutation.mutate({
      ...newAssignment,
      Status: "Not started",
      ParentID: newAssignment.ID,
      CourseID: correspondingCourse?.ID,
      RemoteCourseID: correspondingCourse?.RemoteID,
    } as models.LocalAssignment)

  }

  const courses = useMemo(() => {
    return coursesLinked?.filter((course) => course.UserID != currentUser?.ID)
  }, [coursesLinked, currentUser])


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

  useEffect(() => {
    if (courses?.[0]) {
      setSelectedCourse(courses?.[0])
    }
  }, [courses])

  const handleCourseClick = (course: models.LocalCourse | models.Course) => {
    setSelectedCourse(course as models.Course)
  }


  const filteredAssignments = useMemo(() => {
    return selectedCourse?.Assignments?.filter((assignment) =>
      !assignments?.some((a: models.LocalAssignment) => a.ParentID === assignment.ID)
    )
  }, [assignments, selectedCourse])

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
              value="followers"
              className="flex w-48 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Followers</span>
            </TabsTrigger>
            <TabsTrigger
              value="following"
              className="flex w-48 justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
            >
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline text-sm font-medium">Following</span>
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
            <ExploreView users={users} />
          </TabsContent>

          {/* Followers tab: shows users who follow the current user */}
          <TabsContent value="followers" className="flex flex-col data-[state=active]:flex-1 m-0">
            <FollowersView followers={followers} />
          </TabsContent>

          {/* Following tab: displays users the current user is following */}
          <TabsContent value="following" className="flex flex-col data-[state=active]:flex-1 m-0">
            <FollowingView following={following} />
          </TabsContent>

          <TabsContent value="social" className="flex flex-col data-[state=active]:flex-1 m-0">
            <div className="flex flex-1 gap-6">

              <div className="flex flex-col flex-1 gap-4">
                <h3 className="flex items-end gap-2 text-h3">
                  {selectedCourse?.Name}
                  <p className="text-body text-text-caption align-baseline">{selectedCourse?.Code}</p>
                </h3>
                {filteredAssignments && filteredAssignments.length > 0 && (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filteredAssignments?.map((a) => {
                      const assignment = {
                        ...a,
                        Course: selectedCourse
                      } as models.Assignment
                      return (
                        <AssignmentItem
                          key={assignment.ID}
                          assignment={assignment}
                          variant="outline"
                          mode="user"
                          onCopy={handleCopyAssignment}
                          user={selectedCourse?.User}
                        />
                      )
                    })}
                  </div>
                )}
                {filteredAssignments && filteredAssignments.length === 0 && (
                  <GlassCard variant="board" className="flex-1 items-center justify-center">
                    <EmptyState
                      icon={FileText}
                      title="No assignments found"
                      description="Create a new assignment to get started"
                    />
                  </GlassCard>
                )}
              </div>

              <div className="flex flex-col gap-4">

                {courseInvitations && courseInvitations.length > 0 && (
                  <div className="flex flex-col gap-2">

                    <h4 className="text-h4 self-end">Course Invitations</h4>
                    <div className="flex flex-wrap gap-4" >
                      {courseInvitations?.map((invitation) => (
                        <div key={invitation.ID}>
                          <div>
                            <CourseItem
                              course={invitation.Course!}
                              onEdit={() => { }}
                              onDelete={() => { }}
                              onCourseClick={() => setSelectedCourse(invitation.Course)}
                              size="sm"
                              onAccept={() => handleAcceptCourseInvitation(invitation)}
                              onDecline={() => handleDeclineCourseInvitation(invitation)}
                            />
                          </div>
                        </div>
                      ))}
                    </div>

                  </div>
                )}

                <div className="flex flex-col flex-1 gap-2">
                  <GlassCard
                    variant="board"
                    className="p-4 flex-1"
                  >
                    {coursesLinked && coursesLinked.length > 0 ? (

                      <div className="flex flex-col flex-1 gap-4">
                        {courses?.map((course) => (

                          <CourseItem
                            key={course.ID}
                            course={course}
                            onCourseClick={handleCourseClick}
                            onEdit={() => { }}
                            onDelete={() => { }}
                            size="sm"
                          />
                        ))
                        }
                      </div>

                    ) : (


                      <EmptyState
                        icon={BookOpenIcon}
                        title="No courses linked"
                        description="Link a course to get started"
                        className="flex-1 items-center"
                      />

                    )}
                  </GlassCard>
                </div>

              </div>

            </div>
          </TabsContent>

        </Tabs>
      </div>
    </div >
  )
}
