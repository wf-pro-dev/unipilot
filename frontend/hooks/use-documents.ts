"use client"

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { models } from "@/wailsjs/go/models"
import { LogError } from "@/wailsjs/runtime/runtime"
import {
  GetFileInfo,
  GetSupportDocuments,
  GetSubmissionDocuments,
  GetUserStorageInfo,
  UploadDocument,
  UploadNewDocumentVersion,
  OpenDocument,
  SaveDocumentAs,
  DeleteDocument,
  DownloadDocument,
  UploadDocumentRAG,
  DeleteDocumentRAG,
  GetAssignmentDocumentIDsRAG,
  GetAssignmentStorageInfo
} from "@/wailsjs/go/main/App"
import { toast } from 'sonner'
import { uuidv4 } from 'zod/v4'


// Query keys for consistent cache management
export const documentKeys = {
  all: ['documents'] as const,
  lists: () => [...documentKeys.all, 'list'] as const,
  list: (assignmentId: number) => [...documentKeys.lists(), { assignmentId }] as const,
  support: (assignmentId: number) => [...documentKeys.all, 'support', assignmentId] as const,
  submissions: (assignmentId: number) => [...documentKeys.all, 'submissions', assignmentId] as const,
  rag: (assignmentId: number) => [...documentKeys.all, 'rag', assignmentId] as const,
  storage: () => [...documentKeys.all, 'storage'] as const,
  assignmentStorage: (assignmentID: number) => [...documentKeys.all, 'assignmentStorage', assignmentID] as const,
}

const uploadKey = ['uploads'] as const


