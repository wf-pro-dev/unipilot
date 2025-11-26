import { memo, useState } from "react";
import { document } from "@/wailsjs/go/models";
import { toast } from "sonner";
import { useUploadDocumentRAG } from "@/hooks/use-documents";
import { Check, FileText } from "lucide-react";

interface DocumentCardProps {
    document: document.LocalDocument,
    added: boolean,
}

function BaseDocumentCard(props: DocumentCardProps) {
    const documentRAGMutation = useUploadDocumentRAG()
    const [added, setAdded] = useState(props.added)
        
    const handleAddDocumentToContext = () => {
      if (!props.added) {
        documentRAGMutation.mutate(props.document,{
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

    return (
      <button
        key={props.document.ID}
        className="flex items-center justify-between p-4 glass rounded-md hover:bg-gray-800 transition-colors cursor-pointer"
        onClick={() => handleAddDocumentToContext()}>
        <span className="text-xs">{props.document.FileName}</span>
        {added ? <Check className="w-4 h-4" strokeWidth={1} /> : <FileText className="w-4 h-4" strokeWidth={1} />}

      </button>);
  }

export const AiDocumentCard = memo(BaseDocumentCard, (prevProps, nextProps) => {
    return prevProps.document.ID === nextProps.document.ID 
})