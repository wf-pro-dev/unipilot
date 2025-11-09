"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
  const statuses = Array.from(new Set((assignments || []).map((assignment) => assignment.StatusName)))
  const priorities = Array.from(new Set((assignments || []).map((assignment) => assignment.Priority)))

  const hasActiveFilters = selectedCourse !== "all" || selectedStatus !== "all" || selectedPriority !== "all" || searchTerm !== ""


  // Apply basic filters (simplified for now)
  const filteredAssignments = useMemo(() => (assignments || []).filter((assignment) => {
    const matchesSearch =
      assignment.Title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      assignment.Course?.Name.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesCourse = selectedCourse === "all" || assignment.Course?.Code === selectedCourse
    const matchesStatus = selectedStatus === "all" || assignment.StatusName === selectedStatus
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
    <div className="space-y-4">
      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <Input
                    placeholder="Search assignments by title or course..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10 bg-gray-800/50 border-gray-600"
                  />
                </div>
              </div>

              <div className="flex items-center gap-4">



                <div className="w-36">
                  <CoursesSelect
                    value={selectedCourse}
                    onValueChange={onCourseChange}
                  >
                    <SelectItem value="all">All Courses</SelectItem>
                  </CoursesSelect>
                </div>


                <Select value={selectedStatus} onValueChange={onStatusChange}>
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


                <Select value={selectedPriority} onValueChange={onPriorityChange}>
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
                  {selectedCourse !== "all" && (
                    <Badge variant="secondary" className="bg-green-500/20 text-green-400">
                      {selectedCourse}
                    </Badge>
                  )}
                  {selectedStatus !== "all" && (
                    <Badge variant="secondary" className="bg-purple-500/20 text-purple-400">
                      {selectedStatus}
                    </Badge>
                  )}
                  {selectedPriority !== "all" && (
                    <Badge variant="secondary" className="bg-orange-500/20 text-orange-400">
                      {selectedPriority}
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
        </CardContent>
      </Card>

      {filteredAssignments.length === 0 ? (
        <div className="flex items-center justify-center h-32 text-gray-400">
          <div className="text-center">
            <List className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p>No assignments found</p>
          </div>
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
      )}

    </div>
  )
}
