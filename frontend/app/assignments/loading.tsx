import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

/**
 * Loading skeleton component for the assignments page.
 * 
 * This component is automatically displayed by Next.js during route transitions
 * when a `loading.tsx` file is present in the route directory. It provides a
 * visual placeholder that matches the layout of the actual assignments page
 * to improve perceived performance and user experience.
 * 
 * The skeleton includes:
 * - Page header with title and description placeholders
 * - Filter/search bar card with control skeletons
 * - List of 5 assignment card skeletons with checkbox, title, metadata, and date placeholders
 * 
 * @returns {JSX.Element} A skeleton loading UI matching the assignments page layout
 */
export default function AssignmentsLoading() {
  return (
    <div className="min-h-screen bg-animated p-6">
      <div className="max-w-7xl mx-auto">
        {/* Page header skeleton - title and description */}
        <div className="mb-8">
          <Skeleton className="h-10 w-64 mb-2" />
          <Skeleton className="h-4 w-96" />
        </div>

        <div className="space-y-6">
          {/* Filter/search bar card skeleton */}
          <Card className="glass border-0">
            <CardHeader>
              <Skeleton className="h-6 w-32" />
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Filter controls skeleton */}
              <div className="flex gap-4">
                <Skeleton className="h-10 flex-1" />
                <Skeleton className="h-10 w-40" />
                <Skeleton className="h-10 w-36" />
                <Skeleton className="h-10 w-36" />
              </div>
            </CardContent>
          </Card>

          {/* Assignment list skeleton - renders 5 placeholder cards */}
          <div className="space-y-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <Card key={i} className="glass-dark border-0">
                <CardContent className="p-4">
                  <div className="flex items-start space-x-4">
                    {/* Checkbox skeleton */}
                    <Skeleton className="h-5 w-5 rounded" />
                    <div className="flex-1 space-y-2">
                      {/* Assignment title skeleton */}
                      <Skeleton className="h-5 w-3/4" />
                      {/* Metadata skeletons - course and status */}
                      <div className="flex items-center space-x-3">
                        <Skeleton className="h-4 w-20" />
                        <Skeleton className="h-4 w-16" />
                      </div>
                      {/* Date and action skeletons */}
                      <div className="flex items-center justify-between">
                        <Skeleton className="h-4 w-32" />
                        <Skeleton className="h-4 w-20" />
                      </div>
                    </div>
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
