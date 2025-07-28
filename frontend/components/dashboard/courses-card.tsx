"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { BookOpen, Plus, Loader2 } from "lucide-react"
import Link from "next/link"
import { useCourses } from "@/hooks/use-courses"

export function CoursesCard() {
  const { data: courses = [], isLoading } = useCourses()
  
  return (
    <Card className="glass border-0">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-white flex items-center space-x-2">
          <BookOpen className="h-4 w-4 text-blue-400" />
          <span>My Courses</span>
          {isLoading && <Loader2 className="h-3 w-3 animate-spin ml-2" />}
        </CardTitle>
        <Link href="/courses">
          <Button variant="ghost" size="sm" className="text-blue-400 hover:text-blue-300">
            View All
          </Button>
        </Link>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {(courses || []).slice(0, 3).map((course) => (
            <div key={course.ID} className="flex items-center space-x-3">
              <div className={`w-3 h-3 rounded-full ${course.Color}`} />
              <div className="flex-1">
                <p className="text-sm font-medium text-white">{course.Code}</p>
                <p className="text-xs text-gray-400">{course.Name}</p>
              </div>
            </div>
          ))}
          
          {courses.length === 0 && !isLoading && (
            <div className="text-center py-4">
              <p className="text-xs text-gray-400">No courses yet</p>
            </div>
          )}
          
          <Button 
            variant="outline" 
            size="sm" 
            className="w-full mt-4 border-gray-600 bg-transparent"
            disabled={isLoading}
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Course
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
