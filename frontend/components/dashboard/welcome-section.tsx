"use client"

import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Plus, BookOpen, ClipboardList } from "lucide-react"
import Link from "next/link"
import { useAuthContext } from "../provider/auth-provider"
import { useMemo } from "react"

export function WelcomeSection() {
  const { user } = useAuthContext()
  const currentTime = useMemo(() => new Date(), [])
  const hour = currentTime.getHours()

  let greeting = "Good morning"
  if (hour >= 12 && hour < 17) {
    greeting = "Good afternoon"
  } else if (hour >= 17) {
    greeting = "Good evening"
  }

  return (
    <div className="space-y-6">
      

      <Card className="glass border-0">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between space-y-4 sm:space-y-0">
            <div>
              <h2 className="text-xl font-semibold text-white mb-2">Quick Actions</h2>
              <p className="text-gray-400">Get started with your most common tasks</p>
            </div>
            <div className="flex flex-wrap gap-3">
              <Link href="/assignments">
                <Button variant="outline" className="border-gray-600 bg-transparent">
                  <Plus className="h-4 w-4 mr-2" />
                  Add Assignment
                </Button>
              </Link>
              <Link href="/courses">
                <Button variant="outline" className="border-gray-600 bg-transparent">
                  <BookOpen className="h-4 w-4 mr-2" />
                  View Courses
                </Button>
              </Link>
              <Link href="/assignments">
                <Button className="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700">
                  <ClipboardList className="h-4 w-4 mr-2" />
                  View All Assignments
                </Button>
              </Link>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
