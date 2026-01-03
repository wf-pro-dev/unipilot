"use client"

import type React from "react"
import dynamic from "next/dynamic"
import { Inter } from "next/font/google"
import "./globals.css"
import { ThemeProvider } from "@/components/theme-provider"
import { AuthProvider } from "@/components/provider/auth-provider"
import { QueryProvider } from "@/components/provider/query-provider"
import { NetworkProvider } from "@/components/provider/network-provider"
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { Toaster } from "@/components/ui/sonner"
import { MainSidebar } from "@/components/sidebar/sidebar";
import { usePathname } from 'next/navigation';
import { MeshBackground } from '@/components/ui/mesh-gradient';


const inter = Inter({ subsets: ["latin"] })
// In your layout.tsx, replace BackgroundWrapper with:


function BackgroundWrapper({ children }: { children: React.ReactNode }) {
  return (
    <MeshBackground 
      animated={true}
      interactive={false}
      density="dense"
      showOrbs={true}
      showGrid={true}
    >
      {children}
    </MeshBackground>
  );
}

function AppContent({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isChatPage = pathname?.startsWith('/chat');

  return (

    <div className="w-screen">
      {!isChatPage ? (
        <SidebarProvider>
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
  );
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
          <ThemeProvider attribute="class" defaultTheme="dark" enableSystem disableTransitionOnChange>
            <NetworkProvider>
              <MeshBackground>
                <AuthProvider>
                  <AppContent>{children}</AppContent>
                </AuthProvider>
              </MeshBackground>
              <Toaster />
            </NetworkProvider>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  )
}
