"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { UserItemCompact } from "@/components/community/user-item-compact"
import { AssignmentItemCompact } from "@/components/assignments/assignment-item-compact"
import { NoteItemCompact } from "@/components/notes/note-item-compact"
import { Users, FileText, BookOpen, GraduationCap, Search } from "lucide-react"
import { assignment, note, user } from "@/wailsjs/go/models"
import { useState, useMemo } from "react"
import { Masonry } from "react-plock"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { useCoursesLinked } from "@/hooks/use-courses"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { AssignmentDetailsModal } from "../assignments/assignment-details-modal"

// Mock Data
const MOCK_USERS: user.User[] = [
  {
    ID: 1,
    Username: "Alice Johnson",
    Email: "alice@university.edu",
    University: "Stanford University",
    Avatar: "/placeholder-user.jpg",
    CoursesCode: ["CS101", "MATH202"],
    IsVerified: true,
  } as unknown as user.User,
  {
    ID: 2,
    Username: "Bob Smith",
    Email: "bob@university.edu",
    University: "MIT",
    Avatar: "/placeholder-user.jpg",
    CoursesCode: ["CS101", "PHYS101"],
    IsVerified: false,
  } as unknown as user.User,
  {
    ID: 3,
    Username: "Carol Williams",
    Email: "carol@university.edu",
    University: "Stanford University",
    Avatar: "/placeholder-user.jpg",
    CoursesCode: ["CS101", "BIO101"],
    IsVerified: true,
  } as unknown as user.User,
]

const MOCK_ASSIGNMENTS: assignment.LocalAssignment[] = [
  {
    ID: 101,
    Title: "Midterm Project: AI Ethics",
    StatusName: "In progress",
    Deadline: new Date(2024, 3, 15).toISOString(),
    Priority: "high",
    Type: { Name: "Project", Color: "bg-purple-500" },
    Course: { Code: "CS101", Color: "bg-blue-500" },
    CourseCode: "CS101",
    Link: "https://example.com",
    Todo: "Research paper on ethical implications of AI...",
  } as unknown as assignment.LocalAssignment,
  {
    ID: 102,
    Title: "Algorithm Analysis Essay",
    StatusName: "Done",
    Deadline: new Date(2024, 2, 28).toISOString(),
    Priority: "medium",
    Type: { Name: "Essay", Color: "bg-green-500" },
    Course: { Code: "CS101", Color: "bg-blue-500" },
    CourseCode: "CS101",
    Link: "https://example.com",
    Todo: "Analyze the time complexity of sorting algorithms...",
  } as unknown as assignment.LocalAssignment,
  {
    ID: 103,
    Title: "Calculus Problem Set 5",
    StatusName: "Not started",
    Deadline: new Date(2024, 3, 20).toISOString(),
    Priority: "medium",
    Type: { Name: "Homework", Color: "bg-orange-500" },
    Course: { Code: "MATH202", Color: "bg-red-500" },
    CourseCode: "MATH202",
    Link: "https://example.com",
    Todo: "Complete problems 1-10...",
  } as unknown as assignment.LocalAssignment,
]

const MOCK_NOTES: note.LocalNote[] = [
  {
    ID: 201,
    title: "Lecture 12: Neural Networks",
    subject: "Deep Learning",
    content: "Introduction to backpropagation...",
    course_code: "CS101",
    Course: { Code: "CS101", Color: "bg-blue-500" },
    CreatedAt: new Date(2024, 3, 10).toISOString(),
    UpdatedAt: new Date(2024, 3, 10).toISOString(),
    videos: "[]",
  } as unknown as note.LocalNote,
  {
    ID: 202,
    title: "Study Guide for Midterm",
    subject: "General",
    content: "Key concepts to review...",
    course_code: "CS101",
    Course: { Code: "CS101", Color: "bg-blue-500" },
    CreatedAt: new Date(2024, 3, 12).toISOString(),
    UpdatedAt: new Date(2024, 3, 12).toISOString(),
    videos: "[]",
  } as unknown as note.LocalNote,
  {
    ID: 203,
    title: "Derivatives and Integrals",
    subject: "Calculus",
    content: "Review of integration rules...",
    course_code: "MATH202",
    Course: { Code: "MATH202", Color: "bg-red-500" },
    CreatedAt: new Date(2024, 3, 14).toISOString(),
    UpdatedAt: new Date(2024, 3, 14).toISOString(),
    videos: "[]",
  } as unknown as note.LocalNote,
]

