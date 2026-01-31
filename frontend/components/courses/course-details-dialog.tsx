"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import {
  BookOpen,
  Users,
  Calendar,
  MapPin,
  Edit,
  Trash2,
  TrendingUp,
  FileText,
  Mail,
  Info,
  CheckCircle2,
  Search,
  Share,
  ChevronLeft,
  ChevronRight,
} from "lucide-react"
import Link from "next/link"
import { models } from "@/wailsjs/go/models"
import { useAssignments, useUpdateAssignment } from "@/hooks/use-assignments"
import { CourseEditDialog } from "./course-edit-dialog"
import { memo, useCallback, useEffect, useMemo, useState } from "react"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCourseNotes, useDeleteNote } from "@/hooks/use-notes"
import { NoteItem } from "../notes/note-item"
import { useUpdateNote } from "@/hooks/use-notes"
import { toast } from "sonner"
import { Input } from "../ui/input"
import { useRouter } from "next/navigation"
import useEmblaCarousel from "embla-carousel-react"
import { useCourse } from "@/hooks/use-courses"
import { AssignmentItem } from "../assignments/assignment-item"
import { useDialogContext } from "../provider/dialog-provider"

interface CourseDetailsDialogProps {
  isOpen: boolean
  onClose: () => void
  courseId: number
  courseRO?: models.Course
  mode?: "default" | "readonly"
}

