"use client"

import { useState, useCallback, useEffect } from "react"
import { DocumentUploadDialog } from "../documents/document-upload-dialog"
import {
  FileText,

} from "lucide-react"
import { models } from "@/wailsjs/go/models"
import { useAssignmentDocumentData } from "@/hooks/use-documents"
import useEmblaCarousel from 'embla-carousel-react'
import FileUpload05 from "../file-upload-05"


interface AssignmentDocumentsProps {
  assignment: models.LocalAssignment
  documents?: models.LocalDocument[]
  viewMode?: boolean
}

type DocumentFilter = "All" | "support" | "submission"

export function AssignmentDocuments({ assignment, documents, viewMode = false }: AssignmentDocumentsProps) {
  const [filter, setFilter] = useState<DocumentFilter>("All")
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [isSelectOpen, setIsSelectOpen] = useState(false)


  const documentFilters: DocumentFilter[] = [
    "All",
    "support",
    "submission"
  ]

  // Embla Carousel setup
  const [emblaRef, emblaApi] = useEmblaCarousel({
    align: 'start',
    containScroll: 'trimSnaps',
    dragFree: false,
    skipSnaps: false
  })

  // Track current slide
  const onSelect = useCallback(() => {
    if (!emblaApi) return
    setSelectedIndex(emblaApi.selectedScrollSnap())
  }, [emblaApi])

  useEffect(() => {
    if (!emblaApi) return
    onSelect()
    emblaApi.on('select', onSelect)
    return () => {
      emblaApi.off('select', onSelect)
    }
  }, [emblaApi, onSelect])

  // Reset to first page when filter changes
  useEffect(() => {
    setSelectedIndex(0)
    if (emblaApi) {
      emblaApi.scrollTo(0)
    }
  }, [filter, emblaApi])

  // Use the utility hook to get All document data
  const {
    supportDocuments,
    submissionDocuments,
    isLoading,
    error
  } = useAssignmentDocumentData(assignment.ID)

  // Get filtered documents
  const getFilteredDocuments = () => {
    switch (filter) {
      case "support":
        return supportDocuments.data || []
      case "submission":
        return submissionDocuments.data || []
      default:
        return supportDocuments.data || []
    }
  }

  const filteredDocs = documents || getFilteredDocuments()

  // Group documents into pages of 4 (2x2 grid)
  const documentsPerPage = 3
  const documentPages = []
  for (let i = 0; i < filteredDocs.length; i += documentsPerPage) {
    documentPages.push(filteredDocs.slice(i, i + documentsPerPage))
  }

  // Carousel navigation functions
  const scrollPrev = useCallback(() => {
    if (emblaApi) emblaApi.scrollPrev()
  }, [emblaApi])

  const scrollNext = useCallback(() => {
    if (emblaApi) emblaApi.scrollNext()
  }, [emblaApi])


  const handleUpload = (type: "support" | "submission") => {
    setUploadType(type)
    setUploadDialogOpen(true)
  }

  const handleUploadComplete = () => {
    setUploadDialogOpen(false)
    // The hooks will automatically refetch and update the UI
  }

  // Handle loading state
  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex gap-2 items-center">
          <FileText className="w-5 h-5 text-muted-foreground" />
          <h3 className="text-lg font-semibold">Documents</h3>
        </div>
        <div className="flex justify-center items-center py-8">
          <div className="text-sm text-muted-foreground">Loading documents...</div>
        </div>
      </div>
    )
  }

  // Handle error state
  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex gap-2 items-center">
          <FileText className="w-5 h-5 text-muted-foreground" />
          <h3 className="text-lg font-semibold">Documents</h3>
        </div>
        <div className="flex justify-center items-center py-8">
          <div className="text-sm text-red-500">Failed to load documents</div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1">
      {/* Header */}

      <FileUpload05
        assignment={assignment}
        documents={filteredDocs}
        documentType={uploadType}
        onButtonClick={() => setUploadDialogOpen(true)}
      />


      {/* Upload Dialog */}
      <DocumentUploadDialog
        isOpen={uploadDialogOpen}
        onClose={() => setUploadDialogOpen(false)}
        onUploadComplete={handleUploadComplete}
        assignmentId={assignment.ID}
        remoteAssignmentId={assignment.RemoteID}
        documentType={uploadType}
      />
    </div>
  )
} 