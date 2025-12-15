import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectGroup, SelectLabel } from "@/components/ui/select"
import { SelectProps } from "@radix-ui/react-select"
import { useAuthContext } from "../provider/auth-provider"
import { course as courseType } from "@/wailsjs/go/models"
import { GlassCard } from "../ui/glass-card"


export function CoursesSelect({ value, onValueChange, children }: SelectProps) {
    const { courses } = useAuthContext()

    const getCoursesBySemester = () => {
        const data: Record<string, courseType.LocalCourse[]> = {}
        for (const course of courses ?? []) {
            ; (data[course.Semester] ??= []).push(course)
        }
        return data
    }

    var coursesBySemester = getCoursesBySemester()

    var selectedCourse = courses?.find((course) => course.Code === value)
    console.log("selectedCourse", value, selectedCourse)

    return (
        <Select
            value={value}
            onValueChange={onValueChange}
        >
            <SelectTrigger className="bg-white/5 border-white/10">
                {selectedCourse && (
                    <div className="flex items-center gap-2">
                        <div className={` h-2 w-2 rounded-full ${selectedCourse?.Color}`} />
                        <p className="line-clamp-1">
                            {selectedCourse?.Code}
                        </p>
                    </div>
                )}

                {!selectedCourse && (
                    <SelectValue placeholder="Select course" />
                )}
            </SelectTrigger>

            <SelectContent>
                <GlassCard>
                    {children}
                    {Object.keys(coursesBySemester)?.map((semester) => (
                        <SelectGroup>
                            <SelectLabel>{semester}</SelectLabel>
                            {coursesBySemester[semester].map((course) => (
                                <SelectItem key={course.Code} value={course.Code}>
                                    <div className="flex items-center gap-2">
                                        <div className={` h-2 w-2 rounded-full ${course.Color}`} />
                                        {course.Code} - {course.Name}
                                    </div>
                                </SelectItem>
                            ))}
                        </SelectGroup>
                    ))}
                </GlassCard>
            </SelectContent>
        </Select>
    )
}