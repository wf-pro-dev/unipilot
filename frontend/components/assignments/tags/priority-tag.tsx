import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { Flag } from "lucide-react"


const priorityColors = {
    low: "text-green-400 border-green-400",
    medium: "text-yellow-400 border-yellow-400",
    high: "text-red-400 border-red-400",
}

interface PriorityTagProps {
    assignment: assignment.LocalAssignment
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}



function PriorityTag({ assignment, onEdit }: PriorityTagProps) {
    const handleEdit = (type: string) => {

        onEdit(assignment, "type_name", type)
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="p-0">
                    <Badge
                        variant="outline"
                        className={`text-xs flex flex-row gap-1 ${priorityColors[assignment.Priority as keyof typeof priorityColors]}`}
                    >
                        <Flag className="h-3 w-3 " />
                        {assignment.Priority}
                    </Badge>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass border-gray-600">
                {Object.keys(priorityColors).map((priority) => (
                    <DropdownMenuItem key={priority} onClick={() => handleEdit(priority)}>
                        <Badge variant="outline" className={`text-xs flex flex-row gap-1 ${priorityColors[priority as keyof typeof priorityColors]}`}>
                            <Flag className="h-3 w-3" />
                            {priority}
                        </Badge>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export { PriorityTag }