export function CourseDetailsDialog({
  isOpen,
  onClose,
  courseId,
  courseRO,
  mode = "default",
}: CourseDetailsDialogProps) {


  const { data: courseData } = useCourse(courseId)
  var course = mode == "default" ? courseData as models.LocalCourse : courseRO

  if (!course) return null

  const { SetDialogState } = useDialogContext()

  const router = useRouter()
  const [activeView, setActiveView] = useState("info")
  const [searchTerm, setSearchTerm] = useState("")


  const { data: assignments, isLoading } = useAssignments()
  const notes = useCourseNotes(course as models.LocalCourse)
  const deleteNote = useDeleteNote()

  const [selectedIndex, setSelectedIndex] = useState(0)

  const course_assignments = useMemo(() => {
    return (assignments || []).filter((assignment: models.LocalAssignment) => assignment.CourseCode === course?.Code) || []
  }, [assignments, course])

  const completed_assignments_count = useMemo(() => {
    return course_assignments.filter((assignment: models.LocalAssignment) => assignment.Status === "Done").length
  }, [course_assignments])

  const completionPercentage = useMemo(() => {
    return (completed_assignments_count / course_assignments.length) * 100
  }, [completed_assignments_count, course_assignments])

  const isCompleted = useMemo(() => {
    return completionPercentage === 100
  }, [completionPercentage])


  const handleDeleteNote = async (note: models.LocalNote | models.Note) => {
    const message = "note " + note.Title + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteNote.mutate(note as models.LocalNote, {
      onSuccess: () => {
        toast.success("Note deleted successfully")
      },
      onError: () => {
        toast.error("Note deletion failed")
      }
    })
  }

  const filteredNotes = useMemo(() => {
    return notes.filter((note) => {
      const matchesSearch =
        note.Title.toLowerCase().includes(searchTerm.toLowerCase())
      return matchesSearch
    })
  }, [notes, searchTerm])

  // Embla Carousel setup
  const [emblaRef, emblaApi] = useEmblaCarousel({
    align: 'start',
    containScroll: 'trimSnaps',
    dragFree: false,
    skipSnaps: false
  })

  // Carousel navigation functions
  const scrollPrev = useCallback(() => {
    if (emblaApi) emblaApi.scrollPrev()
  }, [emblaApi])

  const scrollNext = useCallback(() => {
    if (emblaApi) emblaApi.scrollNext()
  }, [emblaApi])

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

  // Reset to first page when filter changes
  useEffect(() => {
    setSelectedIndex(0)
    if (emblaApi) {
      emblaApi.scrollTo(0)
    }
  }, [emblaApi])


  return (
    <div>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="glass border-white/10 text-white max-w-xl max-h-[90vh] overflow-y-auto p-0 gap-0">

          <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
            <div className="flex items-center space-x-4">
              <div className={`w-8 h-8 rounded-full ${course.Color} shadow-lg shadow-black/20`} />
              <div className="space-y-1">
                <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">{course.Code}</p>
                <DialogTitle className="text-h3">{course.Name}</DialogTitle>
              </div>
            </div>
          </DialogHeader>

          <div className="p-6">
            <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl pointer-events-none" />
            <Tabs value={activeView} onValueChange={setActiveView} className="w-full">

              <TabsList className="flex flex-row bg-white/5 p-1 rounded-xl w-full mb-6 border border-white/5">
                <TabsTrigger
                  value="info"
                  className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
                >
                  <Info className="w-4 h-4" />
                  <span className="text-sm font-medium">Info</span>
                </TabsTrigger>
                <TabsTrigger
                  value="assignments"
                  className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
                >
                  <FileText className="w-4 h-4" />
                  <span className="text-sm font-medium">Assignments</span>
                </TabsTrigger>
                <TabsTrigger
                  value="notes"
                  className="flex-1 flex justify-center items-center space-x-2 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
                >
                  <BookOpen className="w-4 h-4" />
                  <span className="text-sm font-medium">Notes</span>
                </TabsTrigger>
              </TabsList>

              <TabsContent value="info" className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-300">
                {/* Key Details: Schedule & Location */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="p-4 rounded-2xl bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 group">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 bg-blue-500/20 rounded-lg text-blue-400 group-hover:bg-blue-500/30 transition-colors">
                        <Calendar className="w-5 h-5" />
                      </div>
                      <span className="text-xs font-medium text-blue-200 uppercase tracking-wider">Schedule</span>
                    </div>
                    <p className="text-h4 text-white leading-tight">{course.Schedule}</p>
                  </div>

                  <div className="p-4 rounded-2xl bg-gradient-to-br from-emerald-500/10 to-teal-500/10 border border-emerald-500/20 group">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 bg-emerald-500/20 rounded-lg text-emerald-400 group-hover:bg-emerald-500/30 transition-colors">
                        <MapPin className="w-5 h-5" />
                      </div>
                      <span className="text-xs font-medium text-emerald-200 uppercase tracking-wider">Location</span>
                    </div>
                    <p className="text-h4 text-white leading-tight">{course.Location || "Online"}</p>
                  </div>
                </div>

                {/* Instructor Section */}
                <div className="p-4 rounded-2xl bg-white/5 border border-white/5 flex items-start gap-4">
                  <div className="p-3 bg-white/5 rounded-full text-gray-400">
                    <Users className="w-6 h-6" />
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Instructor</p>
                    <p className="text-h4 text-white">{course.Instructor}</p>
                    <div className="flex items-center gap-2 text-sm text-blue-400">
                      <Mail className="w-3.5 h-3.5" />
                      <a href={`mailto:${course.InstructorEmail}`} className="hover:underline">{course.InstructorEmail}</a>
                    </div>
                  </div>
                </div>


                {/* Course Metadata Grid */}
                <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider block mb-1.5">Credits</span>
                    <div className="flex">
                      <Badge variant="outline" className="border-white/10 bg-white/5 text-white px-2 py-0.5 text-xs">
                        {course.Credits} credits
                      </Badge>
                    </div>
                  </div>

                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider block mb-1.5">Semester</span>
                    <p className="text-sm font-medium text-white ">{course.Semester}</p>
                  </div>

                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <div className="flex items-center space-x-1.5 mb-1.5">
                      <div className="w-1.5 h-1.5 rounded-full bg-green-400"></div>
                      <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider">Start Date</span>
                    </div>
                    <p className="text-sm font-medium text-white ">{format(course.StartDate, "MMM d, yyyy")}</p>
                  </div>

                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <div className="flex items-center space-x-1.5 mb-1.5">
                      <div className="w-1.5 h-1.5 rounded-full bg-red-400"></div>
                      <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider">End Date</span>
                    </div>
                    <p className="text-sm font-medium text-white ">{format(course.EndDate, "MMM d, yyyy")}</p>
                  </div>
                </div>

              </TabsContent>

              <TabsContent value="assignments" className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-300">
                {course_assignments.length > 0 ? (


                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <label className="text-sm font-medium text-gray-400 uppercase tracking-wider">Recent Assignments</label>

                      <Link href={`/assignments?view=list&course=${course.Code}`}>
                        <Button variant="ghost" size="sm" className="text-blue-400 hover:text-blue-300 hover:bg-blue-400/10 h-8 text-xs">
                          View All
                        </Button>
                      </Link>
                    </div>

                    <div>
                      <div className="space-y-3">
                        <div className="flex justify-between items-center">
                          <div className="flex items-center space-x-2 text-sm">
                            <TrendingUp className="w-4 h-4 text-green-400" />
                            <span className={`${isCompleted ? "text-green-400" : "text-white"} font-medium`}>
                              {completed_assignments_count} of {course_assignments.length} assignments completed
                            </span>
                          </div>
                          <span className="text-sm font-bold text-gray-400">{Math.round(completionPercentage)}%</span>
                        </div>
                        <Progress color={isCompleted ? "green" : "white"} value={completionPercentage} className="h-1.5 bg-white/10" />
                      </div>
                    </div>

                    <div className="grid grid-cols-1">
                      {course_assignments.slice(0, 3).map((assignment, index) => (
                        <AssignmentItem key={assignment.ID} assignmentId={assignment.ID} mode="ghost" />
                      ))}
                    </div>


                  </div>

                ) : (

                  <div className="py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                    <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
                      <CheckCircle2 className="h-8 w-8 text-gray-500" />
                    </div>
                    <h3 className="text-lg font-medium text-white mb-1">No assignments found</h3>
                    <p className="text-gray-400 text-sm">Create an assignment to get started</p>
                  </div>

                )}
              </TabsContent>

              <TabsContent value="notes" className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-300">

                <div className="space-y-4">
                  <div className="flex w-full space-x-3 items-center">
                    <div className="relative flex-1">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                      <Input
                        placeholder="Search notes..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10 bg-white/5 border-white/10    transition-all duration-300 h-10"
                      />
                    </div>
                    <div >
                      <Button
                        variant="outline"
                        size="sm"
                        className="bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-10"
                        onClick={(e) => {
                          e.stopPropagation()
                          router.push(`/notes?course=${course?.Code}`)
                        }}
                      >
                        <FileText className="mr-2 w-3.5 h-3.5" />
                        View All
                      </Button>
                    </div>
                  </div>
                  {filteredNotes.length > 0 ? (
                    <div className="relative group">
                      <div className="overflow-hidden" ref={emblaRef}>
                        <div className="flex -ml-4 py-2">
                          {filteredNotes.map((note) => (
                            <div className="flex-none w-full min-w-0 pl-4" key={note.ID}>
                              <NoteItem note={note} onDelete={handleDeleteNote} mode="default" />
                            </div>
                          ))}
                        </div>
                      </div>
                      {filteredNotes.length > 1 && (
                        <>
                          <Button
                            variant="outline"
                            size="icon"
                            className="left-0 absolute rounded-full top-1/2 -translate-y-1/2 -translate-x-3 z-10 h-8 w-8 bg-black/40 border-white/10 backdrop-blur-sm text-white opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={scrollPrev}
                          >
                            <ChevronLeft className="h-4 w-4" />
                          </Button>

                          <Button
                            variant="outline"
                            size="icon"
                            className="right-0 absolute rounded-full top-1/2 -translate-y-1/2 translate-x-3 z-10 h-8 w-8 bg-black/40 border-white/10 backdrop-blur-sm text-white opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={scrollNext}
                          >
                            <ChevronRight className="h-4 w-4" />
                          </Button>
                        </>
                      )}
                    </div>
                  ) : (
                    <div className="py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                      <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
                        <FileText className="h-8 w-8 text-gray-500" />
                      </div>
                      <h3 className="text-lg font-medium text-white mb-1">No notes found</h3>
                      <p className="text-gray-400 text-sm">Create a note to get started</p>
                    </div>
                  )}


                </div>



              </TabsContent>

            </Tabs>


            {/* Actions */}
            {mode === "default" && (
              <div className="flex gap-3 mt-6">

                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-full"
                  onClick={(e) => {
                    e.stopPropagation()
                    SetDialogState({
                      modelType: "course",
                      dialogType: "edit",
                      id: courseId
                    })
                  }}
                >
                  <Edit className="w-4 h-4" />
                  Edit
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-full"
                  onClick={(e) => {
                    e.stopPropagation()
                    SetDialogState({
                      modelType: "course",
                      dialogType: "linkRequest",
                      id: courseId
                    })
                  }}
                >
                  <Share className="w-4 h-4" />
                  Share
                </Button>
                <Button
                  variant="danger"
                  size="sm"
                  className="rounded-full"
                  onClick={(e) => {
                    e.stopPropagation()
                    SetDialogState({
                      modelType: "course",
                      dialogType: "delete",
                      id: courseId
                    })
                  }}
                >
                  <Trash2 className="w-4 h-4" />
                  Delete
                </Button>

              </div>
            )}
          </div>
        </DialogContent>

      </Dialog>

     

    </div>
  )
}
