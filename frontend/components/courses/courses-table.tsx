import { useState } from "react";
import { BookOpen, List, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { Filter, X } from "lucide-react";
import { models } from "@/wailsjs/go/models";
import { GlassCard } from "../ui/glass-card";
import { useRouter } from "next/navigation";
import { EmptyState } from "../ui/empty-state";
import { CourseItem } from "./course-item";
import { Scroll } from "../core/scroll";
import { CardContent } from "../ui/card";

interface Filter {
    semester: string | null
    instructor: string | null
}

interface CoursesTableProps {
    courses: models.LocalCourse[]
    filter: Filter
}


export default function CoursesTable({ courses, filter }: CoursesTableProps) {
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
        <div className="flex flex-col h-full min-h-0 space-y-4">

            <GlassCard variant="board" className="flex-grow-0 flex-row">
                <CardContent className="flex-1 p-2">
                    <div className="flex flex-col lg:flex-row lg:items-center space-x-2">
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
                </CardContent>
            </GlassCard>

            <div className="flex h-full min-h-0">
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
                            onClick={clearFilters}
                            buttonText="Clear Filters"
                        />

                    </div>
                )}
            </div>

        </div>
    )
}

