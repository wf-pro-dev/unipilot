"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Progress } from "@/components/ui/progress"
import {
  BookOpen,
  Users,
  Calendar,
  GraduationCap,
  MapPin,
  Edit,
  Trash2,
  Plus,
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
import { course as Course } from "@/wailsjs/go/models"
import { useAssignments, useUpdateAssignment } from "@/hooks/use-assignments"
import { formatDeadline } from "@/lib/date-utils"
import { assignment } from "@/wailsjs/go/models"
import { StatusTag } from "@/components/assignments/tags/status-tag"
import { CourseEditDialog } from "./course-edit-dialog"
import { useCallback, useEffect, useMemo, useState } from "react"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCourseNotes, useDeleteNote } from "@/hooks/use-notes"
import { NoteItem } from "../notes/note-item"
import { NoteDetailModal } from "@/components/notes/note-detail-modal"
import { useUpdateNote } from "@/hooks/use-notes"
import { note } from "@/wailsjs/go/models"
import { toast } from "sonner"
import { Input } from "../ui/input"
import { useRouter } from "next/navigation"
import useEmblaCarousel from "embla-carousel-react"
import { LinkRequestModal } from "@/components/community/link-request-modal"

interface CourseDetailsModalProps {
  isOpen: boolean
  onClose: () => void
  courseId: number | null
  courses: Course.LocalCourse[]
  onEdit: (course: Course.LocalCourse, column: string, value: string) => void
  onDelete: (course: Course.LocalCourse) => void
  onLinkRequest: () => void
}

