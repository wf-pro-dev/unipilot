import { Check, X } from "lucide-react";
import { useAuthContext } from "../provider/auth-provider";
import { Button } from "../ui/button";
import { Card, CardHeader, CardContent, CardFooter } from "../ui/card";
import { Dialog, DialogContent } from "../ui/dialog";
import { course } from "@/wailsjs/go/models";

interface LinkAcceptModalProps {
    isOpen: boolean
    onAccept: () => void
    onClose: () => void
    courseData: string
}

export function LinkAcceptModal({
    isOpen,
    onAccept,
    onClose,
    courseData
}: LinkAcceptModalProps) {
    // unmarshal the course data
    const course = JSON.parse(courseData) as course.Course
    

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-0 text-white max-w-l max-h-[50vh] overflow-y-auto">
                <Card className="bg-transparent p-0 border-0 space-y-4">
                    <CardHeader>
                        <div className="flex justify-between items-start">
                            <div className="flex items-center space-x-4">
                                <div className={`w-6 h-6 rounded-full ${course?.Color}`} />
                                <div >
                                    <p className="text-lg font-medium">{course?.Name}</p>
                                    <p className="text-sm text-gray-400">{course?.Code}</p>
                                </div>
                            </div>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm font-medium">Are you sure you want to accept this link?</p>
                        <p className="text-xs text-gray-400">It will be added to your courses and you will be able to view the assignments and documents.</p>
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
        </Dialog>
    )
}