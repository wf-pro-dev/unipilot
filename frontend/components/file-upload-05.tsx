import { useState } from "react";
import { Button } from "@/components/ui/button";
import { FileText, Folder, Info } from "lucide-react";
import { GlassCard } from "./ui/glass-card";
import { OnFileDrop } from "@/wailsjs/runtime/runtime";
import { models } from "@/wailsjs/go/models";
import { useDocuments, useUploadDocument } from "@/hooks/use-documents";
import { toast } from "sonner";
import { PickFile } from "@/wailsjs/go/main/App";
import { Label } from "./ui/label";
import { Badge } from "./ui/badge";
import { DocumentItem } from "./documents/document-item";
import { Select, SelectValue, SelectItem, SelectContent, SelectTrigger } from "./ui/select";
import { EmptyState } from "./ui/empty-state";
import { DocumentStorageInfo } from "./documents/document-storage-info";
import { Checkbox } from "./ui/checkbox";
import { useProgress } from "@/hooks/use-progress";
import { Scroll } from "./core/scroll";
import { Separator } from "./ui/separator";

interface FileUploadProps {
  assignment: models.LocalAssignment;
  mode?: "default" | "readonly"
  includeDocuments?: boolean
  setIncludeDocuments?: (includeDocuments: boolean) => void
}

type DocumentFilter = "All" | "support" | "submission"

export default function FileUpload05({ assignment, mode = "default", includeDocuments = true, setIncludeDocuments }: FileUploadProps) {

  const { data: documentData } = useDocuments(assignment)
  var documents = documentData || assignment.Documents || []


  const [uploadType, setUploadType] = useState<"support" | "submission">("support")
  const [selectedType, setSelectedType] = useState<"support" | "submission">("support")
  const { activeProgress, addProgress, removeProgress, isProgress } = useProgress()


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


  const uploadDocument = useUploadDocument()

  OnFileDrop((x, y, paths) => {
    if (paths.length > 1) {
      toast.error("Please drop only one file")
      return
    }

    handleUpload(paths[0])
  }, true);

  const handleUpload = async (filePath: string) => {


    const documentId = crypto.randomUUID()
    addProgress(documentId)

    try {

      const result = await uploadDocument.mutateAsync({
        documentId: documentId,
        assignmentId: assignment.ID,
        documentType: uploadType,
        filePath: filePath,
      })

    } finally {
      removeProgress(documentId)
    }
  }

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    await handleUpload("")
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
    <div className="flex flex-col  w-full">
      <form onSubmit={onSubmit}>

        <div className="flex flex-col w-full">


          <div className="space-y-4" >


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


            <Separator className="bg-white/10" />


            <div className="flex flex-col flex-1 gap-4">
              <div className="flex flex-col gap-2">
                <div className="flex items-center space-x-2 text-body font-medium text-gray-400 uppercase tracking-wider">
                  <Folder className="w-4 h-4" />
                  <span >Documents</span>
                </div>
                {documents.length > 0 ? (
                  <Scroll
                    data={{ Data: documents, HasMore: false }}
                    renderItem={(document: models.LocalDocument) => (
                      <DocumentItem key={document.ID} document={document} isUploading={isProgress(document.ID)} mode={mode} />

                    )}
                    keyExtractor={(item: models.LocalDocument) => item.ID}
                    numColumns={2}
                    containerClassName="gap-4"
                  />
                ) : (
                  <div className="flex flex-1 border border-dashed border-white/10 rounded-xl bg-white/5">
                    <EmptyState
                      icon={FileText}
                      title="No documents found"
                      description="Upload your first document to get started"
                      className="flex-1 items-center"
                    />

                  </div>
                )}
              </div>

              {mode === "default" && (
                <DocumentStorageInfo assignmentID={assignment.ID} />
              )}

              {mode !== "default" && (
                <div className="flex items-center gap-2">
                  <Checkbox checked={includeDocuments} onCheckedChange={(checked) => setIncludeDocuments?.(checked === "indeterminate" ? false : checked)} />
                  <Label className="text-caption">Do you want to include documents ?</Label>
                </div>
              )}
            </div>




          </div>
        </div>
      </form>

    </div>
  );
}
