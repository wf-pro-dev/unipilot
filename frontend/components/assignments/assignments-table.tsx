"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { AssignmentItem } from "./assignment-item"
import { List, Loader2, X } from "lucide-react"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "../ui/button"
import { Input } from "../ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../ui/select"
import { Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"

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
  onAssignmentClick: (assignment: assignment.LocalAssignment) => void
  filter: Filter
  isLoading?: boolean
}

export function AssignmentsTable({
  assignments,
  onToggleComplete,
  onEdit,
  onDelete,
  onAssignmentClick,
  filter,
  isLoading = false
}: AssignmentsTableProps) {

  const [searchTerm, setSearchTerm] = useState("")
  const [selectedCourse, setSelectedCourse] = useState(filter.course || "all")
  const [selectedStatus, setSelectedStatus] = useState(filter.status || "all")
  const [selectedPriority, setSelectedPriority] = useState(filter.priority || "all")

  const courses = Array.from(new Set((assignments || []).map((assignment) => assignment.Course?.Code || "all")))
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

 

  useEffect(() => {
    setSelectedCourse(filter.course || "all")
    setSelectedStatus(filter.status || "all")
    setSelectedPriority(filter.priority || "all")
  }, [filter])  

  const clearFilters = () => {
    setSearchTerm("")
    setSelectedCourse("all")
    setSelectedStatus("all")
    setSelectedPriority("all")
  }

  return (
    <div className="space-y-6">
      <Card className="glass border-0">
        <CardHeader>
          <CardTitle className="flex flex-row  gap-3 mb-6 items-center">
             
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                <Input
                  placeholder="Search assignments..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 glass border-gray-600 bg-white/5"
                />
              </div>

                <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                  <SelectTrigger className="w-[180px] glass border-gray-600 bg-white/5">
                    <SelectValue placeholder="All Courses" />
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

                <Select value={selectedStatus} onValueChange={setSelectedStatus}>
                  <SelectTrigger className="w-[180px] glass border-gray-600 bg-white/5">
                    <SelectValue placeholder="All Statuses" />
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

                <Select value={selectedPriority} onValueChange={setSelectedPriority}>
                  <SelectTrigger className="w-[180px] glass border-gray-600 bg-white/5">
                    <SelectValue placeholder="All Priorities" />
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
       
          </CardTitle>
        </CardHeader>
        <CardContent>
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
                  onAssignmentClick={onAssignmentClick}
                  disabled={isLoading}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
