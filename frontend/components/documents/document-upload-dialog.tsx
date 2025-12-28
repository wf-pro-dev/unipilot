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
import { toast } from 'sonner'
import { document } from "@/wailsjs/go/models"

interface DocumentUploadDialogProps {
  isOpen: boolean
  onClose: () => void
  onUploadComplete: () => void
  assignmentId: number
  remoteAssignmentId: number
  documentType: "support" | "submission" | "all"
  onSuccess?: (doc: document.LocalDocument) => void
}

export function DocumentUploadDialog({
  isOpen,
  onClose,
  onUploadComplete,
  assignmentId,
  remoteAssignmentId,
  documentType,
  onSuccess
}: DocumentUploadDialogProps) {
  const [selectedType, setSelectedType] = useState<"support" | "submission" | "all">(documentType)

  const uploadDocument = useUploadDocument()

  const handleUpload = async () => {


    const result = await uploadDocument.mutateAsync({
      assignmentId,
      remoteAssignmentId,
      documentType: selectedType
    }, {
      onSuccess: (doc) => {
        toast.success("Document uploaded successfully")
        if (doc) {
          console.log("doc uploaded successfully", doc)
          onSuccess?.(doc)
          setTimeout(() => {
            onUploadComplete()
            handleClose()
          }, 1500)
        }
      },
      onError: () => {
        toast.error("Failed to upload document")
      }
    })


  }

  const handleClose = () => {
    if (!uploadDocument.isPending) {
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
      <DialogContent className="glass border-white/10 text-white sm:max-w-md p-0 overflow-hidden gap-0">
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <DialogTitle className="flex items-center gap-2 text-xl font-semibold">
            <Upload className="h-5 w-5 text-blue-400" />
            Upload Document
          </DialogTitle>
        </DialogHeader>

        <div className="p-6 space-y-6">
          {/* Document Type Selection */}
          <div className="space-y-2">
            <Label htmlFor="document-type" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Document Type</Label>
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
          </div>

          {/* Type Description */}
          <div className="text-xs text-gray-400 bg-white/5 p-3 rounded-lg border border-white/5">
            {selectedType === "support" ? (
              <p>Support documents are reference materials, instructions, or resources related to this assignment.</p>
            ) : (
              <p>Submissions are your completed work or deliverables for this assignment.</p>
            )}
          </div>

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


          {/* Upload Progress */}
          {uploadDocument.isPending && (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-xs text-gray-400">
                <span>Uploading...</span>
                <span>Please wait</span>
              </div>
              <Progress value={undefined} className="h-1.5 bg-white/10" />
            </div>
          )}
        </div>

        <DialogFooter className="p-6 pt-0 grid grid-cols-2 gap-3">
          <Button
            variant="outline"
            size="sm"
            className="flex-1 border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white h-10"
            onClick={handleClose}
          >
            Cancel
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="flex-1 bg-blue-600 hover:bg-blue-500 text-white border-0 h-10 shadow-[0_0_15px_rgba(37,99,235,0.2)]"
            onClick={handleUpload}
            disabled={uploadDocument.isPending}
          >
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