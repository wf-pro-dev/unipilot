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
    assignmentData: string | undefined
}

export function AssignmentAcceptModal({
    isOpen,
    onAccept,
    onClose,
    assignmentData
}: AssignmentAcceptModalProps) {
    // unmarshal the course dat
    if (!assignmentData || assignmentData === undefined) {
        return null
    }
    const fulldAssignmentData = JSON.parse(assignmentData) 
    const assignment = JSON.parse(assignmentData) as assignment.LocalAssignment
    const deadline = parseDeadline(assignment.Deadline)


    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-lg p-0 overflow-hidden gap-0">
                <Card className="bg-transparent p-0 border-0 space-y-0">
                    <CardHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                        <div className="flex justify-between items-start">
                            <div className="space-y-1">
                                <div className="flex items-center space-x-2 mb-1">
                                    <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">{assignment?.Course?.Name}</span>
                                </div>
                                <p className="text-xl font-semibold text-white leading-tight">{assignment?.Title}</p>
                            </div>
                        </div>
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
                                    <p className="text-sm text-white font-medium">{fulldAssignmentData.User?.Username}</p>
                                </div>
                            </div>


                            <div className="space-y-2">
                                <div className="flex items-center space-x-4 text-sm">
                                    <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                                        <Calendar className="w-3.5 h-3.5" />
                                        <span>Deadline</span>
                                    </div>
                                   
                                </div>
                                <div className="bg-white/5 border border-white/10 p-3 rounded-lg">
                                    <p className="text-sm text-white font-medium">{format(deadline,"PPP")}</p>
                                </div>
                            </div>

                        </div>

                        <div className="space-y-2">
                            <div className="flex justify-between items-center">
                                <div className="flex items-center space-x-2 text-xs font-medium text-gray-400 uppercase tracking-wider">
                                    <FileText className="w-3.5 h-3.5" />
                                    <span>Description</span>
                                </div>
                            </div>
                            <div className="bg-white/5 border border-white/10 p-3 rounded-lg max-h-[200px] overflow-y-auto custom-scrollbar">
                                <p className={`whitespace-pre-wrap leading-relaxed text-sm text-gray-300 block`}>{assignment.Todo}</p>
                            </div>
                        </div>
                        <p className="text-xs text-gray-400 bg-blue-500/10 border border-blue-500/20 p-3 rounded-lg flex items-center gap-2">
                            <Check className="w-3 h-3 text-blue-400" />
                            It will be added to your assignments. Make sure to check for similar assignments.
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
                            Accept Assignment
                        </Button>
                    </CardFooter>
                </Card>
            </DialogContent>
        </Dialog >
    )
}