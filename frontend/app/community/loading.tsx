import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

/**
 * Loading skeleton component for the community page.
 * 
 * This component is automatically displayed by Next.js during route transitions
 * when a `loading.tsx` file is present in the route directory. It provides a
 * visual placeholder that matches the layout of the actual community page
 * to improve perceived performance and user experience.
 * 
 * The skeleton includes:
 * - Page header with title and description placeholders
 * - Tab navigation bar skeleton (Explore, Followers, Following)
 * - Grid of 6 user card skeletons with avatar, name, metadata, and action buttons
 * 
 * @returns {JSX.Element} A skeleton loading UI matching the community page layout
 */
export default function CommunityLoading() {
  return (
    <div className="min-h-screen bg-animated p-6">
      <div className="max-w-7xl mx-auto">
        {/* Page header skeleton - title and description matching community page */}
        <div className="mb-8">
          <Skeleton className="h-10 w-64 mb-2" />
          <Skeleton className="h-4 w-96" />
        </div>

        <div className="space-y-6">
          {/* Tab navigation bar skeleton - represents Explore, Followers, Following tabs */}
          <Card className="glass border-0">
            <CardHeader>
              <div className="flex space-x-1">
                <Skeleton className="h-10 w-24" />
                <Skeleton className="h-10 w-32" />
                <Skeleton className="h-10 w-24" />
              </div>
            </CardHeader>
          </Card>

          {/* User cards grid skeleton - responsive grid matching actual layout */}
          {/* 6 cards provide enough visual content to indicate loading state */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i} className="glass-dark border-0">
                <CardContent className="p-6">
                  {/* User avatar and name section */}
                  <div className="flex items-center space-x-4 mb-4">
                    <Skeleton className="h-12 w-12 rounded-full" />
                    <div className="space-y-2">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="h-3 w-16" />
                    </div>
                  </div>
                  {/* User bio/content skeleton */}
                  <div className="space-y-2">
                    <Skeleton className="h-3 w-full" />
                    <Skeleton className="h-3 w-3/4" />
                  </div>
                  {/* Action buttons skeleton - Follow/Message buttons */}
                  <div className="flex space-x-2 mt-4">
                    <Skeleton className="h-6 w-16" />
                    <Skeleton className="h-6 w-20" />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