export function LinkedResources() {
  const { data: coursesLinked } = useCoursesLinked()
  const [activeTab, setActiveTab] = useState("users")
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedAssignment, setSelectedAssignment] = useState<assignment.LocalAssignment | null>(null)
  const [selectedNote, setSelectedNote] = useState<note.LocalNote | null>(null)


  /**
   * Handles assignment card click to open details modal.
   * 
   * @param {assignment.LocalAssignment} assignment - The assignment that was clicked
   */
  const handleAssignmentClick = (assignment: assignment.LocalAssignment) => {
    setSelectedAssignment(assignment)
  }

  const filteredUsers = useMemo(() => {
    if (!searchQuery) return MOCK_USERS
    const query = searchQuery.toLowerCase()
    return MOCK_USERS.filter(user =>
      user.CoursesCode?.some(code => code.toLowerCase().includes(query)) ||
      // Assuming courses have names, but user model only has codes. 
      // In a real app, we'd probably map codes to names or search other user fields.
      user.CoursesCode?.some(code => code.toLowerCase().includes(query))
    )
  }, [searchQuery])

  const filteredAssignments = useMemo(() => {
    if (!searchQuery) return MOCK_ASSIGNMENTS
    const query = searchQuery.toLowerCase()
    return MOCK_ASSIGNMENTS.filter(assignment =>
      assignment.CourseCode?.toLowerCase().includes(query) ||
      assignment.Course?.Name?.toLowerCase().includes(query)
    )
  }, [searchQuery])

  const filteredNotes = useMemo(() => {
    if (!searchQuery) return MOCK_NOTES
    const query = searchQuery.toLowerCase()
    return MOCK_NOTES.filter(note =>
      note.course_code?.toLowerCase().includes(query) ||
      note.Course?.Name?.toLowerCase().includes(query)
    )
  }, [searchQuery])

  // Group items by course code
  const groupedUsers = useMemo(() => {
    const groups: Record<string, user.User[]> = {}
    filteredUsers.forEach(user => {
      user.CoursesCode?.forEach(code => {
        // Filter groups by search query if it matches course code
        if (searchQuery && !code.toLowerCase().includes(searchQuery.toLowerCase())) return

        if (!groups[code]) groups[code] = []
        if (!groups[code].find(u => u.ID === user.ID)) {
          groups[code].push(user)
        }
      })
    })
    return groups
  }, [filteredUsers, searchQuery])

  const groupedAssignments = useMemo(() => {
    const groups: Record<string, assignment.LocalAssignment[]> = {}
    filteredAssignments.forEach(assignment => {
      const code = assignment.CourseCode
      // Filter groups by search query if it matches course code
      if (searchQuery && !code?.toLowerCase().includes(searchQuery.toLowerCase())) return

      if (!groups[code]) groups[code] = []
      groups[code].push(assignment)
    })
    return groups
  }, [filteredAssignments, searchQuery])

  const groupedNotes = useMemo(() => {
    const groups: Record<string, note.LocalNote[]> = {}
    filteredNotes.forEach(note => {
      const code = note.course_code
      // Filter groups by search query if it matches course code
      if (searchQuery && !code?.toLowerCase().includes(searchQuery.toLowerCase())) return

      if (!groups[code]) groups[code] = []
      groups[code].push(note)
    })
    return groups
  }, [filteredNotes, searchQuery])

  const courseCodes = useMemo(() => {
    const codes = Object.keys(coursesLinked || {})
    return codes.sort()
  }, [coursesLinked])


  return (
    <div className="space-y-6">
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

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="bg-white/5 p-1 rounded-xl border border-white/5 w-fit mb-6">
          <TabsTrigger
            value="users"
            className="flex items-center space-x-2 px-4 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
          >
            <Users className="w-4 h-4" />
            <span className="text-sm font-medium">Users ({MOCK_USERS.length})</span>
          </TabsTrigger>
          <TabsTrigger
            value="assignments"
            className="flex items-center space-x-2 px-4 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
          >
            <FileText className="w-4 h-4" />
            <span className="text-sm font-medium">Assignments ({MOCK_ASSIGNMENTS.length})</span>
          </TabsTrigger>
          <TabsTrigger
            value="notes"
            className="flex items-center space-x-2 px-4 py-2 text-gray-400 data-[state=active]:text-white data-[state=active]:bg-white/10 rounded-lg transition-all duration-200"
          >
            <BookOpen className="w-4 h-4" />
            <span className="text-sm font-medium">Notes ({MOCK_NOTES.length})</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="users" className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
          {courseCodes.map(code => {
            const courseData = coursesLinked?.[code]
            const users = courseData?.Users  // lowercase 'user' from API
            if (!users?.length) return null
            return (
              <div key={code} className="space-y-4">
                <div className="flex items-center gap-2 pb-2 border-b border-white/10">
                  <h3 className="text-lg font-bold text-white">{code}</h3>
                  <Badge variant="outline" className="text-xs border-white/10 text-gray-400">
                    {users.length} Users
                  </Badge>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {users.map((u: user.User) => (
                    <UserItemCompact key={u.ID} user={u} />
                  ))}
                </div>
              </div>
            )
          })}
          {courseCodes.length === 0 && (
            <div className="text-center py-12 text-gray-500">No linked users found</div>
          )}
        </TabsContent>

        <TabsContent value="assignments" className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
          {courseCodes.map(code => {
            const courseData = coursesLinked?.[code]
            const assignments = courseData?.Assignments  // lowercase 'assignments' from API
            if (!assignments?.length) return null
            return (
              <div key={code} className="space-y-4">
                <div className="flex items-center gap-2 pb-2 border-b border-white/10">
                  <h3 className="text-lg font-bold text-white">{code}</h3>
                  <Badge variant="outline" className="text-xs border-white/10 text-gray-400">
                    {assignments.length} Assignments
                  </Badge>
                </div>
                <Masonry
                  items={assignments}
                  config={{
                    columns: [1, 2, 3],
                    gap: [16, 16, 16],
                    media: [640, 1024, 1280],
                  }}
                  render={(a: assignment.LocalAssignment, idx: number) => (
                    <div key={a.ID} className="mb-4">
                      <AssignmentItemCompact
                        assignment={a}
                        onClick={handleAssignmentClick}
                        disabled={false}
                        onCopy={(copied) => console.log("Copy assignment", copied)}
                      />
                    </div>
                  )}
                />
              </div>
            )
          })}
          {courseCodes.length === 0 && (
            <div className="text-center py-12 text-gray-500">No linked assignments found</div>
          )}
        </TabsContent>

        <TabsContent value="notes" className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
          {courseCodes.map(code => {
            const courseData = coursesLinked?.[code]
            const notes = courseData?.Notes  // lowercase 'notes' from API
            if (!notes?.length) return null
            return (
              <div key={code} className="space-y-4">
                <div className="flex items-center gap-2 pb-2 border-b border-white/10">
                  <h3 className="text-lg font-bold text-white">{code}</h3>
                  <Badge variant="outline" className="text-xs border-white/10 text-gray-400">
                    {notes.length} Notes
                  </Badge>
                </div>
                <Masonry
                  items={notes}
                  config={{
                    columns: [1, 2, 3],
                    gap: [24, 24, 24],
                    media: [640, 1024, 1280],
                  }}
                  render={(n: note.LocalNote, idx: number) => (
                    <div key={n.ID} className="mb-4">
                      <NoteItemCompact
                        note={n}
                        disabled={false}
                        onCopy={(copied) => console.log("Copy note", copied)}
                      />
                    </div>
                  )}
                />
              </div>
            )
          })}
          {courseCodes.length === 0 && (
            <div className="text-center py-12 text-gray-500">No linked notes found</div>
          )}
        </TabsContent>
      </Tabs>
      <AssignmentDetailsModal
        assignment_id={selectedAssignment?.ID}
        isOpen={!!selectedAssignment}
        onClose={() => setSelectedAssignment(null)}
        onCopy={(copied) => console.log("Copy assignment", copied)}
        assignmentProp={selectedAssignment}
      />
    </div>
  )
}
