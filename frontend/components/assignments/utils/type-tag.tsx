import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { assignment } from "@/wailsjs/go/models"
import { Button } from "@/components/ui/button"
import { BookOpen } from "lucide-react"


const typeColors = {
    HW: "text-blue-400 border-blue-400",
    Exam: "text-red-400 border-red-400",
    Lab: "text-green-400 border-green-400",
    "Group Project": "text-yellow-400 border-yellow-400",
    Quiz: "text-orange-400 border-orange-400",
}

interface TypeTagProps {
    assignment: assignment.LocalAssignment
    onEdit: (assignment: assignment.LocalAssignment, column: string, value: string) => void
}



function TypeTag({ assignment, onEdit }: TypeTagProps) {
    const handleEdit = (type: string) => {

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
                    <DropdownMenuItem key={type} onClick={() => handleEdit(type)}>
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