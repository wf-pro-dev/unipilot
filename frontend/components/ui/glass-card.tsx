import * as React from "react"
import { cn } from "@/lib/utils"
import { Card } from "@/components/ui/card"
import { cva, VariantProps } from "class-variance-authority"

export type GlassCardVariants = VariantProps<typeof glassCardContentVariants>['variant']

export interface GlassCardProps extends React.ComponentProps<typeof Card> {
  variant?: GlassCardVariants
}
export const glassCardContentVariants = cva(
  "flex flex-col flex-1 rounded-xl backdrop-blur-md shadow-lg transition-all duration-300",
  {
    variants: {
      variant: {
        default: "bg-primary/20 text-primary-blue-400 border border-primary-blue-400 hover:bg-gradient-to-br hover:from-primary/25 hover:to-transparent",
        outline: "bg-white/5 border border-white/10 hover:bg-white/10 hover:border-white/15",
        board:"bg-white/5 border-white/10",
        
        
      }
    }
  },

)
export function GlassCard({ className, variant = "default", ...props }: GlassCardProps) {
  return (
    <Card
      className={cn(
        glassCardContentVariants({ variant }),
        className
      )}
      {...props}
    />
  )
}

