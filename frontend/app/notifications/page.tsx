"use client"
import { AddNoteDialog } from "@/components/notes/note-add-dialog"
import { NoteView } from "@/components/notes/note-view"
import { NoteDetailModal } from "@/components/notes/note-detail-modal"
import { useNotes, useDeleteNote, useUpdateNote, useCreateNote } from "@/hooks/use-notes"
import { note } from "@/wailsjs/go/models"
import { useState } from "react"
import { toast } from "sonner"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { addDays, isSameDay, isBefore } from "date-fns"

import { useAuthContext } from "@/components/provider/auth-provider"
import { Card } from "@/components/ui/card"
import { NotificationsItem } from "@/components/notifications/notifications-item"
import { BellOff } from "lucide-react"

export default function NotificationsPage() {
  const { notifications } = useAuthContext()
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
    <div className="page flex flex-col">
      <div className="absolute left-10 top-20 w-72 h-72 rounded-full blur-3xl bg-blue-500/10 animate-float"></div>
      <div className="absolute right-10 bottom-20 w-96 h-96 rounded-full blur-3xl bg-purple-500/10 animate-float-delayed"></div>

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
                      <NotificationsItem key={notification.ID} notification={notification} />
                    ))}
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  )
}
