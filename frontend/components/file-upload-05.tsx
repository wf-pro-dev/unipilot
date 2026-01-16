import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight, FileText, Info } from "lucide-react";
import { GlassCard } from "./ui/glass-card";
import { OnFileDrop } from "@/wailsjs/runtime/runtime";
import { models } from "@/wailsjs/go/models";
import { useAssignmentDocumentData, useUploadDocument } from "@/hooks/use-documents";
import { toast } from "sonner";
import { useCallback, useEffect, useState } from "react";
import { DocumentUploadDialog } from "./documents/document-upload-dialog";
import useEmblaCarousel from 'embla-carousel-react'
import { PickFile } from "@/wailsjs/go/main/App";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs";
import { Label } from "./ui/label";
import { Badge } from "./ui/badge";
import { DocumentItem } from "./documents/document-item";
import { Select, SelectValue, SelectItem, SelectContent, SelectTrigger } from "./ui/select";
import { EmptyState } from "./ui/empty-state";
import { DocumentStorageInfo } from "./documents/document-storage-info";

interface FileUploadProps {
  assignment: models.LocalAssignment;
}

export default function FileUpload05({ assignment }: FileUploadProps) {

  const [filter, setFilter] = useState<DocumentFilter>("All")
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [activeUploads, setActiveUploads] = useState<Set<string>>(new Set());
  const [activeTab, setActiveTab] = useState("upload")
  const [selectedType, setSelectedType] = useState<"support" | "submission">("support")


  const handlePickFile = () => {

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

    const uploadId = crypto.randomUUID()
    setActiveUploads(prev => new Set(prev).add(uploadId))

    try {
      const result = await uploadDocument.mutateAsync({
        assignmentId: assignment.ID,
        remoteAssignmentId: assignment.RemoteID,
        documentType: uploadType,
        filePath: filePath,
        uploadId: uploadId
      })
    } finally {
      setActiveUploads(prev => {
        const next = new Set(prev);
        next.delete(uploadId);
        return next;
      })
    }

  }

  useEffect(() => {
    console.log("Active uploads:", activeUploads)
  }, [activeUploads])

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

  const getSupportedFormats = () => {
    return [
      "PDF documents (.pdf)",
      "Microsoft Word (.doc, .docx)",
      "PowerPoint (.ppt, .pptx)",
      "Excel (.xls, .xlsx)",
      "Text files (.txt, .md)",
      "Images (.png, .jpg, .jpeg, .gif, .bmp, .svg)"
    ]
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
    <div className="flex flex-col flex-1 w-full">
      <form onSubmit={onSubmit}>


        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex flex-col flex-1 w-full">
          <TabsList className="flex flex-row items-center gap-2 mb-4">
            <TabsTrigger
              value="upload"
              className="flex items-baseline text-body text-gray-400 data-[state=active]:text-h6 data-[state=active]:text-white transition-all duration-200"
            >
              <span className="font-normal leading-none uppercase tracking-wider">Upload</span>
            </TabsTrigger>

            <TabsTrigger
              value="documents"
              className="flex items-baseline text-body text-gray-400 data-[state=active]:text-h6 data-[state=active]:text-white transition-all duration-200"
            >
              <span className="font-normal leading-none uppercase tracking-wider">Documents</span>
            </TabsTrigger>


            { activeTab === "documents" && documentPages?.length > 1 && (
              <div className="flex flex-1 items-center gap-2 justify-end">
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

          </TabsList>

          <TabsContent value="upload" key="upload" className="space-y-4" >


            <Select
              value={selectedType}
              onValueChange={(value: "support" | "submission") => setSelectedType(value)}
              disabled={uploadDocument.isPending}

            >
              <SelectTrigger className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="glass border-white/10">
                <SelectItem value="support">
                  <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4 text-blue-400" />
                    <span className="text-sm">Support Document</span>
                  </div>
                </SelectItem>
                <SelectItem value="submission">
                  <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4 text-green-400" />
                    <span className="text-sm">Submission</span>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>


            <GlassCard variant="board"
              style={{ "--wails-drop-target": "drop" } as React.CSSProperties}
              className="items-center justify-center border border-dashed border-input py-10
          [--wails-drop-target:drop] 
         [.wails-drop-target-active_&]:border-blue-500 [.wails-drop-target-active_&]:bg-blue-50 
          "
            >
              <div className="sm:flex sm:items-center sm:gap-x-3">

                <div className="[--wails-drop-target:none] text-body mt-4 flex items-center  leading-6 sm:mt-0">
                  <p className="pointer-events-none text-gray-400 ">Drag and drop or</p>

                  <Button
                    type="button"
                    variant="link"
                    size="sm"
                    className="[--wails-drop-target:none] text-body px-2"
                    onClick={handlePickFile}
                  >
                    Choose a file
                  </Button>

                  <p className="pointer-events-none text-gray-400 ">to upload</p>
                </div>

              </div>

            </GlassCard>

            {/* File Size Limits */}
            <div className="bg-blue-500/10 border border-blue-500/20 p-3 rounded-lg">
              <div className="flex items-start gap-3">
                <Info className="h-4 w-4 text-blue-400 mt-0.5 flex-shrink-0" />
                <div className="space-y-1 text-xs">
                  <p className="font-medium text-blue-200">File Limits:</p>
                  <ul className="space-y-1 text-blue-200/70 list-disc pl-3">
                    <li>Maximum file size: 50 MB</li>
                    <li>Maximum per assignment: 200 MB</li>
                    <li>Total storage limit: 2 GB</li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Supported Formats */}

            <div className="space-y-2">
              <Label className="text-gray-400 text-xs font-medium uppercase tracking-wider">Supported Formats</Label>
              <div className="flex flex-wrap gap-1.5">
                {getSupportedFormats().map((format) => (
                  <Badge key={format} variant="secondary" className="text-[10px] bg-white/5 hover:bg-white/10 text-gray-300 border-white/10 font-normal">
                    {format}
                  </Badge>
                ))}
              </div>
            </div>
          </TabsContent>

          <TabsContent value="documents" key="documents" className="space-y-4" >


            {documentPages?.length > 0 && (
              <div className="flex flex-col flex-1 gap-4">

                <div className="overflow-hidden" ref={emblaRef}>
                  <div className="flex w-full gap-4">
                    {documentPages?.map((page, pageIndex) => (

                      <div className="flex-none w-full min-w-0 px-1" key={pageIndex}>

                        {page.map((document) => (
                          <DocumentItem key={document.ID || document.UploadID} document={document} isUploading={activeUploads.has(document.UploadID)} />
                        ))}

                      </div>

                    ))}
                  </div>
                </div>
                <DocumentStorageInfo assignmentID={assignment.ID} />
              </div>

            )}

            {documentPages?.length === 0 && (
              <GlassCard variant="board" className="flex flex-col py-10 w-full items-center justify-center">
                <EmptyState icon={FileText} title="No documents found" description="Upload your first document to get started" />
              </GlassCard>
            )}
          </TabsContent>
        </Tabs>



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
