"use client"

import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { useState, useEffect } from "react"
import { user } from "@/wailsjs/go/models"

interface AuthState {
  user: user.User | null
}

export function useAuth() {

  const [authState, setAuthState] = useState<AuthState>({
    user: null,
  })

  const register = async (username: string, email: string, password: string, university: string, language: string) => {
    try {
<<<<<<< Updated upstream
      await window.go.main.App.Register(username, email, password, university, language)
=======
      if (!isWailsAvailable()) {
        throw new Error("Wails bindings not available")
      }
      const user = await window.go.main.App.Register(username, email, password, university, language)
>>>>>>> Stashed changes
      setAuthState({
        user: user!,
      })
      return { success: true }
    } catch (error) {
      console.log("Register error: ", error)
      return { success: false, error: error instanceof Error ? error.message : "Register failed" }
    }
  }

  const login = async (username: string, password: string) => {
    try {
<<<<<<< Updated upstream
      await window.go.main.App.Login(username, password)
=======
      if (!isWailsAvailable()) {
        throw new Error("Wails bindings not available")
      }
      const user = await window.go.main.App.Login(username, password)
>>>>>>> Stashed changes
      setAuthState({
        user: user!,
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
<<<<<<< Updated upstream
        if (!authState.isAuthenticated) {
          const creds: storage.LocalCredentials = await window.go.main.App.IsAuthenticated()
=======
        if (!isWailsAvailable()) {
          // If Wails is not available, wait a bit and try again
          setTimeout(checkAuth, 100)
          return
        }
        
        if (!authState.user) {
          const user = await window.go.main.App.IsAuthenticated()
>>>>>>> Stashed changes
          setAuthState({
            user: user!,
          })
        }
      } catch (error) {
        LogError("Error checking auth: " + error)
        // If there's an error checking auth, assume not authenticated
        setAuthState({
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
    register,
  }
}