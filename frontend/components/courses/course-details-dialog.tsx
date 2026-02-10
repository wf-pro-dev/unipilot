"use client"

import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import {
  BookOpen,
  Users,
  Calendar,
  MapPin,
  Edit,
  Trash2,
  FileText,
  Clock,
  Share,
  Search,
  CheckCircle2,
  Award,
  ArrowRight,
  AlertCircle,
  Timer,
  BarChart3
} from "lucide-react"
import Link from "next/link"
import { models } from "@/wailsjs/go/models"
import { useAssignments } from "@/hooks/use-assignments"
import { useMemo, useState } from "react"
import { format, differenceInDays, isPast } from "date-fns"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCourseNotes } from "@/hooks/use-notes"
import { NoteItem } from "../notes/note-item"
import { Input } from "../ui/input"
import { useRouter } from "next/navigation"
import { useCourse } from "@/hooks/use-courses"
import { AssignmentItem } from "../assignments/assignment-item"
import { useDialogContext } from "../provider/dialog-provider"
import { cn } from "@/lib/utils"
import { useGetCourseInvitations } from "@/hooks/use-auth"
import { CourseItem } from "./course-item"

interface CourseDetailsDialogProps {
  isOpen: boolean
  onClose: () => void
  courseId: string
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
  const course = mode === "default" ? courseData as models.LocalCourse : courseRO

  if (!course) return null

  const { SetDialogState } = useDialogContext()
  const router = useRouter()

  const [activeView, setActiveView] = useState("overview")
  const [searchTerm, setSearchTerm] = useState("")

  const { data: assignments } = useAssignments()
  const notes = useCourseNotes(course as models.LocalCourse)

  const { data: courseInvitations } = useGetCourseInvitations()
  const invitations = useMemo(() => {
    return (courseInvitations || []).filter((invitation: models.CourseInvitation) =>
      invitation.Course?.ID === course?.ID && invitation.Status === "pending"
    ) || []
  }, [courseInvitations, course])

  // Filter assignments for this course
  const courseAssignments = useMemo(() => {
    return (assignments || []).filter((assignment: models.LocalAssignment) =>
      assignment.Course?.ID === course?.ID
    ) || []
  }, [assignments, course])

  // --- Statistics Logic ---

  // Completion stats
  const completedCount = useMemo(() => {
    return courseAssignments.filter((a) => a.Status === "Done").length
  }, [courseAssignments])

  const inProgressCount = useMemo(() => {
    return courseAssignments.filter((a) => a.Status === "In progress").length
  }, [courseAssignments])

  const todoCount = useMemo(() => {
    return courseAssignments.filter((a) => a.Status === "To do").length
  }, [courseAssignments])

  const completionPercentage = useMemo(() => {
    return courseAssignments.length > 0
      ? (completedCount / courseAssignments.length) * 100
      : 0
  }, [completedCount, courseAssignments])

  // Next Due Assignment (Urgency Logic)
  const nextAssignment = useMemo(() => {
    const pending = courseAssignments.filter(a => a.Status !== "Done");
    // Sort by Deadline (ascending)
    return pending.sort((a, b) => {
      const dateA = new Date(a.Deadline).getTime();
      const dateB = new Date(b.Deadline).getTime();
      return dateA - dateB;
    })[0]; // Get the first one
  }, [courseAssignments]);

  // Timeline Logic
  const timelineStats = useMemo(() => {
    const start = new Date(course.StartDate);
    const end = new Date(course.EndDate);
    const now = new Date();

    const totalDays = differenceInDays(end, start);
    const daysPassed = differenceInDays(now, start);
    const daysLeft = differenceInDays(end, now);

    const progress = Math.min(Math.max((daysPassed / totalDays) * 100, 0), 100);

    return { daysLeft: Math.max(0, daysLeft), progress };
  }, [course.StartDate, course.EndDate]);

  // Filter notes
  const filteredNotes = useMemo(() => {
    return notes.filter((note) =>
      note.Title.toLowerCase().includes(searchTerm.toLowerCase())
    )
  }, [notes, searchTerm])

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-5xl p-0 gap-0 max-h-[90vh] flex flex-col md:flex-row">

        {/* === LEFT SIDEBAR: Course Identity & Logistics === */}
        <div className="md:w-80 bg-white/5 border-r border-white/5 relative flex flex-col  overflow-y-auto shrink-0">

          <div className="p-6 flex flex-col ">

            {/* Header: Icon & Title */}
            <div className="flex flex-col items-center mb-6 mt-2 text-center">
              <div className={cn(
                "w-24 h-24 rounded-2xl flex items-center justify-center shadow-2xl shadow-black/50 mb-5 ring-2 ring-white/10",
                course.Color
              )}>
                <BookOpen className="w-10 h-10 text-white" />
              </div>

