import { Dialog, DialogContent } from "../ui/dialog";
import { Button } from "../ui/button";

import { useCallback, useMemo, useState } from "react";
import { Check, X } from "lucide-react";
import { useCourse, useCourseShare } from "@/hooks/use-courses";
import { models } from "@/wailsjs/go/models";
import { FriendList } from "./friend-list";
import { UserItem } from "./user-item";
import { useAuthContext } from "../provider/auth-provider";
import { Checkbox } from "../ui/checkbox";
import { useGetClusterStatus } from "@/hooks/use-courses";

interface LinkRequestModalProps {
    courseID: string
    isOpen: boolean
    onClose: () => void

}

export function LinkRequestModal({ isOpen, onClose, courseID }: LinkRequestModalProps) {

    const { data: course } = useCourse(courseID)

    if (!course) return null;

    const [selectedFriends, setSelectedFriends] = useState<string[]>([])

    const { mutate: requestLinkCourse } = useCourseShare()
    const { user: currentUser } = useAuthContext()

    const onCheckedChange = useCallback((user: models.User, checked: boolean) => {
        console.log('🟡 onCheckedChange called', { user: user.ID, checked });
        if (checked) {
            setSelectedFriends(prev => [...prev, user.ID])
        } else {
            setSelectedFriends(prev => prev.filter(id => id !== user.ID))
        }
    }, [])

    const renderActions = useCallback((user: models.User) => (
        <Checkbox
            checked={selectedFriends.includes(user.ID)}
            onCheckedChange={(checked) => onCheckedChange(user, checked as boolean)}
            className="h-5 w-5 border-white/20 data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
        />
    ), [selectedFriends, onCheckedChange])

    const clusterID = useMemo(() => {
        return course.ClusterID || course.ID
    }, [course])

    const renderItem = useCallback((user: models.User) => {
        return (
            <UserItem
                user={user}
                size="compact"
                actions={renderActions}
            />
        )
    }, [renderActions])

    const handleRequestLinkCourse = () => {
        if (!course) { return }
        requestLinkCourse({ c: course, usersID: selectedFriends })
    }

    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent className="glass border-white/10 text-white max-w-lg max-h-[80vh] overflow-y-auto p-0 overflow-hidden gap-0">

                <div className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <div className="flex flex-col items-center space-y-1 text-center">
                        <p className="text-xl font-semibold text-white">Share your course</p>
                        <p className="text-sm text-gray-400">Select friends to invite to this course</p>
                    </div>
                </div>

                <div className="p-6">
                    <FriendList
                        entityID={currentUser?.ID!}
                        numColumns={1}
                        itemsPerPage={6}
                        containerClassName="gap-4"
                        renderItem={renderItem}

                        initialFilters={{
                            "course": clusterID
                        }}
                    />

                </div>
                <div className="p-6 flex gap-3  border-white/5">
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
            </DialogContent>
        </Dialog>
    )
}