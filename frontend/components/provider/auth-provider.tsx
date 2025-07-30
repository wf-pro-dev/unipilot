"use client"

import { createContext, useContext, ReactNode, useEffect } from "react"
import { useAuth } from "@/hooks/use-auth"
import { LoginForm } from "../login/login-form"
import { LogInfo } from "@/wailsjs/runtime/runtime"

interface AuthContextType {
  isAuthenticated: boolean
  isLoading: boolean
  user: any | null
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

  // Show loading spinner while checking authentication
  if (auth.isLoading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="w-32 h-32 rounded-full border-b-2 border-blue-500 animate-spin"></div>
      </div>
    )
  }

  // Show login form if not authenticated
  if (!auth.isAuthenticated) {
    return <LoginForm onLoginSuccess={() => window.location.reload()} />
  }

  // Show the app if authenticated
  return (
    <AuthContext.Provider value={auth}>
      {children}
    </AuthContext.Provider>
  )
} 