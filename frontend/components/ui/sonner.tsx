"use client"

import { useTheme } from "next-themes"
import { Toaster as Sonner } from "sonner"
import { CheckCircle2, XCircle, Info, AlertTriangle } from "lucide-react"

type ToasterProps = React.ComponentProps<typeof Sonner>

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "dark" } = useTheme()

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      position="top-center"
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-white/5 group-[.toaster]:backdrop-blur-xl group-[.toaster]:border group-[.toaster]:border-white/10 group-[.toaster]:text-white group-[.toaster]:shadow-xl group-[.toaster]:shadow-black/40 group-[.toaster]:rounded-xl group-[.toaster]:px-4 group-[.toaster]:py-3 group-[.toaster]:min-w-[320px] group-[.toaster]:max-w-[420px] group-[.toaster]:transition-all group-[.toaster]:duration-300",
          description: "group-[.toast]:text-gray-400 group-[.toast]:text-sm group-[.toast]:mt-1",
          actionButton:
            "group-[.toast]:bg-white/10 group-[.toast]:text-white group-[.toast]:rounded-md group-[.toast]:px-3 group-[.toast]:py-1.5 group-[.toast]:text-sm group-[.toast]:font-medium group-[.toast]:transition-colors group-[.toast]:hover:bg-white/20 group-[.toast]:border group-[.toast]:border-white/10",
          cancelButton:
            "group-[.toast]:bg-white/5 group-[.toast]:text-gray-300 group-[.toast]:rounded-md group-[.toast]:px-3 group-[.toast]:py-1.5 group-[.toast]:text-sm group-[.toast]:font-medium group-[.toast]:transition-colors group-[.toast]:hover:bg-white/10 group-[.toast]:border group-[.toast]:border-white/5",
          success:
            "group-[.toaster]:border-green-400/20",
          error:
            "group-[.toaster]:border-red-400/20",
          warning:
            "group-[.toaster]:border-yellow-400/20",
          info:
            "group-[.toaster]:border-blue-400/20",
          closeButton:
            "group-[.toast]:bg-transparent group-[.toast]:border-0 group-[.toast]:text-gray-400 group-[.toast]:hover:text-white group-[.toast]:hover:bg-white/10 group-[.toast]:rounded-md group-[.toast]:transition-colors group-[.toast]:opacity-70 group-[.toast]:hover:opacity-100",
        },
        duration: 4000,
      }}
      icons={{
        success: <CheckCircle2 className="h-5 w-5 text-green-400/80" />,
        error: <XCircle className="h-5 w-5 text-red-400/80" />,
        warning: <AlertTriangle className="h-5 w-5 text-yellow-400/80" />,
        info: <Info className="h-5 w-5 text-blue-400/80" />,
      }}
      closeButton
      {...props}
    />
  )
}

export { Toaster }
