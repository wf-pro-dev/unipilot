"use client"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Search, Filter, X } from "lucide-react"
import { useRouter } from "next/navigation"
import { Assignment } from "@/types/models"
import { useAssignments } from "@/hooks/use-assignments"

interface Filter {
  course: string | null
  status: string | null
  priority: string | null
}
interface AssignmentFiltersProps {
  searchTerm: string
  setSearchTerm: (term: string) => void
  filter: Filter
}

export function AssignmentFilters({
  filter,
  searchTerm,
  setSearchTerm,
}: AssignmentFiltersProps) {
  
  const router = useRouter()
  const hasActiveFilters =
    filter.course !== "all" || filter.status !== "all" || filter.priority !== "all" || searchTerm !== ""

  const { assignments } = useAssignments()

  const clearFilters = () => {
    setSearchTerm("")
    handleFilterChange({
      course: "all",
      status: "all",
      priority: "all",
    })
  }

  const handleFilterChange = (filter: Filter) => {
    var urlParams = ""
    for (const [key, value] of Object.entries(filter)) {
      if (value && value !== "all") {
        urlParams += `${key}=${value}&`
      }
    }
    const params = new URLSearchParams(urlParams)
    router.push(`/assignments?view=list&${params.toString()}`, { scroll: false })
  }

  const courses = Array.from(new Set(assignments.map((assignment: Assignment) => assignment.Course?.Code)))
  const statuses = Array.from(new Set(assignments.map((assignment: Assignment) => assignment.StatusName)))
  const priorities = Array.from(new Set(assignments.map((assignment: Assignment) => assignment.Priority)))

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <Input
            placeholder="Search assignments..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10 bg-gray-800/50 border-gray-600"
          />
        </div>

        <div className="flex gap-2">
          <Select value={filter.course} onValueChange={(value) => handleFilterChange({ ...filter, course: value })}>
            <SelectTrigger className="w-40 bg-gray-800/50 border-gray-600">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="glass border-gray-600">
            <SelectItem value="all">All Courses</SelectItem>
              {courses.map((course) => (
                <SelectItem key={course} value={course}>
                  {course}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

            <Select value={filter.status} onValueChange={(value) => handleFilterChange({ ...filter, status: value })}>
            <SelectTrigger className="w-36 bg-gray-800/50 border-gray-600">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="glass border-gray-600">
            <SelectItem value="all">All Statuses</SelectItem>
              {statuses.map((status) => (
                <SelectItem key={status} value={status}>
                  {status}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={filter.priority} onValueChange={(value) => handleFilterChange({ ...filter, priority: value })}>
            <SelectTrigger className="w-36 bg-gray-800/50 border-gray-600">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="glass border-gray-600">
              <SelectItem value="all">All Priorities</SelectItem>
              {priorities.map((priority) => (
                <SelectItem key={priority} value={priority}>
                  {priority}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {hasActiveFilters && (
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Filter className="h-4 w-4 text-gray-400" />
            <span className="text-sm text-gray-400">Active filters:</span>
            {searchTerm && (
              <Badge variant="secondary" className="bg-blue-500/20 text-blue-400">
                Search: {searchTerm}
              </Badge>
            )}
            {filter.course !== "all" && (
              <Badge variant="secondary" className="bg-green-500/20 text-green-400">
                {courses.find((c) => c === filter.course)}
              </Badge>
            )}
            {filter.status !== "all" && (
              <Badge variant="secondary" className="bg-purple-500/20 text-purple-400">
                {statuses.find((s) => s === filter.status)}
              </Badge>
            )}
            {filter.priority !== "all" && (
              <Badge variant="secondary" className="bg-orange-500/20 text-orange-400">
                 {priorities.find((p) => p === filter.priority)}
              </Badge>
            )}
          </div>
          <Button variant="ghost" size="sm" onClick={clearFilters} className="text-gray-400 hover:text-white">
            <X className="h-4 w-4 mr-1" />
            Clear
          </Button>
        </div>
      )}
    </div>
  )
}
