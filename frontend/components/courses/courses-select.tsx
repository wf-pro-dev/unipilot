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
  
  interface CourseData {
    code: string
    id: number
    remoteId: number
    name: string
    color: string
    semester: string
  }
  
  interface CoursesSelectProps {
    value: string
    onValueChange: (value: string, courseData?: CourseData) => void
    placeholder?: string
  }
  
  export function CoursesSelect({ 
    value, 
    onValueChange, 
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
    const selectedCourse = courses?.find((course) => course.Code === value)
  
    const handleValueChange = (newValue: string) => {
      const course = courses?.find((c) => c.Code === newValue)
      if (course) {
        const courseData: CourseData = {
          code: course.Code,
          id: course.ID,
          remoteId: course.RemoteID,
          name: course.Name,
          color: course.Color,
          semester: course.Semester
        }
        onValueChange(newValue, courseData)
      } else {
        onValueChange(newValue)
      }
    }
  
    return (
      <Select value={value} onValueChange={handleValueChange}>
        <SelectTrigger className="bg-white/5 border-white/10 h-11 rounded-xl">
          {selectedCourse ? (
            <div className="flex items-center gap-2 w-full">
              <div className={`h-3 w-3 rounded-full ${selectedCourse.Color}`} />
              <div className="flex flex-col items-start flex-1 min-w-0">
                <span className="font-medium text-white truncate">{selectedCourse.Code}</span>
                <span className="text-xs text-gray-400 truncate">{selectedCourse.Name}</span>
              </div>
            </div>
          ) : (
            <SelectValue placeholder={placeholder} />
          )}
        </SelectTrigger>
  
        <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl max-h-80">
          <GlassCard className="border-0 bg-transparent">
            {Object.keys(coursesBySemester).map((semester) => (
              <SelectGroup key={semester} className="space-y-1">
                <SelectLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 px-2 py-1.5">
                  {semester}
                </SelectLabel>
                {coursesBySemester[semester].map((course) => (
                  <SelectItem 
                    key={course.Code} 
                    value={course.Code}
                    className="focus:bg-white/10 focus:text-white cursor-pointer rounded-lg py-2 px-3"
                  >
                    <div className="flex items-center gap-3">
                      <div className={`h-3 w-3 rounded-full ${course.Color}`} />
                        <div className="flex items-center">
                          <span className="font-medium">{course.Code}</span>
                          <span className="text-xs text-gray-400">•</span>
                          <span className="text-sm text-gray-300 truncate">{course.Name}</span>
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