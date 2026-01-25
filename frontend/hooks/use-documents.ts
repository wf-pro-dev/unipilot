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
import { assignmentKeys } from './use-assignments'


// Query keys for consistent cache management
export const documentKeys = {
  storage: () => ['document-storage'] as const,
  assignmentStorage: (id: number) => ['assignment-storage', id] as const,
  rag: (assignmentId: number) => ['document-rag', assignmentId] as const,
}


const uploadKey = ['uploads'] as const

// Hook for fetching documents for an assignment
// Get documents from assignment cache
export function useDocuments(assignment: models.LocalAssignment) {
  const queryClient = useQueryClient()
  
  // Directly subscribe to assignment cache changes
  const { data: assignments } = useQuery({
    queryKey: assignmentKeys.lists(),
  })
  
  const currentAssignment = (assignments as models.LocalAssignment[])?.find(a => a.ID === assignment.ID)
  return {
    data: currentAssignment?.Documents ?? assignment.Documents ?? [],
    isLoading: false,
    isError: false,
  }
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

      const fileInfo = await GetFileInfo(filePath)

      var newDocument = new models.LocalDocument({
        ID: 0,
        AssignmentID: assignmentId,
        RemoteAssignmentID: remoteAssignmentId,
        Type: documentType,
        FilePath: filePath,
        FileName: fileInfo.FileName,
        FileSize: fileInfo.FileSize,
        UploadID: uploadId,
      })

      const previousAssignments = queryClient.getQueryData<models.LocalAssignment[]>(
        assignmentKeys.lists()
      )


      queryClient.setQueryData<models.LocalAssignment[]>(
        assignmentKeys.lists(), 
        (old) => old?.map((assignment: models.LocalAssignment) => 
          assignment.ID === assignmentId
            ? { 
                ...assignment, 
                Documents: [newDocument, ...(assignment.Documents || [])]
              } as models.LocalAssignment
            : assignment
        )
      )

      return { previousAssignments }

    },
    onSuccess: () => {
      toast.success("Document uploaded successfully")
    },
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to upload document: " + err)
      toast.error("Failed to upload document")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
      queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
    },
    // Always refetch to ensure consistency
    retry: false,
  },

  )
}

// Hook for downloading documents
export function useDownloadDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (document: models.LocalDocument) => {
      return await DownloadDocument(document)
    },
    onMutate: async (document) => {

      // Change HasLocalFile to true
      const previousAssignments = queryClient.getQueryData<models.LocalAssignment[]>(assignmentKeys.lists())

      queryClient.setQueryData<models.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return []
        return old.map((assignment: models.LocalAssignment) => 
          assignment.ID === document.AssignmentID
            ? { 
                ...assignment, 
                Documents: [document, ...(assignment.Documents || [])]
              } as models.LocalAssignment
            : assignment
        )
      })

      return { previousAssignments }
    },
    onSuccess: (_,variables) => {
      toast.success(variables.FileName + " downloaded successfully")
    },
    // If the mutation fails, rollback
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to update document: " + err)
      toast.error("Failed to download document")
    },
    onSettled: (_, err,variables) => {
      queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
      queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
    },
  })
}


// // Hook for uploading new document versions
// export function useUploadDocumentVersion() {
//   const queryClient = useQueryClient()

//   return useMutation({
//     mutationFn: async (documentId: number) => {
//       return await UploadNewDocumentVersion(documentId)
//     },

//     onSuccess: (newVersion, documentId) => {
//       if (newVersion) {
//         // Find the assignment ID from existing cache to update the right lists
//         const allQueries = queryClient.getQueriesData<models.LocalDocument[]>({
//           queryKey: documentKeys.lists()
//         })

//         // Update all relevant caches
//         allQueries.forEach(([queryKey, data]) => {
//           if (data && Array.isArray(data)) {
//             const hasDocument = data.some(doc => doc.ID === documentId)
//             if (hasDocument) {
//               queryClient.setQueryData<models.LocalDocument[]>(queryKey, (old) => {
//                 if (!old) return []
//                 return old.map(doc =>
//                   doc.ID === documentId
//                     ? newVersion // Replace with new version
//                     : doc
//                 )
//               })
//             }
//           }
//         })

//         // Invalidate storage info
//         queryClient.invalidateQueries({ queryKey: documentKeys.storage() })

//       }
//     },

//     onError: (err) => {
//       LogError("Failed to upload document version: " + err)
//     },
//   })
// }




// Hook for deleting documents
export function useDeleteDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (document: models.LocalDocument) => {
      console.log("Deleting document: " + document.ID)
      return await DeleteDocument(document.ID)
    },

    // Optimistically remove the document
    onMutate: async (document) => {
      // Find and update all relevant caches
      
      
      const previousAssignments = queryClient.getQueryData<models.LocalAssignment[]>(assignmentKeys.lists())

      queryClient.setQueryData<models.LocalAssignment[]>(assignmentKeys.lists(), (old) => {
        if (!old) return []
        return old.map((assignment: models.LocalAssignment) => 
          assignment.ID === document.AssignmentID
            ? { 
                ...assignment, 
                Documents: assignment.Documents?.filter(doc => doc.ID !== document.ID)
              } as models.LocalAssignment
            : assignment
        )
      })

      return { previousAssignments }
    },
    onSuccess: () => {
      toast.success("Document deleted successfully")
    },
    onError: (err, variables, context) => {
      if (context?.previousAssignments) {
        queryClient.setQueryData(assignmentKeys.lists(), context.previousAssignments)
      }
      LogError("Failed to delete document: " + err)
      toast.error("Failed to delete document" + err)
    },
    onSettled: () => {
      //queryClient.invalidateQueries({ queryKey: assignmentKeys.lists() })
      queryClient.invalidateQueries({ queryKey: documentKeys.storage() })
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

