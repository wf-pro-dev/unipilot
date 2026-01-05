import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight, FileText, Upload } from "lucide-react";
import { GlassCard } from "./ui/glass-card";
import { OnFileDrop } from "@/wailsjs/runtime/runtime";


import { models } from "@/wailsjs/go/models";
import { useAssignmentDocumentData, useUploadDocument } from "@/hooks/use-documents";
import { toast } from "sonner";
import { DocumentItem } from "./documents/document-item";
import { HorizontalEmptyState } from "./ui/empty-state";
import { useCallback, useEffect, useRef, useState } from "react";
import { DocumentUploadDialog } from "./documents/document-upload-dialog";
import useEmblaCarousel from 'embla-carousel-react'
import { PickFile } from "@/wailsjs/go/main/App";

interface FileUploadProps {
  assignment: models.LocalAssignment;
}

export default function FileUpload05({ assignment }: FileUploadProps) {

  const [filter, setFilter] = useState<DocumentFilter>("All")
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")
  const [selectedIndex, setSelectedIndex] = useState(0)


  const handlePickFile =  () => {

    PickFile()
    .then((filePath) => {
      if (filePath) {
        handleUpload(filePath)
      }
    })
    .catch((error) => {
      toast.error("Failed to pick file: " + error)
    })
  }

  // Use the utility hook to get All document data
  const {
    supportDocuments,
    submissionDocuments,
    isLoading,
    error
  } = useAssignmentDocumentData(assignment.ID)

  type DocumentFilter = "All" | "support" | "submission"
  const documentFilters: DocumentFilter[] = [
    "All",
    "support",
    "submission"
  ]

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

  const filteredDocs = getFilteredDocuments() || []

  // Group documents into pages of 4 (2x2 grid)
  const documentsPerPage = 1
  const documentPages: models.LocalDocument[][] = []
  for (let i = 0; i < filteredDocs.length; i += documentsPerPage) {
    documentPages.push(filteredDocs.slice(i, i + documentsPerPage))
  }



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




  // Carousel navigation functions
  const scrollPrev = useCallback(() => {
    if (emblaApi) emblaApi.scrollPrev()
  }, [emblaApi])

  const scrollNext = useCallback(() => {
    if (emblaApi) emblaApi.scrollNext()
  }, [emblaApi])



  const handleUploadComplete = () => {
    setUploadDialogOpen(false)
    // The hooks will automatically refetch and update the UI
  }


  const uploadDocument = useUploadDocument()

  const handleUpload = async (filePath: string) => {

    const result = await uploadDocument.mutateAsync({
      assignmentId: assignment.ID,
      remoteAssignmentId: assignment.RemoteID,
      documentType: uploadType,
      filePath: filePath
    }, {
      onSuccess: (doc) => {
        toast.success("Document uploaded successfully")
      },
      onError: () => {
        toast.error("Failed to upload document")
      }
    })
  }

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    await handleUpload("")
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

  OnFileDrop((x, y, paths) => {
    if (paths.length > 1) {
      toast.error("Please drop only one file")
      return
    }

    handleUpload(paths[0])
  }, true);




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
    <div className="flex flex-col flex-1 w-full ">
      <form onSubmit={onSubmit}>
        <div className="flex items-center justify-between">
          <h5 className="text-h5">File Upload</h5>


          {documentPages?.length > 1 && (
            <div className="flex items-center gap-2">
              {/* Pagination buttons */}
              <Button
                type="button"
                variant="outline"
                size="icon"
                className="rounded-full w-6 h-6  "
                onClick={scrollPrev}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <Button
                type="button"
                variant="outline"
                size="icon"
                className="rounded-full w-6 h-6"
                onClick={scrollNext}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          )}

        </div>

        <GlassCard variant="board"
          style={{ "--wails-drop-target": "drop" } as React.CSSProperties}
          className="mt-4 items-center justify-center border border-dashed border-input py-10
          [--wails-drop-target:drop] 
         [.wails-drop-target-active_&]:border-blue-500 [.wails-drop-target-active_&]:bg-blue-50 
          "
        >
          <div className="sm:flex sm:items-center sm:gap-x-3">

            <div className="[--wails-drop-target:none] text-body text-text-caption mt-4 flex items-center  leading-6 sm:mt-0">
              <p className="pointer-events-none ">Drag and drop or</p>

              <Button
                type="button"
                variant="link"
                size="sm"
                className="[--wails-drop-target:none] text-body px-2"
                onClick={handlePickFile}
              >
                Choose a file
              </Button>

              <p className="pointer-events-none">to upload</p>
            </div>

          </div>

        </GlassCard>

        <p className="mt-2 flex items-center justify-between text-xs leading-5 text-muted-foreground">
          Recommended max. size: 10 MB, Accepted file types: XLSX, XLS, CSV.
        </p>

        <div className="relative mt-6 flex flex-col flex-1 gap-4">

          {documentPages?.length === 0 ? (
            <GlassCard variant="board" className="p-4">
              <HorizontalEmptyState
                icon={Upload}
                title="No documents uploaded yet"
                description="Upload a document to get started"
              />
            </GlassCard>
          ) : (
            <div className="overflow-hidden" ref={emblaRef}>
              <div className="flex w-full gap-4">
                {documentPages?.map((page, pageIndex) => (

                  <div className="flex-none w-full min-w-0 px-1" key={pageIndex}>

                    {page.map((document) => (
                      <DocumentItem key={document.ID} document={document} />
                    ))}

                  </div>

                ))}
              </div>
            </div>
          )}
        </div>

      </form>

      <DocumentUploadDialog
        isOpen={uploadDialogOpen}
        onClose={() => setUploadDialogOpen(false)}
        onUploadComplete={handleUploadComplete}
        assignmentId={assignment.ID}
        remoteAssignmentId={assignment.RemoteID}
        documentType={uploadType}
      />
    </div>
  );
}