export function CourseDetailsModal({ isOpen, onClose, courseId, courses, onEdit, onDelete, onLinkRequest }: CourseDetailsModalProps) {
  const course = courses.find(c => c.ID === courseId) || null
  if (!course) return null
  const router = useRouter()
  const [activeView, setActiveView] = useState("info")
  const [selectedNoteID, setSelectedNoteID] = useState<number | null>(null)
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState("")


  const { data: assignments, isLoading } = useAssignments()
  const notes = useCourseNotes(course)

  const updateMutation = useUpdateAssignment()
  const updateNote = useUpdateNote()
  const deleteNote = useDeleteNote()

  const [selectedIndex, setSelectedIndex] = useState(0)

  var course_assignments = (assignments || []).filter((assignment: assignment.LocalAssignment) => assignment.Course?.Code === course.Code) || []
  var completed_assignments_count = course_assignments.filter((assignment: assignment.LocalAssignment) => assignment.StatusName === "Done").length
  var completionPercentage = (completed_assignments_count / course_assignments.length) * 100
  var isCompleted = completionPercentage === 100
  const [open, setOpen] = useState(false)

  const handleEditAssignment = async (assignment: assignment.LocalAssignment, column: string, value: string) => {
    console.log("Editing assignment:", assignment)
    const message = "assignment " + assignment.ID + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))

    // Use the optimistic update mutation
    updateMutation.mutate({
      assignment,
      column,
      value
    })
  }
  // Mock additional data
  const courseData = {
    ...course,
    location: "Science Building, Room 204",
    email: "instructor@university.edu",
    office: "Faculty Building, Room 301",
    officeHours: "MW 2:00-4:00 PM",
  }




  const handleDeleteNote = async (note: note.LocalNote) => {
    const message = "note " + note.Title + " deleted"
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    deleteNote.mutate(note, {
      onSuccess: () => {
        toast.success("Note deleted successfully")
      },
      onError: () => {
        toast.error("Note deletion failed")
      }
    })
  }

  const handleEditNote = async (note: note.LocalNote, column: string, value: string) => {
    const message = "note " + note.Title + " " + column + " changed to " + value
    LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
    updateNote.mutate({ note, column, value }, {
      onError: () => {
        toast.error(`Note ${column} update failed`)
      }
    })
  }

  const handleCloseDetailModal = () => {
    setIsDetailModalOpen(false)
    setSelectedNoteID(null)
  }

  const filteredNotes = useMemo(() => {
    return notes.filter((note) => {
      const matchesSearch =
        note.Title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        note.Keywords.toLowerCase().includes(searchTerm.toLowerCase())
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
        <DialogContent className="glass border-0 text-white max-w-xl max-h-[90vh] overflow-y-auto">

          <div className="space-y-4">
            {/* Course Header */}
            <div className="flex justify-between items-start">
              <div className="flex items-center space-x-4">

                <div className={`w-6 h-6 rounded-full ${courseData.Color}`} />
                <div className="space-y-2">
                  <p className="text-sm font-medium text-gray-400">{courseData.Code}</p>
                  <h2 className="text-xl font-bold text-white">{courseData.Name}</h2>
                </div>
              </div>
            </div>

            <Tabs value={activeView} onValueChange={setActiveView} className="w-full">

              <TabsList className="flex flex-row border-0 w-full">
                <TabsTrigger value="info" className="flex w-full justify-center items-center space-y-1 space-x-1 py-2 text-gray-400 hover:text-white data-[state=active]:text-white">
                  <Info className="w-4 h-4" />
                  <span className="text-sm">Information</span>
                </TabsTrigger>
                <TabsTrigger value="assignments" className="flex w-full justify-center items-center space-y-1 border-x-[1px] border-gray-700 space-x-1 py-2 text-gray-400 hover:text-white data-[state=active]:text-white">
                  <FileText className="w-4 h-4" />
                  <span className="text-sm">Assignment</span>
                </TabsTrigger>
                <TabsTrigger value="notes" className="flex w-full justify-center items-center space-y-1 border-gray-700 space-x-1 py-2 text-gray-400 hover:text-white data-[state=active]:text-white">
                  <FileText className="w-4 h-4" />
                  <span className="text-sm">Notes</span>
                </TabsTrigger>
              </TabsList>

              <Separator className="bg-gray-700 mt-3 mb-6" />

              <TabsContent value="info" className="space-y-4">
                {/* Course Info Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-2">



                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <Users className="w-4 h-4 text-blue-400" />
                      <span>Instructor</span>
                    </div>
                    <p className="text-white text-sm">{courseData.Instructor}</p>
                  </div>


                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <Mail className="w-4 h-4 text-orange-400" />
                      <span>Email</span>
                    </div>
                    <p className="text-white text-sm">{courseData.InstructorEmail}</p>
                  </div>


                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <Calendar className="w-4 h-4 text-purple-400" />
                      <span>Schedule</span>
                    </div>
                    <p className="text-white text-sm">{courseData.Schedule}</p>
                  </div>


                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <MapPin className="w-4 h-4 text-green-400" />
                      <span>Location</span>
                    </div>
                    <p className="text-white text-sm">{courseData.Location || "No location"}</p>
                  </div>

                </div>




                <Separator className="bg-gray-700 w-[80%] mx-auto" />


                {/* Credits & Course Dates */}

                <div className="grid grid-cols-2 gap-2 md:grid-cols-4">

                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">

                    <span className="text-sm text-gray-400">Credits</span>

                    <Badge variant="outline" className="border-gray-600">
                      {courseData.Credits} credits
                    </Badge>
                  </div>

                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <span className="text-sm text-gray-400">Semester</span>
                    <p className="font-medium text-white text-xs text-center">{courseData.Semester}</p>
                  </div>

                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <Calendar className="w-4 h-4 text-green-400" />
                      <span>Start Date</span>
                    </div>
                    <p className="font-medium text-white text-xs">{format(courseData.StartDate, "MMM d, yyyy")}</p>
                  </div>

                  <div className="flex flex-col items-center space-y-2 border border-gray-700 p-2 rounded-lg bg-gray-800/50">
                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                      <Calendar className="w-4 h-4 text-green-400" />
                      <span>End Date</span>
                    </div>
                    <p className="font-medium text-white text-xs">{format(courseData.EndDate, "MMM d, yyyy")}</p>
                  </div>

                </div>

              </TabsContent>

              <TabsContent value="assignments" className="space-y-4">
                {course_assignments.length > 0 ? (
                  <div className="space-y-4">
                  
                    <div>
                      <div className="flex justify-between items-center">
                        <label className="block mb-3 text-sm font-medium text-gray-400">Recent Assignments</label>

                        <Link href={`/assignments?view=list&course=${courseData.Code}`}>
                          <Button variant="ghost" size="sm" className="text-blue-400 hover:text-blue-300">
                            View All
                          </Button>
                        </Link>
                      </div>

                      <div className="grid gap-2 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-2">
                        {course_assignments.slice(0, 4).map((assignment, index) => (
                          <div key={index} className="flex items-center p-3 rounded-lg border border-gray-600 bg-gray-800/50">

                            <span className="w-2/3 text-sm text-white line-clamp-2">{assignment.Title}</span>
                            <div className="flex flex-col items-end space-y-2 grow">
                              <span className="text-xs text-gray-400">{formatDeadline(assignment.Deadline)}</span>
                              <StatusTag assignment={assignment} onEdit={handleEditAssignment} />
                            </div>
                          </div>
                        ))}
                      </div>

                    </div>

                    <Separator className="bg-gray-700 w-[80%] mx-auto" />

                    <div className="p-4 border border-gray-700 bg-gray-800/50 rounded-lg">
                      <div className="space-y-3">
                        <div className="flex justify-between items-center">
                          <div className="flex items-center space-x-2 text-sm">
                            <TrendingUp className="w-4 h-4 text-green-400" />
                            <span className={`${isCompleted ? "text-green-400" : "text-white"}`}>
                              {completed_assignments_count} of {course_assignments.length} assignments completed
                            </span>
                          </div>
                          <span className="text-sm text-gray-400">{Math.round(completionPercentage)}%</span>
                        </div>
                        <Progress color={isCompleted ? "green" : "white"} value={completionPercentage} className="h-1.5" />
                      </div>
                    </div>
                  </div>

                ) : (

                  <div className="py-12 text-center">
                    <CheckCircle2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                    <h3 className="text-lg font-medium text-white mb-2">No assignments found</h3>
                    <p className="text-gray-400">Create an assignment to get started</p>
                  </div>

                )}
              </TabsContent>

              <TabsContent value="notes" className="space-y-4">

                <div className="space-y-4">
                  <div className="flex w-full space-x-2 items-center">
                    <div className="relative flex-1">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                      <Input
                        placeholder="Search notes by title or keywords..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10 bg-gray-800/50 border-gray-600"
                      />
                    </div>
                    <div >
                      <Button
                        variant="outline"
                        size="sm"
                        className="w-32 bg-transparent border-gray-600"
                        onClick={(e) => {
                          e.stopPropagation()
                          router.push(`/notes?course=${course.Code}`)
                        }}
                      >
                        <FileText className="mr-2 w-3 h-3" />
                        View All
                      </Button>
                    </div>
                  </div>
                  {filteredNotes.length > 0 ? (
                    <div className="relative">
                      <div className="overflow-hidden" ref={emblaRef}>
                        <div className="flex">
                          {filteredNotes.map((note) => (
                            <div className="flex-none w-full min-w-0" key={note.ID}>
                              <NoteItem note={note} onNoteClick={setSelectedNoteID} onEdit={handleEditNote} onDelete={handleDeleteNote} />
                            </div>
                          ))}
                        </div>
                      </div>
                      {filteredNotes.length > 1 && (
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
                  ) : (
                    <div className="py-8 text-center">
                      <CheckCircle2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                      <h3 className="text-lg font-medium text-white mb-2">No notes found</h3>
                      <p className="text-gray-400">Create a note to get started</p>
                    </div>
                  )}


                </div>



              </TabsContent>
            
            </Tabs>

            <Separator className="bg-gray-700" />

            {/* Actions */}
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="flex-1 bg-transparent border-gray-600"
                onClick={(e) => {
                  e.stopPropagation()
                  setOpen(true)
                }}
              >
                <Edit className="mr-1 w-3 h-3" />
                Edit
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="flex-1 bg-transparent border-gray-600"
                onClick={(e) => {
                  e.stopPropagation()
                  onLinkRequest()
                }}
              >
                <Share className="mr-1 w-3 h-3" />
                Share
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10"
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete(course)
                }}
              >
                <Trash2 className="mr-1 w-3 h-3" />
                Delete
              </Button>

            </div>
          </div>
        </DialogContent>

      </Dialog>
      <CourseEditDialog
        open={open}
        setOpen={setOpen}
        course={course}
        onEdit={onEdit}
      />
      <NoteDetailModal
        key={selectedNoteID}
        noteID={selectedNoteID}
        isOpen={!!selectedNoteID}
        onClose={handleCloseDetailModal}
        onEdit={handleEditNote}
        onDelete={handleDeleteNote}
      />

    </div>
  )
}
