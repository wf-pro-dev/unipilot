"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { document } from "@/wailsjs/go/models"
import { LogError } from "@/wailsjs/runtime/runtime"
import { 
  GetAssignmentDocuments,
  GetSupportDocuments,
  GetSubmissionDocuments,
  GetUserStorageInfo,
  UploadDocument,
  UploadNewDocumentVersion,
  OpenDocument,
  SaveDocumentAs,
  DeleteDocument
} from "@/wailsjs/go/main/App"

// Query keys for consistent cache management
export const documentKeys = {
  all: ['documents'] as const,
  lists: () => [...documentKeys.all, 'list'] as const,
  list: (assignmentId: number) => [...documentKeys.lists(), { assignmentId }] as const,
  support: (assignmentId: number) => [...documentKeys.all, 'support', assignmentId] as const,
  submissions: (assignmentId: number) => [...documentKeys.all, 'submissions', assignmentId] as const,
  storage: () => [...documentKeys.all, 'storage'] as const,
}

// Hook for fetching all documents for an assignment
export function useAssignmentDocuments(assignmentId: number) {
  return useQuery({
    queryKey: documentKeys.list(assignmentId),
    queryFn: async (): Promise<document.LocalDocument[]> => {
      try {
        const docs = await GetAssignmentDocuments(assignmentId)
        return docs || []
      } catch (error) {
        LogError("Failed to fetch assignment documents: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch documents")
      }
    },
    enabled: !!assignmentId,
    staleTime: 1 * 60 * 1000, // Consider fresh for 1 minute
    gcTime: 5 * 60 * 1000,    // Keep in cache for 5 minutes
  })
}

// Hook for fetching support documents
export function useSupportDocuments(assignmentId: number) {
  return useQuery({
    queryKey: documentKeys.support(assignmentId),
    queryFn: async (): Promise<document.LocalDocument[]> => {
      try {
        const docs = await GetSupportDocuments(assignmentId)
        return docs || []
      } catch (error) {
        LogError("Failed to fetch support documents: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch support documents")
      }
    },
    enabled: !!assignmentId,
    staleTime: 1 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  })
}

// Hook for fetching submission documents
export function useSubmissionDocuments(assignmentId: number) {
  return useQuery({
    queryKey: documentKeys.submissions(assignmentId),
    queryFn: async (): Promise<document.LocalDocument[]> => {
      try {
        const docs = await GetSubmissionDocuments(assignmentId)
        return docs || []
      } catch (error) {
        LogError("Failed to fetch submission documents: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch submission documents")
      }
    },
    enabled: !!assignmentId,
    staleTime: 1 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  })
}

// Hook for fetching user storage information
export function useUserStorageInfo() {
  return useQuery({
    queryKey: documentKeys.storage(),
    queryFn: async (): Promise<document.StorageInfo> => {
      try {
        return await GetUserStorageInfo()
      } catch (error) {
        LogError("Failed to fetch storage info: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch storage info")
      }
    },
    staleTime: 5 * 60 * 1000, // Consider fresh for 5 minutes
    gcTime: 15 * 60 * 1000,   // Keep in cache for 15 minutes
  })
}

// Hook for uploading documents
export function useUploadDocument() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async ({ 
      assignmentId, 
      documentType 
    }: { 
      assignmentId: number
      documentType: string 
    }) => {
      return await UploadDocument(assignmentId, documentType)
    },
    
    // Optimistically update the cache
    onMutate: async ({ assignmentId }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: documentKeys.list(assignmentId) })
      await queryClient.cancelQueries({ queryKey: documentKeys.support(assignmentId) })
      await queryClient.cancelQueries({ queryKey: documentKeys.submissions(assignmentId) })
      
      // Note: We don't do optimistic updates for uploads since we need the actual file data
      // Instead, we just prepare for the refetch
    },
    
    onSuccess: (newDocument, { assignmentId, documentType }) => {
      // Add the new document to the appropriate caches
      if (newDocument) {
        // Update all documents list
        queryClient.setQueryData<document.LocalDocument[]>(
          documentKeys.list(assignmentId), 
          (old) => old ? [newDocument, ...old] : [newDocument]
        )
        
        // Update specific type list
        const typeKey = documentType === 'support' 
          ? documentKeys.support(assignmentId)
          : documentKeys.submissions(assignmentId)
          
        queryClient.setQueryData<document.LocalDocument[]>(
          typeKey, 
          (old) => old ? [newDocument, ...old] : [newDocument]
        )
        
        // Invalidate storage info to refresh quota
        queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
      }
    },
    
    onError: (err) => {
      LogError("Failed to upload document: " + err)
    },
    
    // Always refetch to ensure consistency
    onSettled: (data, error, { assignmentId }) => {
      queryClient.invalidateQueries({ queryKey: documentKeys.list(assignmentId) })
      queryClient.invalidateQueries({ queryKey: documentKeys.support(assignmentId) })
      queryClient.invalidateQueries({ queryKey: documentKeys.submissions(assignmentId) })
    },
  })
}

