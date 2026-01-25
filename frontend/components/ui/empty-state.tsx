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
      <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary-blue-500/30 to-transparent flex items-center justify-center ring-1 ring-primary-blue-400/50">
        <Icon className="h-6 w-6 text-primary-blue-400" strokeWidth={1.5} />
      </div>

      <div className="flex flex-col items-center gap-2">
        <h5 className="text-h5 font-medium">{title}</h5>
        <p className="pl-2 text-caption text-gray-400 max-w-sm">{description}</p>
      </div>
      {onClick && (
        <Button variant="outline" onClick={onClick} className="text-caption text-gray-400">
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
      <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary-blue-500/30 to-transparent flex items-center justify-center ring-1 ring-primary-blue-400/50">
        <Icon className="h-6 w-6 text-primary-blue-400" strokeWidth={1.5} />
      </div>
      <div className="flex flex-col items-start gap-2">
        <h5 className="text-h5 font-medium">{title}</h5>
        <p className="pl-2 text-caption text-gray-400 max-w-sm">{description}</p>
      </div>
      {onClick && (
        <Button variant="outline" onClick={onClick} className="text-caption text-gray-400">
          {buttonText}
        </Button>
      )}
    </div>

  )
}
