import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Assignment } from "@/types/models"
import { Button } from "../ui/button"

const statusColors = {
    "Not started": "bg-gray-500/20 text-gray-400",
    "In progress": "bg-blue-500/20 text-blue-400",
    "Done": "bg-green-500/20 text-green-400",
}

interface StatusTagProps {
    assignment: Assignment
    onEdit: (assignment: Assignment, column: string, value: string) => void
}

function StatusTag({ assignment, onEdit }: StatusTagProps) {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="p-0">
                    <Badge variant="outline" className={`text-xs ${statusColors[assignment.StatusName as keyof typeof statusColors]}`}>
                        {assignment.StatusName}
                    </Badge>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass border-gray-600">
                {Object.keys(statusColors).map((status) => (
                    <DropdownMenuItem key={status} onClick={() => onEdit(assignment, "status_name", status)}>
                        <Badge variant="outline" className={`text-xs ${statusColors[status as keyof typeof statusColors]}`}>
                            {status}
                        </Badge>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export { StatusTag }