import { useEffect, useState } from "react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs"
import { useCurrentUser, useGetCourseInvitations } from "@/hooks/use-auth"
import { CourseItem } from "@/components/courses/course-item"
import { useAcceptCourseInvitation, useCourses, useCoursesLinked, useDeclineCourseInvitation } from "@/hooks/use-courses"
import { EmptyState } from "@/components/ui/empty-state"
import { GlassCard } from "@/components/ui/glass-card"
import { useMemo } from "react"
import { AssignmentItem } from "@/components/assignments/assignment-item"
import { useCreateAssignment, useAssignments } from "@/hooks/use-assignments"
import { useCreateNote, useNotes } from "@/hooks/use-notes"
import { NoteItem } from "@/components/notes/note-item"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { models } from "@/wailsjs/go/models"
import { FileText, BookOpenIcon } from "lucide-react"
import { Separator } from "@radix-ui/react-separator"

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

    const createMutation = useCreateAssignment()
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
    const handleCopyAssignment = async (assignment: models.Assignment) => {
        const message = "[Frontend] assignment " + assignment.Title + " added"
        var correspondingCourse = userCourses?.find((c) => c.Code == assignment.CourseCode)
        var newAssignment = models.LocalAssignment.createFrom(assignment)
        createMutation.mutate({
            ...newAssignment,
            Status: "Not started",
            ParentID: newAssignment.ID,
            CourseID: correspondingCourse?.ID,
            RemoteCourseID: correspondingCourse?.RemoteID,
        } as models.LocalAssignment)

    }

    const handleCopyNote = async (note: models.Note) => {
        const message = "[Frontend] note " + note.Title + " added"
        var correspondingCourse = userCourses?.find((c) => c.Code == note.CourseCode)
        var newNote = models.LocalNote.createFrom(note)
        createNoteMutation.mutate({
            ...newNote,
            ParentID: newNote.ID,
            CourseID: correspondingCourse?.ID,
            RemoteCourseID: correspondingCourse?.RemoteID,
        } as models.LocalNote)
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
        <div className="flex flex-1 gap-6">
            {!filteredAssignments?.length && !filteredNotes?.length && (
                <GlassCard variant="board" className="flex-1 items-center justify-center">
                    <EmptyState
                        icon={FileText}
                        title="No resources found"
                        description="Create a new assignment or note to get started"
                    />
                </GlassCard>
            )}
            {selectedCluster && (filteredAssignments?.length > 0 || filteredNotes?.length > 0) && (
                <Tabs defaultValue="assignments" className="flex flex-col flex-1">

                    <div className="flex flex-col gap-2 mb-6">

                        <h3 className="flex items-end gap-2 text-h3">
                            {selectedCluster?.course.Name}
                            <p className="text-h5 text-text-caption align-baseline">{selectedCluster?.course.Code}</p>

                        </h3>

                        <TabsList className="flex flex-row gap-4 items-center w-fit bg-transparent">
                            <TabsTrigger
                                value="assignments"
                                className="flex items-baseline text-body text-gray-400 data-[state=active]:text-h6 data-[state=active]:text-white transition-all duration-200"
                            >
                                <span className="font-normal leading-none uppercase tracking-wider">Assignments</span>
                            </TabsTrigger>
                            <TabsTrigger
                                value="notes"
                                className="flex items-baseline text-body text-gray-400 data-[state=active]:text-h6 data-[state=active]:text-white transition-all duration-200"
                            >
                                <span className="font-normal leading-none uppercase tracking-wider">Notes</span>
                            </TabsTrigger>
                            {selectedCluster && selectedCluster?.users.length > 0 && (
                                <div className="flex items-center gap-4"> 
                                <Separator orientation="vertical" className="h-2 w-px bg-gray-300" />
                                    <div className="flex items-center gap-2">
                                        <div className="flex items-center">
                                            {selectedCluster?.users.map((user: models.User) => {
                                                return (
                                                    <div className="last:mr-0 mr-[-5px]">
                                                        <Avatar className="h-5 w-5 rounded-full overflow-hidden border border-white/10">
                                                            <AvatarImage src={user?.Avatar || "/placeholder-user.jpg"} />
                                                            <AvatarFallback className="text-[10px]">
                                                                {user?.Username?.split(" ").map((n: string) => n[0]).join("")}
                                                            </AvatarFallback>
                                                        </Avatar>
                                                    </div>
                                                )
                                            })}
                                        </div>
                                        <p className="text-body text-text-caption align-baseline">{selectedCluster?.users.length} users join this course</p>
                                    </div>
                                </div>
                            )}
                        </TabsList>

                    </div>


                    <TabsContent value="assignments" className="flex flex-col data-[state=active]:flex-1 m-0">
                        {filteredAssignments?.length > 0 && (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {filteredAssignments?.map((a) => {
                                    return (
                                        <AssignmentItem
                                            assignment={a}
                                            variant="outline"
                                            mode="user"
                                            onCopy={handleCopyAssignment}
                                            user={a.User}
                                        />
                                    )
                                })}


                            </div>
                        )}
                    </TabsContent>


                    <TabsContent value="notes" className="flex flex-col data-[state=active]:flex-1 m-0">
                        {filteredNotes?.length > 0 && (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {filteredNotes?.map((n) => {
                                    return (
                                        <NoteItem
                                            note={n}
                                            onCopy={handleCopyNote}
                                            mode="user"
                                            user={n.User}
                                        />
                                    )
                                })}
                            </div>
                        )}
                    </TabsContent>
                </Tabs>
            )}

            <div className="flex flex-col gap-4">

                {courseInvitations && courseInvitations.length > 0 && (
                    <div className="flex flex-col gap-2">

                        <h5 className="text-h5 font-medium text-gray-400 self-end">Course Invitations</h5>
                        <div className="flex flex-wrap w-full" >
                            {courseInvitations?.map((invitation) => (

                                <CourseItem
                                    course={invitation.Course!}
                                    onEdit={() => { }}
                                    onDelete={() => { }}
                                    size="sm"
                                    onAccept={() => handleAcceptCourseInvitation(invitation)}
                                    onDecline={() => handleDeclineCourseInvitation(invitation)}
                                />

                            ))}
                        </div>

                    </div>
                )}

                <div className="flex flex-col flex-1 gap-2">
                    <GlassCard
                        variant="board"
                        className="p-4 flex-1"
                    >
                        {courses && courses.length > 0 ? (

                            <div className="flex flex-col flex-1 gap-4">
                                {courses?.map((course) => (
                                    <div key={course.course.Code}>
                                        <CourseItem
                                            course={course.course}
                                            onCourseClick={() => handleCourseClick(course)}
                                            size="sm"
                                        />
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <EmptyState
                                icon={BookOpenIcon}
                                title="No courses linked"
                                description="Link a course to get started"
                                className="flex-1 items-center"
                            />

                        )}
                    </GlassCard>
                </div>

            </div>

        </div>
    )
}