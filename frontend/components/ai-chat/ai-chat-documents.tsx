import { memo, useState } from "react";
import { document } from "@/wailsjs/go/models";
import { toast } from "sonner";
import { useUploadDocumentRAG } from "@/hooks/use-documents";
import { Check, EllipsisVertical } from "lucide-react";

interface DocumentCardProps {
  document: document.LocalDocument,
  added: boolean,
}

function BaseDocumentCard(props: DocumentCardProps) {
  const documentRAGMutation = useUploadDocumentRAG()
  const [added, setAdded] = useState(props.added)

  const handleAddDocumentToContext = () => {
    if (!props.added) {
      documentRAGMutation.mutate(props.document, {
        onSuccess: () => {
          toast.success(props.document.FileName + " added to context")
          setAdded(true)
        },
        onError: () => {
          toast.error(props.document.FileName + " failed to add to RAG")
        }
      })

    }
    else {
      toast.error(props.document.FileName + " removed from context")
      setAdded(false)
    }

  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 B"
    const k = 1024
    const sizes = ["B", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
  }

  return (
    <div
      key={props.document.ID}
      className={`bg-white/5 flex items-center gap-4 p-4 border-white/5 shadow-lg shadow-black/40 hover:border-white/10 hover:translate-y-1 rounded-xl transition-all duration-300 cursor-pointer group w-full text-left border relative overflow-hidden group/document-card`}
      onClick={() => handleAddDocumentToContext()}>

      {/* Shine effect on hover - inspired by courses-schedule.tsx */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent opacity-0 group-hover/document-card:opacity-100 transition-opacity duration-300" />

      <div className="flex-1 min-w-0 flex flex-col relative z-10">
        <span className={`text-body-small font-semibold text-white truncate transition-colors duration-300 `}>
          {props.document.FileName}
        </span>
        <span className="text-caption text-white/50 truncate transition-colors">
          {formatFileSize(props.document.FileSize)}
        </span>
      </div>

      <div className=" flex items-center justify-center p-2 rounded-full bg-white/10 border border-white/10 shadow-lg shadow-black/40">
        <EllipsisVertical className="w-4 h-4 text-white" strokeWidth={1.5} />
      </div>
      
    </div>);
}

export const AiDocumentCard = memo(BaseDocumentCard, (prevProps, nextProps) => {
  return prevProps.document.ID === nextProps.document.ID
})