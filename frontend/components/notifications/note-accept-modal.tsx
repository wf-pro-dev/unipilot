import { BookOpen, Check, Tag, User, X } from "lucide-react";
import { Button } from "../ui/button";
import { Card, CardHeader, CardContent, CardFooter } from "../ui/card";
import { Dialog, DialogContent } from "../ui/dialog";
import { Badge } from "../ui/badge";
import { note } from "@/wailsjs/go/models";


interface NoteAcceptModalProps {
    isOpen: boolean
    onAccept: () => void
    onClose: () => void
    noteData: string | undefined
}

export function NoteAcceptModal({
    isOpen,
    onAccept,
    onClose,
    noteData
}: NoteAcceptModalProps) {
    // unmarshal the course dat
    if (!noteData || noteData === undefined) {
        return null
    }

    const fulldNoteData = JSON.parse(noteData)


    // Parse keywords if they're stored as JSON string
    console.log("note keywords", Object.keys(fulldNoteData), fulldNoteData["keywords"])
    const keywords = fulldNoteData["keywords"] ?
        (fulldNoteData["keywords"].startsWith('[') ? JSON.parse(fulldNoteData["keywords"]) : fulldNoteData["keywords"].split(',')) :
        []

    console.log("note keywords", keywords)

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-0 text-white max-w-l max-h-[50vh] overflow-y-auto">
                <Card className="bg-transparent p-0 border-0 space-y-4">
                    <CardHeader>
                        <p className="text-lg font-medium">{fulldNoteData["title"]}</p>
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
                                    <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{fulldNoteData.User?.Username}</p>
                                </div>
                            </div>


                            <div className="space-y-2">
                                <div className="flex items-center space-x-4 text-sm">
                                    <div className="flex items-center space-x-2 text-gray-400">
                                        <BookOpen className="w-4 h-4" />
                                        <span>Course</span>
                                    </div>

                                </div>
                                <div className="bg-gray-800/50 border border-gray-600 p-2 rounded-lg max-h-[200px] overflow-y-auto">
                                    <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{fulldNoteData.Course?.Name}</p>
                                </div>
                            </div>

                        </div>

                        <div className="space-y-2">
                            <div className="flex justify-between items-center">
                                <div className="flex items-center space-x-2 text-sm text-gray-400">
                                    <Tag className="w-4 h-4" />
                                    <span>Keywords</span>
                                </div>
                            </div>
                            <div className="flex flex-wrap gap-1.5 p-3 rounded-lg border border-gray-600">
                                {keywords.map((keyword: string, index: number) => (
                                    <Badge key={index} variant="secondary" className="text-xs">
                                        {keyword}
                                    </Badge>
                                ))} 
                            </div>
                        </div>
                        <p className="text-xs text-gray-400">It will be added to your notes. Make sure to check for similar notes.</p>
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