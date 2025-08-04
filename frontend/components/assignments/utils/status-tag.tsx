import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"


const statusColors = {
    "Not started": "bg-gray-500/20 text-gray-400",
    "In progress": "bg-blue-500/20 text-blue-400",
    "Done": "bg-green-500/20 text-green-400",
}

interface StatusTagProps {
    assignment: assignment.LocalAssignment
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}



function StatusTag({ assignment, onEdit }: StatusTagProps) {
    const handleEdit = (status: string) => {
    
        onEdit(assignment, "status_name", status)
    }

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
                    <DropdownMenuItem key={status} onClick={() => handleEdit(status)}>
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