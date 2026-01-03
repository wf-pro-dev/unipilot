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
import { models } from "@/wailsjs/go/models"
import {
  useOpenDocument,
  useSaveDocumentAs,
  useDeleteDocument,
  useUploadDocumentVersion,
  useDownloadDocument
} from "@/hooks/use-documents"
import { format } from "date-fns"
import { toast } from "sonner"
import { GlassCard } from "../ui/glass-card"

interface DocumentItemProps {
  document: models.LocalDocument
}

export function DocumentItem({ document: doc }: DocumentItemProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  // Document action hooks
  const openDocument = useOpenDocument()
  const saveDocumentAs = useSaveDocumentAs()
  const deleteDocument = useDeleteDocument()
  const uploadVersion = useUploadDocumentVersion()
  const downloadDocument = useDownloadDocument()

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

  const handleDelete = async (documentId: number) => {
    try {
      await deleteDocument.mutateAsync(documentId)
      setDeleteDialogOpen(false)
    } catch (error) {
      console.error("Failed to delete document:", error)
    }
  }

  const handleDownload = async () => {
    try {
      await downloadDocument.mutateAsync(doc,
        {
          onSuccess: () => {
            toast.success("Document downloaded successfully")
          },
          onError: (error) => {
            toast.error("Failed to download document: " + error)
          }
        })
    } catch (error) {
      console.error("Failed to download document:", error)
    }
  }

  const isLoading = openDocument.isPending || saveDocumentAs.isPending ||
    deleteDocument.isPending || uploadVersion.isPending

  return (
    <>
      <GlassCard
        className="
          grid grid-cols-[auto,1fr,auto] items-center gap-3
          p-3
         group relative"
      >
        {/* File Icon */}
        <div className="flex-shrink-0 p-2.5 rounded-lg bg-white/5 border border-white/5">
          {getFileIcon(doc.FileName)}
        </div>

        {/* File Info */}
        <div className="min-w-0 flex flex-col gap-1.5">
          <p className="text-sm font-medium truncate text-gray-200 group-hover:text-white transition-colors">
            {doc.FileName}
          </p>

          <div className="flex flex-wrap gap-2 items-center">
            <Badge
              variant="secondary"
              className={`text-[10px] border-0 px-1.5 py-0 font-medium h-5 ${getDocumentTypeColor(doc.Type)}`}
            >
              {doc.Type === "support" ? "Support" : "Submission"}
            </Badge>

            <Badge variant="outline" className="text-[10px] border-white/10 bg-white/5 text-gray-400 px-1.5 py-0 h-5">
              v{doc.Version}
            </Badge>

            <span className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">{formatFileSize(doc.FileSize)}</span>
            
            <span className="text-[10px] text-gray-500 flex items-center gap-1 ml-auto sm:ml-0">
              <Clock className="h-3 w-3" />
              {format(new Date(doc.UpdatedAt), "MMM d")}
            </span>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="sm" variant="ghost" className="p-0 h-8 w-8 hover:bg-white/10 text-gray-400 hover:text-white rounded-lg transition-colors" disabled={isLoading}>
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass border-white/10 bg-black/90 backdrop-blur-xl">
              {(doc.HasLocalFile) && (
                <>
                  <DropdownMenuItem onClick={handleOpen} disabled={!doc.HasLocalFile} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                    <Eye className="h-4 w-4 mr-2" />
                    Open
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={handleSaveAs} disabled={!doc.HasLocalFile} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                    <Download className="h-4 w-4 mr-2" />
                    Save As...
                  </DropdownMenuItem>
                </>
              )}

              {!doc.HasLocalFile && (
                <DropdownMenuItem onClick={handleDownload} disabled={doc.HasLocalFile} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                  <Download className="h-4 w-4 mr-2" />
                  Download
                </DropdownMenuItem>
              )}

              <DropdownMenuItem onClick={handleUploadNewVersion} className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">
                <Upload className="h-4 w-4 mr-2" />
                Upload New Version
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => setDeleteDialogOpen(true)}
                className="text-red-400 focus:text-red-300 focus:bg-red-500/10 cursor-pointer"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </GlassCard >

      {/* Delete Confirmation Dialog */}
      < AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen} >
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
              onClick={()=>handleDelete(doc.ID)}
              className="bg-red-600 hover:bg-red-700"
              disabled={isLoading}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog >
    </>
  )
} 