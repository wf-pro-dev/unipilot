"use client"

import { memo, useEffect, useState } from "react"
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
} from "lucide-react"
import { models, progress } from "@/wailsjs/go/models"
import {
  useOpenDocument,
  useSaveDocumentAs,
  useDeleteDocument,
  useDownloadDocument,
  useUpload
} from "@/hooks/use-documents"
import { toast } from "sonner"
import { GlassCard } from "../ui/glass-card"
import { EventsOn } from "@/wailsjs/runtime/runtime"
import { Progress } from "../ui/progress"


interface DocumentItemProps {
  document: models.LocalDocument
  isUploading: boolean
  mode?: "default" | "readonly"

}

function BaseDocumentItem({ document: doc, isUploading, mode = "default" }: DocumentItemProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [progress, setProgress] = useState(0);
  const [status, setStatus] = useState('');
  const [error, setError] = useState('');

  const { addUpload, removeUpload } = useUpload()

  useEffect(() => {

    if (!isUploading) return
    var KeyProgress = "upload:progress:" + doc.UploadID
    var KeyStatus = "upload:status:" + doc.UploadID
    var KeyComplete = "upload:complete:" + doc.UploadID
    var KeyError = "upload:error:" + doc.UploadID


    EventsOn(KeyProgress, (progressData: progress.TrackerSnapshot) => {
      if (progressData.percentage > progress) {
        setProgress(progressData.percentage)
      }
    })

    EventsOn(KeyStatus, (status: string) => {
      setStatus(status)
    })

    EventsOn(KeyComplete, () => {
      setProgress(100)
    })
    EventsOn(KeyError, (error: string) => {
      setError(error)
    })
  }, [isUploading])

  // Document action hooks
  const openDocument = useOpenDocument()
  const saveDocumentAs = useSaveDocumentAs()
  const deleteDocument = useDeleteDocument()
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

  const handleDelete = async (document: models.LocalDocument) => {
    try {
      await deleteDocument.mutateAsync(document)
      setDeleteDialogOpen(false)
    } catch (error) {
      console.error("Failed to delete document:", error)
    }
  }

  const handleDownload = async () => {
    const uploadId = crypto.randomUUID()
    addUpload(uploadId)

    try {
      await downloadDocument.mutateAsync(new models.LocalDocument({
        ...doc,
        UploadID: uploadId,
        HasLocalFile: true
      }))
    } finally {
      removeUpload(uploadId)
    }

  }

  const isLoading = openDocument.isPending || saveDocumentAs.isPending ||
    deleteDocument.isPending

  return (
    <>
      <GlassCard
        id={doc.ID.toString()}
        variant="outline"
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
          <p className={`text-body line-clamp-1 tracking-tight`}>
            {doc.FileName}
          </p>

          {!isUploading && (
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

              <span className="text-caption font-medium uppercase tracking-wider">{formatFileSize(doc.FileSize)}</span>
              {/* 
              <p className={`text-caption flex items-center gap-1 line-clamp-1 leading-relaxed`}  >
                <Clock className="w-3.5 h-3.5" />
                {format(new Date(doc.UpdatedAt) || new Date(), "MMM d")}
              </p> */}
            </div>
          )}

          {isUploading && (
            <div className="flex flex-1 w-full">
              <Progress value={progress} className="h-1.5 bg-white/10 w-full" />
            </div>
          )}
        </div>

        {/* Actions */}
        {mode === "default" && (
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
        )}
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
              onClick={() => handleDelete(doc)}
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

export const DocumentItem = memo(BaseDocumentItem, (prevProps, nextProps) => {
  return prevProps.isUploading === nextProps.isUploading
})