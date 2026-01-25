import { useState } from "react";
import { BookOpen, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Filter, X } from "lucide-react";
import { models } from "@/wailsjs/go/models";
import { CardContent } from "../ui/card";
import { GlassCard } from "../ui/glass-card";
import { useRouter } from "next/navigation";
import { EmptyState } from "../ui/empty-state";
import { CourseItem } from "./course-item";

interface Filter {
    semester: string | null
    instructor: string | null
}

interface CoursesTableProps {
    courses: models.LocalCourse[]
    filter: Filter
    onCourseClick: (course: models.LocalCourse) => void
    onEdit: (course: models.LocalCourse, column: string, value: string) => void
    onDelete: (course: models.LocalCourse) => void
}


export default function CoursesTable({ courses, filter, onCourseClick, onEdit, onDelete }: CoursesTableProps) {
    const router = useRouter()
    const [searchTerm, setSearchTerm] = useState("")
    const [selectedSemester, setSelectedSemester] = useState(filter.semester)
    const [selectedInstructor, setSelectedInstructor] = useState(filter.instructor)

    const filteredCourses = (courses || []).filter((course) => {
        const matchesSearch =
            course.Code.toLowerCase().includes(searchTerm.toLowerCase()) ||
            course.Name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            course.Instructor.toLowerCase().includes(searchTerm.toLowerCase())
        const matchesSemester = selectedSemester === null || course.Semester === selectedSemester
        const matchesInstructor = selectedInstructor === null || course.Instructor === selectedInstructor

        return matchesSearch && matchesSemester && matchesInstructor
    })

    const onSemesterChange = (semester: string) => {
        router.push(`/courses?view=list&semester=${semester}&instructor=${selectedInstructor}`)
        setSelectedSemester(semester)
    }
    const onInstructorChange = (instructor: string) => {
        router.push(`/courses?view=list&semester=${selectedSemester}&instructor=${instructor}`)
        setSelectedInstructor(instructor)
    }


    const clearFilters = () => {
        setSearchTerm("")
        setSelectedSemester(null)
        setSelectedInstructor(null)
        router.push("/courses?view=list")
    }

    const semesters = Array.from(new Set((courses || []).map((course) => course.Semester)))
    const instructors = Array.from(new Set((courses || []).map((course) => course.Instructor)))
    const hasActiveFilters = selectedSemester !== null || selectedInstructor !== null || searchTerm !== ""

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
    return (
        <div className="flex flex-col flex-1 space-y-6">
            <div className="flex flex-col gap-2">
                <div className="flex w-full lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-4">
                    <div className="relative flex-1">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                        <Input
                            placeholder="Search courses by code, name or instructor..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="pl-10 bg-white/5 border-white/10 transition-all duration-300 h-10"
                        />
                    </div>

                    <div className="flex flex-wrap items-center gap-4">

                        <Select value={selectedSemester || undefined} onValueChange={onSemesterChange}>
                            <SelectTrigger className="w-48 h-10 bg-white/5 border-white/10 text-gray-400 hover:bg-white/10 focus:bg-white/10 transition-all duration-300">
                                <SelectValue placeholder="All Semester" />
                            </SelectTrigger>
                            <SelectContent className="glass border-gray-600">

                                {semesters.map((semester) => (
                                    <SelectItem key={semester} value={semester}>
                                        {semester}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>




                        <Select value={selectedInstructor || undefined} onValueChange={onInstructorChange}>
                            <SelectTrigger className="w-48 h-10 bg-white/5 border-white/10 text-gray-400 hover:bg-white/10 focus:bg-white/10 transition-all duration-300">
                                <SelectValue placeholder="All Instructors" />
                            </SelectTrigger>
                            <SelectContent className="glass border-gray-600">
                                {instructors.map((instructor) => (
                                    <SelectItem key={instructor} value={instructor}>
                                        {instructor}
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
                            {selectedSemester !== null && (
                                <Badge variant="secondary" className="bg-green-500/20 text-green-400">
                                    {selectedSemester}
                                </Badge>
                            )}
                            {selectedInstructor !== null && (
                                <Badge variant="secondary" className="bg-purple-500/20 text-purple-400">
                                    {selectedInstructor}
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


            {filteredCourses.length === 0 ? (
                <GlassCard variant="board" className="flex-1">
                    <EmptyState
                        icon={BookOpen}
                        title="No courses found"
                        description="Adjust your filters or search terms"
                        className="flex-1 items-center"
                        onClick={clearFilters}
                        buttonText="Clear Filters"
                    />
                </GlassCard>
            ) : (
                <div className="flex-1 grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">

                    {filteredCourses.map((course) => {
                        return (
                            <CourseItem
                                course={course}
                                onEdit={onEdit}
                                onDelete={onDelete}
                            />
                        )
                    })}

                </div>
            )}
        </div>
    )
}

