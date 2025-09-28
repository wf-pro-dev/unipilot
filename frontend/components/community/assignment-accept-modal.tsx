import { Badge, Calendar, Check, FileText, User, X } from "lucide-react";
import { useAuthContext } from "../provider/auth-provider";
import { Button } from "../ui/button";
import { Card, CardHeader, CardContent, CardFooter } from "../ui/card";
import { Dialog, DialogContent } from "../ui/dialog";
import { assignment, course } from "@/wailsjs/go/models";
import { calculateDaysDifference, getDueDescription, isOverdue, parseDeadline } from "@/lib/date-utils";
import { format } from "date-fns";

interface AssignmentAcceptModalProps {
    isOpen: boolean
    onAccept: () => void
    onClose: () => void
    assignmentData: string
}

export function AssignmentAcceptModal({
    isOpen,
    onAccept,
    onClose,
    assignmentData
}: AssignmentAcceptModalProps) {
    // unmarshal the course dat
    const fulldAssignmentData = JSON.parse(assignmentData) 
    const assignment = JSON.parse(assignmentData) as assignment.LocalAssignment
    const deadline = parseDeadline(assignment.Deadline)
    const isOverdueStatus = isOverdue(deadline, assignment.StatusName)
    const daysUntilDue = calculateDaysDifference(deadline)

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-0 text-white max-w-l max-h-[50vh] overflow-y-auto">
                <Card className="bg-transparent p-0 border-0 space-y-4">
                    <CardHeader>
                        <div className="flex justify-between items-start">
                            <div className="flex items-center space-x-4">

                                <div >
                                    <p className="text-lg font-medium">{assignment?.Title}</p>
                                    <p className="text-sm text-gray-400">{assignment?.Course.Name}</p>
                                </div>
                            </div>
                        </div>
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
                                    <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{fulldAssignmentData.User?.Username}</p>
                                </div>
                            </div>


                            <div className="space-y-2">
                                <div className="flex items-center space-x-4 text-sm">
                                    <div className="flex items-center space-x-2 text-gray-400">
                                        <Calendar className="w-4 h-4" />
                                        <span>Deadline</span>
                                    </div>
                                   
                                </div>
                                <div className="bg-gray-800/50 border border-gray-600 p-2 rounded-lg max-h-[200px] overflow-y-auto">
                                    <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{format(deadline,"PPP")}</p>
                                </div>
                            </div>

                        </div>

                        <div className="space-y-2">
                            <div className="flex justify-between items-center">
                                <div className="flex items-center space-x-2 text-sm text-gray-400">
                                    <FileText className="w-4 h-4" />
                                    <span>Description</span>
                                </div>
                            </div>
                            <div className="bg-gray-800/50 border border-gray-600 p-3 rounded-lg max-h-[200px] overflow-y-auto">
                                <p className={`whitespace-pre-wrap leading-relaxed text-sm text-white block`}>{assignment.Todo}</p>
                            </div>
                        </div>
                        <p className="text-xs text-gray-400">It will be added to your assignments. Make sure to check for similar assignments.</p>
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