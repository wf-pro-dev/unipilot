"use client"

import { createContext, useContext, ReactNode, useEffect } from "react"
import { useCurrentUser, useGetAuthToken } from "@/hooks/use-auth"
import AuthPage from "../auth/page"
import { models } from "@/wailsjs/go/models"
import { useFollowers, useFollowing } from "@/hooks/use-follows"
import { useUsers } from "@/hooks/use-users"
import { useCourses } from "@/hooks/use-courses"
import { useAssignments } from "@/hooks/use-assignments"
import { useNotes } from "@/hooks/use-notes"

interface AuthContextType {
  user: models.User | undefined
  token: string | undefined
  followers: models.User[] | undefined
  following: models.User[] | undefined
  users: models.User[] | undefined
  courses: models.LocalCourse[] | undefined
  assignments: models.LocalAssignment[] | undefined
  notes: models.LocalNote[] | undefined
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
  const { data: token,refetch: refetchToken } = useGetAuthToken()
  const { data: followers, refetch: refetchFollowers } = useFollowers(user?.ID as number)
  const { data: following, refetch: refetchFollowing } = useFollowing(user?.ID as number)
  const { data: users, refetch: refetchUsers } = useUsers()
  const { data: courses, refetch: refetchCourses } = useCourses()
  const { data: assignments, refetch: refetchAssignments } = useAssignments()
  const { data: notes, refetch: refetchNotes } = useNotes()
  useEffect(() => {
    if (user) {
        refetchToken()
        refetchFollowers()
        refetchFollowing()
        refetchUsers()
        refetchCourses()
        refetchAssignments()
        refetchNotes()
    }
  }, [user])
  
  if (isLoading) {
    return <div>Loading...</div> // Or a proper loading component
  }

  if (!user) {
    return <AuthPage onLoginSuccess={() => {}} />
  }

  return (
    <AuthContext.Provider value={{ user, token, followers, following, users, courses, assignments, notes }}>
      {children}
    </AuthContext.Provider>
  )
} 