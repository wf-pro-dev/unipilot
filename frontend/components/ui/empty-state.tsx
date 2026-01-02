import { LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "./button"

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  onClick?: () => void
  buttonText?: string
  className?: string
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  onClick,
  buttonText,
  className
}: EmptyStateProps) {
  return (
    <div className={cn("flex flex-col items-center justify-center text-center gap-4 animate-in fade-in zoom-in duration-500", className)}>
      <div className="w-20 h-20 rounded-full bg-gradient-to-br from-white/10 to-transparent flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-black/20">
        <Icon className="h-8 w-8 text-text-caption" strokeWidth={1.5} />
      </div>

      <div className="flex flex-col items-center gap-2">
        <h5 className="text-h5">{title}</h5>
        <p className="pl-2 text-caption max-w-sm">{description}</p>
      </div>
      {onClick && (
        <Button variant="outline" onClick={onClick} className="text-caption">
          {buttonText}
        </Button>
      )}
    </div>
  )
}

export function HorizontalEmptyState({
  icon: Icon,
  title,
  description,
  onClick,
  buttonText,
  className
}: EmptyStateProps) {
  return (
    <div className={cn("flex items-center gap-4 justify-center text-center animate-in fade-in zoom-in duration-500", className)}>
      <div className="w-20 h-20 rounded-full bg-gradient-to-br from-white/10 to-transparent flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-black/20">
        <Icon className="h-8 w-8 text-text-caption" strokeWidth={1.5} />
      </div>
      <div className="flex flex-col items-start gap-2">
        <h5 className="text-h5">{title}</h5>
        <p className="pl-2 text-caption max-w-sm">{description}</p>
      </div>
      {onClick && (
        <Button variant="outline" onClick={onClick} className="text-caption">
          {buttonText}
        </Button>
      )}
    </div>

  )
}
