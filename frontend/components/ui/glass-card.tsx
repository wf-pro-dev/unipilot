import * as React from "react"
import { cn } from "@/lib/utils"
import { Card } from "@/components/ui/card"

interface GlassCardProps extends React.ComponentProps<typeof Card> {
  variant?: "default" | "hover" | "interactive"
}

export function GlassCard({ className, variant = "default", ...props }: GlassCardProps) {
  return (
    <Card
      className={cn(
        "rounded-2xl bg-gray-900/40 backdrop-blur-xl border-white/10 shadow-lg shadow-black/20",
        variant === "hover" && "transition-all duration-300 hover:bg-gray-800/60 hover:border-white/20",
        variant === "interactive" && "transition-all duration-300 hover:scale-[1.01] hover:shadow-2xl hover:shadow-black/30 hover:border-white/15 cursor-pointer",
        className
      )}
      {...props}
    />
  )
}

