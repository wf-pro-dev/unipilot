import { createContext, ReactNode, useContext, useState } from "react"
import { AssignmentDetailsDialog } from "../assignments/assignment-details-dialog"
import { AssignmentEditDialog } from "../assignments/assignment-edit-dialog"
import { CourseEditDialog } from "../courses/course-edit-dialog"
import { CourseDetailsDialog  } from "../courses/course-details-dialog"
import { CourseDeleteDialog } from "@/app/courses/course-delete-dialog"
import { AssignmentAddDialog } from "../assignments/assignment-add-dialog"
import { CourseAddDialog } from "../courses/course-add-dialog"

interface DialogProviderProps {
    children: ReactNode
}

type modelTypes = "assignment" | "course" | "note" | "user"
type dialogTypes = "details" | "edit" | "add" | "delete" | "linkRequest"

interface DialogProps {
    modelType: modelTypes
    dialogType: dialogTypes
    id?: number
    open?: boolean
}
interface DialogContextType { 
    SetDialogState: (props: DialogProps) => void
}

const DialogContext = createContext<DialogContextType | undefined>(undefined)

export function DialogProvider({ children }: DialogProviderProps) {
    const [assignmentDetailsID, setAssignmentDetailsID] = useState<number | undefined>(undefined)
    const [assignmentEditID, setAssignmentEditID] = useState<number | undefined>(undefined)
    const [assignmentAdd, setAssignmentAdd] = useState<boolean>(false)

    const [courseDetailsID, setCourseDetailsID] = useState<number | undefined>(undefined)
    const [courseEditID, setCourseEditID] = useState<number | undefined>(undefined)
    const [courseAdd, setCourseAdd] = useState<boolean>(false)
    const [courseDeleteID, setCourseDeleteID] = useState<number | undefined>(undefined)
    const [courseLinkRequestID, setCourseLinkRequestID] = useState<number | undefined>(undefined)

    const [noteAdd, setNoteAdd] = useState<boolean>(false)
    const [noteDetailsID, setNoteDetailsID] = useState<number | undefined>(undefined)
    const [noteEditID, setNoteEditID] = useState<number | undefined>(undefined)
    
    const [userDetailsID, setUserDetailsID] = useState<number | undefined>(undefined)


    function  SetDialogState ({
        modelType,
        dialogType,
        id,
        open = true,
    }: {
        modelType: modelTypes
        dialogType: dialogTypes
        id?: number
        open?: boolean
    }) {
        switch (modelType) {
            case "assignment":
                switch (dialogType) {
                    case "details":
                        open ? setAssignmentDetailsID(id) : setAssignmentDetailsID(undefined)
                        break
                    case "edit":
                        open ? setAssignmentEditID(id) : setAssignmentEditID(undefined)
                        break
                    case "add":
                        open ? setAssignmentAdd(true) : setAssignmentAdd(false)
                        break
                }
                break
            case "course":
                switch (dialogType) {
                    case "details":
                        open ? setCourseDetailsID(id) : setCourseDetailsID(undefined)
                        break
                    case "edit":
                        open ? setCourseEditID(id) : setCourseEditID(undefined)
                        break
                    case "delete":
                        open ? setCourseDeleteID(id) : setCourseDeleteID(undefined)
                        break
                    case "linkRequest":
                        open ? setCourseLinkRequestID(id) : setCourseLinkRequestID(undefined)
                        break
                    case "add":
                        open ? setCourseAdd(true) : setCourseAdd(false)
                        break
                }
                break
            case "note":
                switch (dialogType) {
                    case "details":
                        open ? setNoteDetailsID(id) : setNoteDetailsID(undefined)
                        break
                    case "edit":
                        open ? setNoteEditID(id) : setNoteEditID(undefined)
                        break
                }
                break
            case "user":
                switch (dialogType) {
                    case "details":
                        open ? setUserDetailsID(id) : setUserDetailsID(undefined)
                        break
                }
                break
            default:
                break
        }
    }

    
    return (
        <DialogContext.Provider value={{ SetDialogState }}>
            {children}

            <AssignmentDetailsDialog
                key={assignmentDetailsID}
                assignmentId={assignmentDetailsID!}
                isOpen={assignmentDetailsID !== undefined}
                onClose={() => setAssignmentDetailsID(undefined)}
                handleEditOpen={() => setAssignmentEditID(assignmentDetailsID)}
            />
            <AssignmentEditDialog
                key={assignmentEditID}
                assignmentId={assignmentEditID!}
                isOpen={assignmentEditID !== undefined}
                onClose={() => setAssignmentEditID(undefined)}
            />

            <AssignmentAddDialog
                isOpen={assignmentAdd}
                onClose={() => setAssignmentAdd(false)}
            />

            <CourseDetailsDialog
                key={courseDetailsID}
                courseId={courseDetailsID!}
                isOpen={courseDetailsID !== undefined}
                onClose={() => setCourseDetailsID(undefined)}
                onEdit={() => setCourseEditID(courseDetailsID)}
                onDelete={() => setCourseDeleteID(courseDetailsID)}
                onLinkRequest={() => setCourseLinkRequestID(courseDetailsID)}
            />

            <CourseEditDialog
                key={courseEditID}
                courseId={courseEditID!}
                isOpen={courseEditID !== undefined}
                onClose={() => setCourseEditID(undefined)}

            />

            <CourseAddDialog
                isOpen={courseAdd}
                onClose={() => setCourseAdd(false)}
            />
           

            <CourseDeleteDialog
                key={courseDeleteID}
                courseId={courseDeleteID!}
                isOpen={courseDeleteID !== undefined}
                onClose={() => setCourseDeleteID(undefined)}
                onDelete={() => setCourseDeleteID(courseDeleteID)}
            />

        </DialogContext.Provider>
    )
}

export function useDialogContext() {
    const context = useContext(DialogContext)
    if (context === undefined) {
        throw new Error("useDialogContext must be used within a DialogProvider")
    }
    return context
}