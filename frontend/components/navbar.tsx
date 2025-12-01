"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Home, BookOpen, ClipboardList, FileText, Users, Settings, LogOut, User, Bell } from "lucide-react"
import { useAuthContext } from "./provider/auth-provider"
import { OfflineIndicator } from "./ui/offline-indicator"
import { useGetAvatarUrl, useLogout } from "@/hooks/use-auth"
import Image from "next/image"

const navItems = [
  { href: "/", label: "Dashboard", icon: Home },
  { href: "/courses", label: "Courses", icon: BookOpen },
  { href: "/assignments", label: "Assignments", icon: ClipboardList },
  { href: "/notes", label: "Notes", icon: FileText },
  { href: "/community", label: "Community", icon: Users },
]

export function Navbar() {
  const pathname = usePathname()
  const { user } = useAuthContext()
  const { mutate: logout } = useLogout()
  const { data: avatarUrl } = useGetAvatarUrl()

  const finalAvatarUrl = avatarUrl || "/placeholder.svg?height=40&width=40"

  const handleLogout = () => {
    logout()
    window.location.reload()
  }

  console.log("navbar user", user)
  console.log("navbar user avatar", user?.Avatar)


  return (
    <nav className="fixed top-0 right-0 left-0 z-50 border-0 border-b backdrop-blur-xl glass border-white/10">
      <div className="px-4 mx-auto max-w-7xl sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <Image src="/icon.png" alt="Unipilot" width={32} height={32} className="rounded-lg" />
              <span className="text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400">
                UniPilot
              </span>
            </Link>
          </div>

          <div className="hidden items-center space-x-2 md:flex">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = pathname === item.href
              return (
                <Link key={item.href} href={item.href}>
                  <Button
                    variant={isActive ? "secondary" : "ghost"}
                    className={`flex items-center space-x-1 transition-all duration-300 relative ${
                      isActive 
                        ? "text-white bg-white/10 shadow-[0_0_15px_rgba(59,130,246,0.2)] after:content-[''] after:absolute after:bottom-0 after:left-0 after:right-0 after:h-[2px] after:bg-blue-400 after:rounded-full" 
                        : "text-gray-300 hover:text-white hover:bg-white/5"
                    }`}
                  >
                    <Icon className={`w-4 h-4 ${isActive ? "text-blue-400" : ""}`} />
                    <span>{item.label}</span>
                  </Button>
                </Link>
              )
            })}
          </div>

          <div className="flex items-center space-x-4">
            
            <Link href="/notifications" className="relative w-8 h-8 rounded-full glass" >
              <Button variant="ghost" className="relative w-8 h-8 rounded-full">
                <Bell className="w-6 h-6" />
              </Button>
            </Link>

            <OfflineIndicator variant="icon" />
            
            <DropdownMenu>
              
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative w-10 h-10 rounded-full transition-smooth hover:scale-110 hover:ring-2 hover:ring-blue-500/50">
                  <Avatar className="w-10 h-10 ring-2 ring-white/10">
                    <AvatarImage src={finalAvatarUrl} alt="User" />
                    <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-600 text-white font-semibold">
                      {user?.Username?.split(" ").map((n: string) => n[0]).join("") || "U"}
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              
              <DropdownMenuContent className="w-56 glass border-white/20" align="end" forceMount>
                <DropdownMenuLabel className="font-normal p-3">
                  <div className="flex flex-col space-y-1.5">
                    <p className="text-body-small font-semibold leading-none text-white">
                      {user?.Username || "User"}
                    </p>
                    <p className="text-caption leading-none text-gray-400">
                      {user?.Email || "user@student.acc.edu"}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator className="bg-white/20" />
                <Link href="/profile">
                  <DropdownMenuItem className="text-body-small text-gray-300 cursor-pointer hover:text-white hover:bg-blue-500/10 transition-smooth p-3">
                    <User className="mr-3 w-4 h-4" />
                    <span>Profile</span>
                  </DropdownMenuItem>
                </Link>
                <DropdownMenuItem className="text-body-small text-gray-300 hover:text-white hover:bg-blue-500/10 transition-smooth p-3">
                  <Settings className="mr-3 w-4 h-4" />
                  <span>Settings</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator className="bg-white/20" />
                <DropdownMenuItem 
                  className="text-body-small text-red-400 cursor-pointer hover:text-red-300 hover:bg-red-500/10 transition-smooth p-3"
                  onClick={handleLogout}
                >
                  <LogOut className="mr-3 w-4 h-4" />
                  <span>Log out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

          </div>
        </div>
      </div>
    </nav>
  )
}