// Hook for fetching support documents
export function useSupportDocuments(assignmentId: number) {
  return useQuery({
    queryKey: documentKeys.support(assignmentId),
    queryFn: async (): Promise<models.LocalDocument[]> => {
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
    queryFn: async (): Promise<models.LocalDocument[]> => {
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
    queryFn: async (): Promise<models.DocumentStorage> => {
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

export function useAssignmentStorageInfo(assignmentID: number) {
  return useQuery({
    queryKey: documentKeys.assignmentStorage(assignmentID),
    queryFn: async (): Promise<models.LocalAssignmentStorage> => {
      return await GetAssignmentStorageInfo(assignmentID)
    },
  })
}


interface FileInfo {
  FileName: string
  FileSize: number
}

// Hook for uploading documents
export function useUploadDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      assignmentId,
      remoteAssignmentId,
      documentType,
      filePath,
      uploadId
    }: {
      assignmentId: number
      remoteAssignmentId: number
      documentType: string
      filePath: string
      uploadId: string
    }) => {
      // Generate a unique upload ID
      return await UploadDocument(assignmentId, remoteAssignmentId, documentType, filePath, uploadId)
    },

    // Optimistically update the cache
    onMutate: async ({ assignmentId, remoteAssignmentId, documentType, filePath, uploadId }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: documentKeys.list(assignmentId) })
      await queryClient.cancelQueries({ queryKey: documentKeys.support(assignmentId) })
      await queryClient.cancelQueries({ queryKey: documentKeys.submissions(assignmentId) })

      const previousDocuments = queryClient.getQueryData<models.LocalDocument[]>(documentKeys.list(assignmentId))


      const fileInfo = await GetFileInfo(filePath)

      console.log("FileInfo", fileInfo)

      var newDocument = new models.LocalDocument({
        ID: 0,
        AssignmentID: assignmentId,
        RemoteAssignmentID: remoteAssignmentId,
        DocumentType: documentType,
        FilePath: filePath,
        FileName: fileInfo.FileName,
        FileSize: fileInfo.FileSize,
        UploadID: uploadId,
      })

      queryClient.setQueryData<models.LocalDocument[]>(documentKeys.list(assignmentId), (old) => {
        if (!old) return [newDocument]
        return [newDocument, ...old]
      })

      const typeKey = documentType === 'support'
        ? documentKeys.support(assignmentId)
        : documentKeys.submissions(assignmentId)

      queryClient.setQueryData<models.LocalDocument[]>(
        typeKey,
        (old) => old ? [newDocument, ...old] : [newDocument]
      )



      return { previousDocuments }

      // Note: We don't do optimistic updates for uploads since we need the actual file data
    },
    onSuccess: (data) => {
      toast.success(data?.FileName + " uploaded successfully")
    },
    onError: (err, variables, context) => {
      if (context?.previousDocuments) {
        queryClient.setQueryData(documentKeys.list(variables.assignmentId), context.previousDocuments)
      }
      LogError("Failed to upload document: " + err)
      toast.error("Failed to upload document")
    },
    // Always refetch to ensure consistency
    onSettled: (data, error, { assignmentId }) => {
      queryClient.invalidateQueries({ queryKey: documentKeys.list(assignmentId) })
      queryClient.invalidateQueries({ queryKey: documentKeys.support(assignmentId) })
      queryClient.invalidateQueries({ queryKey: documentKeys.submissions(assignmentId) })
      queryClient.invalidateQueries({ queryKey: documentKeys.assignmentStorage(assignmentId) })
    },
    retry: false,
  },

  )
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
        const allQueries = queryClient.getQueriesData<models.LocalDocument[]>({
          queryKey: documentKeys.lists()
        })

        // Update all relevant caches
        allQueries.forEach(([queryKey, data]) => {
          if (data && Array.isArray(data)) {
            const hasDocument = data.some(doc => doc.ID === documentId)
            if (hasDocument) {
              queryClient.setQueryData<models.LocalDocument[]>(queryKey, (old) => {
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

// Hook for downloading documents
export function useDownloadDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (document: models.LocalDocument) => {
      return await DownloadDocument(document)
    },
    onMutate: async (document) => {

      await queryClient.cancelQueries({ queryKey: documentKeys.list(document.AssignmentID) })

      // Change HasLocalFile to true
      const previousDocuments = queryClient.getQueryData<models.LocalDocument[]>(documentKeys.list(document.AssignmentID))

      queryClient.setQueryData<models.LocalDocument[]>(documentKeys.list(document.AssignmentID), (old) => {
        if (!old) return []
        return old.map(d => d.ID === document.ID ? {
          ...d,
          HasLocalFile: true
        } : d) as models.LocalDocument[]
      })

      return { previousDocuments }
    },
    // If the mutation fails, rollback
    onError: (err, variables, context) => {
      if (context?.previousDocuments) {
        queryClient.setQueryData(documentKeys.list(variables.AssignmentID), context.previousDocuments)
      }
      LogError("Failed to update document: " + err)
    },
    onSettled: () => {
      // Invalidate all queries
      queryClient.invalidateQueries({ queryKey: documentKeys.all })
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
      const allQueries = queryClient.getQueriesData<models.LocalDocument[]>({
        queryKey: documentKeys.lists()
      })

      const previousData: Array<[unknown, models.LocalDocument[] | undefined]> = []
      var deletedDocument: models.LocalDocument | undefined;
      allQueries.forEach(([queryKey, data]) => {
        if (data && Array.isArray(data)) {
          deletedDocument = data.find(doc => doc.ID === documentId)
          if (deletedDocument) {
            previousData.push([queryKey, data])
            queryClient.setQueryData<models.LocalDocument[]>(queryKey, (old) => {
              if (!old) return []
              return old.filter(doc => doc.ID !== documentId)
            })
          }
        }
      })

      return { previousData, deletedDocument }
    },

    // If the mutation fails, rollback
    onError: (err, variables, context) => {
      if (context?.previousData) {
        context.previousData.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey as readonly unknown[], data)
        })
      }
      toast.error("Failed to delete document")
    },

    onSuccess: () => {
      // Invalidate storage info to refresh quota
      queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
      toast.success("Document deleted successfully")
    },

    // Always refetch after error or success to ensure consistency
    onSettled: (data, error, variables, context) => {
      queryClient.invalidateQueries({ queryKey: documentKeys.all })
      if (context?.deletedDocument) {
        queryClient.invalidateQueries({ queryKey: documentKeys.assignmentStorage(context.deletedDocument.AssignmentID) })
      }

    },
  })
}

