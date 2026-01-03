import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { FileSpreadsheet, Upload, X } from "lucide-react";
import { GlassCard } from "./ui/glass-card";
import { models } from "@/wailsjs/go/models";
import { useSupportDocuments, useUploadDocument } from "@/hooks/use-documents";
import { toast } from "sonner";
import { DocumentItem } from "./documents/document-item";
import { EmptyState, HorizontalEmptyState } from "./ui/empty-state";

interface FileUploadProps {
  assignment: models.LocalAssignment;
  documents: models.LocalDocument[];
  documentType?: "support" | "submission";
  onButtonClick: () => void;
}

export default function FileUpload05({ assignment, documents, documentType = "support", onButtonClick }: FileUploadProps) {

  const uploadDocument = useUploadDocument()

  const handleUpload = async () => {

    const result = await uploadDocument.mutateAsync({
      assignmentId: assignment.ID,
      remoteAssignmentId: assignment.RemoteID,
      documentType
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


  return (
    <div className="flex flex-col flex-1 items-center justify-center  w-full ">
      <form onSubmit={onSubmit}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">File Upload</h3>
          <Button
            type="button"
            variant="default"
            size="sm"
            className="rounded-full"
            onClick={onButtonClick}
          >
            Upload
          </Button>
        </div>

        <GlassCard variant="board" className="mt-4 items-center justify-center rounded-md border border-dashed border-input py-10">
          <div className="sm:flex sm:items-center sm:gap-x-3">

            <div className="p-2 rounded-xl bg-gradient-to-br from-white/15 to-transparent border border-white/15 shadow-inner">
              <Upload className="w-3 h-3 text-white" />
            </div>

            <div className="mt-4 flex text-sm leading-6 text-foreground sm:mt-0">
              <p>Drag and drop or</p>
              <Label
                htmlFor="file-upload-4"
                className="relative cursor-pointer rounded-sm pl-1 font-medium text-primary hover:underline hover:underline-offset-4 flex items-center"
              >
                <span>choose file</span>
                <input
                  id="file-upload-4"
                  name="file-upload-4"
                  type="file"
                  className="sr-only"
                />
              </Label>
              <p className="pl-1">to upload</p>
            </div>
          </div>

        </GlassCard>

        <p className="mt-2 flex items-center justify-between text-xs leading-5 text-muted-foreground">
          Recommended max. size: 10 MB, Accepted file types: XLSX, XLS, CSV.
        </p>

        <div className="relative mt-8 flex flex-col flex-1 gap-4">
          {documents?.length === 0 ? (
            <GlassCard variant="board" className="p-4">
              <HorizontalEmptyState
                icon={Upload}
                title="No documents uploaded yet"
                description="Upload a document to get started"
              />
            </GlassCard>
          ) : (
            <div className="flex flex-col gap-4">
              {documents?.map((document) => (
                <DocumentItem key={document.ID} document={document} />
              ))}
            </div>
          )}
        </div>

      </form>
    </div>
  );
}
