"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { LogError } from "@/wailsjs/runtime/runtime"
import { user } from "@/wailsjs/go/models"
import { 
  GetCurrentUser, 
  Login, 
  UploadProfilePicture, 
  Register, 
  Logout, 
  UpdateUser, 
  GetAuthToken,
  GetFileAsDataURL
} from "@/wailsjs/go/main/App"
import { useAuthContext } from '@/components/provider/auth-provider'
import { courseKeys } from './use-courses'

// Query keys for auth
export const authKeys = {
  user: ['auth', 'user'] as const,
  followers: ['auth', 'followers'] as const,
  following: ['auth', 'following'] as const,
  token: ['auth', 'token'] as const,
}

// Main hook for fetching current user
export function useCurrentUser() {
  return useQuery({
    queryKey: authKeys.user,
    queryFn: async (): Promise<user.User | null> => {
      try {
        return await GetCurrentUser()
      } catch (error) {
        LogError("Failed to check authentication: " + error)
        return null
      }
    },
    retry: false, // Don't retry if authentication fails ! IMPORTANT
  })
}

export function useGetAuthToken() {
  return useQuery({
    queryKey: authKeys.token,
    queryFn: async () : Promise<string> => {
      try {
        var token = await GetAuthToken()
        console.log("useGetAuthToken", token)
        return token
      } catch (error) {
        LogError("Failed to get auth token: " + error)
        throw error
      }
    },
    retry: false, // Don't retry if authentication fails ! IMPORTANT
  })
}

// Login mutation
export function useLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ username, password }: { username: string; password: string }) => {
      return await Login(username, password)
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
      language,
      semester,
      year
    }: {
      username: string
      email: string
      password: string
      university: string
      language: string
      semester: string
      year: string
    }) => {
      return await Register({ 
        Username: username  , 
        Email: email, 
        PasswordHash: password, 
        University: university, 
        Language: language, 
        Semester: semester, 
        Year: year 
      } as user.User)
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
      return await Logout()
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
    mutationFn: async ({ column, key, value }: { column: string; key: string; value: string }) => {
      return await UpdateUser(column, value)
    },
    onMutate: async ({ column, key, value, }: { column: string; key: string; value: string }) => {
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
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: courseKeys.linked() })
    },
  })
}

export function useUploadProfilePicture() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      return await UploadProfilePicture()
    },
    onSuccess: (data) => {
      const previousUser = queryClient.getQueryData<user.User>(authKeys.user)
      queryClient.setQueryData(authKeys.user, { ...previousUser, Avatar: data })
      // Invalidate avatar queries to refetch with new path
      queryClient.invalidateQueries({ queryKey: ['avatar'] })
    },
    onError: (error) => {
      LogError("Failed to upload profile picture: " + error)
      const previousUser = queryClient.getQueryData<user.User>(authKeys.user)
      queryClient.setQueryData(authKeys.user, previousUser)
    },
  })
}

export function useGetAvatarUrl() {
  const { user: currentUser } = useAuthContext()
  return useQuery({
    queryKey: ['avatar', currentUser?.Avatar],
    queryFn: async () => {
      if (!currentUser?.Avatar) return "/placeholder.svg?height=40&width=40"

      // If it's already a URL (http/https/data), return as is
      if (currentUser.Avatar.startsWith("http://") ||
        currentUser.Avatar.startsWith("https://") ||
        currentUser.Avatar.startsWith("data:")) {
        return currentUser.Avatar
      }

      // If it's a local file path, convert to data URL
      if (currentUser.Avatar.startsWith("/")) {
        try {
          return await GetFileAsDataURL(currentUser.Avatar)
        } catch (error) {
          console.error("Failed to load avatar:", error)
          return "/placeholder.svg?height=40&width=40"
        }
      }


      // Fallback: treat as relative path
      return currentUser.Avatar
    },
    enabled: !!currentUser?.Avatar,
    // Remove staleTime: Infinity so the query refetches when Avatar path changes
    // The query key includes currentUser?.Avatar, so React Query will automatically
    // refetch when the Avatar path changes
    refetchOnMount: true,
    refetchOnWindowFocus: false,
  })
}