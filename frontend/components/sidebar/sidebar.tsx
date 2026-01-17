"use client"

import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { Home, BookOpen, ClipboardList, FileText, Users, Settings, LogOut, User, Bell, ChevronDown, ChevronRight } from "lucide-react"
import { useAuthContext } from "@/components/provider/auth-provider"
import { useGetAvatarUrl, useLogout } from "@/hooks/use-auth"
import { OfflineIndicator } from "@/components/ui/offline-indicator"
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
import Image from "next/image"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarGroupContent,
  SidebarRail,
  SidebarGroupLabel
} from "@/components/ui/sidebar"
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from "@/components/ui/collapsible"
import { SidebarMenuSub, SidebarMenuSubItem, SidebarMenuSubButton } from "@/components/ui/sidebar"
import { GlassCard } from "@/components/ui/glass-card"
import { useNextAssignments } from "@/hooks/use-assignments"
import { AiAssignmentCard } from "@/components/ai-chat/ai-chat-assignments"

const navItems = [
  { href: "/", label: "Dashboard", icon: Home },
  {
    href: "/courses", label: "Courses", icon: BookOpen, items: [
      { href: "/courses?view=schedule", label: "Schedule" },
      { href: "/courses?view=linked", label: "Linked" },
      { href: "/courses?view=list", label: "All" },
    ]
  },
  {
    href: "/assignments", label: "Assignments", icon: ClipboardList, items: [
      { href: "/assignments?view=today", label: "Today" },
      { href: "/assignments?view=calendar", label: "Calendar" },
      { href: "/assignments?view=list", label: "All" },
    ]
  },
  {
    href: "/notes", label: "Notes", icon: FileText, items: [
      { href: "/notes", label: "All" },
      { href: "/notes", label: "Videos" },
    ]
  },
  {
    href: "/community", label: "Community", icon: Users, items: [
      { href: "/community", label: "Explore" },
      { href: "/community", label: "Followers" },
      { href: "/community", label: "Following" },

    ]
  },

]



