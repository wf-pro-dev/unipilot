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
import { Home, BookOpen, ClipboardList, FileText, Users, Settings, LogOut, User } from "lucide-react"
import { useAuthContext } from "./provider/auth-provider"
import { OfflineIndicator } from "./ui/offline-indicator"
import { useLogout } from "@/hooks/use-auth"
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

  const handleLogout = () => {
    logout()
    window.location.reload()
  }


  return (
    <nav className="fixed top-0 right-0 left-0 z-50 border-0 border-b backdrop-blur-xl glass border-white/10">
      <div className="px-4 mx-auto max-w-7xl sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <Image src="/icon.png" alt="Unipilot" width={24} height={24} className="rounded-lg" />
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
                    className={`flex items-center space-x-1 ${
                      isActive ? "text-white bg-white/10" : "text-gray-300 hover:text-white hover:bg-white/5"
                    }`}
                  >
                    <Icon className="w-4 h-4" />
                    <span>{item.label}</span>
                  </Button>
                </Link>
              )
            })}
          </div>

          <div className="flex items-center space-x-2">
            <OfflineIndicator variant="icon" />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative w-8 h-8 rounded-full">
                  <Avatar className="w-8 h-8">
                    <AvatarImage src="/placeholder.svg?height=32&width=32" alt="User" />
                    <AvatarFallback>JD</AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56 glass border-white/10" align="end" forceMount>
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none text-white">
                      {user?.Email || "User"}
                    </p>
                    <p className="text-xs leading-none text-gray-400">
                      {user?.Email || "user@student.acc.edu"}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator className="bg-white/10" />
                <Link href="/profile">
                  <DropdownMenuItem className="text-gray-300 cursor-pointer hover:text-white hover:bg-white/5">
                    <User className="mr-2 w-4 h-4" />
                    <span>Profile</span>
                  </DropdownMenuItem>
                </Link>
                <DropdownMenuItem className="text-gray-300 hover:text-white hover:bg-white/5">
                  <Settings className="mr-2 w-4 h-4" />
                  <span>Settings</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator className="bg-white/10" />
                <DropdownMenuItem 
                  className="text-gray-300 cursor-pointer hover:text-white hover:bg-white/5"
                  onClick={handleLogout}
                >
                  <LogOut className="mr-2 w-4 h-4" />
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