// Hook for uploading new document versions
export function useUploadDocumentVersion() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (documentId: number) => {
      return await UploadNewDocumentVersion(documentId)
    },
    
    onSuccess: (newVersion, documentId) => {
      if (newVersion) {
        // Find the assignment ID from existing cache to update the right lists
        const allQueries = queryClient.getQueriesData<document.LocalDocument[]>({ 
          queryKey: documentKeys.lists() 
        })
        
        // Update all relevant caches
        allQueries.forEach(([queryKey, data]) => {
          if (data && Array.isArray(data)) {
            const hasDocument = data.some(doc => doc.ID === documentId)
            if (hasDocument) {
              queryClient.setQueryData<document.LocalDocument[]>(queryKey, (old) => {
                if (!old) return []
                return old.map(doc => 
                  doc.ID === documentId 
                    ? newVersion // Replace with new version
                    : doc
                )
              })
            }
          }
        })
        
        // Invalidate storage info
        queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
      }
    },
    
    onError: (err) => {
      LogError("Failed to upload document version: " + err)
    },
  })
}

// Hook for deleting documents
export function useDeleteDocument() {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (documentId: number) => {
      return await DeleteDocument(documentId)
    },
    
    // Optimistically remove the document
    onMutate: async (documentId) => {
      // Find and update all relevant caches
      const allQueries = queryClient.getQueriesData<document.LocalDocument[]>({ 
        queryKey: documentKeys.lists() 
      })
      
      const previousData: Array<[unknown, document.LocalDocument[] | undefined]> = []
      
      allQueries.forEach(([queryKey, data]) => {
        if (data && Array.isArray(data)) {
          const hasDocument = data.some(doc => doc.ID === documentId)
          if (hasDocument) {
            previousData.push([queryKey, data])
            queryClient.setQueryData<document.LocalDocument[]>(queryKey, (old) => {
              if (!old) return []
              return old.filter(doc => doc.ID !== documentId)
            })
          }
        }
      })
      
      return { previousData }
    },
    
    // If the mutation fails, rollback
    onError: (err, variables, context) => {
      if (context?.previousData) {
        context.previousData.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey as readonly unknown[], data)
        })
      }
      LogError("Failed to delete document: " + err)
    },
    
    onSuccess: () => {
      // Invalidate storage info to refresh quota
      queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
    },
    
    // Always refetch after error or success to ensure consistency
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: documentKeys.all })
    },
  })
}

// Hook for opening documents
export function useOpenDocument() {
  return useMutation({
    mutationFn: async (documentId: number) => {
      return await OpenDocument(documentId)
    },
    
    onError: (err) => {
      LogError("Failed to open document: " + err)
    },
  })
}

// Hook for saving documents
export function useSaveDocumentAs() {
  return useMutation({
    mutationFn: async (documentId: number) => {
      return await SaveDocumentAs(documentId)
    },
    
    onError: (err) => {
      LogError("Failed to save document: " + err)
    },
  })
}

// Utility hook to get all document-related data for an assignment
export function useAssignmentDocumentData(assignmentId: number) {
  const allDocuments = useAssignmentDocuments(assignmentId)
  const supportDocuments = useSupportDocuments(assignmentId)
  const submissionDocuments = useSubmissionDocuments(assignmentId)
  const storageInfo = useUserStorageInfo()
  
  return {
    allDocuments,
    supportDocuments, 
    submissionDocuments,
    storageInfo,
    isLoading: allDocuments.isLoading || supportDocuments.isLoading || submissionDocuments.isLoading,
    error: allDocuments.error || supportDocuments.error || submissionDocuments.error,
  }
} 