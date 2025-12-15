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
    const message = "note " + note.title + " deleted"
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
    const message = "note " + note.title + " " + column + " changed to " + value
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
        note.title.toLowerCase().includes(searchTerm.toLowerCase())
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
        <DialogContent className="glass border-white/10 text-white max-w-xl max-h-[90vh] overflow-y-auto p-0 overflow-hidden gap-0">
          
          <div className="p-6 pb-4 border-b border-white/5 bg-white/5">
            <div className="flex justify-between items-start">
              <div className="flex items-center space-x-4">
                <div className={`w-8 h-8 rounded-full ${courseData.Color} shadow-lg shadow-black/20`} />
                <div className="space-y-1">
                  <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">{courseData.Code}</p>
                  <h2 className="text-xl font-bold text-white tracking-tight">{courseData.Name}</h2>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6">
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
                    <p className="text-h4 text-white leading-tight">{courseData.Schedule}</p>
                  </div>

                  <div className="p-4 rounded-2xl bg-gradient-to-br from-emerald-500/10 to-teal-500/10 border border-emerald-500/20 group">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="p-2 bg-emerald-500/20 rounded-lg text-emerald-400 group-hover:bg-emerald-500/30 transition-colors">
                        <MapPin className="w-5 h-5" />
                      </div>
                      <span className="text-xs font-medium text-emerald-200 uppercase tracking-wider">Location</span>
                    </div>
                    <p className="text-h4 text-white leading-tight">{courseData.Location || "Online"}</p>
                  </div>
                </div>

                {/* Instructor Section */}
                <div className="p-4 rounded-2xl bg-white/5 border border-white/5 flex items-start gap-4">
                  <div className="p-3 bg-white/5 rounded-full text-gray-400">
                    <Users className="w-6 h-6" />
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">Instructor</p>
                    <p className="text-h4 text-white">{courseData.Instructor}</p>
                    <div className="flex items-center gap-2 text-sm text-blue-400">
                      <Mail className="w-3.5 h-3.5" />
                      <a href={`mailto:${courseData.InstructorEmail}`} className="hover:underline">{courseData.InstructorEmail}</a>
                    </div>
                  </div>
                </div>


                {/* Course Metadata Grid */}
                <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider block mb-1.5">Credits</span>
                    <div className="flex">
                      <Badge variant="outline" className="border-white/10 bg-white/5 text-white px-2 py-0.5 text-xs">
                        {courseData.Credits} credits
                      </Badge>
                    </div>
                  </div>

                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider block mb-1.5">Semester</span>
                    <p className="text-sm font-medium text-white ">{courseData.Semester}</p>
                  </div>

                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <div className="flex items-center space-x-1.5 mb-1.5">
                      <div className="w-1.5 h-1.5 rounded-full bg-green-400"></div>
                      <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider">Start Date</span>
                    </div>
                    <p className="text-sm font-medium text-white ">{format(courseData.StartDate, "MMM d, yyyy")}</p>
                  </div>

                  <div className="p-3 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors">
                    <div className="flex items-center space-x-1.5 mb-1.5">
                      <div className="w-1.5 h-1.5 rounded-full bg-red-400"></div>
                      <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider">End Date</span>
                    </div>
                    <p className="text-sm font-medium text-white ">{format(courseData.EndDate, "MMM d, yyyy")}</p>
                  </div>
                </div>

              </TabsContent>

              <TabsContent value="assignments" className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-300">
                {course_assignments.length > 0 ? (
                  <div className="space-y-6">
                  
                    <div>
                      <div className="flex justify-between items-center mb-4">
                        <label className="text-sm font-medium text-gray-400 uppercase tracking-wider">Recent Assignments</label>

                        <Link href={`/assignments?view=list&course=${courseData.Code}`}>
                          <Button variant="ghost" size="sm" className="text-blue-400 hover:text-blue-300 hover:bg-blue-400/10 h-8 text-xs">
                            View All
                          </Button>
                        </Link>
                      </div>

                      <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-2">
                        {course_assignments.slice(0, 4).map((assignment, index) => (
                          <div key={index} className="flex items-center p-3 rounded-xl border border-white/5 bg-white/5 hover:bg-white/10 transition-all duration-200 group">
                            <span className="w-2/3 text-sm text-white line-clamp-1 font-medium group-hover:text-blue-300 transition-colors">{assignment.Title}</span>
                            <div className="flex flex-col items-end space-y-1 grow">
                              <span className="text-[10px] uppercase tracking-wider text-gray-500 font-medium">{formatDeadline(assignment.Deadline)}</span>
                              <StatusTag assignment={assignment} onEdit={handleEditAssignment} />
                            </div>
                          </div>
                        ))}
                      </div>

                    </div>

                    <div className="p-4 border border-white/5 bg-white/5 rounded-xl">
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
                          router.push(`/notes?course=${course.Code}`)
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
            <div className="flex gap-3 mt-6">
              <Button
                variant="outline"
                size="sm"
                className="flex-1 bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-9"
                onClick={(e) => {
                  e.stopPropagation()
                  setOpen(true)
                }}
              >
                <Edit className="mr-2 w-3.5 h-3.5" />
                Edit
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="flex-1 bg-white/5 border-white/10 hover:bg-white/10 hover:text-white h-9"
                onClick={(e) => {
                  e.stopPropagation()
                  onLinkRequest()
                }}
              >
                <Share className="mr-2 w-3.5 h-3.5" />
                Share
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-red-400 bg-red-500/5 border-red-500/20 hover:bg-red-500/10 hover:text-red-300 h-9"
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete(course)
                }}
              >
                <Trash2 className="mr-2 w-3.5 h-3.5" />
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
