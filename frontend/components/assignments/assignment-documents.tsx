"use client"

import { useState } from "react"
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
  Send
} from "lucide-react"
import { Assignment } from "@/types/models"
import { useAssignmentDocumentData } from "@/hooks/use-documents"

interface AssignmentDocumentsProps {
  assignment: Assignment
}

type DocumentFilter = "all" | "support" | "submission"

export function AssignmentDocuments({ assignment }: AssignmentDocumentsProps) {
  const [filter, setFilter] = useState<DocumentFilter>("all")
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")

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
      <div className="flex justify-between items-center">
        <div className="flex gap-2 items-center">
          <FileText className="w-5 h-5 text-muted-foreground" />
          <h3 className="text-lg font-semibold">Documents</h3>
          <Badge variant="secondary" className="ml-2">
            {allDocuments.data?.length || 0}
          </Badge>
        </div>
        <DocumentStorageInfo />
      </div>

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

      <Separator />

      {/* Documents List */}
      <div className="space-y-2">
        {filteredDocs.length > 0 ? (
          filteredDocs.map((doc) => (
            <DocumentItem
              key={doc.ID}
              document={doc}
            />
          ))
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
              Upload {filter === "submission" ? "Submission" : "Support Doc"}
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