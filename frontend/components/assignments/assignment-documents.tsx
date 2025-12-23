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
  ChevronLeft,
  ChevronRight
} from "lucide-react"
import { assignment, document } from "@/wailsjs/go/models"
import { useAssignmentDocumentData } from "@/hooks/use-documents"
import useEmblaCarousel from 'embla-carousel-react'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../ui/select"


interface AssignmentDocumentsProps {
  assignment: assignment.LocalAssignment
  documents?: document.LocalDocument[]
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

  const filteredDocs = documents || getFilteredDocuments()

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
    <div className="space-y-4">
      {/* Header */}

      <div
        className="
          bg-white/5 
          border border-white/5 
          p-6
          rounded-xl
          space-y-6"
      >
        {!viewMode && (
          <div className="flex items-center space-x-6">
            {/* Filter Tabs */}

            <div className="w-full grid grid-cols-1 md:grid-cols-3 gap-4 items-center">

              <Select open={isSelectOpen} onOpenChange={setIsSelectOpen} value={filter} onValueChange={(value) => setFilter(value as DocumentFilter)}>

                <SelectTrigger onClick={() => setIsSelectOpen(!isSelectOpen)} className="flex-1 bg-white/5 border-white/10 hover:bg-white/10 text-white flex space-x-2 h-10 transition-All focus:outline-none" >
                  <div className="h-10 w-full flex items-center  transition-All">

                    <span className="text-sm">{filter}</span>
                  </div>
                </SelectTrigger>

                <SelectContent className="glass border-white/10 bg-black/90 backdrop-blur-xl">
                  {documentFilters.map((filter: DocumentFilter) => (
                    <SelectItem key={filter} value={filter} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                      {filter}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Button variant="outline"
                size="sm"
                className="flex-1 bg-white/5 border-white/10 hover:bg-white/10 text-white flex items-center justify-between h-10 transition-All"
                onClick={() => handleUpload(filter === "submission" ? "submission" : "support")}>


                <span className="text-sm">Upload</span>
                <Upload className="w-1 h-1 text-slate-500" />
              </Button>

              <DocumentStorageInfo />

            </div>


            {/* Page indicators */}

            {documentPages.length > 1 && (
              <Badge variant="outline" className="flex items-center p-0 h-8 px-3 bg-white/5 border-white/10 rounded-lg">
                <p className="text-xs text-gray-400 font-medium">{selectedIndex + 1} / <span className="text-white">{documentPages.length}</span></p>
              </Badge>
            )}
          </div>
        )}


        {/* Documents List */}
        <div className="flex flex-col">
          {filteredDocs.length > 0 ? (
            <div className="relative group/carousel">

              {/* Carousel container */}
              <div className="overflow-hidden" ref={emblaRef}>
                <div className="flex">
                  {documentPages.map((page, pageIndex) => (
                    <div
                      key={pageIndex}
                      className="flex-none w-full min-w-0 px-1"
                    >
                      <div className="grid grid-cols-1 gap-3 min-h-[224px]">
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
                    className="left-0 absolute rounded-full top-1/2 -translate-y-1/2 -translate-x-3 z-10 h-8 w-8 bg-black/40 border-white/10 backdrop-blur-sm text-white opacity-0 group-hover/carousel:opacity-100 transition-opacity"
                    onClick={scrollPrev}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>

                  <Button
                    variant="outline"
                    size="icon"
                    className="right-0 absolute rounded-full top-1/2 -translate-y-1/2 translate-x-3 z-10 h-8 w-8 bg-black/40 border-white/10 backdrop-blur-sm text-white opacity-0 group-hover/carousel:opacity-100 transition-opacity"
                    onClick={scrollNext}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </>
              )}

            </div>
          ) : (
            <div className="flex flex-col justify-center items-center py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
              <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mb-4">
                <FileText className="w-8 h-8 text-gray-500" strokeWidth={1.5} />
              </div>
              <p className="text-sm text-gray-400 mb-4">
                {filter === "All"
                  ? "No documents uploaded yet"
                  : filter === "support"
                    ? "No support documents"
                    : "No submissions yet"
                }
              </p>
              {!viewMode && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleUpload(filter === "submission" ? "submission" : "support")}
                  className="py-2 px-4 gap-2 text-blue-400 bg-blue-500/10 border-blue-500/20 hover:bg-blue-500/20 hover:text-blue-300 h-9"
                >
                  <Upload className="w-3.5 h-3.5" />
                  Upload Document
                </Button>
              )}
            </div>
          )}
        </div>
      </div>


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