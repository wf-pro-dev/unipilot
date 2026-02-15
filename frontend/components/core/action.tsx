import { LucideIcon } from "lucide-react"
import { GlassCard } from "../ui/glass-card"
import { CardContent } from "../ui/card"
import { cn } from "@/lib/utils"

interface ActionProps {
    label: string
    Icon: LucideIcon
    variant?: "default" | "outline"
    onClick: () => void
    className?: string
}

export function Action({
    label,
    Icon,
    variant = "outline",
    onClick,
    className = "",
}: ActionProps) {
    return (
        <GlassCard
            variant={variant}
            className={cn("flex flex-col items-start p-4 cursor-pointer gap-2", className)}
            onClick={onClick}
        >

            <div className="p-2 rounded-lg bg-white/5 border border-white/5">
                <Icon className="w-4 h-4 text-text-caption" />
            </div>

            <p className="text-sm font-medium uppercase tracking-wider">{label}</p>

        </GlassCard>
    )
}