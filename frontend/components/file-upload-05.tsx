import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { ChevronLeft, ChevronRight, FileText, Upload } from "lucide-react";
import { GlassCard } from "./ui/glass-card";
import { models } from "@/wailsjs/go/models";
import { useAssignmentDocumentData, useUploadDocument } from "@/hooks/use-documents";
import { toast } from "sonner";
import { DocumentItem } from "./documents/document-item";
import { HorizontalEmptyState } from "./ui/empty-state";
import { useCallback, useEffect, useState } from "react";
import { DocumentUploadDialog } from "./documents/document-upload-dialog";
import useEmblaCarousel from 'embla-carousel-react'

interface FileUploadProps {
  assignment: models.LocalAssignment;
}

export default function FileUpload05({ assignment }: FileUploadProps) {

  const [filter, setFilter] = useState<DocumentFilter>("All")
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [isSelectOpen, setIsSelectOpen] = useState(false)
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

  const handleUpload = async () => {

    const result = await uploadDocument.mutateAsync({
      assignmentId: assignment.ID,
      remoteAssignmentId: assignment.RemoteID,
      documentType: uploadType
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
    await handleUpload()
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
    <div className="flex flex-col flex-1 items-center justify-center  w-full ">
      <form onSubmit={onSubmit}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">File Upload</h3>


          {documentPages?.length > 1 && (
            <div className="flex items-center gap-2">
              {/* Pagination buttons */}
              <Button
                type="button"
                variant="outline"
                size="icon"
                className="rounded-full"
                onClick={scrollPrev}
              >
                <ChevronLeft className="w-4 h-4" />
              </Button>
              <Button
                type="button"
                variant="outline"
                size="icon"
                className="rounded-full"
                onClick={scrollNext}
              >
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          )}

        </div>

        <GlassCard variant="board" className="mt-4 items-center justify-center border border-dashed border-input py-10">
          <div className="sm:flex sm:items-center sm:gap-x-3">

            <div className="p-2 rounded-xl bg-gradient-to-br from-white/15 to-transparent border border-white/15 shadow-inner">
              <Upload className="w-3 h-3 text-white" />
            </div>

            <div className="mt-4 flex items-center text-sm leading-6 text-foreground sm:mt-0">
              <p>Drag and drop or</p>
              <Button
                type="button"
                variant="link"
                size="sm"
                className="text-primary-blue-500 hover:text-primary-blue-600 px-2"
                onClick={() => setUploadDialogOpen(true)}
              >
                Choose a file
              </Button>
              <p className="pl-1">to upload</p>
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
