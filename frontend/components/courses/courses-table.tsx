"use client"

import { BookOpen, List } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useRouter } from "next/navigation"
import { CourseItem } from "./course-item"
import { EmptyState } from "../ui/empty-state"
import { Scroll } from "../core/scroll"
import { GlassCard } from "../ui/glass-card"
import { SearchFilter, FilterDefinition, SearchConfig } from "../core/search-filter/search-filter"

interface Filter {
  semester: string | null
  instructor: string | null
}

interface CoursesTableProps {
  courses: models.LocalCourse[]
  filter: Filter
}

export function CoursesTable({ courses, filter }: CoursesTableProps) {
  const router = useRouter()

  // Early return for no courses
  if (courses.length === 0) {
    return (
      <GlassCard variant="board">
        <EmptyState
          icon={BookOpen}
          title="No courses found"
          description="Create a new course to get started"
          className="flex-1 items-center"
        />
      </GlassCard>
    )
  }

  // Search configuration
  const searchConfig: SearchConfig<models.LocalCourse> = {
    placeholder: "Search courses by code, name or instructor...",
    searchableFields: ["Code", "Name", "Instructor"],
    enabled: true
  }

  // Filter definitions
  const filterDefinitions: FilterDefinition<models.LocalCourse>[] = [
    {
      field: "Semester",
      label: "Semester",
      type: "select",
      placeholder: "All Semesters",
      width: "w-48"
    },
    {
      field: "Instructor",
      label: "Instructors",
      type: "select",
      placeholder: "All Instructors",
      width: "w-48"
    }
  ]

  // Initial filter values from URL params
  const initialFilters = {
    Semester: filter.semester || "all",
    Instructor: filter.instructor || "all"
  }

  // Handlers
  const handleFilterChange = (filters: Record<string, string>) => {
    const semesterFilter = filters.Semester === "all" ? "" : filters.Semester
    const instructorFilter = filters.Instructor === "all" ? "" : filters.Instructor
    
    const params = new URLSearchParams()
    params.set("view", "list")
    if (semesterFilter) params.set("semester", semesterFilter)
    if (instructorFilter) params.set("instructor", instructorFilter)
    
    router.push(`/courses?${params.toString()}`)
  }

  const handleSearchChange = (searchTerm: string) => {
    // If you want to persist search in URL, add it here
    console.log("Search term:", searchTerm)
  }

  const handleClearAll = () => {
    router.push("/courses?view=list")
  }

  return (
    <SearchFilter
      data={courses}
      searchConfig={searchConfig}
      filterDefinitions={filterDefinitions}
      initialFilters={initialFilters}
      onFilterChange={handleFilterChange}
      onSearchChange={handleSearchChange}
      onClearAll={handleClearAll}
      layout="horizontal"
    >
      {(filteredCourses) => (
        <>
          {filteredCourses.length > 0 ? (
            <Scroll
              data={{ Data: filteredCourses, HasMore: false }}
              renderItem={(course: models.LocalCourse) => (
                <CourseItem
                  key={course.ID}
                  courseId={course.ID}
                />
              )}
              keyExtractor={(item: models.LocalCourse) => item.ID}
              numColumns={3}
              containerClassName="gap-4"
            />
          ) : (
            <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
              <EmptyState
                icon={List}
                title="No courses found"
                description="Try adjusting your filters or search terms"
                className="flex-1 items-center"
                onClick={handleClearAll}
                buttonText="Clear Filters"
              />
            </div>
          )}
        </>
      )}
    </SearchFilter>
  )
}