"use client"

import { useState, useCallback, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { DocumentItem } from "../documents/document-item"
import { DocumentUploadDialog } from "../documents/document-upload-dialog"
import { DocumentStorageInfo } from "../documents/document-storage-info"
import {
  FileText,
  Upload,
  Folder,
  FileCheck,
  Send,
  ChevronLeft,
  ChevronRight
} from "lucide-react"
import { assignment, document } from "@/wailsjs/go/models"
import { useAssignmentDocumentData } from "@/hooks/use-documents"
import useEmblaCarousel from 'embla-carousel-react'

interface AssignmentDocumentsProps {
  assignment: assignment.LocalAssignment
}

type DocumentFilter = "all" | "support" | "submission"

export function AssignmentDocuments({ assignment }: AssignmentDocumentsProps) {
  const [filter, setFilter] = useState<DocumentFilter>("all")
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")
  const [selectedIndex, setSelectedIndex] = useState(0)

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

  // Use the utility hook to get all document data
  const {
    allDocuments,
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
        return allDocuments.data || []
    }
  }

  const filteredDocs = getFilteredDocuments()

  // Group documents into pages of 4 (2x2 grid)
  const documentsPerPage = 4
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

  const scrollTo = useCallback((index: number) => {
    if (emblaApi) emblaApi.scrollTo(index)
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
    <div className="space-y-4 ">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div className="flex gap-2 items-center">
          <FileText className="w-5 h-5 text-muted-foreground" />
          <h3 className="text-lg font-medium">Documents</h3>
          <Badge variant="secondary" className="ml-2">
            {allDocuments.data?.length || 0}
          </Badge>
        </div>
        <div className="flex gap-2 items-center">
          <Button variant="outline" size="sm" className="h-8" onClick={() => handleUpload(filter === "submission" ? "submission" : "support")}>
            <Upload className="w-4 h-4" />
          </Button>
          <DocumentStorageInfo />
        </div>
      </div>

      <div className="flex justify-between items-center">
        {/* Filter Tabs */}
        <div className="flex gap-2 items-center">
          <Button
            variant={filter === "all" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("all")}
            className="h-8"
          >
            <Folder className="mr-1 w-4 h-4" />
            All ({allDocuments.data?.length || 0})
          </Button>
          <Button
            variant={filter === "support" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("support")}
            className="h-8"
          >
            <FileCheck className="mr-1 w-4 h-4" />
            Support ({supportDocuments.data?.length || 0})
          </Button>
          <Button
            variant={filter === "submission" ? "default" : "outline"}
            size="sm"
            onClick={() => setFilter("submission")}
            className="h-8"
          >
            <Send className="mr-1 w-4 h-4" />
            Submissions ({submissionDocuments.data?.length || 0})
          </Button>
        </div>

        {/* Page indicators */}

        {documentPages.length > 1 && (
          <Badge variant="outline" className="flex items-center p-2 bg-gray-800/50 border border-gray-600 rounded-full">
            <p className="text-sm text-muted-foreground">{selectedIndex + 1} / <span className="text-white">{documentPages.length}</span></p>
          </Badge>
        )}
      </div>

      <Separator />

      {/* Documents List */}
      <div className="flex flex-col">
        {filteredDocs.length > 0 ? (
          <div className="relative">


            {/* Carousel container */}
            <div className="overflow-hidden" ref={emblaRef}>
              <div className="flex">
                {documentPages.map((page, pageIndex) => (
                  <div
                    key={pageIndex}
                    className="flex-none w-full min-w-0"
                  >
                    <div className="grid grid-cols-2 gap-4 p-4">
                      {page.map((document) => (
                        <div key={document.ID}>
                          <DocumentItem
                            document={document}
                          />
                        </div>
                      ))}


                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Navigation buttons */}
            {documentPages.length > 1 && (
              <>
                <Button
                  variant="outline"
                  size="icon"
                  className="left-0 absolute rounded-full top-1/2 -translate-y-1/2 z-10 h-8 w-8 bg-gray-800/50 border border-gray-600"
                  onClick={scrollPrev}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>

                <Button
                  variant="outline"
                  size="icon"
                  className="right-0 absolute rounded-full top-1/2 -translate-y-1/2 z-10 h-8 w-8 bg-gray-800/50 border border-gray-600"
                  onClick={scrollNext}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </>
            )}

          </div>
        ) : (
          <div className="flex flex-col justify-center items-center py-8 text-center">
            <FileText className="mb-2 w-12 h-12 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              {filter === "all"
                ? "No documents uploaded yet"
                : filter === "support"
                  ? "No support documents uploaded yet"
                  : "No submissions uploaded yet"
              }
            </p>
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleUpload(filter === "submission" ? "submission" : "support")}
              className="mt-2 h-8"
            >
              <Upload className="mr-1 w-4 h-4" />
              Upload Document
            </Button>
          </div>
        )}
      </div>

      {/* Upload Dialog */}
      <DocumentUploadDialog
        isOpen={uploadDialogOpen}
        onClose={() => setUploadDialogOpen(false)}
        onUploadComplete={handleUploadComplete}
        assignmentId={assignment.ID}
        documentType={uploadType}
      />
    </div>
  )
} 