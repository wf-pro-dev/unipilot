"use client"

import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import {
  BookOpen,
  X,
  Users,
  MapPin,
  Edit,
  Trash2,
  Clock,
  Share,
  Search,
  ArrowRight,
  List,
  Check,
  Bot,

} from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useAssignments } from "@/hooks/use-assignments"
import { useCallback, useMemo, useState } from "react"
import { format, differenceInDays, isPast } from "date-fns"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCourseNotes } from "@/hooks/use-notes"
import { NoteItem } from "../notes/note-item"
import { Input } from "../ui/input"
import { useRouter } from "next/navigation"
import { useAcceptCourseInvitation, useCourse, useDeclineCourseInvitation } from "@/hooks/use-courses"
import { AssignmentItem } from "../assignments/assignment-item"
import { useDialogContext } from "../provider/dialog-provider"
import { cn } from "@/lib/utils"
import { useGetCourseInvitations } from "@/hooks/use-auth"
import { SearchFilter } from "../core/search-filter/search-filter"
import { SearchConfig } from "../core/search-filter/types"
import { Scroll } from "../core/scroll"
import { EmptyState, } from "../ui/empty-state"
import { UserItem } from "../community/user-item"
import { ClusterUsersList } from "../community/cluster-users-list"
import { useAuthContext } from "../provider/auth-provider"
import { Action } from "../core/action"

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

  const { user: currentUser } = useAuthContext()
  const { SetDialogState } = useDialogContext()
  const router = useRouter()

  const [activeView, setActiveView] = useState("overview")
  const [searchTerm, setSearchTerm] = useState("")

  const { data: assignments } = useAssignments()
  const { mutate: acceptCourseInvitation } = useAcceptCourseInvitation()
  const { mutate: cancelCourseInvitation } = useDeclineCourseInvitation()
  const notes = useCourseNotes(course as models.LocalCourse)


  const { data: courseInvitations } = useGetCourseInvitations()

  const clusterID = useMemo(() => {
    return course?.ClusterID || course?.ID
  }, [course])

  const invitationsShare = useMemo(() => {
    return (courseInvitations || []).filter((invitation: models.CourseInvitation) =>
      invitation.Course?.ID === course?.ID &&
      invitation.Status === "pending" &&
      invitation.SenderID === currentUser?.ID
    ) || []
  }, [courseInvitations, course])

  const invitationsRequest = useMemo(() => {
    return (courseInvitations || []).filter((invitation: models.CourseInvitation) =>
      invitation.Course?.ID === course?.ID &&
      invitation.Status === "pending" &&
      invitation.ReceiverID === currentUser?.ID
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

  // Search configuration
  const searchConfig: SearchConfig<models.LocalAssignment> = {
    placeholder: "Search assignments by title or course...",
    searchableFields: ["Title", "Course"],
    enabled: true
  }


  const handleSearchChange = (searchTerm: string) => {
    // If you want to persist search in URL, add it here
    console.log("Search term:", searchTerm)
  }


  const renderShare = useCallback((invitation: models.CourseInvitation) => {
    return (
      <UserItem
        user={invitation.Receiver!}
        size="compact"
        actions={(user: models.User) => (
          <Button
            variant="danger"
            size="sm"
            className="rounded-lg text-red-400/80 hover:text-red-400 hover:bg-red-500/30 border-red-400/80"
            onClick={() => {
              cancelCourseInvitation(invitation)
            }}
          >
            <X className="w-4 h-4" />
          </Button>
        )}
      />
    )
  }, [])


  const renderRequest = useCallback((invitation: models.CourseInvitation) => {
    return (
      <UserItem
        user={invitation.Sender!}
        size="compact"
        actions={(user: models.User) => (
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              className="rounded-lg text-white/80 hover:text-white hover:bg-white/10 border-white/80"
              onClick={() => {
                acceptCourseInvitation({ invitation })
              }}
            >
              <Check className="w-4 h-4" />

            </Button>
            <Button
              variant="danger"
              size="sm"
              className="rounded-lg text-red-400/80 hover:text-red-400 hover:bg-red-500/30 border-red-400/80"
              onClick={() => {
                cancelCourseInvitation(invitation)
              }}
            >
              <X className="w-4 h-4" />

            </Button>

          </div>
        )}
      />
    )
  }, [])

  const renderItem = useCallback((user: models.User) => {
    return <UserItem user={user} size="compact" />
  }, [currentUser]);




  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-4xl p-0 gap-0 max-h-[90vh] flex flex-col md:flex-row">

        {/* === LEFT SIDEBAR: Course Identity & Logistics === */}
        <div className="md:w-80 bg-white/5 border-r border-white/5 relative flex flex-col  overflow-y-auto shrink-0">

          <div className="p-6 flex flex-col ">

            {/* Header: Icon & Title */}
            <div className="flex flex-col items-center mb-6 mt-2 text-center">
              <div className={cn(
                "w-20 h-20 rounded-2xl flex items-center justify-center shadow-2xl shadow-black/50 mb-5 ring-2 ring-white/10",
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

          
          </div>
        </div>

        {/* === RIGHT CONTENT: Workspace === */}
        <div className="flex-1 bg-bg-base/30">

          <Tabs value={activeView} onValueChange={setActiveView} className=" flex flex-col h-full ">

            {/* Top Navigation Bar */}
            <div className="px-6 border-b border-white/5 bg-white/5 flex-shrink-0">
              <TabsList className="flex gap-6 bg-transparent h-14 p-0">
                {['Overview', 'Assignments', 'Notes', 'Share'].map((tab) => (
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
                    {tab}
                  </TabsTrigger>
                ))}
              </TabsList>
            </div>

            {/* Scrollable Content Area */}
            <div className="flex flex-col min-h-0 p-6">

              {/* --- OVERVIEW TAB --- */}
              <TabsContent value="overview" className="mt-0 space-y-6 outline-none animate-in fade-in-50 duration-300 overflow-y-auto">


                {/* Actions */}
                <div className="grid grid-cols-2 gap-3">
                  <Action
                    label="Edit Course"
                    Icon={Edit}
                    onClick={() => SetDialogState({
                      modelType: "course",
                      dialogType: "edit",
                      id: courseId
                    })}
                  />
                  <Action
                    label="Delete Course"
                    Icon={Trash2}
                    onClick={() => SetDialogState({
                      modelType: "course",
                      dialogType: "delete",
                      id: courseId
                    })}
                  />
                </div>
                {/* 1. NEXT UP HERO SECTION */}
                {nextAssignment && (

                  <section>
                    <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider mb-3 flex items-center gap-2">
                      Next Assignment
                    </h3>

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

                  </section>


                )}

                {/* 2. COURSE METRICS GRID */}
                <div className="gap-6">

                  {/* Timeline Stats */}
                  <div className="space-y-3">
                    <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider flex items-center gap-2">
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


                </div>



              </TabsContent>

              {/* --- ASSIGNMENTS TAB --- */}
              <TabsContent value="assignments" className="mt-0  flex flex-col outline-none animate-in fade-in-50 duration-300">


                <SearchFilter
                  data={assignments || []}
                  searchConfig={searchConfig}
                  onSearchChange={handleSearchChange}
                  layout="vertical"

                  filterDefinitions={[
                    {
                      field: "Status",
                      label: "Status",
                      type: "select",
                      placeholder: "All Statuses",
                      width: "w-32",
                      customOptions: [
                        { label: "Not started", value: "Not started" },
                        { label: "In progress", value: "In progress" },
                        { label: "Done", value: "Done" }
                      ]
                    },
                    {
                      field: "Priority",
                      label: "Priority",
                      type: "select",
                      placeholder: "All Priorities",
                      width: "w-32",
                      customOptions: [
                        { label: "Low", value: "low" },
                        { label: "Medium", value: "medium" },
                        { label: "High", value: "high" }
                      ]
                    }
                  ]}

                >
                  {(filteredAssignments) => (
                    <>
                      {filteredAssignments.length > 0 ? (

                        <Scroll
                          data={{ Data: filteredAssignments, HasMore: false }}
                          renderItem={(assignment: models.LocalAssignment) => (
                            <AssignmentItem key={assignment.ID} assignmentId={assignment.ID} variant="outline" size="sm" />
                          )}
                          keyExtractor={(item: models.LocalAssignment) => item.ID}
                          numColumns={1}
                          containerClassName="gap-4"
                        />
                      ) : (
                        <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
                          <EmptyState
                            icon={List}
                            title="No assignments found"
                            description="Try adjusting your filters or search terms"
                            className="flex-1 items-center"
                            onClick={() => router.push(`/assignments?view=all`)}
                            buttonText="View All Assignments"
                          />
                        </div>
                      )}
                    </>
                  )}
                </SearchFilter>
              </TabsContent>

              {/* --- NOTES TAB --- */}
              <TabsContent value="notes" className="mt-0 flex flex-col outline-none animate-in fade-in-50 duration-300">
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

              {/* --- SHARE TAB --- */}
              <TabsContent value="share" className="mt-0  flex flex-col outline-none animate-in fade-in-50 duration-300 gap-4">

                {mode === "default" && (
                  <div className="space-y-4">

                    <div className="bg-white/5 border border-white/5 rounded-xl p-4 flex flex-col sm:flex-row items-center justify-between gap-4">

                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-full bg-primary-purple-500/10 flex items-center justify-center text-primary-purple-400 border border-primary-purple-500/20">
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
                          className="sm:flex-none border-white/10 hover:bg-white/10 rounded-full"
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

                    {invitationsShare.length > 0 && (
                      <div className="space-y-4">

                        <div className="space-y-2">
                          <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider flex items-center gap-2">
                            Shared by You
                          </h3>
                          <Scroll
                            data={{ Data: invitationsShare, HasMore: false }}
                            renderItem={renderShare}
                            keyExtractor={(item: models.CourseInvitation) => item.ID}
                            numColumns={2}
                            containerClassName="gap-4"
                          />
                        </div>
                        <Separator className="bg-white/10" />
                      </div>
                    )}




                    {invitationsRequest.length > 0 && (
                      <div className="space-y-4">
                        <div className="space-y-2">
                          <h3 className="text-sm font-medium text-text-caption uppercase tracking-wider flex items-center gap-2">
                            Shared with You
                          </h3>
                          <Scroll
                            data={{ Data: invitationsRequest, HasMore: false }}
                            renderItem={renderRequest}
                            keyExtractor={(item: models.CourseInvitation) => item.ID}
                            numColumns={1}
                            containerClassName="gap-4"
                          />
                        </div>
                        <Separator className="bg-white/10" />
                      </div>

                    )}


                  </div>
                )}

                <ClusterUsersList
                  userID={currentUser?.ID!}
                  courseID={clusterID}
                  numColumns={1}
                  containerClassName="gap-4"
                  renderItem={renderItem}
                />



              </TabsContent>
            </div>
          </Tabs>
        </div>

      </DialogContent >
    </Dialog >
  )
}