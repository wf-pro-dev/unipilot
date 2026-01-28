import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { models } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { Flag } from "lucide-react"
import { memo, useEffect, useState } from "react"
import { GlassCard, GlassCardVariants } from "@/components/ui/glass-card"
import { cn } from "@/lib/utils"
import { useUpdateAssignment } from "@/hooks/use-assignments"




const priorities = [
    { value: "low", label: "Low" },
    { value: "medium", label: "Medium" },
    { value: "high", label: "High" },
]

interface PriorityTagProps {
    assignment: models.LocalAssignment
    variant?: GlassCardVariants
    className?: string
}



function BasePriorityTag({ assignment, variant = "default", className = "" }: PriorityTagProps) {
    
    const updateMutation = useUpdateAssignment()

    const [priority, setPriority] = useState(assignment.Priority)

    useEffect(() => {
        setPriority(assignment.Priority)
    }, [assignment.Priority])

    const handleEdit = (e: React.MouseEvent<HTMLDivElement>, priority: string) => {
        e.stopPropagation()
        updateMutation.mutate({ assignment, column: "priority", value: priority })
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant={variant == "default" ? "primary" : "outline"} size="tag" className={cn("text-caption font-normal", className)}>
                    {priority}
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="bg-transparent border-none">
                <GlassCard variant="board">
                    {priorities.map((priority) => (
                        <DropdownMenuItem key={priority.value} onClick={(e) => handleEdit(e, priority.value)}>
                            <Badge variant="outline" className={`text-caption text-white font-normal`}>
                                {priority.label}
                            </Badge>
                        </DropdownMenuItem>
                    ))}
                </GlassCard>
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export const PriorityTag = memo(BasePriorityTag, (prevProps, nextProps) => {
    return prevProps.assignment.Priority === nextProps.assignment.Priority
})