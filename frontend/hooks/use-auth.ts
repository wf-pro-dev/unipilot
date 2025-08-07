"use client"

import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { useState, useEffect } from "react"
import { storage } from "@/wailsjs/go/models"

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

  const register = async (username: string, email: string, password: string, university: string, language: string) => {
    try {
      await window.go.main.App.Register(username, email, password, university, language)
      setAuthState({
        isAuthenticated: true,
        isLoading: false,
        user: { username },
      })
      return { success: true }
    } catch (error) {
      console.log("Register error: ", error)
      return { success: false, error: error instanceof Error ? error.message : "Register failed" }
    }
  }

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
        if (!authState.isAuthenticated) {
          const creds: storage.LocalCredentials = await window.go.main.App.IsAuthenticated()
          setAuthState({
            isAuthenticated: creds.is_authenticated,
            isLoading: false,
            user: creds.is_authenticated ? { username: creds.user.username } : null,
          })
        }
      } catch (error) {
        //LogError("Error checking auth: " + error)
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
    register,
  }
}