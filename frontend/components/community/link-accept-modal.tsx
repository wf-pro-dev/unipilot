import { Check, X } from "lucide-react";
import { Button } from "../ui/button";
import { Card, CardHeader, CardContent, CardFooter } from "../ui/card";
import { Dialog, DialogContent } from "../ui/dialog";
import { course } from "@/wailsjs/go/models";

interface LinkAcceptModalProps {
    isOpen: boolean
    onAccept: () => void
    onClose: () => void
    courseData: string | undefined
}

export function LinkAcceptModal({
    isOpen,
    onAccept,
    onClose,
    courseData
}: LinkAcceptModalProps) {

    if (!courseData || courseData === undefined) {
        return null
    }
    // unmarshal the course data
    const course = JSON.parse(courseData) as course.Course
    

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-md p-0 overflow-hidden gap-0">
                <Card className="bg-transparent p-0 border-0 space-y-0">
                    <CardHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                        <div className="flex justify-between items-start">
                            <div className="flex items-center space-x-4">
                                <div className={`w-10 h-10 rounded-full ${course?.Color} shadow-lg shadow-black/20`} />
                                <div className="space-y-1">
                                    <p className="text-lg font-bold text-white leading-tight">{course?.Name}</p>
                                    <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">{course?.Code}</p>
                                </div>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent className="p-6 space-y-4">
                        <div className="bg-white/5 border border-white/10 p-4 rounded-lg">
                            <p className="text-sm font-medium text-white mb-1">Are you sure you want to accept this link?</p>
                            <p className="text-xs text-gray-400 leading-relaxed">It will be added to your courses and you will be able to view the assignments and documents.</p>
                        </div>
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
                            Accept Link
                        </Button>
                    </CardFooter>
                </Card>
            </DialogContent>
        </Dialog>
    )
}