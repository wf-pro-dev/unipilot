import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { useState, memo, useEffect } from "react"


const statusColors = {
    "Not started": "bg-gray-500/20 text-gray-400",
    "In progress": "bg-blue-500/20 text-blue-400",
    "Done": "bg-green-500/20 text-green-400",
}

interface StatusTagProps {
    assignment: assignment.LocalAssignment
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}



function BaseStatusTag({ assignment, onEdit }: StatusTagProps) {
    const [status, setStatus] = useState(assignment.StatusName)


    useEffect(() => {
        setStatus(assignment.StatusName)
    }, [assignment.StatusName])

    const handleEdit = (e: React.MouseEvent<HTMLButtonElement>, status: string) => {
        e.stopPropagation()
        setStatus(status)
        onEdit(assignment, "status_name", status)
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="p-0 w-fit rounded-full">
                    <Badge variant="outline" className={`text-xs ${statusColors[status as keyof typeof statusColors]}`}>
                        {status}
                    </Badge>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass border-gray-600">
                {Object.keys(statusColors).map((status) => (
                    <DropdownMenuItem key={status} onClick={(e) => handleEdit(e, status)}>
                        <Badge variant="outline" className={`text-xs ${statusColors[status as keyof typeof statusColors]}`}>
                            {status}
                        </Badge>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export const StatusTag = memo(BaseStatusTag, (prevProps, nextProps) => {
    console.log("StatusTag memo", prevProps.assignment.StatusName, nextProps.assignment.StatusName)
    return prevProps.assignment.StatusName === nextProps.assignment.StatusName
})