import { memo, useEffect, useState } from "react";
import { document } from "@/wailsjs/go/models";
import { toast } from "sonner";
import { useUploadDocumentRAG } from "@/hooks/use-documents";
import { Check, Edit, Eye, Plus, X } from "lucide-react";
import { DropdownMenuContent, DropdownMenuItem } from "../ui/dropdown-menu";
import { DropdownMenu, DropdownMenuTrigger } from "../ui/dropdown-menu";
import { useDeleteDocumentRAG } from "@/hooks/use-documents";

interface DocumentCardProps {
  document: document.LocalDocument,
  added: boolean,
}

function BaseDocumentCard(props: DocumentCardProps) {
  const { mutate: documentRAGMutation } = useUploadDocumentRAG()
  const { mutate: deleteDocumentRAG } = useDeleteDocumentRAG()
  // ✅ Remove local state - rely on props.added which comes from the query

  const handleAddDocumentToContext = () => {
    if (!props.added) {
      documentRAGMutation(props.document, {
        onSuccess: () => {
          toast.success(props.document.FileName + " added to context")
          // ✅ No need to set state - query will update via invalidation
        },
        onError: () => {
          toast.error(props.document.FileName + " failed to add to RAG")
        }
      })
    }
  }

  const handleDeleteDocumentFromContext = () => {
    if (props.added) {
      deleteDocumentRAG(props.document, {
        onSuccess: () => {
          toast.success(props.document.FileName + " removed from context")
          // ✅ No need to set state - query will update via invalidation
        },
        onError: () => {
          toast.error(props.document.FileName + " failed to remove from RAG")
        }
      })
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


    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <div className="flex items-center gap-2">
          <div
            key={props.document.ID}
            className={`bg-white/5 flex items-center gap-4 p-4 border-white/5 shadow-lg shadow-black/40 hover:border-white/10 hover:translate-y-1 rounded-xl transition-all duration-300 cursor-pointer group w-full text-left border relative overflow-hidden group/document-card`}
          >

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

            <div className=" flex items-center justify-center p-2 rounded-full bg-white/10 border border-white/10 shadow-lg shadow-black/40 hover:-translate-y-1 transition-all duration-300">
              {props.added ? <Check className="w-4 h-4 text-white" strokeWidth={1.5} /> : <Plus className="w-4 h-4 text-white" strokeWidth={1.5} />}
            </div>

          </div>


        </div>

      </DropdownMenuTrigger>
      {/* ... existing dropdown content ... */}
      <DropdownMenuContent align="end" className="bg-white/5 backdrop-blur-xl border border-white/15 shadow-lg shadow-black/40">

        <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent" />

        <DropdownMenuItem
          onClick={() => { }}
          className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer">

          <Eye className="h-4 w-4 mr-2" />
          Open
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => handleAddDocumentToContext()}
          className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
          disabled={props.added}
        >
          <Edit className="h-4 w-4 mr-2" />
          Add to context
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={(e) => {
            e.stopPropagation()
            handleDeleteDocumentFromContext()
          }}
          className="text-red-400 focus:text-red-300 focus:bg-red-500/10 cursor-pointer"
          disabled={!props.added}
        >
          <X className="h-4 w-4 mr-2" />
          Remove from context
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export const AiDocumentCard = memo(BaseDocumentCard, (prevProps, nextProps) => {
  console.log(prevProps.added, nextProps.added)
  return prevProps.document.ID === nextProps.document.ID && prevProps.added === nextProps.added
})