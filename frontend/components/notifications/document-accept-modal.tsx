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
            <DialogContent className="glass border-0 text-white max-w-l max-h-[50vh] overflow-y-auto">
                <Card className="bg-transparent p-0 border-0 space-y-4">
                    <CardHeader>
                        <p className="text-lg font-medium">{document.FileName}</p>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid grid-cols-2 space-x-4">
                            <div className="space-y-2">
                                <div className="flex justify-between items-center">
                                    <div className="flex items-center space-x-2 text-sm text-gray-400">
                                        <User className="w-4 h-4" />
                                        <span>Created By</span>
                                    </div>
                                </div>
                                <div className="bg-gray-800/50 border border-gray-600 p-2 rounded-lg max-h-[200px] overflow-y-auto">
                                    <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{fulldDocumentData.User?.Username}</p>
                                </div>
                            </div>


                            <div className="space-y-2">
                                <div className="flex items-center space-x-4 text-sm">
                                    <div className="flex items-center space-x-2 text-gray-400">
                                        <Calendar className="w-4 h-4" />
                                        <span>Size</span>
                                    </div>

                                </div>
                                <div className="bg-gray-800/50 border border-gray-600 p-2 rounded-lg max-h-[200px] overflow-y-auto">
                                    <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{formatFileSize(document.FileSize)}</p>
                                </div>
                            </div>

                        </div>

                        <div className="space-y-2">
                            <div className="flex justify-between items-center">
                                <div className="flex items-center space-x-2 text-sm text-gray-400">
                                    <FileText className="w-4 h-4" />
                                    <span>Assignment</span>
                                </div>
                            </div>
                            <div className="bg-gray-800/50 border border-gray-600 p-3 rounded-lg max-h-[200px] overflow-y-auto">
                                <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{assignment?.Title}</p>
                            </div>
                        </div>
                        <p className="text-xs text-gray-400">It will be added to your documents. Make sure to check for similar assignments.</p>
                    </CardContent>
                    <CardFooter className="flex space-x-4">
                        <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 text-blue-400 bg-transparent border-blue-600 hover:bg-blue-600/10 space-x-2"
                            onClick={(e) => {
                                e.stopPropagation()
                                onAccept()
                            }}>
                            <Check className="w-4 h-4" />
                            Accept
                        </Button>



                        <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10 space-x-2"
                            onClick={(e) => {
                                e.stopPropagation()
                                onClose()
                            }}
                        >
                            <X className="w-4 h-4" />
                            Cancel
                        </Button>
                    </CardFooter>
                </Card>
            </DialogContent>
        </Dialog >
    )
}