import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  SelectGroup,
  SelectLabel
} from "@/components/ui/select"
import { useAuthContext } from "../provider/auth-provider"
import { models } from "@/wailsjs/go/models"
import { GlassCard } from "../ui/glass-card"


interface CoursesSelectProps {
  value: string
  onValueChange: (value: string) => void
  selectedCourse: models.LocalCourse | undefined
  placeholder?: string
}

export function CoursesSelect({
  value,
  onValueChange,
  selectedCourse,
  placeholder = "Select course"
}: CoursesSelectProps) {
  const { courses } = useAuthContext()

  const getCoursesBySemester = () => {
    const data: Record<string, models.LocalCourse[]> = {}
    for (const course of courses ?? []) {
      ; (data[course.Semester] ??= []).push(course)
    }
    return data
  }

  const coursesBySemester = getCoursesBySemester()



  return (
    <Select value={value} onValueChange={onValueChange}>
      <SelectTrigger className="bg-white/5 border-white/10 text-gray-400 h-10 rounded-xl">
        {selectedCourse && ( 
          <div className="flex items-center gap-2">
            <div className={` h-2 w-2 rounded-full ${selectedCourse?.Color}`} />
            <p className="line-clamp-1 text-body text-white">
              {selectedCourse?.Code}
            </p>
          </div>
        )}

        {!selectedCourse && (
          <SelectValue placeholder={placeholder} />
        )}
      </SelectTrigger>

      <SelectContent className="max-h-80 bg-transparent border-none">
        <GlassCard variant="board">
          <SelectItem value="all" className="text-text-body">All Courses</SelectItem>
          {Object.keys(coursesBySemester).map((semester) => (
            <SelectGroup key={semester} className="space-y-1">
              <SelectLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 px-2 py-1.5">
                {semester}
              </SelectLabel>
              {coursesBySemester[semester].map((course) => (
                <SelectItem
                  key={course.ID}
                  value={course.ID}
                  className="focus:bg-white/10 focus:text-white cursor-pointer rounded-lg py-2 px-3"
                >
                  <div className="flex items-center gap-3">
                    <div className={`h-3 w-3 rounded-full ${course.Color}`} />
                    <div className="flex items-center gpa-2">
                      <span className="text-body-small">{course.Code}</span>
                      <span className="text-xs text-gray-400">•</span>
                      <span className="text-text-caption">{course.Name}</span>
                    </div>
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