"use client"

import type React from "react"
import { Inter } from "next/font/google"
import "./globals.css"
import { ThemeProvider } from "@/components/theme-provider"
import { AuthProvider } from "@/components/provider/auth-provider"
import { QueryProvider } from "@/components/provider/query-provider"
import { NetworkProvider } from "@/components/provider/network-provider"
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from "@/components/ui/sonner"
import { MainSidebar } from "@/components/sidebar/sidebar"
import { usePathname } from 'next/navigation'
import { MeshBackground } from '@/components/ui/mesh-gradient'
import { useEffect } from "react"

const inter = Inter({ subsets: ["latin"] })


// This component MUST be inside QueryProvider since it uses MainSidebar
function AppContent({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const isChatPage = pathname?.startsWith('/chat')

  return (
    <div className="w-screen">

      {!isChatPage ? (
        <SidebarProvider>
          <Toaster />
          <MainSidebar />
          <main className="flex flex-col flex-1 w-full p-12">
            <SidebarTrigger className="fixed top-4 left-4 z-50 md:hidden" />
            {children}
          </main>
        </SidebarProvider>
      ) : (
        <main className="w-full h-full">
          {children}
        </main>
      )}
    </div>
  )
}

function TestEffect() {
  console.log("🧪 TestEffect rendering")
  useEffect(() => {
    console.log("✅ TestEffect useEffect DID RUN!")
  }, [])
  return null
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>

      <body className={inter.className}>
        <QueryProvider>
          <TestEffect />
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <MeshBackground>
              <AuthProvider>
                <NetworkProvider>
                  <AppContent>{children}</AppContent>
                </NetworkProvider>
              </AuthProvider>
            </MeshBackground>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  )
}