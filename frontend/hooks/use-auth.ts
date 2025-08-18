"use client"

import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { useState, useEffect } from "react"
import { user } from "@/wailsjs/go/models"
import { useFollowers, useFollowing } from "./use-follows"

interface AuthState {
  user: user.User | null,
  followers: user.User[] | [],
  following: user.User[] | [],
}

// Helper function to check if Wails bindings are available
const isWailsAvailable = (): boolean => {
  return typeof window !== 'undefined' &&
    !!window.go &&
    !!window.go.main &&
    !!window.go.main.App
}

export function useAuth() {

  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    followers: [],
    following: [],
  })

  const register = async (username: string, email: string, password: string, university: string, language: string) => {
    try {
      const user = await window.go.main.App.Register(username, email, password, university, language)
      setAuthState({
        ...authState,
        user: user!,
      })
      return { success: true }
    } catch (error) {

      return { success: false, error: error instanceof Error ? error.message : "Register failed" }
    }
  }

  const login = async (username: string, password: string) => {
    try {
      const user = await window.go.main.App.Login(username, password)
      setAuthState({
        ...authState,
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
      if (!isWailsAvailable()) {
        throw new Error("Wails bindings not available")
      }
      await window.go.main.App.Logout()
      setAuthState({
        user: null,
        followers: [],
        following: [],
      })
      return { success: true }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Logout failed"
      }
    }
  }

  const useUpdateUser = async ( column:string, value: string) : Promise<user.User | null> => {
    const oldUser = authState.user
    const key = column.slice(0, 1).toUpperCase() + column.slice(1)
    try {
      
      const user = await window.go.main.App.UpdateUser(column, value)
      return user

    } catch (error) {
      LogError("Error updating user: " + error)
      setAuthState({
        ...authState,
        user: oldUser,
      })
      return null
    }
  }

  const setFollowers = (followers: user.User[]) => {
    setAuthState({
      ...authState,
      followers: followers,
    })
  }
  
  const setFollowing = (following: user.User[]) => {
    setAuthState({
      ...authState,
      following: following,
    })
  }

  // Check authentication status on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        if (!authState.user) {
          const user = await window.go.main.App.IsAuthenticated()
          setAuthState({
            user: user!,
            followers: [],
            following: [],
          })
        }
      } catch (error) {
        LogError("Error checking auth: " + error)
        // If there's an error checking auth, assume not authenticated
        setAuthState({
          ...authState,
          user: null
        })

      }
    }

    checkAuth()
  }, [])

  return {
    ...authState,
    setAuthState,
    setFollowers,
    setFollowing,
    useUpdateUser,
    login,
    logout,
    register,
  }
}