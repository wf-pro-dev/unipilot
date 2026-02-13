"use client"

import { List } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useRouter, useSearchParams } from "next/navigation"
import { AssignmentItem } from "./assignment-item"
import { EmptyState } from "../ui/empty-state"
import { Scroll } from "../core/scroll"
import { SearchFilter } from "../core/search-filter/search-filter"
import { FilterDefinition, SearchConfig } from "../core/search-filter/types"

interface Filter {
  course: string | null
  status: string | null
  priority: string | null
}

interface AssignmentsTableProps {
  assignments: models.LocalAssignment[]
  filter: Filter
  isLoading?: boolean
}

export function AssignmentsTable({
  assignments,
  filter,
  isLoading = false
}: AssignmentsTableProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const currentView = searchParams.get("view") || "week"

  // Search configuration
  const searchConfig: SearchConfig<models.LocalAssignment> = {
    placeholder: "Search assignments...",
    searchableFields: ["Title", "Course"],
    enabled: true
  }

  // Filter definitions
  const filterDefinitions: FilterDefinition<models.LocalAssignment>[] = [
    {
      field: "Course",
      label: "Courses",
      type: "select",
      placeholder: "All Courses",
      width: "w-36",
      extractOptions: (data) => {
        // Extract unique course codes from nested Course object
        return Array.from(new Set(
          data
            .filter(a => a.Course?.Code)
            .map(a => a.Course?.Code || "")
        ))
      },
      customOptions: undefined // Or provide custom options if needed
    },
    {
      field: "Status",
      label: "Statuses",
      type: "select",
      placeholder: "All Statuses",
      width: "w-36"
    },
    {
      field: "Priority",
      label: "Priorities",
      type: "select",
      placeholder: "All Priorities",
      width: "w-36"
    }
  ]

  // Initial filter values from URL params
  const initialFilters = {
    Course: filter.course || "all",
    Status: filter.status || "all",
    Priority: filter.priority || "all"
  }

  // Handlers
  const handleFilterChange = (filters: Record<string, string>) => {
    const courseFilter = filters.Course === "all" ? "" : filters.Course
    const statusFilter = filters.Status === "all" ? "" : filters.Status
    const priorityFilter = filters.Priority === "all" ? "" : filters.Priority
    
    const params = new URLSearchParams()
    params.set("view", currentView)
    if (courseFilter) params.set("course", courseFilter)
    if (statusFilter) params.set("status", statusFilter)
    if (priorityFilter) params.set("priority", priorityFilter)
    
    router.push(`/assignments?${params.toString()}`)
  }

  const handleSearchChange = (searchTerm: string) => {
    // If you want to persist search in URL, add it here
    console.log("Search term:", searchTerm)
  }

  const handleClearAll = () => {
    router.push(`/assignments?view=${currentView}`)
  }

  return (
    <SearchFilter
      data={assignments}
      searchConfig={searchConfig}
      filterDefinitions={filterDefinitions}
      initialFilters={initialFilters}
      onFilterChange={handleFilterChange}
      onSearchChange={handleSearchChange}
      onClearAll={handleClearAll}
      isLoading={isLoading}
      layout="horizontal"
    >
      {(filteredAssignments) => (
        <>
          {filteredAssignments.length > 0 ? (
            <Scroll
              data={{ Data: filteredAssignments, HasMore: false }}
              renderItem={(assignment: models.LocalAssignment) => (
                <AssignmentItem
                  key={assignment.ID}
                  assignmentId={assignment.ID}
                  disabled={isLoading}
                  variant="outline"
                />
              )}
              keyExtractor={(item: models.LocalAssignment) => item.ID}
              numColumns={3}
              containerClassName="gap-4"
            />
          ) : (
            <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
              <EmptyState
                icon={List}
                title="No assignments found"
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