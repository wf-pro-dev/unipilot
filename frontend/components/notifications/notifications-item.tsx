import { Card, CardHeader, CardContent } from "../ui/card"
import { Button } from "../ui/button"
import { useFollow } from "@/hooks/use-follows"

import { notifications } from "@/wailsjs/go/models"
import { Check, X } from "lucide-react"
import { useAuthContext } from "../provider/auth-provider"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { format } from "date-fns"
import { useDeleteNotification } from "@/hooks/use-notifications"
import { toast } from "sonner"

export function NotificationsItem({ notification }: { notification: notifications.LocalNotification }) {
    const acceptButtonText: { [key: string]: string } = {
        "follow": "Follow",
        "sync": "Link",
    }
    const { user: currentUser, following } = useAuthContext()
    // Check if current user is following this user by checking if current user is in the followers list
    const isFollowed = following?.some((following) => following.ID === notification.sender_id) 
    const followMutation = useFollow(currentUser!, false)
    const deleteMutation = useDeleteNotification()

    const handleFollow = () => {
        const message = "following " + notification.title
        LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
        followMutation.mutate(notification.sender!) , {
            onSuccess: () => {
                toast.success("You are now following " + notification.title)
            },
            onError: () => {
                toast.error("Failed to follow " + notification.title)
            }
        }
    }

    const handleDelete = () => {
        const message = "deleting notification " + notification.title
        LogInfo(message + " " + format(new Date(), "yyyy/MM/dd HH:mm:ssxxx"))
        deleteMutation.mutate(notification)
    }



    return (
        <Card
            key={notification.ID}
            className="border-0 transition-all duration-300 cursor-pointer glass group"
        >
            <CardHeader className="pb-2">
               <h2 className="text-lg font-medium">{notification.title}</h2>
            </CardHeader>

            <CardContent className="space-y-2">
                <p className="text-sm text-gray-400">{notification.message}</p>


                <div className="flex pt-2 space-x-2">
                    {!isFollowed && (
                    <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 text-blue-400 bg-transparent border-blue-600 hover:bg-blue-600/10 space-x-2"
                        onClick={(e) => {
                            e.stopPropagation()
                            handleFollow()
                            handleDelete()
                        }}>
                        <Check className="w-4 h-4" />
                        {acceptButtonText[notification.type]}
                    </Button>
                    )} 

                    <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 text-red-400 bg-transparent border-red-600 hover:bg-red-600/10 space-x-2"
                        onClick={(e) => {
                            e.stopPropagation()
                            handleDelete()
                        }}
                    >
                        <X className="w-4 h-4" />
                        Ignore
                    </Button>
                </div>
            </CardContent>
        </Card>
    )
}