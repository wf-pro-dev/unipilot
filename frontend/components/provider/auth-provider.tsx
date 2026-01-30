"use client"

import { createContext, useContext, ReactNode, useEffect, useState } from "react"
import { useCurrentUser } from "@/hooks/use-auth"
import AuthPage from "../auth/page"
import { models } from "@/wailsjs/go/models"
import { useFollowers, useFollowing } from "@/hooks/use-follows"
import { useUsers } from "@/hooks/use-users"
import { useCourses } from "@/hooks/use-courses"
import { useAssignments } from "@/hooks/use-assignments"
import { useNotes } from "@/hooks/use-notes"

interface AuthContextType {
  user: models.User | null | undefined
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

  const { data: user, isLoading: isLoadingUser } = useCurrentUser()
  const { data: followers, refetch: refetchFollowers } = useFollowers(user?.ID ?? 0)
  const { data: following, refetch: refetchFollowing } = useFollowing(user?.ID ?? 0)
  const { data: users, refetch: refetchUsers } = useUsers()
  const { data: courses, refetch: refetchCourses } = useCourses()
  const { data: assignments, refetch: refetchAssignments } = useAssignments()
  const { data: notes, refetch: refetchNotes } = useNotes()


  useEffect(() => {
    if (user) {
      refetchFollowers()
      refetchFollowing()
      refetchUsers()
      refetchCourses()
      refetchAssignments()
      refetchNotes()
    }
    console.log("user",user)
  }, [user])

  if (isLoadingUser) {
    return <div>Loading...</div>
  }

  return (

    <div>
      {
        user ? (
          <AuthContext.Provider value={{ user, followers, following, users, courses, assignments, notes }
          }>
            {children}
          </AuthContext.Provider >
        ) : (
          <AuthPage onLoginSuccess={() => { }} />
        )}
    </div>
  )

} 