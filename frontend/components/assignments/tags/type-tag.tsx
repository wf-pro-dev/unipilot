import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { models } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { memo, useEffect, useState } from "react"
import { GlassCard, GlassCardVariants } from "@/components/ui/glass-card"
import { cn } from "@/lib/utils"




const types = [
    { value: "HW", label: "HW" },
    { value: "Exam", label: "Exam" },
    { value: "Lab", label: "Lab" },
    { value: "Group Project", label: "Group Project" },
    { value: "Quiz", label: "Quiz" },
]

interface TypeTagProps {
    assignment: models.LocalAssignment
    onEdit: (assignment: models.LocalAssignment, column: string, value: string) => void
    variant?: GlassCardVariants
    className?: string
}



function BaseTypeTag({ assignment, onEdit, variant = "default", className = "" }: TypeTagProps) {

    const [type, setType] = useState(assignment.Type)

    useEffect(() => {
        setType(assignment.Type)
    }, [assignment.Type])

    const handleEdit = (e: React.MouseEvent<HTMLDivElement>, type: string) => {
        e.stopPropagation()
        setType(type)
        onEdit(assignment, "type", type)
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant={variant == "default" ? "primary" : "outline"} size="tag" className={cn("text-caption", className)}>
                    {type}
                </Button> 
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="bg-transparent border-none">
                <GlassCard variant="board">
                {types.map((type) => (
                    <DropdownMenuItem key={type.value} onClick={(e) => handleEdit(e, type.value)}>
                            <Badge variant="outline" className={`text-caption`}>
                                {type.label}
                            </Badge>
                    </DropdownMenuItem>
                ))}
                </GlassCard>
            </DropdownMenuContent>
        </DropdownMenu>
    )
}
export const TypeTag = memo(BaseTypeTag, (prevProps, nextProps) => {
    return prevProps.assignment.Type === nextProps.assignment.Type
})