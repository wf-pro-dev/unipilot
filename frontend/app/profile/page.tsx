"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Progress } from "@/components/ui/progress"
import {
  User,
  MapPin,
  Calendar,
  BookOpen,
  GraduationCap,
  Mail,
  Phone,
  Edit,
  Save,
  Camera,
  TrendingUp,
  ChevronLeft,
  ChevronRight,
  X,
} from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { useCurrentUser, useUpdateUser } from "@/hooks/use-auth"
import { useAssignments, useCompletedAssignments } from "@/hooks/use-assignments"
import { useCourses, useDeleteCourse } from "@/hooks/use-courses"
import { CourseItem } from "@/components/courses/course-item"
import { CourseDetailsModal } from "@/components/courses/course-details-modal"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { differenceInDays, format, isSameDay } from "date-fns"
import { useUpdateCourse } from "@/hooks/use-courses"
import { course, user } from "@/wailsjs/go/models"
import { CourseDeleteDialog } from "../courses/course-delete-dialog"
import useEmblaCarousel from "embla-carousel-react"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { toast } from "sonner"
import { useAuthContext } from "@/components/provider/auth-provider"



const years = [
  "Freshman",
  "Sophomore",
  "Junior",
  "Senior",
  "Graduate"
]

const semesters = [
  { name: "SUMMER 2028", value: "SUMMER 2028" },
  { name: "SPRING 2028", value: "SPRING 2028" },
  { name: "FALL 2027", value: "FALL 2027" },
  { name: "SUMMER 2027", value: "SUMMER 2027" },
  { name: "SPRING 2027", value: "SPRING 2027" },
  { name: "FALL 2026", value: "FALL 2026" },
  { name: "SUMMER 2026", value: "SUMMER 2026" },
  { name: "SPRING 2026", value: "SPRING 2026" },
  { name: "FALL 2025", value: "FALL 2025" },
  { name: "SUMMER 2025", value: "SUMMER 2025" },
  { name: "SPRING 2025", value: "SPRING 2025" },
  { name: "FALL 2024", value: "FALL 2024" },
]

const universities = [
  "Austin Community College",
  "Harvard University",
  "Stanford University",
  "MIT",
  "University of California, Berkeley",
  "University of Oxford",
  "University of Cambridge",
  "Yale University",
  "Princeton University",
  "Columbia University",
  "University of Chicago",
  "Other"
]

// Common languages
const languages = [
  { code: "en", name: "English" },
  { code: "es", name: "Spanish" },
  { code: "fr", name: "French" },
  { code: "de", name: "German" },
  { code: "it", name: "Italian" },
  { code: "pt", name: "Portuguese" },
  { code: "zh", name: "Chinese" },
  { code: "ja", name: "Japanese" },
  { code: "ko", name: "Korean" },
  { code: "ar", name: "Arabic" }
]

