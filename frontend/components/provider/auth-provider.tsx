"use client"

import { createContext, useContext, ReactNode } from "react"
import { useAuth } from "@/hooks/use-auth"
import AuthPage from "../auth/page"
import { user } from "@/wailsjs/go/models"
import { LogInfo } from "@/wailsjs/runtime/runtime"


interface AuthContextType {
  user: user.User | null
  login: (username: string, password: string) => Promise<{ success: boolean; error?: string }>
  logout: () => Promise<{ success: boolean; error?: string }>
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

  const auth = useAuth()

  if (auth.user === null) {
    return <AuthPage />
  }

  // Show the app if authenticated
  return (
    <AuthContext.Provider value={auth}>
      {children}
    </AuthContext.Provider>
  )
} 