"use client"

import { useState } from "react"
import { CoursesGrid } from "@/components/courses/courses-grid"
import { AddCourseDialog } from "@/components/courses/add-course-dialog"
import { CourseDetailsModal } from "@/components/courses/course-details-modal"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Search, Filter, X, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useCourses } from "@/hooks/use-courses"
import { Course } from "@/types/models"

export default function CoursesPage() {
  const { data: courses = [], isLoading, error } = useCourses()
  const [selectedCourse, setSelectedCourse] = useState<Course | null>(null)
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedSemester, setSelectedSemester] = useState("all")
  const [selectedInstructor, setSelectedInstructor] = useState("all")

  const filteredCourses = (courses || []).filter((course) => {
    const matchesSearch =
      course.Code.toLowerCase().includes(searchTerm.toLowerCase()) ||
      course.Name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      course.Instructor.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesSemester = selectedSemester === "all" || course.Semester === selectedSemester
    const matchesInstructor = selectedInstructor === "all" || course.Instructor === selectedInstructor

    return matchesSearch && matchesSemester && matchesInstructor
  })

  const semesters = Array.from(new Set((courses || []).map((course) => course.Semester)))
  const instructors = Array.from(new Set((courses || []).map((course) => course.Instructor)))

  const hasActiveFilters = selectedSemester !== "all" || selectedInstructor !== "all" || searchTerm !== ""

  const clearFilters = () => {
    setSearchTerm("")
    setSelectedSemester("all")
    setSelectedInstructor("all")
  }

  // Show loading state
  if (isLoading) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin mr-2" />
          <span>Loading courses...</span>
        </div>
      </div>
    )
  }

  // Show error state
  if (error) {
    return (
      <div className="page">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500">
            Error loading courses: {error.message}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="relative z-10">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              My Courses
            </h1>
            <p className="text-gray-400 mt-2">
              Manage your enrolled courses ({courses.length} total)
            </p>
          </div>
          <AddCourseDialog />
        </div>

        {/* Search and Filters */}
        <div className="flex flex-col md:flex-row gap-4 mb-6">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search courses..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 glass border-gray-600 bg-white/5"
            />
          </div>
          
          <div className="flex gap-2">
            <Select value={selectedSemester} onValueChange={setSelectedSemester}>
              <SelectTrigger className="w-[180px] glass border-gray-600 bg-white/5">
                <SelectValue placeholder="All Semesters" />
              </SelectTrigger>
              <SelectContent className="glass border-gray-600">
                <SelectItem value="all">All Semesters</SelectItem>
                {semesters.map((semester) => (
                  <SelectItem key={semester} value={semester}>
                    {semester}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={selectedInstructor} onValueChange={setSelectedInstructor}>
              <SelectTrigger className="w-[180px] glass border-gray-600 bg-white/5">
                <SelectValue placeholder="All Instructors" />
              </SelectTrigger>
              <SelectContent className="glass border-gray-600">
                <SelectItem value="all">All Instructors</SelectItem>
                {instructors.map((instructor) => (
                  <SelectItem key={instructor} value={instructor}>
                    {instructor}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {hasActiveFilters && (
              <Button
                variant="outline"
                size="sm"
                onClick={clearFilters}
                className="glass border-gray-600"
              >
                <X className="h-4 w-4 mr-2" />
                Clear
              </Button>
            )}
          </div>
        </div>

        {/* Active Filters */}
        {hasActiveFilters && (
          <div className="flex items-center gap-2 mb-6">
            <Filter className="h-4 w-4 text-gray-400" />
            <span className="text-sm text-gray-400">Active filters:</span>
            {searchTerm && (
              <Badge variant="secondary" className="text-xs">
                Search: {searchTerm}
              </Badge>
            )}
            {selectedSemester !== "all" && (
              <Badge variant="secondary" className="text-xs">
                Semester: {selectedSemester}
              </Badge>
            )}
            {selectedInstructor !== "all" && (
              <Badge variant="secondary" className="text-xs">
                Instructor: {selectedInstructor}
              </Badge>
            )}
          </div>
        )}

        <CoursesGrid 
          courses={filteredCourses} 
          onCourseClick={setSelectedCourse} 
        />

        <CourseDetailsModal
          isOpen={!!selectedCourse}
          onClose={() => setSelectedCourse(null)}
          course={selectedCourse}
        />
      </div>
    </div>
  )
}