export default function ProfilePage() {
  const [isEditing, setIsEditing] = useState(false)
  const [selectedCourseId, setSelectedCourseId] = useState<number | null>(null)
  const [selectedDeleteCourseId, setSelectedDeleteCourseId] = useState<number | null>(null)
  const [selectedIndex, setSelectedIndex] = useState(0)

  const { data: user } = useCurrentUser()
  const [editedData, setEditedData] = useState({
    Email: user?.Email || "",
    Username: user?.Username || "",
    University: user?.University || "",
    Semester: user?.Semester || "",
    Year: user?.Year || "",
    Language: user?.Language || "",
    Avatar: user?.Avatar || "",
  })
  const { data: assignments } = useAssignments()
  const { data: completedAssignments } = useCompletedAssignments()
  const { data: courses } = useCourses()
  const { followers, following } = useAuthContext()
  const { mutate: updateUser } = useUpdateUser()

  const updateMutation = useUpdateCourse()
  const deleteMutation = useDeleteCourse()

  const completionPercentage = ((completedAssignments || []).length / (assignments || []).length) * 100

  const coursesPerPage = 4
  const coursePages = []
  for (let i = 0; i < (courses?.length || 0); i += coursesPerPage) {
    coursePages.push(courses?.sort((a, b) => differenceInDays(b.StartDate, a.StartDate)).slice(i, i + coursesPerPage))
  }

  useEffect(() => {
    if (!user) return
    setEditedData({
      Email: user?.Email || "",
      Username: user?.Username || "",
      University: user?.University || "",
      Semester: user?.Semester || "",
      Year: user?.Year || "",
      Language: user?.Language || "",
      Avatar: user?.Avatar || "",
    })
  }, [user])

  const handleCancel = () => {
    setEditedData({
      Email: user?.Email || "",
      Username: user?.Username || "",
      University: user?.University || "",
      Semester: user?.Semester || "",
      Year: user?.Year || "",
      Language: user?.Language || "",
      Avatar: user?.Avatar || "",
    })
    setIsEditing(false)
  }

  const handleCourseClick = (course: course.LocalCourse) => {
    setSelectedCourseId(course.ID)
  }

  const handleEditCourse = async (courseData: course.LocalCourse, column: string, value: string) => {
    const message = "course " + courseData.Code + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      course: courseData,
      column,
      value
    })
  }

  const handleDeleteCourseClick = (course: course.LocalCourse) => {
    setSelectedDeleteCourseId(course.ID)
  }

  const handleDeleteCourse = async (course: course.LocalCourse) => {
    const message = "course " + course.Code + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteMutation.mutate(course)
  }

  // Embla Carousel setup
  const [emblaRef, emblaApi] = useEmblaCarousel({
    align: 'start',
    containScroll: 'trimSnaps',
    dragFree: false,
    skipSnaps: false
  })

  // Track current slide
  const onSelect = useCallback(() => {
    if (!emblaApi) return
    setSelectedIndex(emblaApi.selectedScrollSnap())
  }, [emblaApi])

  useEffect(() => {
    if (!emblaApi) return
    onSelect()
    emblaApi.on('select', onSelect)
    return () => {
      emblaApi.off('select', onSelect)
    }
  }, [emblaApi, onSelect])

  const scrollPrev = useCallback(() => {
    if (emblaApi) emblaApi.scrollPrev()
  }, [emblaApi])

  const scrollNext = useCallback(() => {
    if (emblaApi) emblaApi.scrollNext()
  }, [emblaApi])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const key_to_column = {
      Email: "email",
      Username: "username",
      University: "university",
      Semester: "semester",
      Year: "year",
      Language: "language",
      Avatar: "avatar",
    }

    var isChanged = false


    // Process updates sequentially to avoid race conditions
    for (const [key, value] of Object.entries(editedData)) {
      const column = key_to_column[key as keyof typeof key_to_column]
      if (value != user?.[key as keyof user.User]) {
        const message = "user " + user?.Username + " " + column + ": " + user?.[key as keyof user.User] + " -> " + value
        LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
        
        // Wait for each update to complete before proceeding
        await updateUser({ column, key, value })
        isChanged = true
      } else {  
        const message = `No changes to ${column} value: ${value}`
        LogInfo(message)
      }
    }

    if (isChanged) {
      toast.success("Profile updated successfully")
    } else {
      toast.info("No changes to profile")
    }

    setIsEditing(false)
  }

  return (
    <div className="page">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>

      <div className="max-w-6xl mx-auto relative z-10">
        <div className="mb-8">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            My Profile
          </h1>
          <p className="text-gray-400 mt-2">Manage your account information and academic progress</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Profile Info */}
          <div className="lg:col-span-1 space-y-6">
            {/* Profile Card */}
            <Card className="glass border-0">
              <CardContent className="p-6">
                <div className="flex flex-col items-center space-y-4">
                  <div className="relative">
                    <Avatar className="h-24 w-24">
                      <AvatarImage src={"/placeholder.svg"} alt={user?.Username} />
                      <AvatarFallback className="text-lg">
                        {user?.Username
                          .split(" ")
                          .map((n) => n[0])
                          .join("")}
                      </AvatarFallback>
                    </Avatar>
                    <Button
                      size="sm"
                      className="absolute -bottom-2 -right-2 h-8 w-8 rounded-full p-0"
                      disabled={!isEditing}
                    >
                      <Camera className="h-4 w-4" />
                    </Button>
                  </div>


                  <div className="text-center">
                    <h2 className="text-xl font-bold text-white">{user?.Username}</h2>
                    <p className="text-blue-400">{user?.Email}</p>
                  </div>

                  <Separator orientation="horizontal" className="w-20 bg-gray-600" />

                  <div className="flex flex-row w-full justify-evenly text-center items-center">
                    <div>
                      <div className="text-xl font-bold text-white">{followers?.length}</div>
                      <div className="text-xs text-gray-400">Followers</div>
                    </div>

                    <Separator orientation="vertical" className="h-10 bg-gray-600" />

                    <div>
                      <div className="text-xl font-bold text-white">{following?.length}</div>
                      <div className="text-xs text-gray-400">Following</div>
                    </div>

                  </div>

                </div>
              </CardContent>
            </Card>

            {/* Stats Card */}
            <Card className="glass border-0">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2 text-white">
                  <TrendingUp className="h-5 w-5 text-green-400" />
                  <span>Academic Stats</span>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-gray-400">Assignments Completed</span>
                    <span className="text-sm text-white">{completedAssignments?.length}/{assignments?.length || 1}</span>
                  </div>
                  <Progress value={(completedAssignments?.length / (assignments?.length || 0)) * 100} className="h-2" />
                </div>


              </CardContent>
            </Card>


          </div>

          {/* Right Column - Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Personal Information */}
            <Card className="glass border-0">
              <form onSubmit={handleSubmit}>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between text-white">
                    <div className="flex items-center space-x-2">
                      <User className="h-5 w-5 text-blue-400" />
                      <span>Personal Information</span>
                    </div>
                    <div>
                      {!isEditing ? (
                        <Button
                          variant="outline"
                          size="sm"
                          className="flex-1 bg-transparent border-gray-600"
                          onClick={() => setIsEditing(true)}>
                          <Edit className="h-4 w-4 mr-1" />
                          Edit
                        </Button>
                      ) : (
                        <div className="flex items-center space-x-2">
                          <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10"
                            onClick={handleCancel}
                          >
                            <X className="h-4 w-4 mr-1" />
                            Cancel
                          </Button>
                        <Button
                          type="submit"
                          variant="outline"
                          size="sm"
                          className="flex-1 text-blue-400 bg-transparent border-blue-600 hover:bg-blue-600/10"
                        >
                          <Save className="h-4 w-4 mr-1" />
                          Save
                        </Button>
                        </div>
                      )}
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <label htmlFor="email" className="text-sm font-medium text-gray-400 block mb-2">Email</label>
                        {isEditing ? (
                          <Input
                            id="email"
                            value={editedData?.Email}
                            placeholder={user?.Email}
                            onChange={(e) => setEditedData({ ...editedData, Email: e.target.value })}
                            className="bg-gray-800/50 border-gray-600"
                          />
                        ) : (
                          <div className="flex items-center space-x-2 text-white">
                            <Mail className="h-4 w-4 text-blue-400" />
                            <span>{user?.Email}</span>
                          </div>
                        )}
                      </div>

                      <div>
                        <label className="text-sm font-medium text-gray-400 block mb-2">University</label>
                        {!isEditing ? (
                          <div className="text-white">{user?.University}</div>
                        ) : (
                          <Select value={editedData?.University} onValueChange={(value) => setEditedData({ ...editedData, University: value })}>
                            <SelectTrigger className="bg-gray-800/50 border-gray-600">
                              <SelectValue placeholder={editedData?.University} />
                            </SelectTrigger>
                            <SelectContent className="glass">
                              {universities.map((university) => (
                                <SelectItem key={university} value={university}>{university}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      </div>


                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">


                      <div>
                        <label className="text-sm font-medium text-gray-400 block mb-2">Joined</label>
                        <div className="flex items-center space-x-2 text-white">
                          <Calendar className="h-4 w-4 text-purple-400" />
                          <span>{format(new Date(user?.CreatedAt || new Date()), "MMMM d, yyyy")}</span>
                        </div>
                      </div>

                      <div>
                        <label className="text-sm font-medium text-gray-400 block mb-2">Semester</label>
                        {!isEditing ? (
                          <Badge variant="outline" className="px-2 py-1 bg-gray-800/50 border border-gray-600 rounded-full">
                            <span className="text-xs text-white font-medium">{user?.Semester}</span>
                          </Badge>
                        ) : (
                          <Select value={editedData?.Semester} onValueChange={(value) => setEditedData({ ...editedData, Semester: value })}>
                            <SelectTrigger className="bg-gray-800/50 border-gray-600">
                              <SelectValue placeholder={editedData?.Semester} />
                            </SelectTrigger>
                            <SelectContent className="glass">
                              {semesters.map((semester) => (
                                <SelectItem key={semester.value} value={semester.value}>{semester.name}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      </div>

                      <div>
                        <label className="text-sm font-medium text-gray-400 block mb-2">Year</label>
                        {!isEditing ? (
                          <Badge variant="outline" className=" px-2 py-1 bg-gray-800/50 border border-gray-600 rounded-full">
                            <span className="text-xs text-white font-medium">{user?.Year}</span>
                          </Badge>
                        ) : (
                          <Select value={editedData?.Year} onValueChange={(value) => setEditedData({ ...editedData, Year: value })}>
                            <SelectTrigger className="bg-gray-800/50 border-gray-600">
                              <SelectValue placeholder={editedData?.Year} />
                            </SelectTrigger>
                            <SelectContent className="glass">
                              {years.map((year) => (
                                <SelectItem key={year} value={year}>{year}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      </div>

                    </div>
                  </div>
                </CardContent>
              </form>
            </Card>

            {/* Current Courses */}
            <div className="space-y-2">
              <Card className="flex items-center justify-between text-white p-4 border-0 glass">
                <div className="flex items-center space-x-2">
                  <BookOpen className="h-5 w-5 text-green-400" />
                  <span>Courses</span>
                </div>
                <div className="flex items-center">
                  {coursePages.length > 1 && (
                    <Badge variant="outline" className="flex items-center p-2 bg-gray-800/50 border border-gray-600 rounded-full">
                      <p className="text-sm text-muted-foreground">{selectedIndex + 1} / <span className="text-white">{coursePages.length}</span></p>
                    </Badge>
                  )}
                </div>
              </Card>

              <div className="flex flex-col relative">

                <div className="overflow-hidden" ref={emblaRef}>
                  <div className="flex">
                    {coursePages.map((page, pageIndex) => {
                      return (
                        <div
                          key={pageIndex}
                          className="flex-none w-full min-w-0"
                        >
                          <div className="grid grid-cols-2 gap-4">
                            {page?.map((course) => {
                              return (
                                <CourseItem
                                  course={course}
                                  onEdit={() => { }}
                                  onDelete={handleDeleteCourseClick}
                                  onCourseClick={handleCourseClick}
                                />
                              )
                            })}
                          </div>
                        </div>
                      )
                    })}
                  </div>

                </div>

                {coursePages.length > 1 && (
                  <>
                    <Button
                      variant="outline"
                      size="icon"
                      className="left-0 absolute rounded-full top-1/2 -translate-y-1/2 -translate-x-1/2 z-10 h-8 w-8 bg-gray-800/50 border border-gray-600"
                      onClick={scrollPrev}
                    >
                      <ChevronLeft className="h-4 w-4" />
                    </Button>

                    <Button
                      variant="outline"
                      size="icon"
                      className="right-0 absolute rounded-full top-1/2 -translate-y-1/2 translate-x-1/2 z-10 h-8 w-8 bg-gray-800/50 border border-gray-600"
                      onClick={scrollNext}
                    >
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </>
                )}

              </div>


            </div>
          </div>
        </div>

        <CourseDeleteDialog
          isOpen={!!selectedDeleteCourseId}
          onClose={() => setSelectedDeleteCourseId(null)}
          courseId={selectedDeleteCourseId}
          courses={courses || []}
          onDelete={handleDeleteCourse}
        />

        <CourseDetailsModal
          isOpen={!!selectedCourseId}
          courseId={selectedCourseId}
          courses={courses || []}
          onClose={() => setSelectedCourseId(null)}
          onEdit={handleEditCourse}
          onDelete={handleDeleteCourseClick}
        />
      </div>
    </div>
  )
}
