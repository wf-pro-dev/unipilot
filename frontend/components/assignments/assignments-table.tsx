"use client"

import { CardContent } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
import { AssignmentItem } from "./assignment-item"
import { List, X } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { Input } from "../ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../ui/select"
import { Filter, Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { Button } from "../ui/button"
import { Badge } from "../ui/badge"
import { useRouter, useSearchParams } from "next/navigation"
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
  assignments: models.LocalAssignment[]
  onEdit: (assignment: models.LocalAssignment, column: string, value: string) => void
  onDelete: (assignment: models.LocalAssignment) => void
  onOpenEdit: (assignment: models.LocalAssignment) => void
  onAssignmentClick: (assignment: models.LocalAssignment) => void
  filter: Filter
  isLoading?: boolean
}

export function AssignmentsTable({
  assignments,
  onEdit,
  onDelete,
  onOpenEdit,
  onAssignmentClick,
  filter,
  isLoading = false
}: AssignmentsTableProps) {
  const router = useRouter()
  const [searchTerm, setSearchTerm] = useState("")
  const [selectedCourseData, setSelectedCourseData] = useState<models.LocalCourse |  undefined>(undefined)
  const [selectedCourse, setSelectedCourse] = useState(filter.course || "all")
  const [selectedStatus, setSelectedStatus] = useState(filter.status || "all")
  const [selectedPriority, setSelectedPriority] = useState(filter.priority || "all")

  const { user } = useAuthContext()
  const { data: courses } = useCoursesBySemester(user?.Semester || "FALL 2025")
  
  const searchParams = useSearchParams()
  const currentView = searchParams.get("view") || "week"

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
    router.push(`/assignments?view=${currentView}&course=${course}&status=${selectedStatus}&priority=${selectedPriority}`)
  }

  const onStatusChange = (status: string) => {
    router.push(`/assignments?view=${currentView}&course=${selectedCourse}&status=${status}&priority=${selectedPriority}`)
  }

  const onPriorityChange = (priority: string) => {
    router.push(`/assignments?view=${currentView}&course=${selectedCourse}&status=${selectedStatus}&priority=${priority}`)
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
    setSelectedCourseData(undefined)
    router.push(`/assignments?view=${currentView}`)
  }

  return (
    <div className="flex flex-col flex-1 space-y-4">

      <GlassCard variant="board" className="flex-grow-0 flex-row">
        <CardContent className="flex-1 p-5 ">
          <div className="space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center space-x-4">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="Search assignments..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-white/5 border-white/10  rounded-xl transition-all duration-300 h-10"
                />
              </div>

              <div className="flex items-center gap-3 m-0 space-y-0">

                <div className="w-36">
                  <CoursesSelect
                    value={selectedCourse}
                    onValueChange={onCourseChange}
                    onCourseChange={setSelectedCourseData}
                    selectedCourse={selectedCourseData}
                  />
                </div>


                <Select value={selectedStatus} onValueChange={onStatusChange}>
                  <SelectTrigger className="w-36 bg-white/5 border-white/10 h-11 rounded-xl">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-transparent border-none">
                    <GlassCard variant="board">
                      <SelectItem value="all" className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">All Statuses</SelectItem>
                      {statuses.map((status) => (
                        <SelectItem key={status} value={status} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                          {status}
                        </SelectItem>
                      ))}
                    </GlassCard>
                  </SelectContent>

                </Select>


                <Select value={selectedPriority} onValueChange={onPriorityChange}>
                  <SelectTrigger className="w-36 bg-white/5 border-white/10 h-11 rounded-xl">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-transparent border-none">
                    <GlassCard variant="board">
                      <SelectItem value="all" className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">All Priorities</SelectItem>
                      {priorities.map((priority) => (
                        <SelectItem key={priority} value={priority} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                          {priority}
                        </SelectItem>
                      ))}
                    </GlassCard>
                  </SelectContent>
                </Select>
              </div>
            </div>
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredAssignments.map((assignment) => (
            <AssignmentItem
              key={assignment.ID}
              assignment={assignment}
              onEdit={onEdit}
              onDelete={(assignment) => onDelete(assignment as models.LocalAssignment)}
              onOpenEdit={(assignment) => onOpenEdit(assignment as models.LocalAssignment)}
              disabled={isLoading}
              variant="outline"
            />
          ))}
        </div>
      )
      }

    </div >
  )
}
