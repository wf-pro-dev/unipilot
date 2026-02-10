import { useEffect, useState } from "react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCurrentUser, useGetCourseInvitations } from "@/hooks/use-auth"
import { CourseItem } from "@/components/courses/course-item"
import { useAcceptCourseInvitation, useCourses, useCoursesLinked, useDeclineCourseInvitation } from "@/hooks/use-courses"
import { EmptyState } from "@/components/ui/empty-state"
import { GlassCard } from "@/components/ui/glass-card"
import { useMemo } from "react"
import { AssignmentItem } from "@/components/assignments/assignment-item"
import { useCreateAssignment, useAssignments, useCopyAssignment } from "@/hooks/use-assignments"
import { useCreateNote, useNotes } from "@/hooks/use-notes"
import { NoteItem } from "@/components/notes/note-item"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { models } from "@/wailsjs/go/models"
import { FileText, BookOpenIcon, Bell, Users2, CheckCircle2, XCircle, Info } from "lucide-react"
import { Separator } from "@radix-ui/react-separator"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Scroll } from "@/components/core/scroll"
import { v4 as uuidv4 } from 'uuid';

interface Cluster {
    course: models.Course
    assignments: models.Assignment[]
    notes: models.Note[]
    users: models.User[]
}

export function SocialTab() {

    const [selectedCluster, setSelectedCluster] = useState<Cluster | undefined>(undefined)


    const { data: userCourses } = useCourses()
    const { data: courseInvitations } = useGetCourseInvitations()
    const { data: coursesLinked } = useCoursesLinked()
    const { data: currentUser } = useCurrentUser()
    const { data: assignments } = useAssignments()
    const { data: notes } = useNotes()

    const copyAssignmentMutation = useCopyAssignment()
    const createNoteMutation = useCreateNote()
    const AcceptCourseInvitation = useAcceptCourseInvitation()
    const DeclineCourseInvitation = useDeclineCourseInvitation()


    /**
      
    * Handles assignment creation with optimistic UI updates.
      * 
      * Creates a new assignment and provides immediate UI feedback. Logs the creation
      * and shows success/error toast notifications.
      * 
      * @param {assignment.LocalAssignment} assignment - The assignment to create
      * @returns {Promise<void>}
      */
    const handleCopyAssignment = async (assignment: models.Assignment, includeDocuments: boolean) => {
        const message = "[Frontend] assignment " + assignment.Title + " added"
        var correspondingCourse = userCourses?.find((c) => c.Code == assignment.Course?.Code)
        var newAssignment = models.LocalAssignment.createFrom(assignment)
        const id = uuidv4().toString()
        copyAssignmentMutation.mutate({
            assignment: {
                ...newAssignment,
                ID: id,
                Status: "Not started",
                ParentID: newAssignment.ID,
                CourseID: correspondingCourse?.ID,
            } as unknown as models.LocalAssignment,
            includeDocuments
        })

    }

    const handleCopyNote = async (note: models.Note) => {
        const message = "[Frontend] note " + note.Title + " added"
        var correspondingCourse = userCourses?.find((c) => c.Code == note.Course?.Code)
        var newNote = models.LocalNote.createFrom(note)
        createNoteMutation.mutate({
            ...newNote,
            CourseID: correspondingCourse?.ID,
        } as unknown as models.LocalNote)
    }

    const handleAcceptCourseInvitation = (invitation: models.CourseInvitation) => {
        AcceptCourseInvitation.mutate({ invitation })
    }

    const handleDeclineCourseInvitation = (invitation: models.CourseInvitation) => {
        DeclineCourseInvitation.mutate(invitation)
    }

    const handleCourseClick = (course: Cluster) => {
        setSelectedCluster(course)
    }

    const courses = useMemo(() => {

        var clusters: Map<string, Cluster> = new Map<string, Cluster>()

        // Create assignment and note lists for each course code
        coursesLinked?.forEach((course) => {

            if (course.UserID != currentUser?.ID) {

                var assignments: models.Assignment[] = []
                if (course.Assignments) {
                    assignments = course.Assignments.map((assignment) => {
                        return {
                            ...assignment,
                            Course: course,
                            User: course.User!
                        } as models.Assignment
                    })
                }

                var notes: models.Note[] = []
                if (course.Notes) {
                    notes = course.Notes.map((note) => {
                        return {
                            ...note,
                            Course: course,
                            User: course.User!
                        } as models.Note
                    })
                }

                var cluster: Cluster | undefined = clusters.get(course.Code)
                if (!cluster) {
                    cluster = {
                        course: course,
                        assignments: [],
                        notes: [],
                        users: []
                    }
                }
                cluster = {
                    ...cluster,
                    assignments: cluster.assignments.concat(assignments),
                    notes: cluster.notes.concat(notes),
                    users: [...cluster.users, course.User!]
                }
                clusters.set(course.Code, cluster)
            }

        })

        return Array.from(clusters.values())
    }, [coursesLinked])
    console.log(courses)

    useEffect(() => {
        if (courses?.[0]) {
            setSelectedCluster(courses?.[0])
        }
    }, [courses])

    const filteredAssignments = useMemo(() => {
        return (selectedCluster?.assignments || []).filter((assignment) =>
            !assignments?.some((a: models.LocalAssignment) => a.ParentID === assignment.ID)
        )
    }, [assignments, selectedCluster])

    const filteredNotes = useMemo(() => {
        return (selectedCluster?.notes || []).filter((note) =>
            !notes?.some((n: models.LocalNote) => n.ParentID === note.ID)
        )
    }, [notes, selectedCluster])

    return (
        <div className="flex flex-1 h-full gap-0 overflow-hidden">

            {/* LEFT SIDEBAR: Course Navigation */}
            <div className="w-80 border-r border-white/5 flex flex-col bg-white/[0.02]">

                {/* Invitations Section - Priority placement at top */}
                {courseInvitations && courseInvitations.length > 0 && (
                    <div className="border-b border-white/5 bg-amber-500/5">
                        <div className="p-4 space-y-3">
                            <div className="flex items-center gap-2">
                                <div className="p-1.5 rounded-lg bg-amber-500/10 border border-amber-500/20">
                                    <Bell className="w-3.5 h-3.5 text-amber-400" />
                                </div>
                                <h4 className="text-caption font-semibold text-amber-400 uppercase tracking-wider">
                                    Pending Invitations
                                </h4>
                                <Badge variant="outline" className="ml-auto bg-amber-500/10 border-amber-500/20 text-amber-400 text-[10px] px-1.5 py-0">
                                    {courseInvitations.length}
                                </Badge>
                            </div>

                            <div className="space-y-2 max-h-64 overflow-y-auto">
                                {courseInvitations.map((invitation) => (
                                    <div key={invitation.ID} className="group">
                                        <GlassCard variant="outline" className="p-3 hover:bg-white/5 transition-colors">
                                            <div className="space-y-2.5">
                                                {/* Course Info */}
                                                <div className="flex items-start gap-2.5">
                                                    <div className={cn("w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0", invitation.Course?.Color)}>
                                                        <BookOpenIcon className="w-4 h-4 text-white" />
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <p className="text-body font-semibold text-white truncate">
                                                            {invitation.Course?.Code}
                                                        </p>
                                                        <p className="text-caption text-text-caption truncate">
                                                            {invitation.Course?.Name}
                                                        </p>
                                                    </div>
                                                </div>

                                                {/* Action Buttons */}
                                                <div className="flex gap-2">
                                                    <Button
                                                        size="sm"
                                                        variant="default"
                                                        className="flex-1 h-8 bg-green-500/20 hover:bg-green-500/30 text-green-400 border-green-500/30"
                                                        onClick={() => handleAcceptCourseInvitation(invitation)}
                                                    >
                                                        <CheckCircle2 className="w-3.5 h-3.5 mr-1.5" />
                                                        Accept
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        className="flex-1 h-8 text-red-400 hover:bg-red-500/10 border-white/10"
                                                        onClick={() => handleDeclineCourseInvitation(invitation)}
                                                    >
                                                        <XCircle className="w-3.5 h-3.5 mr-1.5" />
                                                        Decline
                                                    </Button>
                                                </div>
                                            </div>
                                        </GlassCard>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                )}

                {/* Shared Courses Section */}
                <div className="flex-1 flex flex-col overflow-hidden">
                    <div className="p-4 pb-3 border-b border-white/5">
                        <div className="flex items-center gap-2">
                            <div className="p-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20">
                                <Users2 className="w-3.5 h-3.5 text-blue-400" />
                            </div>
                            <h4 className="text-caption font-semibold text-blue-400 uppercase tracking-wider">
                                Shared Courses
                            </h4>
                            <Badge variant="outline" className="ml-auto bg-blue-500/10 border-blue-500/20 text-blue-400 text-[10px] px-1.5 py-0">
                                {courses?.length || 0}
                            </Badge>
                        </div>
                    </div>

                    <div className="flex-1 overflow-y-auto p-3">
                        {courses && courses.length > 0 ? (
                            <div className="space-y-2">
                                {courses.map((cluster) => {
                                    const isSelected = selectedCluster?.course.Code === cluster.course.Code
                                    const totalItems = cluster.assignments.length + cluster.notes.length

                                    return (
                                        <button
                                            key={cluster.course.Code}
                                            onClick={() => handleCourseClick(cluster)}
                                            className={cn(
                                                "w-full text-left group transition-all duration-200",
                                                "rounded-lg p-3 border",
                                                isSelected
                                                    ? "bg-white/10 border-blue-400/50 shadow-lg"
                                                    : "bg-white/[0.02] border-white/5 hover:bg-white/5 hover:border-white/10"
                                            )}
                                        >
                                            <div className="space-y-2.5">
                                                {/* Course Header */}
                                                <div className="flex items-center gap-2.5">
                                                    <div className={cn(
                                                        "w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 transition-transform",
                                                        cluster.course.Color,
                                                        isSelected && "scale-105"
                                                    )}>
                                                        <BookOpenIcon className="w-5 h-5 text-white" />
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <p className={cn(
                                                            "text-body font-bold truncate transition-colors",
                                                            isSelected ? "text-white" : "text-gray-300"
                                                        )}>
                                                            {cluster.course.Code}
                                                        </p>
                                                        <p className="text-caption text-text-caption truncate">
                                                            {cluster.course.Name}
                                                        </p>
                                                    </div>
                                                </div>

                                                {/* Stats */}
                                                <div className="flex items-center justify-between text-caption">
                                                    <div className="flex items-center gap-3">
                                                        <span className="text-text-caption">
                                                            <span className="font-medium text-white">{cluster.assignments.length}</span> assignments
                                                        </span>
                                                        <span className="text-gray-600">•</span>
                                                        <span className="text-text-caption">
                                                            <span className="font-medium text-white">{cluster.notes.length}</span> notes
                                                        </span>
                                                    </div>
                                                </div>

                                                {/* Contributors */}
                                                {cluster.users.length > 0 && (
                                                    <div className="flex items-center gap-2 pt-1">
                                                        <div className="flex items-center -space-x-2">
                                                            {cluster.users.slice(0, 3).map((user, idx) => (
                                                                <Avatar key={idx} className="h-5 w-5 rounded-full border-2 border-gray-900">
                                                                    <AvatarImage src={user?.Avatar || "/placeholder-user.jpg"} />
                                                                    <AvatarFallback className="text-[9px] bg-gray-700">
                                                                        {user?.Username?.split(" ").map((n: string) => n[0]).join("")}
                                                                    </AvatarFallback>
                                                                </Avatar>
                                                            ))}
                                                        </div>
                                                        <span className="text-[11px] text-text-caption">
                                                            {cluster.users.length} {cluster.users.length === 1 ? 'contributor' : 'contributors'}
                                                        </span>
                                                    </div>
                                                )}
                                            </div>
                                        </button>
                                    )
                                })}
                            </div>
                        ) : (
                            <div className="h-full flex items-center justify-center p-6">
                                <EmptyState
                                    icon={BookOpenIcon}
                                    title="No shared courses"
                                    description="Courses shared with you will appear here"
                                />
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* MAIN CONTENT AREA */}
            <div className="flex-1 flex flex-col overflow-hidden">
                {!selectedCluster ? (
                    <div className="flex-1 flex items-center justify-center">
                        <EmptyState
                            icon={FileText}
                            title="No course selected"
                            description="Select a shared course to view its resources"
                        />
                    </div>
                ) : (
                    <>
                        {/* Course Header */}
                        <div className="h-full flex flex-col border-b border-white/5 bg-white/[0.02]">
                            <div className="p-6 pb-5">
                                <div className="space-y-4">
                                    {/* Course Title */}
                                    <div className="flex items-center gap-4">
                                        <div className={cn("w-14 h-14 rounded-xl flex items-center justify-center flex-shrink-0", selectedCluster.course.Color)}>
                                            <BookOpenIcon className="w-7 h-7 text-white" />
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <div className="flex items-baseline gap-3">
                                                <h2 className="text-h3 font-bold text-white">
                                                    {selectedCluster.course.Name}
                                                </h2>
                                                <span className="text-h6 font-medium text-text-caption">
                                                    {selectedCluster.course.Code}
                                                </span>
                                            </div>
                                            {selectedCluster.course.Instructor && (
                                                <p className="text-caption text-text-caption mt-1">
                                                    Instructor: {selectedCluster.course.Instructor}
                                                </p>
                                            )}
                                        </div>
                                    </div>

                                    {/* Contributors Banner */}
                                    {selectedCluster.users.length > 0 && (
                                        <div className="flex items-center gap-3 p-3 rounded-lg bg-blue-500/5 border border-blue-500/10">
                                            <div className="p-2 rounded-lg bg-blue-500/10">
                                                <Info className="w-4 h-4 text-blue-400" />
                                            </div>
                                            <div className="flex items-center gap-3 flex-1">
                                                <div className="flex items-center -space-x-2">
                                                    {selectedCluster.users.map((user, idx) => (
                                                        <Avatar key={idx} className="h-6 w-6 rounded-full border-2 border-gray-900">
                                                            <AvatarImage src={user?.Avatar || "/placeholder-user.jpg"} />
                                                            <AvatarFallback className="text-[10px] bg-gray-700">
                                                                {user?.Username?.split(" ").map((n: string) => n[0]).join("")}
                                                            </AvatarFallback>
                                                        </Avatar>
                                                    ))}
                                                </div>
                                                <p className="text-caption text-blue-400">
                                                    <span className="font-medium">{selectedCluster.users.length}</span> {selectedCluster.users.length === 1 ? 'person is' : 'people are'} sharing content for this course
                                                </p>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Tabs */}
                            <Tabs defaultValue="assignments" className="flex flex-col flex-1">
                                <TabsList className="flex flex-row gap-6 px-6 bg-transparent border-t border-white/5">
                                    <TabsTrigger
                                        value="assignments"
                                        className="relative px-0 pb-3 pt-4 text-body font-medium text-gray-400 data-[state=active]:text-white transition-colors border-b-2 border-transparent data-[state=active]:border-blue-400"
                                    >
                                        <span className="uppercase tracking-wider">Assignments</span>
                                        <Badge variant="outline" className="ml-2 bg-white/5 border-white/10 text-white text-[10px] px-1.5 py-0">
                                            {filteredAssignments.length}
                                        </Badge>
                                    </TabsTrigger>
                                    <TabsTrigger
                                        value="notes"
                                        className="relative px-0 pb-3 pt-4 text-body font-medium text-gray-400 data-[state=active]:text-white transition-colors border-b-2 border-transparent data-[state=active]:border-blue-400"
                                    >
                                        <span className="uppercase tracking-wider">Notes</span>
                                        <Badge variant="outline" className="ml-2 bg-white/5 border-white/10 text-white text-[10px] px-1.5 py-0">
                                            {filteredNotes.length}
                                        </Badge>
                                    </TabsTrigger>
                                </TabsList>

                                {/* Tab Content */}
                                <div className="flex-1 overflow-y-auto">
                                    <TabsContent value="assignments" className="m-0 p-6 data-[state=active]:h-full data-[state=active]:flex data-[state=active]:flex-col">

                                        <div className="flex h-full min-h-0">
                                            {filteredAssignments.length > 0 ? (
                                                <Scroll
                                                    data={{ Data: filteredAssignments, HasMore: false }}
                                                    renderItem={(assignment: models.Assignment) => (
                                                        <AssignmentItem
                                                            key={assignment.ID}
                                                            assignmentId={assignment.ID}
                                                            assignment={assignment}
                                                            variant="outline"
                                                            mode="user"
                                                            onCopy={handleCopyAssignment}
                                                            user={assignment.User}
                                                        />
                                                    )}
                                                    keyExtractor={(item: models.Assignment) => item.ID}
                                                    numColumns={2}
                                                    containerClassName="gap-4"
                                                />
                                            ) : (
                                                <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
                                                    <EmptyState
                                                        icon={FileText}
                                                        title="No assignments shared"
                                                        description="Assignments from this course will appear here"
                                                        className="flex-1 items-center"
                                                    />

                                                </div>
                                            )}
                                        </div>
                                    </TabsContent>

                                    <TabsContent value="notes" className="m-0 p-6 data-[state=active]:h-full data-[state=active]:flex data-[state=active]:flex-col">
                                        <div className="flex h-full min-h-0">
                                            {filteredNotes.length > 0 ? (
                                                <Scroll
                                                    data={{ Data: filteredNotes, HasMore: false }}
                                                    renderItem={(note: models.Note) => (
                                                        <NoteItem
                                                            key={note.ID}
                                                            noteID={note.ID}
                                                            noteRO={note}
                                                            mode="user"
                                                            onCopy={handleCopyNote}
                                                            user={note.User}
                                                        />
                                                    )}
                                                    keyExtractor={(item: models.Note) => item.ID}
                                                    numColumns={2}
                                                    containerClassName="gap-4"
                                                />
                                            ) : (
                                                <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
                                                    <EmptyState
                                                        icon={BookOpenIcon}
                                                        title="No notes shared"
                                                        description="Notes from this course will appear here"
                                                        className="flex-1 items-center"
                                                    />

                                                </div>
                                            )}
                                        </div>
                                    </TabsContent>
                                </div>
                            </Tabs>
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}