"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError, LogInfo } from "@/wailsjs/runtime/runtime"
import { user } from "@/wailsjs/go/models"

// Query keys for auth
export const authKeys = {
  user: ['auth', 'user'] as const,
  followers: ['auth', 'followers'] as const,
  following: ['auth', 'following'] as const,
}

// Main hook for fetching current user
export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.user,
    queryFn: async (): Promise<user.User | null> => {
      try {
        return await window.go.main.App.IsAuthenticated()
      } catch (error) {
        LogError("Failed to check authentication: " + error)
        return null
      }
    },
    retry: false,
  })
}

// Login mutation
export function useLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ username, password }: { username: string; password: string }) => {
      return await window.go.main.App.Login(username, password)
    },
    onSuccess: (user) => {
      // Update the user cache
      queryClient.setQueryData(authKeys.user, user)
      // Invalidate followers/following to refetch them
      queryClient.invalidateQueries({ queryKey: authKeys.followers })
      queryClient.invalidateQueries({ queryKey: authKeys.following })
    },
    onError: (error) => {
      LogError("Login failed: " + error)

    },
  })
}

// Register mutation
export function useRegister() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      username,
      email,
      password,
      university,
      language
    }: {
      username: string
      email: string
      password: string
      university: string
      language: string
    }) => {
      return await window.go.main.App.Register(username, email, password, university, language)
    },
    onSuccess: (user) => {
      queryClient.setQueryData(authKeys.user, user)
    },
    onError: (error) => {
      LogError("Registration failed: " + error)
    },
  })
}

// Logout mutation
export function useLogout() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      return await window.go.main.App.Logout()
    },
    onSuccess: () => {
      // Clear all auth-related cache
      queryClient.setQueryData(authKeys.user, null)
      queryClient.removeQueries({ queryKey: authKeys.followers })
      queryClient.removeQueries({ queryKey: authKeys.following })
    },
    onError: (error) => {
      LogError("Logout failed: " + error)
    },
  })
}

// Update user mutation
export function useUpdateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ column, key , value }: { column: string; key:string ;value: string }) => {
      return await window.go.main.App.UpdateUser(column, value)
    },
    onMutate: async ({ column, key, value, }: { column: string; key: string; value: string  }) => {
      await queryClient.cancelQueries({ queryKey: authKeys.user })
      const previousUser = queryClient.getQueryData<user.User>(authKeys.user)
      queryClient.setQueryData(authKeys.user, { ...previousUser, [key]: value })
      return { previousUser }
    },
    onError: (error, variables, context) => {
      if (context?.previousUser) {
        queryClient.setQueryData(authKeys.user, context.previousUser)
      }
      LogError("Failed to update user: " + error)
    },
  })
}