              <div className="flex items-center gap-2 mb-2">
                <Badge variant="outline" className="text-caption border-white/20 bg-white/5 text-text-caption px-2 py-0.5 h-6">
                  {course.Code}
                </Badge>
                {course.Credits > 0 && (
                  <Badge variant="outline" className="text-caption border-white/20 bg-white/5 text-text-caption px-2 py-0.5 h-6">
                    {course.Credits} Credits
                  </Badge>
                )}
              </div>

              <DialogTitle className="text-h3 text-text-title mb-1 leading-tight">
                {course.Name}
              </DialogTitle>

              <p className="text-caption text-text-caption font-medium">
                {course.Semester}
              </p>
            </div>

            {/* Progress Section */}
            {courseAssignments.length > 0 && (
              <div className="mb-6 bg-white/5 rounded-xl p-4 border border-white/5">
                <div className="flex justify-between items-end mb-2">
                  <span className="text-caption text-text-caption uppercase tracking-wider font-semibold">Progress</span>
                  <span className="text-h5 font-bold text-primary-blue-400">{Math.round(completionPercentage)}%</span>
                </div>
                <Progress value={completionPercentage} className="h-2 bg-white/10" />
                <p className="text-[11px] text-text-caption mt-2 text-right">
                  {completedCount} / {courseAssignments.length} Completed
                </p>
              </div>
            )}

            <Separator className="bg-white/10 mb-6" />

            {/* Key Logistics Info */}
            <div className="space-y-4 ">
              <div className="flex items-start gap-3">
                <div className="p-2 rounded-lg bg-white/5 border border-white/5 shrink-0">
                  <Users className="w-4 h-4 text-text-caption" />
                </div>
                <div className=" min-w-0">
                  <p className="text-caption text-text-caption uppercase tracking-wider mb-0.5">
                    Instructor
                  </p>
                  <p className="text-body text-white font-medium truncate">
                    {course.Instructor}
                  </p>
                  {course.InstructorEmail && (
                    <a href={`mailto:${course.InstructorEmail}`} className="text-xs text-primary-blue-400 hover:text-primary-blue-300 truncate block mt-0.5">
                      {course.InstructorEmail}
                    </a>
                  )}
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="p-2 rounded-lg bg-white/5 border border-white/5 shrink-0">
                  <Clock className="w-4 h-4 text-text-caption" />
                </div>
                <div className=" min-w-0">
                  <p className="text-caption text-text-caption uppercase tracking-wider mb-0.5">
                    Schedule
                  </p>
                  <p className="text-body text-white font-medium">
                    {course.Schedule || "No schedule"}
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="p-2 rounded-lg bg-white/5 border border-white/5 shrink-0">
                  <MapPin className="w-4 h-4 text-text-caption" />
                </div>
                <div className=" min-w-0">
                  <p className="text-caption text-text-caption uppercase tracking-wider mb-0.5">
                    Location
                  </p>
                  <p className="text-body text-white font-medium">
                    {course.Location || "Remote"}
                  </p>
                </div>
              </div>
            </div>

