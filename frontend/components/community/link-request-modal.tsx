import { Dialog, DialogContent } from "../ui/dialog";
import { Button } from "../ui/button";
import { useAuthContext } from "../provider/auth-provider";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { useState } from "react";
import { Check, Users, X } from "lucide-react";
import { useCourse, useCourseShare } from "@/hooks/use-courses";
import { LogInfo } from "@/wailsjs/runtime/runtime";
import { models } from "@/wailsjs/go/models";

interface LinkRequestModalProps {
    courseID: string
    isOpen: boolean
    onClose: () => void

}

export function LinkRequestModal({ isOpen, onClose, courseID }: LinkRequestModalProps) {
    const { data: course } = useCourse(courseID)
    
    const { friends } = useAuthContext()
    const [selectedFriends, setSelectedFriends] = useState<string[]>([])

    const { mutate: requestLinkCourse } = useCourseShare()

    const handleShare = (friend: models.User) => {
        if (selectedFriends.includes(friend.ID)) {
            setSelectedFriends(selectedFriends.filter((f) => f !== friend.ID))
        } else {
            setSelectedFriends([...selectedFriends, friend.ID])
        }
    }

    const handleRequestLinkCourse = () => {
        if (!course) { return }
        requestLinkCourse({ c: course, usersID: selectedFriends })
    }

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-xl max-h-[90vh] overflow-y-auto p-0 overflow-hidden gap-0">

                <div className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <div className="flex flex-col items-center space-y-1 text-center">
                        <p className="text-xl font-semibold text-white">Share your course</p>
                        <p className="text-sm text-gray-400">Select friends to invite to this course</p>
                    </div>
                </div>

                <div className="p-6 space-y-6">
                    {friends && friends.length > 0 ? (
                        <div className="grid grid-cols-3 gap-3">
                            {friends?.map((friend) => (
                                <div
                                    key={friend.ID}
                                    className={`flex flex-col items-center justify-center p-4 rounded-xl transition-all duration-200 cursor-pointer border ${selectedFriends.includes(friend.ID) ? 'bg-blue-600/20 border-blue-500/50' : 'bg-white/5 border-white/5 hover:bg-white/10 hover:border-white/10'}`}
                                    onClick={() => handleShare(friend)}
                                >
                                    <div className="relative">
                                        <Avatar className="w-10 h-10 mb-2 ring-2 ring-transparent transition-all duration-200">
                                            <AvatarImage src={friend.Avatar || "/placeholder.svg"} alt={friend.Username} />
                                            <AvatarFallback className="bg-gray-700 text-white">
                                                {friend.Username
                                                    .split(" ")
                                                    .map((n: string) => n[0])
                                                    .join("")}
                                            </AvatarFallback>
                                        </Avatar>
                                        {selectedFriends.includes(friend.ID) && (
                                            <div className="absolute -top-1 -right-1 bg-blue-500 rounded-full p-0.5 ring-2 ring-[#0f172a]">
                                                <Check className="w-3 h-3 text-white" />
                                            </div>
                                        )}
                                    </div>
                                    <p className={`text-xs font-medium truncate w-full text-center ${selectedFriends.includes(friend.ID) ? 'text-blue-200' : 'text-gray-300'}`}>{friend.Username}</p>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="py-12 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                            <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mx-auto mb-4">
                                <Users className="h-8 w-8 text-gray-500" />
                            </div>
                            <h3 className="text-lg font-medium text-white mb-1">No friends found</h3>
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
                            disabled={selectedFriends.length === 0}
                        >
                            <Check className="mr-2 w-4 h-4" />
                            Share with {selectedFriends.length > 0 ? `${selectedFriends.length} ` : ''}Friend{selectedFriends.length !== 1 ? 's' : ''}
                        </Button>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}