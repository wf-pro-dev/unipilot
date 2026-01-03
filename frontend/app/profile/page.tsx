"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { GlassCard } from "@/components/ui/glass-card"
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
import { useCurrentUser, useGetAvatarUrl, useUpdateUser, useUploadProfilePicture } from "@/hooks/use-auth"
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
  const { data: avatarUrl } = useGetAvatarUrl()

  const updateMutation = useUpdateCourse()
  const deleteMutation = useDeleteCourse()
  const uploadProfilePictureMutation = useUploadProfilePicture()

  const finalAvatarUrl = avatarUrl || "/placeholder.svg?height=40&width=40"

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
    <div className="">
      <div className="max-w-6xl mx-auto relative z-10">
        <div className="mb-8">
          <h1 className="text-h1 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            My Profile
          </h1>
          <p className="text-body-small text-gray-400 mt-3">Manage your account information and academic progress</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Profile Info */}
          <div className="lg:col-span-1 space-y-6">
            {/* Profile Card */}
            <GlassCard variant="board">
              <CardContent className="p-6">
                <div className="flex flex-col items-center space-y-6">
                  <div className="relative">
                    <Avatar className="h-28 w-28 border-2 border-white/10 shadow-lg shadow-black/20">
                      <AvatarImage src={finalAvatarUrl} alt={user?.Username} />
                      <AvatarFallback className="text-h3 bg-gradient-to-br from-blue-500/20 to-purple-500/20 text-white">
                        {user?.Username
                          .split(" ")
                          .map((n) => n[0])
                          .join("")}
                      </AvatarFallback>
                    </Avatar>
                    <Button
                      size="sm"
                      className="absolute -bottom-1 -right-1 h-9 w-9 rounded-full p-0 bg-white/10 border border-white/10 hover:bg-white/20 hover:border-white/20 transition-smooth"
                      disabled={!isEditing}
                      onClick={() => uploadProfilePictureMutation.mutate()}
                    >
                      <Camera className="h-4 w-4 text-white" />
                    </Button>
                  </div>

                  <div className="text-center space-y-1">
                    <h2 className="text-h3 text-white font-semibold tracking-tight">{user?.Username}</h2>
                    <p className="text-body-small text-blue-400">{user?.Email}</p>
                  </div>

                  <Separator orientation="horizontal" className="w-20 bg-white/10" />

                  <div className="flex flex-row w-full justify-evenly text-center items-center">
                    <div className="space-y-1">
                      <div className="text-h4 text-white font-semibold">{followers?.length || 0}</div>
                      <div className="text-caption text-gray-400 uppercase tracking-wider">Followers</div>
                    </div>

                    <Separator orientation="vertical" className="h-12 bg-white/10" />

                    <div className="space-y-1">
                      <div className="text-h4 text-white font-semibold">{following?.length || 0}</div>
                      <div className="text-caption text-gray-400 uppercase tracking-wider">Following</div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </GlassCard>

            {/* Stats Card */}
            <GlassCard variant="board">
              <CardHeader className="pb-4">
                <CardTitle className="flex items-center space-x-2 text-white">
                  <div className="p-1.5 rounded-lg bg-green-500/10 border border-green-500/20">
                    <TrendingUp className="h-4 w-4 text-green-400" />
                  </div>
                  <span className="text-h4 font-semibold">Academic Stats</span>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-caption text-gray-400 uppercase tracking-wider font-medium">Assignments Completed</span>
                    <span className="text-body-small text-white font-semibold">{completedAssignments?.length || 0}/{assignments?.length || 0}</span>
                  </div>
                  <div className="space-y-1.5">
                    <Progress 
                      value={(completedAssignments?.length || 0) / ((assignments?.length || 0) || 1) * 100} 
                      className="h-2 bg-white/10"
                    />
                    <div className="flex justify-end">
                      <span className="text-caption text-gray-400 font-medium">
                        {Math.round(((completedAssignments?.length || 0) / ((assignments?.length || 0) || 1)) * 100)}%
                      </span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </GlassCard>
          </div>

          {/* Right Column - Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Personal Information */}
            <GlassCard variant="board">
              <form onSubmit={handleSubmit}>
                <CardHeader className="pb-4">
                  <CardTitle className="flex items-center justify-between text-white">
                    <div className="flex items-center space-x-2">
                      <div className="p-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20">
                        <User className="h-4 w-4 text-blue-400" />
                      </div>
                      <span className="text-h4 font-semibold">Personal Information</span>
                    </div>
                    <div>
                      {!isEditing ? (
                        <Button
                          variant="outline"
                          size="sm"
                          className="bg-white/5 border-white/10 hover:bg-white/10 hover:border-white/20 text-gray-300 hover:text-white transition-smooth"
                          onClick={() => setIsEditing(true)}
                        >
                          <Edit className="h-4 w-4 mr-1.5" />
                          <span className="text-body-small">Edit</span>
                        </Button>
                      ) : (
                        <div className="flex items-center space-x-2">
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-red-400 bg-red-500/10 border-red-500/20 hover:bg-red-500/20 hover:border-red-500/30 transition-smooth"
                            onClick={handleCancel}
                          >
                            <X className="h-4 w-4 mr-1.5" />
                            <span className="text-body-small">Cancel</span>
                          </Button>
                          <Button
                            type="submit"
                            variant="outline"
                            size="sm"
                            className="text-blue-400 bg-blue-500/10 border-blue-500/20 hover:bg-blue-500/20 hover:border-blue-500/30 transition-smooth"
                          >
                            <Save className="h-4 w-4 mr-1.5" />
                            <span className="text-body-small">Save</span>
                          </Button>
                        </div>
                      )}
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-2">
                        <label htmlFor="email" className="text-caption text-gray-400 uppercase tracking-wider font-medium block">Email</label>
                        {isEditing ? (
                          <Input
                            id="email"
                            value={editedData?.Email}
                            placeholder={user?.Email}
                            onChange={(e) => setEditedData({ ...editedData, Email: e.target.value })}
                            className="bg-white/5 border-white/10 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 text-white placeholder:text-gray-500 transition-smooth"
                          />
                        ) : (
                          <div className="flex items-center space-x-2.5 bg-white/5 border border-white/5 rounded-lg px-3 py-2.5">
                            <Mail className="h-4 w-4 text-blue-400 flex-shrink-0" />
                            <span className="text-body text-white">{user?.Email}</span>
                          </div>
                        )}
                      </div>

                      <div className="space-y-2">
                        <label className="text-caption text-gray-400 uppercase tracking-wider font-medium block">University</label>
                        {!isEditing ? (
                          <div className="flex items-center space-x-2.5 bg-white/5 border border-white/5 rounded-lg px-3 py-2.5">
                            <GraduationCap className="h-4 w-4 text-purple-400 flex-shrink-0" />
                            <span className="text-body text-white">{user?.University}</span>
                          </div>
                        ) : (
                          <Select value={editedData?.University} onValueChange={(value) => setEditedData({ ...editedData, University: value })}>
                            <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500/50 text-white transition-smooth">
                              <SelectValue placeholder={editedData?.University} />
                            </SelectTrigger>
                            <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                              {universities.map((university) => (
                                <SelectItem key={university} value={university} className="text-white focus:bg-white/10">{university}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                      <div className="space-y-2">
                        <label className="text-caption text-gray-400 uppercase tracking-wider font-medium block">Joined</label>
                        <div className="flex items-center space-x-2.5 bg-white/5 border border-white/5 rounded-lg px-3 py-2.5">
                          <Calendar className="h-4 w-4 text-purple-400 flex-shrink-0" />
                          <span className="text-body text-white">{format(new Date(user?.CreatedAt || new Date()), "MMMM d, yyyy")}</span>
                        </div>
                      </div>

                      <div className="space-y-2">
                        <label className="text-caption text-gray-400 uppercase tracking-wider font-medium block">Semester</label>
                        {!isEditing ? (
                          <Badge variant="outline" className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg">
                            <span className="text-body-small text-white font-medium">{user?.Semester}</span>
                          </Badge>
                        ) : (
                          <Select value={editedData?.Semester} onValueChange={(value) => setEditedData({ ...editedData, Semester: value })}>
                            <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500/50 text-white transition-smooth">
                              <SelectValue placeholder={editedData?.Semester} />
                            </SelectTrigger>
                            <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                              {semesters.map((semester) => (
                                <SelectItem key={semester.value} value={semester.value} className="text-white focus:bg-white/10">{semester.name}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      </div>

                      <div className="space-y-2">
                        <label className="text-caption text-gray-400 uppercase tracking-wider font-medium block">Year</label>
                        {!isEditing ? (
                          <Badge variant="outline" className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg">
                            <span className="text-body-small text-white font-medium">{user?.Year}</span>
                          </Badge>
                        ) : (
                          <Select value={editedData?.Year} onValueChange={(value) => setEditedData({ ...editedData, Year: value })}>
                            <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500/50 text-white transition-smooth">
                              <SelectValue placeholder={editedData?.Year} />
                            </SelectTrigger>
                            <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                              {years.map((year) => (
                                <SelectItem key={year} value={year} className="text-white focus:bg-white/10">{year}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </form>
            </GlassCard>

            {/* Current Courses */}
            <div className="space-y-4">
              <GlassCard variant="board">
                <CardContent className="p-5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2.5">
                      <div className="p-1.5 rounded-lg bg-green-500/10 border border-green-500/20">
                        <BookOpen className="h-4 w-4 text-green-400" />
                      </div>
                      <span className="text-h4 font-semibold text-white">Courses</span>
                    </div>
                    <div className="flex items-center">
                      {coursePages.length > 1 && (
                        <Badge variant="outline" className="flex items-center px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg">
                          <p className="text-body-small text-gray-400">{selectedIndex + 1} / <span className="text-white font-semibold">{coursePages.length}</span></p>
                        </Badge>
                      )}
                    </div>
                  </div>
                </CardContent>
              </GlassCard>

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
                      className="left-0 absolute rounded-full top-1/2 -translate-y-1/2 -translate-x-1/2 z-10 h-9 w-9 bg-white/5 border border-white/10 hover:bg-white/10 hover:border-white/20 transition-smooth"
                      onClick={scrollPrev}
                    >
                      <ChevronLeft className="h-4 w-4 text-white" />
                    </Button>

                    <Button
                      variant="outline"
                      size="icon"
                      className="right-0 absolute rounded-full top-1/2 -translate-y-1/2 translate-x-1/2 z-10 h-9 w-9 bg-white/5 border border-white/10 hover:bg-white/10 hover:border-white/20 transition-smooth"
                      onClick={scrollNext}
                    >
                      <ChevronRight className="h-4 w-4 text-white" />
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
          onLinkRequest={() => {}}
        />
      </div>
    </div>
  )
}
