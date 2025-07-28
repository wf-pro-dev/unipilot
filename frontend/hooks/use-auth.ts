"use client"

import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { useState, useEffect } from "react"

interface AuthState {
  isAuthenticated: boolean
  isLoading: boolean
  user: any | null
}

export function useAuth() {

  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    user: null,
  })

  const login = async (username: string, password: string) => {
    try {
      await window.go.main.App.Login(username, password)
      setAuthState({
        isAuthenticated: true,
        isLoading: false,
        user: { username },
      })
      return { success: true }
    } catch (error) {
      return { 
        success: false, 
        error: error instanceof Error ? error.message : "Login failed" 
      }
    }
  }

  const logout = async () => {
    try {
      await window.go.main.App.Logout()
      setAuthState({
        isAuthenticated: false,
        isLoading: false,
        user: null,
      })
      return { success: true }
    } catch (error) {
      return { 
        success: false, 
        error: error instanceof Error ? error.message : "Logout failed" 
      }
    }
  }

  // Check authentication status on mount
  useEffect(() => {

    const checkAuth = async () => {
      try {
        const isAuthenticated = await window.go.main.App.IsAuthenticated()
        LogInfo("IsAuthenticated: " + isAuthenticated)
        setAuthState({
          isAuthenticated,
          isLoading: false,
          user: isAuthenticated ? { username: "User" } : null,
        })
      } catch (error) {
        LogError("Error checking auth: " + error)
        // If there's an error checking auth, assume not authenticated
        setAuthState({
          isAuthenticated: false,
          isLoading: false,
          user: null,
        })
      }
    }

    checkAuth()
  }, [])

  return {
    ...authState,
    login,
    logout,
  }
} 