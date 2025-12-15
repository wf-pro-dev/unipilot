import { Calendar, Check, FileText, User, X } from "lucide-react";
import { Button } from "../ui/button";
import { Card, CardHeader, CardContent, CardFooter } from "../ui/card";
import { Dialog, DialogContent } from "../ui/dialog";
import { document } from "@/wailsjs/go/models";
import { useAuthContext } from "../provider/auth-provider";


interface DocumentAcceptModalProps {
    isOpen: boolean
    onAccept: () => void
    onClose: () => void
    documentData: string | undefined
}

export function DocumentAcceptModal({
    isOpen,
    onAccept,
    onClose,
    documentData
}: DocumentAcceptModalProps) {
    // unmarshal the course dat
    if (!documentData || documentData === undefined) {
        return null
    }

    const { assignments } = useAuthContext()

    const fulldDocumentData = JSON.parse(documentData)
    const document = JSON.parse(documentData) as document.LocalDocument
    const assignment = assignments?.find((assignment) => assignment.ParentID === fulldDocumentData.AssignmentID)

    const formatFileSize = (bytes: number): string => {
        if (bytes === 0) return "0 B"
        const k = 1024
        const sizes = ["B", "KB", "MB", "GB"]
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
      }

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-lg p-0 overflow-hidden gap-0">
                <Card className="bg-transparent p-0 border-0 space-y-0">
                    <CardHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                        <p className="text-xl font-semibold text-white">{document.FileName}</p>
                    </CardHeader>
                    <CardContent className="p-6 space-y-6">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <div className="flex justify-between items-center">
                                    <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                                        <User className="w-3.5 h-3.5" />
                                        <span>Created By</span>
                                    </div>
                                </div>
                                <div className="bg-white/5 border border-white/10 p-3 rounded-lg">
                                    <p className="text-sm text-white font-medium">{fulldDocumentData.User?.Username}</p>
                                </div>
                            </div>


                            <div className="space-y-2">
                                <div className="flex items-center space-x-4 text-sm">
                                    <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                                        <Calendar className="w-3.5 h-3.5" />
                                        <span>Size</span>
                                    </div>

                                </div>
                                <div className="bg-white/5 border border-white/10 p-3 rounded-lg">
                                    <p className="text-sm text-white font-medium">{formatFileSize(document.FileSize)}</p>
                                </div>
                            </div>

                        </div>

                        <div className="space-y-2">
                            <div className="flex justify-between items-center">
                                <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                                    <FileText className="w-3.5 h-3.5" />
                                    <span>Assignment</span>
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 p-3 rounded-lg">
                                <p className="text-sm text-white font-medium">{assignment?.Title}</p>
                            </div>
                        </div>
                        <p className="text-xs text-gray-400 bg-blue-500/10 border border-blue-500/20 p-3 rounded-lg flex items-center gap-2">
                            <Check className="w-3 h-3 text-blue-400" />
                            It will be added to your documents. Make sure to check for similar assignments.
                        </p>
                    </CardContent>
                    <CardFooter className="flex space-x-3 p-6 pt-0">
                        <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white h-10"
                            onClick={(e) => {
                                e.stopPropagation()
                                onClose()
                            }}
                        >
                            <X className="w-4 h-4 mr-2" />
                            Cancel
                        </Button>
                        <Button
                            size="sm"
                            className="flex-1 bg-blue-600 hover:bg-blue-500 text-white border-0 h-10 shadow-[0_0_15px_rgba(37,99,235,0.2)]"
                            onClick={(e) => {
                                e.stopPropagation()
                                onAccept()
                            }}>
                            <Check className="w-4 h-4 mr-2" />
                            Accept Document
                        </Button>
                    </CardFooter>
                </Card>
            </DialogContent>
        </Dialog >
    )
}