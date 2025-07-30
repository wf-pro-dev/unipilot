"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { 
  FileText, 
  MoreVertical, 
  Download, 
  Eye, 
  Trash2, 
  Upload,
  Clock,
  CheckCircle2
} from "lucide-react"
import { document } from "@/wailsjs/go/models"
import { 
  useOpenDocument,
  useSaveDocumentAs,
  useDeleteDocument,
  useUploadDocumentVersion
} from "@/hooks/use-documents"
import { format } from "date-fns"

interface DocumentItemProps {
  document: document.LocalDocument
}

export function DocumentItem({ document: doc }: DocumentItemProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  // Document action hooks
  const openDocument = useOpenDocument()
  const saveDocumentAs = useSaveDocumentAs()
  const deleteDocument = useDeleteDocument()
  const uploadVersion = useUploadDocumentVersion()

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 B"
    const k = 1024
    const sizes = ["B", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
  }

  const getFileIcon = (fileName: string) => {
    const ext = fileName.split('.').pop()?.toLowerCase()
    return <FileText className="h-4 w-4" />
  }

  const getDocumentTypeColor = (type: string) => {
    return type === "support" 
      ? "bg-blue-500/20 text-blue-400" 
      : "bg-green-500/20 text-green-400"
  }

  const handleOpen = async () => {
    if (!doc.HasLocalFile) return
    try {
      await openDocument.mutateAsync(doc.ID)
    } catch (error) {
      console.error("Failed to open document:", error)
    }
  }

  const handleSaveAs = async () => {
    if (!doc.HasLocalFile) return
    try {
      await saveDocumentAs.mutateAsync(doc.ID)
    } catch (error) {
      console.error("Failed to save document:", error)
    }
  }

  const handleUploadNewVersion = async () => {
    try {
      await uploadVersion.mutateAsync(doc.ID)
    } catch (error) {
      console.error("Failed to upload new version:", error)
    }
  }

  const handleDelete = async () => {
    try {
      await deleteDocument.mutateAsync(doc.ID)
      setDeleteDialogOpen(false)
    } catch (error) {
      console.error("Failed to delete document:", error)
    }
  }

  const isLoading = openDocument.isPending || saveDocumentAs.isPending || 
                   deleteDocument.isPending || uploadVersion.isPending

  return (
    <>
      <div className="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          {/* File Icon */}
          <div className="flex-shrink-0">
            {getFileIcon(doc.FileName)}
          </div>

          {/* File Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <p className="text-sm font-medium truncate">
                {doc.FileName}
              </p>
              <Badge 
                variant="secondary" 
                className={`text-xs ${getDocumentTypeColor(doc.Type)}`}
              >
                {doc.Type === "support" ? "Support" : "Submission"}
              </Badge>
              {doc.Version > 1 && (
                <Badge variant="outline" className="text-xs">
                  v{doc.Version}
                </Badge>
              )}
            </div>
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              <span>{formatFileSize(doc.FileSize)}</span>
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {format(new Date(doc.UpdatedAt), "MMM d, yyyy")}
              </span>
             
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1">
          <Button
            size="sm"
            variant="ghost"
            onClick={handleOpen}
            disabled={isLoading || !doc.HasLocalFile}
            className="h-8 w-8 p-0"
          >
            <Eye className="h-4 w-4" />
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="sm" variant="ghost" className="h-8 w-8 p-0" disabled={isLoading}>
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleOpen} disabled={!doc.HasLocalFile}>
                <Eye className="h-4 w-4 mr-2" />
                Open
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleSaveAs} disabled={!doc.HasLocalFile}>
                <Download className="h-4 w-4 mr-2" />
                Save As...
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleUploadNewVersion}>
                <Upload className="h-4 w-4 mr-2" />
                Upload New Version
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setDeleteDialogOpen(true)}
                className="text-red-600 hover:text-red-600"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Document</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{doc.FileName}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-red-600 hover:bg-red-700"
              disabled={isLoading}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
} 