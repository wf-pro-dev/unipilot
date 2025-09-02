import { Dialog, DialogContent, DialogPortal, DialogOverlay } from "../ui/dialog";
import { Button } from "../ui/button";
import { Separator } from "../ui/separator";
import { useAuthContext } from "../provider/auth-provider";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Card } from "../ui/card";
import { useState } from "react";
import { user } from "@/wailsjs/go/models"
import { Check, CheckCircle2, X } from "lucide-react";

export function LinkRequestModal({ isOpen, onClose }: { isOpen: boolean, onClose: () => void }) {
    const { followers } = useAuthContext()
    const [selectedFollowers, setSelectedFollowers] = useState<user.User[]>([])

    const handleShare = (follower: user.User) => {
        if (selectedFollowers.includes(follower)) {
            setSelectedFollowers(selectedFollowers.filter((f) => f.ID !== follower.ID))
        } else {
            setSelectedFollowers([...selectedFollowers, follower])
        }
    }

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-0 text-white max-w-xl max-h-[90vh] overflow-y-auto">
                <div className="space-y-6">
                    {/* Course Header */}

                    <div className="flex flex-col items-center space-y-2">
                        <p className="text-lg font-medium text-white ">Share your course</p>
                        <p className="text-sm text-gray-300">Allow your followers to join your course</p>
                    </div>

                    <Separator className="bg-gray-700" />

                    {followers && followers.length > 0 ? (
                        <div className="grid grid-cols-3 gap-4">
                            {followers?.map((follower) => (
                                <Card key={follower.ID} className={`flex items-center space-x-2 p-4 transition-all duration-300 cursor-pointer ${selectedFollowers.includes(follower) ? 'bg-blue-500/50 hover:bg-blue-500/70' : 'glass hover:bg-white/5'}`} onClick={() => handleShare(follower)}>
                                    <Avatar className="w-6 h-6">
                                        <AvatarImage src={follower.Avatar || "/placeholder.svg"} alt={follower.Username} />
                                        <AvatarFallback>
                                            {follower.Username
                                                .split(" ")
                                                .map((n: string) => n[0])
                                                .join("")}
                                        </AvatarFallback>
                                    </Avatar>
                                    <p className="text-sm font-medium text-white truncate">{follower.Username}</p>

                                </Card>
                            ))}

                        </div>
                    ) : (
                        <div className="py-8 text-center">
                            <CheckCircle2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                            <h3 className="text-lg font-medium text-white mb-2">No followers found</h3>
                            <p className="text-gray-400">Follow someone to get started</p>
                        </div>
                    )}


                    <Separator className="bg-gray-700" />

                    {/* Actions */}
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 bg-transparent border-gray-600"
                            onClick={(e) => {
                                e.stopPropagation()
                                onClose()
                            }}
                        >
                            <Check className="mr-1 w-3 h-3" />
                            Confirm
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10"
                            onClick={(e) => {
                                e.stopPropagation()
                                onClose()
                            }}
                        >
                            <X className="mr-1 w-3 h-3" />
                            Cancel
                        </Button>


                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}