// Hook to get the current upload state
export function useUpload() {
  const queryClient = useQueryClient();

  const { data: activeUploads = new Set() } = useQuery({
    queryKey: uploadKey,
    queryFn: () => new Set(),
    staleTime: Infinity,
    gcTime: Infinity,
  });

  const addUpload = (uploadId: string) => {
    queryClient.setQueryData(uploadKey, (old: Set<string> = new Set()) => {
      const newSet = new Set(old);
      newSet.add(uploadId);
      return newSet;
    });
  };

  const removeUpload = (uploadId: string) => {
    queryClient.setQueryData(uploadKey, (old: Set<string> = new Set()) => {
      const newSet = new Set(old);
      newSet.delete(uploadId);
      return newSet;
    });
  };

  const isUploading = (uploadId: string) => {
    return activeUploads.has(uploadId);
  };

  return {
    activeUploads,
    addUpload,
    removeUpload,
    isUploading,
    uploadCount: activeUploads.size,
  };
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

export function useAcceptDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (documentData: string) => {
      return await window.go.main.App.AcceptDocument(documentData)
    },

    onSuccess: async (newDocument) => {
      await queryClient.cancelQueries({ queryKey: documentKeys.lists() })

      const previousAssignments = queryClient.getQueryData<models.LocalDocument[]>(documentKeys.lists())

      queryClient.setQueryData<models.LocalDocument[]>(documentKeys.lists(), (old) => {
        if (!old) return [newDocument]
        return [newDocument, ...old]
      })

      return { previousAssignments }
    },

    onError: (err) => {
      LogError("Failed to accept document: " + err)
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: documentKeys.lists() })
    },
  })
}


// Utility hook to get all document-related data for an assignment
export function useAssignmentDocumentData(assignmentId: number) {
  const supportDocuments = useSupportDocuments(assignmentId)
  const submissionDocuments = useSubmissionDocuments(assignmentId)

  return {
    supportDocuments,
    submissionDocuments,
    isLoading: supportDocuments.isLoading || submissionDocuments.isLoading,
    error: supportDocuments.error || submissionDocuments.error,
  }
}

// Hook for uploading documents to RAG
export function useUploadDocumentRAG() {
  const queryClient = useQueryClient()

  return useMutation({
    onMutate: async (document) => {
      await queryClient.cancelQueries({ queryKey: documentKeys.rag(document.RemoteAssignmentID) })

      // Update all queries that match the base key
      queryClient.setQueriesData<number[]>({ queryKey: documentKeys.rag(document.RemoteAssignmentID) }, (old) => {
        if (!old) return [document.RemoteID]
        return old.includes(document.RemoteID) ? old : [...old, document.RemoteID]
      })
    },
    mutationFn: async (document: models.LocalDocument) => {
      return await UploadDocumentRAG(document)
    },
    onError: (err, variables, context) => {
      // Since we updated multiple queries potentially, invalidation is the safest rollback
      queryClient.invalidateQueries({ queryKey: documentKeys.rag(variables.RemoteAssignmentID) })
      LogError("Failed to upload document to RAG: " + err)
    },
    onSettled: (data, error, variables, context) => {
      if (!error) {
        queryClient.invalidateQueries({ queryKey: documentKeys.rag(variables.RemoteAssignmentID) })
      }
    },
  })
}

// Hook for deleting documents from RAG
export function useDeleteDocumentRAG() {
  const queryClient = useQueryClient()
  return useMutation({
    onMutate: async (document) => {
      await queryClient.cancelQueries({ queryKey: documentKeys.rag(document.RemoteAssignmentID) })

      // Update all queries that match the base key
      queryClient.setQueriesData<number[]>({ queryKey: documentKeys.rag(document.RemoteAssignmentID) }, (old) => {
        if (!old) return []
        return old.filter(id => id !== document.RemoteID)
      })
    },
    mutationFn: async (document: models.LocalDocument) => {
      return await DeleteDocumentRAG(document.RemoteAssignmentID, document.RemoteID)
    },
    onError: (err, variables, context) => {
      queryClient.invalidateQueries({ queryKey: documentKeys.rag(variables.RemoteAssignmentID) })
      LogError("Failed to delete document from RAG: " + err)
    },
    onSettled: (data, error, variables, context) => {
      if (!error) {
        queryClient.invalidateQueries({ queryKey: documentKeys.rag(variables.RemoteAssignmentID) })
      }
    },
  })
}

// Hook for fetching all documents for an assignment
export function useAssignmentDocumentIDsRAG(assignmentId: number) {
  const queryClient = useQueryClient()
  return useQuery({
    queryKey: [...documentKeys.rag(assignmentId)],
    queryFn: async (): Promise<number[]> => {
      try {
        const docIds = await GetAssignmentDocumentIDsRAG(assignmentId)
        return docIds
      } catch (error) {
        LogError("Failed to fetch assignment document IDs: " + error)
        throw new Error(error instanceof Error ? error.message : "Failed to fetch document IDs")
      }
    },
    enabled: !!assignmentId,
    staleTime: 60 * 60 * 1000, // Consider fresh for 1 hour
    gcTime: 120 * 60 * 1000,    // Keep in cache for 2 hours
    refetchOnMount: true, // ✅ Added: Always refetch when component mounts
  })
}

