import { LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  action?: React.ReactNode
  className?: string
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  className
}: EmptyStateProps) {
  return (
    <div className={cn("flex flex-col items-center justify-center p-8 text-center animate-in fade-in zoom-in duration-500", className)}>
      <div className="w-20 h-20 rounded-full bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center mb-4 ring-1 ring-white/10">
        <Icon className="h-8 w-8 text-white/50" strokeWidth={1.5} />
      </div>
      <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
      <p className="text-muted-foreground max-w-sm mb-6 text-sm leading-relaxed">{description}</p>
      {action}
    </div>
  )
}

export function HorizontalEmptyState({
  icon: Icon,
  title,
  description,
  action,
  className
}: EmptyStateProps) {
  return (
    <div className={cn("flex items-center gap-4 justify-center p-8 text-center animate-in fade-in zoom-in duration-500", className)}>
      <div className="w-20 h-20 rounded-full bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center mb-4 ring-1 ring-white/10">
        <Icon className="h-8 w-8 text-white/50" strokeWidth={1.5} />
      </div>
      <div className="flex flex-col items-start gap-1">
        <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
        <p className="pl-2 text-muted-foreground max-w-sm mb-6 text-sm leading-relaxed">{description}</p>
      </div>
      {action}
    </div>

  )
}
