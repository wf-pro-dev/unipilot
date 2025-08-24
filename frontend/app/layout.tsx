"use client"

import type React from "react"
import { Inter } from "next/font/google"
import "./globals.css"
import { Navbar } from "@/components/navbar"
import { ThemeProvider } from "@/components/theme-provider"
import { AuthProvider } from "@/components/provider/auth-provider"
import { QueryProvider } from "@/components/provider/query-provider"
import { NetworkProvider } from "@/components/provider/network-provider"
import { Toaster } from "sonner"

const inter = Inter({ subsets: ["latin"] })

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
                <Navbar />
                <main>{children}</main>

              </AuthProvider>
              <Toaster position="top-center" />
            </NetworkProvider>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  )
}
