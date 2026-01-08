import { Dialog, DialogContent } from "../ui/dialog";
import { Button } from "../ui/button";
import { useAuthContext } from "../provider/auth-provider";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { useState } from "react";
import { user } from "@/wailsjs/go/models"
import { Check, Users, X} from "lucide-react";
import { useCourseShare } from "@/hooks/use-courses";
import { LogInfo } from "@/wailsjs/runtime/runtime";
import { course } from "@/wailsjs/go/models";

interface LinkRequestModalProps {
    isOpen: boolean
    onClose: () => void
    courseID: number
}

export function LinkRequestModal({ isOpen, onClose, courseID }: LinkRequestModalProps) {
    const { courses } = useAuthContext()
    const { followers } = useAuthContext()
    const [selectedFollowers, setSelectedFollowers] = useState<number[]>([])

    const { mutate: requestLinkCourse } = useCourseShare()

    const handleShare = (follower: user.User) => {
        if (selectedFollowers.includes(follower.ID)) {
            setSelectedFollowers(selectedFollowers.filter((f) => f !== follower.ID))
        } else {
            setSelectedFollowers([...selectedFollowers, follower.ID])
        }
    }

    const handleRequestLinkCourse = () => {
        LogInfo("Requesting to link course " + courseID + " to " + selectedFollowers.length + " followers")
        
        var targetCourse = courses?.find((c) => c.ID === courseID) as course.LocalCourse
        if (!targetCourse) {
            return
        }
        requestLinkCourse({ c: targetCourse, usersID: selectedFollowers })
    }

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-xl max-h-[90vh] overflow-y-auto p-0 overflow-hidden gap-0">
               
                <div className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <div className="flex flex-col items-center space-y-1 text-center">
                        <p className="text-xl font-semibold text-white">Share your course</p>
                        <p className="text-sm text-gray-400">Select followers to invite to this course</p>
                    </div>
                </div>

                <div className="p-6 space-y-6">
                    {followers && followers.length > 0 ? (
                        <div className="grid grid-cols-3 gap-3">
                            {followers?.map((follower) => (
                                <div 
                                    key={follower.ID} 
                                    className={`flex flex-col items-center justify-center p-4 rounded-xl transition-all duration-200 cursor-pointer border ${selectedFollowers.includes(follower.ID) ? 'bg-blue-600/20 border-blue-500/50' : 'bg-white/5 border-white/5 hover:bg-white/10 hover:border-white/10'}`} 
                                    onClick={() => handleShare(follower)}
                                >
                                    <div className="relative">
                                        <Avatar className="w-10 h-10 mb-2 ring-2 ring-transparent transition-all duration-200">
                                            <AvatarImage src={follower.Avatar || "/placeholder.svg"} alt={follower.Username} />
                                            <AvatarFallback className="bg-gray-700 text-white">
                                                {follower.Username
                                                    .split(" ")
                                                    .map((n: string) => n[0])
                                                    .join("")}
                                            </AvatarFallback>
                                        </Avatar>
                                        {selectedFollowers.includes(follower.ID) && (
                                            <div className="absolute -top-1 -right-1 bg-blue-500 rounded-full p-0.5 ring-2 ring-[#0f172a]">
                                                <Check className="w-3 h-3 text-white" />
                                            </div>
                                        )}
                                    </div>
                                    <p className={`text-xs font-medium truncate w-full text-center ${selectedFollowers.includes(follower.ID) ? 'text-blue-200' : 'text-gray-300'}`}>{follower.Username}</p>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                            <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
                                <Users className="h-8 w-8 text-gray-500" />
                            </div>
                            <h3 className="text-lg font-medium text-white mb-1">No followers found</h3>
                            <p className="text-gray-400 text-sm">Follow someone to get started</p>
                        </div>
                    )}

                    <div className="flex gap-3 pt-2 border-t border-white/5 mt-6">
                        <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white h-10"
                            onClick={(e) => {
                                e.stopPropagation()
                                onClose()
                            }}
                        >
                            <X className="mr-2 w-4 h-4" />
                            Cancel
                        </Button>
                        <Button
                            size="sm"
                            className="flex-1 bg-blue-600 hover:bg-blue-500 text-white border-0 h-10 shadow-[0_0_15px_rgba(37,99,235,0.2)]"
                            onClick={(e) => {
                                e.stopPropagation()
                                handleRequestLinkCourse()
                            }}
                            disabled={selectedFollowers.length === 0}
                        >
                            <Check className="mr-2 w-4 h-4" />
                            Share with {selectedFollowers.length > 0 ? `${selectedFollowers.length} ` : ''}Friend{selectedFollowers.length !== 1 ? 's' : ''}
                        </Button>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}