            {/* Bottom Actions */}
            {mode === "default" && (
              <div className="mt-6 pt-4 border-t border-white/10 space-y-2">
                <div className="grid grid-cols-2 gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full rounded-lg border-white/10 hover:bg-white/10"
                    onClick={() => {
                      SetDialogState({
                        modelType: "course",
                        dialogType: "edit",
                        id: courseId
                      })
                    }}
                  >
                    <Edit className="w-4 h-4 mr-2" />
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full rounded-lg border-white/10 hover:bg-white/10"
                    onClick={() => {
                      SetDialogState({
                        modelType: "course",
                        dialogType: "linkRequest",
                        id: courseId
                      })
                    }}
                  >
                    <Share className="w-4 h-4 mr-2" />
                    Share
                  </Button>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full text-status-error hover:text-status-error hover:bg-status-error/10"
                  onClick={() => {
                    SetDialogState({
                      modelType: "course",
                      dialogType: "delete",
                      id: courseId
                    })
                  }}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete Course
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* === RIGHT CONTENT: Workspace === */}
        <div className="w-full flex flex-col bg-bg-base/30 grow-0">

          <Tabs value={activeView} onValueChange={setActiveView} className=" flex flex-col h-full ">

            {/* Top Navigation Bar */}
            <div className="px-6 border-b border-white/5 bg-white/5 flex-shrink-0">
              <TabsList className="flex gap-6 bg-transparent h-14 p-0">
                {['Overview', 'Assignments', 'Notes'].map((tab) => (
                  <TabsTrigger
                    key={tab}
                    value={tab.toLowerCase()}
                    className="
                      relative  flex items-center gap-2 px-1
                      text-sm font-medium text-text-caption 
                      data-[state=active]:text-white 
                      border-b-2 border-transparent data-[state=active]:border-primary-blue-500
                      transition-colors bg-transparent
                    "
                  >
                    {tab === 'Overview' && <Award className="w-4 h-4" />}
                    {tab === 'Assignments' && <FileText className="w-4 h-4" />}
                    {tab === 'Notes' && <BookOpen className="w-4 h-4" />}
                    {tab}
                  </TabsTrigger>
                ))}
              </TabsList>
            </div>

            {/* Scrollable Content Area */}
            <div className="flex flex-col min-h-0 p-6">

              {/* --- OVERVIEW TAB --- */}
              <TabsContent value="overview" className="mt-0 space-y-8 outline-none animate-in fade-in-50 duration-300 overflow-y-auto">

                {/* 1. NEXT UP HERO SECTION */}
                <section>
                  <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider mb-3 flex items-center gap-2">
                    <AlertCircle className="w-4 h-4" />
                    Action Required
                  </h3>

                  {nextAssignment ? (
                    <div className="group relative bg-white/5 border border-white/10 hover:border-primary-blue-500/50 rounded-xl p-5 transition-all">
                      <div className="absolute top-0 left-0 w-1  bg-primary-blue-500 rounded-l-xl" />
                      <div className="flex items-start justify-between gap-4">
                        <div className="space-y-1">
                          <div className="flex items-center gap-2 mb-1">
                            <Badge variant="secondary" className="bg-primary-blue-500/10 text-primary-blue-300 border-primary-blue-500/20">
                              Up Next
                            </Badge>
                            <span className="text-xs text-text-caption font-medium">
                              Due {format(new Date(nextAssignment.Deadline), "EEEE, MMM d")}
                            </span>
                          </div>
                          <h4 className="text-xl font-semibold text-white group-hover:text-primary-blue-300 transition-colors">
                            {nextAssignment.Title}
                          </h4>
                          <p className="text-sm text-text-caption line-clamp-2 max-w-xl">
                            {nextAssignment.Todo || "No description provided."}
                          </p>
                        </div>
                        <Button className="shrink-0 rounded-full" onClick={() => setActiveView("assignments")}>
                          View Details <ArrowRight className="w-4 h-4 ml-2" />
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className="bg-gradient-to-br from-status-success/20 to-transparent border border-status-success/20 rounded-xl p-6 text-center">
                      <div className="w-12 h-12 bg-status-success/20 rounded-full flex items-center justify-center mx-auto mb-3 text-status-success">
                        <CheckCircle2 className="w-6 h-6" />
                      </div>
                      <h4 className="text-lg font-semibold text-white">All Caught Up!</h4>
                      <p className="text-text-caption text-sm">You have no pending assignments for this course.</p>
                    </div>
                  )}
                </section>

                <Separator className="bg-white/10" />

                {/* 2. COURSE METRICS GRID */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

                  {/* Timeline Stats */}
                  <div className="space-y-3">
                    <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider flex items-center gap-2">
                      <Timer className="w-4 h-4" />
                      Timeline
                    </h3>
                    <div className="bg-white/5 border border-white/5 rounded-xl p-5">
                      <div className="flex justify-between items-end mb-2">
                        <div>
                          <p className="text-2xl font-bold text-white">{timelineStats.daysLeft}</p>
                          <p className="text-xs text-text-caption uppercase">Days Remaining</p>
                        </div>
                        <Badge variant="outline" className="bg-white/5 text-text-caption border-white/10">
                          {Math.round(timelineStats.progress)}% Elapsed
                        </Badge>
                      </div>
                      <Progress value={timelineStats.progress} className="h-2 mb-3 bg-white/10" />
                      <div className="flex justify-between text-xs text-text-caption">
                        <span>{format(new Date(course.StartDate), "MMM d")}</span>
                        <span>{format(new Date(course.EndDate), "MMM d, yyyy")}</span>
                      </div>
                    </div>
                  </div>

                  {/* Workload Breakdown */}
                  <div className="space-y-3">
                    <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider flex items-center gap-2">
                      <BarChart3 className="w-4 h-4" />
                      Workload
                    </h3>
                    <div className="bg-white/5 border border-white/5 rounded-xl p-5 grid grid-cols-3 gap-2 text-center divide-x divide-white/10">
                      <div>
                        <p className="text-xl font-bold text-white">{todoCount}</p>
                        <p className="text-[10px] text-text-caption uppercase mt-1">To Do</p>
                      </div>
                      <div>
                        <p className="text-xl font-bold text-primary-blue-400">{inProgressCount}</p>
                        <p className="text-[10px] text-text-caption uppercase mt-1">Active</p>
                      </div>
                      <div>
                        <p className="text-xl font-bold text-status-success">{completedCount}</p>
                        <p className="text-[10px] text-text-caption uppercase mt-1">Done</p>
                      </div>
                    </div>
                  </div>
                </div>

                {/* 3. COLLABORATION SECTION */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider flex items-center gap-2">
                      <Users className="w-4 h-4" />
                      Collaboration & Access
                    </h3>
                    {/* Badge for pending requests placeholder */}
                    <Badge variant="outline" className="bg-white/5 border-white/10 text-text-caption">
                      0 Pending
                    </Badge>
                  </div>

                  <div className="bg-white/5 border border-white/5 rounded-xl p-5 flex flex-col sm:flex-row items-center justify-between gap-4">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-full bg-secondary-purple-500/10 flex items-center justify-center text-secondary-purple-400 border border-secondary-purple-500/20">
                        <Share className="w-5 h-5" />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">Share Course Material</p>
                        <p className="text-xs text-text-caption">Invite peers to view assignments and notes.</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 w-full sm:w-auto">
                      <Button
                        variant="outline"
                        size="sm"
                        className=" sm:flex-none border-white/10 hover:bg-white/10"
                        onClick={() => {
                          SetDialogState({
                            modelType: "course",
                            dialogType: "linkRequest",
                            id: courseId
                          })
                        }}
                      >
                        Manage Access
                      </Button>
                    </div>
                  </div>
                  {invitations.length > 0 && (
                    invitations.map((invitation) => (
                      <CourseItem
                        key={invitation.ID}
                        courseId={invitation.Course?.ID!}
                        courseRO={invitation.Course!}
                        mode="readonly"
                        size="compact"
                      />
                    ))
                  )}
                </div>

              </TabsContent>

              {/* --- ASSIGNMENTS TAB --- */}
              <TabsContent value="assignments" className="mt-0  flex flex-col outline-none animate-in fade-in-50 duration-300">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-body font-medium text-text-caption uppercase tracking-wider">
                    All Tasks ({courseAssignments.length})
                  </h3>
                  <Button variant="outline" size="sm" className="bg-white/5 border-white/10 hover:bg-white/10" asChild>
                    <Link href={`/assignments?view=list&course=${course.Code}`}>
                      Full View
                    </Link>
                  </Button>
                </div>

                {courseAssignments.length > 0 ? (
                  <div className="space-y-3 pb-4">
                    {courseAssignments.map((assignment) => (
                      <div key={assignment.ID} className="bg-white/5 border border-white/5 rounded-lg overflow-hidden hover:border-white/20 transition-colors">
                        <AssignmentItem assignmentId={assignment.ID} mode="ghost" />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className=" flex flex-col items-center justify-center text-center p-8 border-2 border-dashed border-white/10 rounded-xl bg-white/5">
                    <FileText className="w-12 h-12 text-text-muted mb-4" />
                    <h3 className="text-lg font-medium text-white">No assignments</h3>
                    <p className="text-text-caption max-w-xs mt-2">
                      There are no assignments for this course yet.
                    </p>
                  </div>
                )}
              </TabsContent>

              {/* --- NOTES TAB --- */}
              <TabsContent value="notes" className="mt-0  flex flex-col outline-none animate-in fade-in-50 duration-300">
                <div className="mb-4 space-y-4">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-caption" />
                    <Input
                      placeholder="Search notes..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-9 bg-white/5 border-white/10 focus:border-primary-blue-500/50 h-10"
                    />
                  </div>
                </div>

                {filteredNotes.length > 0 ? (
                  <div className="space-y-3 pb-4">
                    {filteredNotes.map((note) => (
                      <div key={note.ID} className="bg-white/5 border border-white/5 rounded-lg overflow-hidden hover:border-white/20 transition-colors">
                        <NoteItem noteID={note.ID} mode="default" />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className=" flex flex-col items-center justify-center text-center p-8 border-2 border-dashed border-white/10 rounded-xl bg-white/5">
                    <BookOpen className="w-12 h-12 text-text-muted mb-4" />
                    <h3 className="text-lg font-medium text-white">
                      {searchTerm ? 'No matches found' : 'No notes'}
                    </h3>
                    <p className="text-text-caption max-w-xs mt-2">
                      {searchTerm ? 'Try a different keyword.' : 'Start taking notes to track your learning.'}
                    </p>
                    <Button
                      variant="outline"
                      className="mt-4 border-white/10 bg-white/5"
                      onClick={() => router.push(`/notes?course=${course?.Code}`)}
                    >
                      Create Note
                    </Button>
                  </div>
                )}
              </TabsContent>

            </div>
          </Tabs>
        </div>

      </DialogContent>
    </Dialog>
  )
}