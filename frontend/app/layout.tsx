"use client"

import type React from "react"
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

const inter = Inter({ subsets: ["latin"] })

function AppContent({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isChatPage = pathname?.startsWith('/chat');

  return (
    <div className="page w-screen">
      <div className="fixed top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float" />
      <div className="fixed bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed" />
      
      {!isChatPage ? (
        <SidebarProvider>
          <MainSidebar />
          <main className="flex-1 w-full p-12">
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
              <AuthProvider>
                <AppContent>{children}</AppContent>
              </AuthProvider>
              <Toaster />
            </NetworkProvider>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  )
}
