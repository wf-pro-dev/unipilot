import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { GlassCard, GlassCardVariants } from "@/components/ui/glass-card"
import { cn } from "@/lib/utils"
import { models } from "@/wailsjs/go/models"
import { useState, memo, useEffect } from "react"



const statuses = [
    { value: "Not started", label: "Not started" },
    { value: "In progress", label: "In progress" },
    { value: "Done", label: "Done" },
]

interface StatusTagProps {
    assignment: models.LocalAssignment
    onEdit: (assignment: models.LocalAssignment, column: string, value: string) => void
    variant?: GlassCardVariants
    className?: string
}



function BaseStatusTag({ assignment, onEdit, variant = "default", className = "" }: StatusTagProps) {
    const [status, setStatus] = useState(assignment.Status)


    useEffect(() => {
        setStatus(assignment.Status)
    }, [assignment.Status])

    const handleEdit = (e: React.MouseEvent<HTMLDivElement>, status: string) => {
        e.stopPropagation()
        setStatus(status)
        onEdit(assignment, "status", status)
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant={variant == "default" ? "primary" : "outline"} size="tag" className={cn("text-caption font-normal", className)}>
                    {status}
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="bg-transparent border-none">
                <GlassCard variant="board">
                    {statuses.map((status) => (
                        <DropdownMenuItem key={status.value} onClick={(e) => handleEdit(e, status.value)}>
                            <Badge variant="outline" className={`text-caption text-white font-normal`}>
                                {status.label}
                            </Badge>
                        </DropdownMenuItem>
                    ))}
                </GlassCard>
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export const StatusTag = memo(BaseStatusTag, (prevProps, nextProps) => {
    return prevProps.assignment.Status === nextProps.assignment.Status
})