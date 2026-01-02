"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { UserItemCompact } from "@/components/community/user-item-compact"
import { AssignmentItemCompact } from "@/components/assignments/assignment-item-compact"
import { NoteItemCompact } from "@/components/notes/note-item-compact"
import { Users, FileText, BookOpen, Search, User } from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useState, useMemo, useCallback } from "react"
import { Masonry } from "react-plock"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { useCoursesLinked } from "@/hooks/use-courses"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { AssignmentDetailsModal } from "../assignments/assignment-details-modal"
import { toast } from "sonner"
import { useAssignments, useCreateAssignment } from "@/hooks/use-assignments"
import { format } from "date-fns"
import { EmptyState } from "../ui/empty-state"
import { useRouter } from "next/navigation"


export function LinkedResources() {
  const router = useRouter()
  const { data: coursesLinked } = useCoursesLinked()
  const [activeTab, setActiveTab] = useState("users")
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedAssignment, setSelectedAssignment] = useState<models.LocalAssignment | null>(null)

  const createMutation = useCreateAssignment()

  const { data: userAssignments } = useAssignments()

  /**
   * Handles assignment card click to open details modal.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment that was clicked
   */
  const handleAssignmentClick = (assignment: models.LocalAssignment) => {
    setSelectedAssignment(assignment)
  }

  /**
  * Handles assignment creation with optimistic UI updates.
  * 
  * Creates a new assignment and provides immediate UI feedback. Logs the creation
  * and shows success/error toast notifications.
  * 
  * @param {assignment.LocalAssignment} assignment - The assignment to create
  * @returns {Promise<void>}
  */
  const handleCopyAssignment = async (assignment: models.LocalAssignment) => {
    const message = "[Frontend] assignment " + assignment.Title + " added"
    LogInfo(format(new Date(), "yyyy/MM/dd HH:mm:ssxxx") + " " + message)
    createMutation.mutate({
      ...assignment,
      Status: "Not started",
      ParentID: assignment.ID
    } as models.LocalAssignment, {
      onSuccess: () => {
        toast.success("Assignment added successfully")
        setSelectedAssignment(null)
      },
      onError: () => {
        toast.error("Failed to add assignment")
      }
    })
  }

  const filteredUsers = useCallback((code: string) => {
    return (coursesLinked?.find(c => c.Code === code)?.User || []).filter((u: models.User) =>
      !userAssignments?.some(a => a.ParentID === u.ID) && ( // If the assignment is already in the user's assignments, don't show it
        u.Username.toLowerCase().includes(searchQuery.toLowerCase()) ||
        u.University.toLowerCase().includes(searchQuery.toLowerCase()) ||
        u.Semester.toLowerCase().includes(searchQuery.toLowerCase())
      )
    )
  }, [coursesLinked, searchQuery, userAssignments])

  const filteredAssignments = useCallback((code: string) => {
    return (coursesLinked?.[code]?.Assignments || []).filter(assignment =>
      !userAssignments?.some(a => a.ParentID === assignment.ID) && ( // If the assignment is already in the user's assignments, don't show it
        assignment.CourseCode?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        assignment.Course?.Name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        assignment.Title.toLowerCase().includes(searchQuery.toLowerCase())
      )
    )
  }, [coursesLinked, searchQuery, userAssignments])

  const filteredNotes = useCallback((code: string) => {
    return (coursesLinked?.find(c => c.Code === code)?.Notes || []).filter(note =>
      note.CourseCode?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      note.Course?.Name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      note.Title.toLowerCase().includes(searchQuery.toLowerCase())
    )
  }, [coursesLinked, searchQuery, userAssignments])





  const courseCodes = useMemo(() => {
    const codes = Object.keys(coursesLinked || {})
    return codes.sort()
  }, [coursesLinked])

  var uersrsCount = useMemo(() => {
    return new Set(coursesLinked?.map(course => course.Children?.map(user => user.ID))).size
  }, [coursesLinked])
  var assignmentsCount = useMemo(() => {
    var count = 0
    coursesLinked?.map(course => {
      course.Children?.map(child => {
        count += child.Assignments?.length || 0
      })
    })
    return count
  }, [coursesLinked])
  var notesCount = useMemo(() => {
    var count = 0
    coursesLinked?.map(course => {
      course.Children?.map(child => {
        count += child.Notes?.length || 0
      })
    })
    return count
  }, [coursesLinked])


  return (
    <div className="flex flex-col flex-1 space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search by course code or name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 bg-white/5 border-white/10 text-white placeholder:text-gray-400 focus:bg-white/10 transition-all"
          />
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex flex-col flex-1 w-full">
        <TabsList className="bg-white/5 p-1 rounded-xl border border-white/5 w-fit">
          <TabsTrigger
            value="users"
            className="flex items-center space-x-2 px-4 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
          >
            <Users className="w-4 h-4" />
            <span className="text-sm font-medium">Users ({uersrsCount})</span>
          </TabsTrigger>
          <TabsTrigger
            value="assignments"
            className="flex items-center space-x-2 px-4 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
          >
            <FileText className="w-4 h-4" />
            <span className="text-sm font-medium">Assignments ({assignmentsCount})</span>
          </TabsTrigger>
          <TabsTrigger
            value="notes"
            className="flex items-center space-x-2 px-4 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
          >
            <BookOpen className="w-4 h-4" />
            <span className="text-sm font-medium">Notes ({notesCount})</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="users" className="flex flex-col data-[state=active]:flex-1 m-0 space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
          {courseCodes.map(code => {
         
            if (!filteredUsers(code)?.length) return null
            return (
              <div key={code} className="space-y-4">
                <div className="flex items-center gap-2 pb-2 border-b border-white/10">
                  <h3 className="text-lg font-bold text-white">{code}</h3>
                  <Badge variant="outline" className="text-xs border-white/10 text-gray-400">
                    {filteredUsers(code).length} Users
                  </Badge>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {filteredUsers(code).map((u: models.User) => (
                    <UserItemCompact key={u.ID} user={u} />
                  ))}
                </div>
              </div>
            )
          })}
          {courseCodes.length === 0 && (
            <EmptyState
              icon={User}
              title="No linked users found"
              description="Share your courses with other users"
              className="flex-1 items-center"
              onClick={() => router.push("/courses")}
              buttonText="Go to Courses"
            />
          )}
        </TabsContent>

        <TabsContent value="assignments" className="flex flex-col data-[state=active]:flex-1 m-0 space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
          {courseCodes.map(code => {
            if (!filteredAssignments(code)?.length) return null
            return (
              <div key={code} className="space-y-4">
                <div className="flex items-center gap-2 pb-2 border-b border-white/10">
                  <h3 className="text-lg font-bold text-white">{code}</h3>
                  <Badge variant="outline" className="text-xs border-white/10 text-gray-400">
                    {filteredAssignments(code).length} Assignments
                  </Badge>
                </div>
                <Masonry
                  items={filteredAssignments(code)}
                  config={{
                    columns: [1, 2, 3],
                    gap: [16, 16, 16],
                    media: [640, 1024, 1280],
                  }}
                  render={(a: models.LocalAssignment, idx: number) => (
                    <div key={a.ID} className="mb-4">
                      <AssignmentItemCompact
                        assignment={a}
                        onClick={handleAssignmentClick}
                        disabled={false}
                        onCopy={handleCopyAssignment}
                      />
                    </div>
                  )}
                />
              </div>
            )
          })}
          {courseCodes.length === 0 && (
            <EmptyState
              icon={FileText}
              title="No linked assignments found"
              description="Create a new assignment in a shared course"
              className="flex-1 items-center"
              onClick={() => router.push("/assignments")}
              buttonText="Go to Assignments"
            />
          )}
        </TabsContent>

        <TabsContent value="notes" className="flex flex-col data-[state=active]:flex-1 m-0 space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
          {courseCodes.map(code => {
            if (!filteredNotes(code)?.length) return null
            return (
              <div key={code} className="space-y-4">
                <div className="flex items-center gap-2 pb-2 border-b border-white/10">
                  <h3 className="text-lg font-bold text-white">{code}</h3>
                  <Badge variant="outline" className="text-xs border-white/10 text-gray-400">
                    {filteredNotes(code).length} Notes
                  </Badge>
                </div>
                <Masonry
                  items={filteredNotes(code)}
                  config={{
                    columns: [1, 2, 3],
                    gap: [24, 24, 24],
                    media: [640, 1024, 1280],
                  }}
                  render={(n: models.LocalNote, idx: number) => (
                    <div key={n.ID} className="mb-4">
                      <NoteItemCompact
                        note={n}
                        disabled={false}
                        onCopy={() => { }}
                      />
                    </div>
                  )}
                />
              </div>
            )
          })}
          {courseCodes.length === 0 && (
            <EmptyState
              icon={BookOpen}
              title="No linked notes found"
              description="Create a new note in a shared course"
              className="flex-1 items-center"
              onClick={() => router.push("/notes")}
              buttonText="Go to Notes"
            />
          )}  
        </TabsContent>
      </Tabs>
      <AssignmentDetailsModal
        assignment_id={selectedAssignment?.ID}
        isOpen={!!selectedAssignment}
        onClose={() => setSelectedAssignment(null)}
        onCopy={handleCopyAssignment}
        assignmentProp={selectedAssignment}
      />
    </div>
  )
}
