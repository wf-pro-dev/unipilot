"use client"
import { useState } from "react"
import { toast } from "sonner"
import { LogError } from "@/wailsjs/runtime/runtime"
import { addDays, isSameDay, isBefore } from "date-fns"

import { useAuthContext } from "@/components/provider/auth-provider"
import { Card } from "@/components/ui/card"
import { NotificationsItem } from "@/components/notifications/notifications-item"
import { BellOff } from "lucide-react"
import { DocumentAcceptModal } from "@/components/notifications/document-accept-modal"
import { LinkAcceptModal } from "@/components/community/link-accept-modal"
import { AssignmentAcceptModal } from "@/components/community/assignment-accept-modal"
import { useAcceptLink } from "@/hooks/use-courses"
import { useAcceptAssignment } from "@/hooks/use-assignments"
import { useAcceptDocument } from "@/hooks/use-documents"
import { notifications } from "@/wailsjs/go/models"
import { useDeleteNotification } from "@/hooks/use-notifications"
import { NoteAcceptModal } from "@/components/notifications/note-accept-modal"
import { useAcceptNote } from "@/hooks/use-notes"

export default function NotificationsPage() {
  const { notifications } = useAuthContext()
  const [selectedNotification, setSelectedNotification] = useState<notifications.LocalNotification | null>(null)
  const deleteNotificationMutation = useDeleteNotification()
  const acceptLinkMutation = useAcceptLink()
  const acceptAssignmentMutation = useAcceptAssignment()
  const acceptDocumentMutation = useAcceptDocument()
  const acceptNoteMutation = useAcceptNote()
  
  const handleAcceptLink = () => {
    setSelectedNotification(null)
    acceptLinkMutation.mutate({ courseData: selectedNotification?.data! }, {
      onSuccess: () => {
        toast.success("Course successfully linked")
        setSelectedNotification(null)
        deleteNotificationMutation.mutate(selectedNotification!)
      },
      onError: (error: any) => {
        LogError("Failed to accept link request: " + error)
        toast.error("Failed to link course")
      }
    })
  }

  const handleAcceptAssignment = () => {
    acceptAssignmentMutation.mutate(selectedNotification?.data!, {
      onSuccess: () => {
        toast.success("Assignment successfully accepted")
        setSelectedNotification(null)
        deleteNotificationMutation.mutate(selectedNotification!)
      },
      onError: (error: any) => {
        LogError("Failed to accept assignment: " + error)
        toast.error("Failed to accept assignment")
      }
    })
  }

  const handleAcceptDocument = () => {
    acceptDocumentMutation.mutate(selectedNotification?.data!, {
      onSuccess: () => {
        toast.success("Document successfully accepted")
        setSelectedNotification(null)
        deleteNotificationMutation.mutate(selectedNotification!)
      },
      onError: (error: any) => {
        LogError("Failed to accept document: " + error)
        toast.error("Failed to accept document")
      }
    })
  }

  const handleAcceptNote = () => {
    acceptNoteMutation.mutate(selectedNotification?.data!, {
      onSuccess: () => {
        toast.success("Note successfully accepted")
        setSelectedNotification(null)
        deleteNotificationMutation.mutate(selectedNotification!)
      },
      onError: (error: any) => {
        LogError("Failed to accept note: " + error)
        toast.error("Failed to accept note")
      }
    })
  }
  const handleClose = () => {
    setSelectedNotification(null)
  }

  var categories = [
    {
      label: "Today",
      notifications: notifications?.filter((notification) => {
        return isSameDay(new Date(notification.CreatedAt), new Date())
      })
    },

    {
      label: "This week",
      notifications: notifications?.filter((notification) => {
        return !isSameDay(new Date(notification.CreatedAt), new Date()) && isBefore(new Date(notification.CreatedAt), addDays(new Date(), 7))
      })
    }
  ]

  return (
    <div className=" flex flex-col">

      <div className="relative z-10 flex flex-col flex-1">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
              Notifications
            </h1>
            <p className="mt-2 text-gray-400">
              What's new ({notifications?.length} total)
            </p>
          </div>
        </div>

        {notifications?.length === 0 ? (
          <div className="flex flex-col flex-1 items-center justify-center">
            <div className="glass flex flex-col items-center justify-center rounded-lg p-10 text-center">
              <BellOff className="h-12 w-12 text-white/20 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-white mb-2">No notifications found</h3>
              <p className="text-gray-400">Interact with other users to get notifications</p>

            </div>
          </div>
        ) : (

          <div className="flex flex-col flex-1 space-y-8">
            {categories.map((category) => (
              category.notifications && category.notifications?.length > 0 && (
                <div key={category.label} className="space-y-4">

                  <Card className="flex items-center justify-between glass border-0 p-4">
                    <h2 className="text-lg font-medium">{category.label}</h2>
                    <p>{category.notifications?.length}</p>
                  </Card>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    {category.notifications?.map((notification) => (
                      <NotificationsItem
                        key={notification.ID}
                        notification={notification}
                        setSelectedNotification={setSelectedNotification}
                      />
                    ))}
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </div>


      <LinkAcceptModal
        isOpen={selectedNotification !== null && selectedNotification.type === "sync"}
        onAccept={handleAcceptLink}
        onClose={handleClose}
        courseData={selectedNotification?.data}
      />
      <AssignmentAcceptModal
        isOpen={selectedNotification !== null && selectedNotification.type === "assignment_update"}
        onAccept={handleAcceptAssignment}
        onClose={handleClose}
        assignmentData={selectedNotification?.data}
      />

      <DocumentAcceptModal
        isOpen={selectedNotification !== null && selectedNotification.type === "document_update"}
        onAccept={handleAcceptDocument}
        onClose={handleClose}
        documentData={selectedNotification?.data}
      />

      <NoteAcceptModal
        isOpen={selectedNotification !== null && selectedNotification.type === "note_update"}
        onAccept={handleAcceptNote}
        onClose={handleClose}
        noteData={selectedNotification?.data}
      />

    </div>
  )
}
