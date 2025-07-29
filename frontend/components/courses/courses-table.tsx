import { useState } from "react";
import { CoursesGrid } from "./courses-grid";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Filter, X } from "lucide-react";
import { Course } from "@/types/models";

interface Filter {
    semester: string | null
    instructor: string | null
  }

interface CoursesTableProps {
    courses: Course[]
    filter: Filter
    onCourseClick: (course: Course) => void
}


export default function CoursesTable({ courses, filter, onCourseClick }: CoursesTableProps) {
    
    const [searchTerm, setSearchTerm] = useState("")
    const [selectedSemester, setSelectedSemester] = useState(filter.semester || "all")
    const [selectedInstructor, setSelectedInstructor] = useState(filter.instructor || "all")
    
    const filteredCourses = (courses || []).filter((course) => {
        const matchesSearch =
            course.Code.toLowerCase().includes(searchTerm.toLowerCase()) ||
            course.Name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            course.Instructor.toLowerCase().includes(searchTerm.toLowerCase())
        const matchesSemester = selectedSemester === "all" || course.Semester === selectedSemester
        const matchesInstructor = selectedInstructor === "all" || course.Instructor === selectedInstructor

        return matchesSearch && matchesSemester && matchesInstructor
    })


  const clearFilters = () => {
    setSearchTerm("")
    setSelectedSemester("all")
    setSelectedInstructor("all")
  }


    const semesters = Array.from(new Set((courses || []).map((course) => course.Semester)))
    const instructors = Array.from(new Set((courses || []).map((course) => course.Instructor)))

    const hasActiveFilters = selectedSemester !== "all" || selectedInstructor !== "all" || searchTerm !== ""
    return (
        <div className="space-y-6">
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
                onCourseClick={onCourseClick}
            />
        </div>
    )
}