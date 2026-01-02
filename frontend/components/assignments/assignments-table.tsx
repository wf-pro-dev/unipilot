"use client"

import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { AssignmentItem } from "./assignment-item"
import { List, X } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { Input } from "../ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../ui/select"
import { Filter, Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { useRouter } from "next/navigation"
import { useCoursesBySemester } from "@/hooks/use-courses"
import { useAuthContext } from "../provider/auth-provider"
import { CoursesSelect } from "../courses/courses-select"
import { EmptyState } from "../ui/empty-state"

interface Filter {
  course: string | null
  status: string | null
  priority: string | null
}

interface AssignmentsTableProps {
  assignments: assignment.LocalAssignment[]
  onToggleComplete: (assignment: assignment.LocalAssignment) => void
  onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: assignment.LocalAssignment) => void
  onOpenEdit: (assignment: assignment.LocalAssignment) => void
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  filter: Filter
  isLoading?: boolean
}

export function AssignmentsTable({
  assignments,
  onToggleComplete,
  onEdit,
  onDelete,
  onOpenEdit,
  onAssignmentClick,
  filter,
  isLoading = false
}: AssignmentsTableProps) {
  const router = useRouter()
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedCourse, setSelectedCourse] = useState(filter.course || "all")
  const [selectedStatus, setSelectedStatus] = useState(filter.status || "all")
  const [selectedPriority, setSelectedPriority] = useState(filter.priority || "all")

  const { user } = useAuthContext()
  const { data: courses } = useCoursesBySemester(user?.Semester || "FALL 2025")

  const courseCodes = Array.from(new Set(courses?.map((course) => course.Code) || []))
  const statuses = Array.from(new Set((assignments || []).map((assignment) => assignment.Status)))
  const priorities = Array.from(new Set((assignments || []).map((assignment) => assignment.Priority)))

  const hasActiveFilters = selectedCourse !== "all" || selectedStatus !== "all" || selectedPriority !== "all" || searchTerm !== ""


  // Apply basic filters (simplified for now)
  const filteredAssignments = useMemo(() => (assignments || []).filter((assignment) => {
    const matchesSearch =
      assignment.Title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      assignment.Course?.Name.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesCourse = selectedCourse === "all" || assignment.Course?.Code === selectedCourse
    const matchesStatus = selectedStatus === "all" || assignment.Status === selectedStatus
    const matchesPriority = selectedPriority === "all" || assignment.Priority === selectedPriority
    return matchesSearch && matchesCourse && matchesStatus && matchesPriority
  }), [assignments, searchTerm, selectedCourse, selectedStatus, selectedPriority])

  const onCourseChange = (course: string) => {
    router.push(`/assignments?view=list&course=${course}&status=${selectedStatus}&priority=${selectedPriority}`)
  }

  const onStatusChange = (status: string) => {
    router.push(`/assignments?view=list&course=${selectedCourse}&status=${status}&priority=${selectedPriority}`)
  }

  const onPriorityChange = (priority: string) => {
    router.push(`/assignments?view=list&course=${selectedCourse}&status=${selectedStatus}&priority=${priority}`)
  }


  useEffect(() => {
    if (filter.course) {
      setSelectedCourse(filter.course || "all")
    }
    if (filter.status) {
      setSelectedStatus(filter.status || "all")
    }
    if (filter.priority) {
      setSelectedPriority(filter.priority || "all")
    }
  }, [filter])

  const clearFilters = () => {
    setSearchTerm("")
    setSelectedCourse("all")
    setSelectedStatus("all")
    setSelectedPriority("all")
    router.push("/assignments?view=list")
  }

  return (
    <div className="flex flex-col flex-1 space-y-4">

      <GlassCard className="border-white/5 bg-white/5 backdrop-blur-xl shadow-lg shadow-black/20">
        <CardContent className="p-5">
          <div className="space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search assignments..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-white/5 border-white/10    transition-all duration-300 h-10"
                />
              </div>

              <div className="flex items-center gap-3">

                <div className="w-36">
                  <CoursesSelect
                    value={selectedCourse}
                    onValueChange={onCourseChange}
                  >
                    <SelectItem value="all">All Courses</SelectItem>
                  </CoursesSelect>
                </div>


                <Select value={selectedStatus} onValueChange={onStatusChange}>
                  <SelectTrigger className="w-36 bg-white/5 border-white/10 h-10 focus:border-blue-500 focus:ring-blue-500/20">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                    <SelectItem value="all" className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">All Statuses</SelectItem>
                    {statuses.map((status) => (
                      <SelectItem key={status} value={status} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                        {status}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>


                <Select value={selectedPriority} onValueChange={onPriorityChange}>
                  <SelectTrigger className="w-36 bg-white/5 border-white/10 h-10 focus:border-blue-500 focus:ring-blue-500/20">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                    <SelectItem value="all" className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">All Priorities</SelectItem>
                    {priorities.map((priority) => (
                      <SelectItem key={priority} value={priority} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                        {priority}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            {hasActiveFilters && (
              <div className="flex items-center justify-between pt-2 border-t border-white/5">
                <div className="flex items-center space-x-2 flex-wrap gap-y-2">
                  <div className="flex items-center text-xs font-medium text-gray-400 uppercase tracking-wider mr-2">
                    <Filter className="h-3.5 w-3.5 mr-1.5" />
                    Filters:
                  </div>
                  {searchTerm && (
                    <Badge variant="secondary" className="text-[10px] bg-blue-500/10 text-blue-400 border border-blue-500/20 px-2 py-0.5 font-medium">
                      Search: {searchTerm}
                    </Badge>
                  )}
                  {selectedCourse !== "all" && (
                    <Badge variant="secondary" className="text-[10px] bg-green-500/10 text-green-400 border border-green-500/20 px-2 py-0.5 font-medium">
                      {selectedCourse}
                    </Badge>
                  )}
                  {selectedStatus !== "all" && (
                    <Badge variant="secondary" className="text-[10px] bg-purple-500/10 text-purple-400 border border-purple-500/20 px-2 py-0.5 font-medium">
                      {selectedStatus}
                    </Badge>
                  )}
                  {selectedPriority !== "all" && (
                    <Badge variant="secondary" className="text-[10px] bg-orange-500/10 text-orange-400 border border-orange-500/20 px-2 py-0.5 font-medium">
                      {selectedPriority}
                    </Badge>
                  )}
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearFilters}
                  className="text-gray-400 hover:text-white hover:bg-white/10 h-8 text-xs"
                >
                  <X className="h-3.5 w-3.5 mr-1.5" />
                  Clear All
                </Button>
              </div>
            )}
          </div>
        </CardContent>
      </GlassCard>

      {filteredAssignments.length === 0 ? (
        <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
          <EmptyState
            icon={List}
            title="No assignments found"
            description="Try adjusting your filters or search terms"
            className="flex-1 items-center"
            onClick={clearFilters}
            buttonText="Clear Filters"
          />

        </div>

      ) : (
        <div className="grid md:grid-cols-1 lg:grid-cols-2 gap-4">
          {filteredAssignments.map((assignment) => (
            <AssignmentItem
              key={assignment.ID}
              assignment={assignment}
              onToggleComplete={onToggleComplete}
              onEdit={onEdit}
              onDelete={onDelete}
              onOpenEdit={onOpenEdit}
              onAssignmentClick={onAssignmentClick}
              disabled={isLoading}
            />
          ))}
        </div>
      )
      }

    </div >
  )
}
