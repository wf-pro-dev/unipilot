import { useState } from "react";
import { CoursesGrid } from "./courses-grid";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Filter, X } from "lucide-react";
import { course as Course } from "@/wailsjs/go/models";
import { Card, CardContent } from "../ui/card";
import { useRouter } from "next/navigation";

interface Filter {
    semester: string | null
    instructor: string | null
}

interface CoursesTableProps {
    courses: Course.LocalCourse[]
    filter: Filter
    onCourseClick: (course: Course.LocalCourse) => void
    onEdit: (course: Course.LocalCourse, column: string, value: string) => void
    onDelete: (course: Course.LocalCourse) => void
}


export default function CoursesTable({ courses, filter, onCourseClick, onEdit, onDelete }: CoursesTableProps) {
    const router = useRouter()
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
        setSelectedSemester("all")
        setSelectedInstructor("all")
        router.push("/courses?view=list")
    }


    const semesters = Array.from(new Set((courses || []).map((course) => course.Semester)))
    const instructors = Array.from(new Set((courses || []).map((course) => course.Instructor)))

    const hasActiveFilters = selectedSemester !== "all" || selectedInstructor !== "all" || searchTerm !== ""
    return (
        <div className="space-y-6">
            <Card className="glass border-0">
                <CardContent className="p-6">
                    <div className="space-y-4">
                        <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-6">
                            <div className="flex-1">
                                <div className="relative">
                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                                    <Input
                                        placeholder="Search courses by code, name or instructor..."
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 bg-gray-800/50 border-gray-600"
                                    />
                                </div>
                            </div>

                            <div className="flex flex-wrap items-center gap-4">

                                <Select value={selectedSemester} onValueChange={onSemesterChange}>
                                    <SelectTrigger className="w-48 bg-gray-800/50 border-gray-600">
                                        <SelectValue />
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


                                <Select value={selectedInstructor} onValueChange={onInstructorChange}>
                                    <SelectTrigger className="w-48 bg-gray-800/50 border-gray-600">
                                        <SelectValue />
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
                                    {selectedSemester !== "all" && (
                                        <Badge variant="secondary" className="bg-green-500/20 text-green-400">
                                            {selectedSemester}
                                        </Badge>
                                    )}
                                    {selectedInstructor !== "all" && (
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
                </CardContent>
            </Card>

            <CoursesGrid
                courses={filteredCourses}
                onCourseClick={onCourseClick}
                onEdit={onEdit}
                onDelete={onDelete}
            />
        </div>
    )
}