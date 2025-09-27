"use client"

import { createContext, useContext, ReactNode, useEffect } from "react"
import { useCurrentUser } from "@/hooks/use-auth"
import AuthPage from "../auth/page"
import { assignment, course, note, user, notifications } from "@/wailsjs/go/models"
import { useFollowers, useFollowing } from "@/hooks/use-follows"
import { useUsers } from "@/hooks/use-users"
import { useCourses } from "@/hooks/use-courses"
import { useAssignments } from "@/hooks/use-assignments"
import { useNotes } from "@/hooks/use-notes"
import { useNotifications } from "@/hooks/use-notifications"
import { LogInfo } from "@/wailsjs/runtime/runtime"


interface AuthContextType {
  user: user.User | undefined
  followers: user.User[] | undefined
  following: user.User[] | undefined
  users: user.User[] | undefined
  courses: course.LocalCourse[] | undefined
  assignments: assignment.LocalAssignment[] | undefined
  notes: note.LocalNote[] | undefined
  notifications: notifications.LocalNotification[] | undefined
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function useAuthContext() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuthContext must be used within an AuthProvider")
  }
  return context
}

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const { data: user, isLoading } = useCurrentUser()
  const { data: followers, refetch: refetchFollowers } = useFollowers(user?.ID as number)
  const { data: following, refetch: refetchFollowing } = useFollowing(user?.ID as number)
  const { data: users, refetch: refetchUsers } = useUsers()
  const { data: courses, refetch: refetchCourses } = useCourses()
  const { data: assignments, refetch: refetchAssignments } = useAssignments()
  const { data: notes, refetch: refetchNotes } = useNotes()
  const { data: notifications, refetch: refetchNotifications } = useNotifications() 
  useEffect(() => {
    if (user) {
        refetchFollowers()
        refetchFollowing()
        refetchUsers()
        refetchCourses()
        refetchAssignments()
        refetchNotes()
        refetchNotifications()
    }
    LogInfo("AuthProvider: Refetching data for user: \n" + user)
  }, [user])
  
  if (isLoading) {
    return <div>Loading...</div> // Or a proper loading component
  }

  if (!user) {
    return <AuthPage />
  }

  return (
    <AuthContext.Provider value={{ user, followers, following, users, courses, assignments, notes, notifications }}>
      {children}
    </AuthContext.Provider>
  )
} 