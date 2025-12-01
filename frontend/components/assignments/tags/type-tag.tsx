import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { BookOpen } from "lucide-react"


const typeColors = {
    HW: "text-gray-400 border-gray-600 bg-gray-800/30",
    Exam: "text-red-400 border-red-400 bg-red-900/20",
    Lab: "text-gray-400 border-gray-600 bg-gray-800/30",
    "Group Project": "text-gray-400 border-gray-600 bg-gray-800/30",
    Quiz: "text-gray-400 border-gray-600 bg-gray-800/30",
}

interface TypeTagProps {
    assignment: assignment.LocalAssignment
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}



function TypeTag({ assignment, onEdit }: TypeTagProps) {
   
    const handleEdit = (e: React.MouseEvent<HTMLButtonElement>, type: string) => {
        e.stopPropagation()
        onEdit(assignment, "type_name", type)
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="p-0">
                    <Badge
                        variant="outline"
                        className={`text-xs flex flex-row gap-1 ${typeColors[assignment.TypeName as keyof typeof typeColors]}`}
                    >
                        <BookOpen className="h-3 w-3 " />
                        {assignment.TypeName}
                    </Badge>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="glass border-gray-600">
                {Object.keys(typeColors).map((type) => (
                    <DropdownMenuItem key={type} onClick={(e) => handleEdit(e, type)}>
                        <Badge variant="outline" className={`text-xs flex flex-row gap-1 ${typeColors[type as keyof typeof typeColors]}`}>
                            <BookOpen className="h-3 w-3" />
                            {type}
                        </Badge>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    )
}

export { TypeTag }