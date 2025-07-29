"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Progress } from "@/components/ui/progress"
import { Badge } from "@/components/ui/badge"
import { 
  Upload, 
  FileText, 
  CheckCircle2, 
  AlertCircle,
  Info
} from "lucide-react"
import { useUploadDocument } from "@/hooks/use-documents"

interface DocumentUploadDialogProps {
  isOpen: boolean
  onClose: () => void
  onUploadComplete: () => void
  assignmentId: number
  documentType: "support" | "submission"
}

export function DocumentUploadDialog({
  isOpen,
  onClose,
  onUploadComplete,
  assignmentId,
  documentType
}: DocumentUploadDialogProps) {
  const [selectedType, setSelectedType] = useState<"support" | "submission">(documentType)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const uploadDocument = useUploadDocument()

  const handleUpload = async () => {
    setError(null)
    setSuccess(null)

    try {
      const result = await uploadDocument.mutateAsync({ 
        assignmentId, 
        documentType: selectedType 
      })
      
      if (result) {
        setSuccess(`Successfully uploaded "${result.FileName}"`)
        setTimeout(() => {
          onUploadComplete()
          handleClose()
        }, 1500)
      }
    } catch (error) {
      console.error("Upload failed:", error)
      setError(error instanceof Error ? error.message : "Upload failed")
    }
  }

  const handleClose = () => {
    if (!uploadDocument.isPending) {
      setError(null)
      setSuccess(null)
      onClose()
    }
  }

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

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Upload className="h-5 w-5" />
            Upload Document
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Document Type Selection */}
          <div className="space-y-2">
            <Label htmlFor="document-type">Document Type</Label>
            <Select
              value={selectedType}
              onValueChange={(value: "support" | "submission") => setSelectedType(value)}
              disabled={uploadDocument.isPending}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="support">
                  <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4" />
                    Support Document
                  </div>
                </SelectItem>
                <SelectItem value="submission">
                  <div className="flex items-center gap-2">
                    <FileText className="h-4 w-4" />
                    Submission
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Type Description */}
          <div className="text-sm text-muted-foreground">
            {selectedType === "support" ? (
              <p>Support documents are reference materials, instructions, or resources related to this assignment.</p>
            ) : (
              <p>Submissions are your completed work or deliverables for this assignment.</p>
            )}
          </div>

          {/* File Size Limits */}
          <div className="rounded-lg border p-3 bg-muted/50">
            <div className="flex items-start gap-2">
              <Info className="h-4 w-4 text-blue-500 mt-0.5 flex-shrink-0" />
              <div className="space-y-1 text-xs">
                <p className="font-medium">File Limits:</p>
                <ul className="space-y-0.5 text-muted-foreground">
                  <li>• Maximum file size: 50 MB</li>
                  <li>• Maximum per assignment: 200 MB</li>
                  <li>• Total storage limit: 2 GB</li>
                </ul>
              </div>
            </div>
          </div>

          {/* Supported Formats */}
          <div className="space-y-2">
            <Label className="text-sm font-medium">Supported Formats</Label>
            <div className="flex flex-wrap gap-1">
              {getSupportedFormats().map((format) => (
                <Badge key={format} variant="secondary" className="text-xs">
                  {format}
                </Badge>
              ))}
            </div>
          </div>

          {/* Status Messages */}
          {error && (
            <div className="flex items-center gap-2 p-3 rounded-lg border border-red-200 bg-red-50 text-red-700">
              <AlertCircle className="h-4 w-4" />
              <p className="text-sm">{error}</p>
            </div>
          )}

          {success && (
            <div className="flex items-center gap-2 p-3 rounded-lg border border-green-200 bg-green-50 text-green-700">
              <CheckCircle2 className="h-4 w-4" />
              <p className="text-sm">{success}</p>
            </div>
          )}

          {/* Upload Progress */}
          {uploadDocument.isPending && (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>Uploading...</span>
                <span>Please wait</span>
              </div>
              <Progress value={undefined} className="h-2" />
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={uploadDocument.isPending}>
            Cancel
          </Button>
          <Button onClick={handleUpload} disabled={uploadDocument.isPending}>
            {uploadDocument.isPending ? (
              <>
                <Upload className="h-4 w-4 mr-2 animate-spin" />
                Uploading...
              </>
            ) : (
              <>
                <Upload className="h-4 w-4 mr-2" />
                Select & Upload
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
} 