export function MainSidebar() {
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const { user } = useAuthContext()
  const { mutate: logout } = useLogout()
  const { data: avatarUrl } = useGetAvatarUrl()
  const { data: nextAssignments } = useNextAssignments()

  const finalAvatarUrl = avatarUrl || user?.Avatar || "/placeholder.svg?height=40&width=40"

  const handleLogout = () => {
    logout()
    window.location.reload()
  }

  const upcomingAssignments = nextAssignments?.slice(0, 3) || []

  // Helper to check if a link is active, handling query parameters
  const isLinkActive = (href: string) => {
    const [path, query] = href.split('?')

    // Strict check for root
    if (path === "/" && pathname !== "/") return false

    // Check path match (exact or startswith for nested routes)
    const isPathMatch = path === pathname || (path !== "/" && pathname?.startsWith(path))

    if (!isPathMatch) return false

    // If no query in link, it's a match based on path
    if (!query) return true

    // If query in link, check params
    const linkParams = new URLSearchParams(query)
    for (const [key, value] of linkParams.entries()) {
      if (searchParams?.get(key) !== value) return false
    }

    return true
  }

  return (
    <Sidebar collapsible="icon" className="h-screen shadow-2xl border-none bg-transparent" variant="sidebar">
      <GlassCard variant="board" className="flex flex-col flex-1 rounded-none h-full p-0 overflow-hidden">
        <SidebarHeader className="p-4 pb-2 border-b border-white/5 bg-white/5 backdrop-blur-3xl">
          <Link href="/" className="flex items-center space-x-2 w-full overflow-hidden">
            <Image src="/icon.png" width={32} height={32} alt="Unipilot" className="rounded-lg shrink-0 w-8 h-8 group-data-[collapsible=icon]:w-6 group-data-[collapsible=icon]:h-6" />
            <span className="text-xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-400 truncate group-data-[collapsible=icon]:hidden transition-all duration-300">
              UniPilot
            </span>
          </Link>
        </SidebarHeader>

        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu className="gap-2">
                {navItems.map((item) => (
                  item.items == undefined ? (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton
                        asChild
                        isActive={isLinkActive(item.href)}
                        tooltip={item.label}
                        className="transition-all duration-200"
                      >
                        <Link href={item.href}>
                          <item.icon width={16} height={16} strokeWidth={1.5} />
                          <span className="group-data-[collapsible=icon]:hidden">{item.label}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ) : (
                    <Collapsible key={item.href} defaultOpen={pathname?.includes(item.href)} className="group/collapsible">
                      <SidebarMenuItem>
                        <SidebarMenuButton
                          isActive={isLinkActive(item.href)}
                          tooltip={item.label}
                          className="flex items-center justify-between"
                        >
                          <Link href={item.href} className="flex items-center w-full gap-2">
                            <item.icon width={16} height={16} strokeWidth={1.5} />
                            <span className="group-data-[collapsible=icon]:hidden">{item.label}</span>
                          </Link>
                          <CollapsibleTrigger asChild>
                            <div role="button" className="hover:bg-white/10 rounded-sm cursor-pointer group-data-[collapsible=icon]:hidden">
                              <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                            </div>
                          </CollapsibleTrigger>

                        </SidebarMenuButton>


                        <CollapsibleContent>
                          <SidebarMenuSub className="border-white/20">
                            {item.items.map((subItem) => (
                              <SidebarMenuSubItem key={subItem.href}>
                                <SidebarMenuSubButton
                                  asChild
                                  isActive={isLinkActive(subItem.href)}
                                  className="text-gray-400 hover:text-white hover:bg-white/5 data-[active=true]:bg-white/10 data-[active=true]:text-white transition-colors"
                                >
                                  <Link href={subItem.href}>
                                    <span className="text-sm">{subItem.label}</span>
                                  </Link>
                                </SidebarMenuSubButton>
                              </SidebarMenuSubItem>
                            ))}
                          </SidebarMenuSub>
                        </CollapsibleContent>
                      </SidebarMenuItem>
                    </Collapsible>
                  )

                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup className="p-0 group-data-[collapsible=icon]:hidden">
            <div className="flex items-center justify-between mb-3 px-1">
              <SidebarGroupLabel className="text-xs uppercase tracking-widest text-white/40 font-bold">
                AI Chat Quick Access
              </SidebarGroupLabel>
            </div>
            <div className="flex flex-col gap-2.5">
              {upcomingAssignments.map((assignment) => (

                <AiAssignmentCard assignment={assignment} />

              ))}
              {upcomingAssignments.length === 0 && (
                <div className="px-4 py-8 text-center border border-dashed border-white/10 rounded-xl bg-white/5">
                  <p className="text-[10px] text-muted-foreground/60">No active assignments</p>
                </div>
              )}
            </div>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter className="p-0 border-t border-white/5 bg-gradient-to-t from-black/20 to-transparent">
          <div className="p-4 flex flex-col gap-4">
            <div className="flex items-center justify-between group-data-[collapsible=icon]:flex-col group-data-[collapsible=icon]:gap-2">
              <Link href="/notifications" className="relative">
                <Button variant="ghost" size="icon" className="rounded-full w-8 h-8 hover:bg-white/10 text-gray-400 hover:text-white transition-colors">
                  <Bell className="w-5 h-5" />
                </Button>
              </Link>
              <OfflineIndicator variant="icon" />
            </div>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="w-full justify-start px-2 group-data-[collapsible=icon]:justify-center h-12 hover:bg-white/5 transition-colors rounded-xl">
                  <Avatar className="w-8 h-8 mr-2 group-data-[collapsible=icon]:mr-0 transition-all duration-200 ring-2 ring-white/10">
                    <AvatarImage src={finalAvatarUrl} alt="User" />
                    <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-600 text-white font-semibold text-xs">
                      {user?.Username?.split(" ").map((n: string) => n[0]).join("") || "U"}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col items-start truncate group-data-[collapsible=icon]:hidden transition-all duration-200">
                    <span className="text-sm font-medium truncate w-full text-left text-gray-200">
                      {user?.Username || "User"}
                    </span>
                    <span className="text-[10px] text-gray-500 truncate w-full text-left">
                      {user?.Email}
                    </span>
                  </div>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56 glass border-white/20" align="start" side="right" forceMount>
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
        </SidebarFooter>
      </GlassCard>
      <SidebarRail />
    </Sidebar>
  )
}
