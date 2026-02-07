"use client"

import type React from "react"

import { useEffect, useState, useRef } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { ArrowLeft, BookOpen, Check, Loader2, Plus } from "lucide-react"

import { models } from "@/wailsjs/go/models"
import { CoursesSelect } from "../courses/courses-select"
import { useStreamNote } from "@/hooks/use-stream-notes"
import { NoteStreamModal } from "./note-stream-modal"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { Form } from "@/components/ui/form"
import { noteSchema, NoteValues } from "./schema"
import { FormControl, FormItem, FormLabel, FormField } from "../ui/form"
import { FormErrorMessage } from "../auth/form-error-message"
import { GlassCard } from "../ui/glass-card"
import { toast } from "sonner"

const subjects = [
    { value: "Mathematics", label: "Mathematics", color: "text-blue-400 border-blue-400" },
    { value: "Science", label: "Science", color: "text-yellow-400 border-yellow-400" },
    { value: "History", label: "History", color: "text-red-400 border-red-400" },
    { value: "English", label: "English", color: "text-orange-400 border-orange-400" },
    { value: "Computer Science", label: "Computer Science", color: "text-green-400 border-green-400" },
    { value: "Physics", label: "Physics", color: "text-purple-400 border-purple-400" },
    { value: "Chemistry", label: "Chemistry", color: "text-pink-400 border-pink-400" },
    { value: "Biology", label: "Biology", color: "text-gray-400 border-gray-400" },
    { value: "Social Studies", label: "Social Studies", color: "text-brown-400 border-brown-400" },
    { value: "Art", label: "Art", color: "text-indigo-400 border-indigo-400" },
    { value: "Music", label: "Music", color: "text-teal-400 border-teal-400" },
]

interface NoteAddDialogProps {
    isOpen: boolean
    onClose: () => void
}

export function NoteAddDialog({ isOpen, onClose }: NoteAddDialogProps) {

    const [showStreamModal, setShowStreamModal] = useState(false)
    const [selectedCourse, setSelectedCourse] = useState<models.LocalCourse | undefined>(undefined)
    const [isSubmitting, setIsSubmitting] = useState(false)
    // Store note data separately so it persists after form reset
    const [streamNoteData, setStreamNoteData] = useState<models.LocalNote | null>(null)

    const {
        content,
        isStreaming,
        error,
        startStream,
        stopStream,
        reset
    } = useStreamNote()

    const form = useForm<NoteValues>({
        resolver: zodResolver(noteSchema),
        mode: "onChange",
        defaultValues: {
            title: "",
            subject: "",
            course_id: "",
        },
    })


    const handleCloseStreamModal = () => {
        // Stop streaming if still active
        if (isStreaming) {
            stopStream()
        }
        // Reset state
        reset()
        // Close modal
        setShowStreamModal(false)
        // Clear stream note data
        setStreamNoteData(null)
        // Reset form
        form.reset()
        setSelectedCourse(undefined)
    }

    const onSubmit = async (data: NoteValues) => {
        setIsSubmitting(true)
        try {
            const noteData: models.LocalNote = {
                Title: data.title,
                Subject: data.subject,
                CourseID: data.course_id,
            } as unknown as models.LocalNote
            setStreamNoteData(noteData)
            // Close form dialog
            onClose()

            // Reset streaming state and show stream modal
            reset()
            setShowStreamModal(true)

            // Start streaming
            setIsSubmitting(false)
            await startStream(noteData)

        } catch (error) {
            console.log("error", error)
            toast.error("Failed to add note")
        }
    }

    const handleOpenChange = (open: boolean) => {

        if (!open) {
            form.reset()
        }
        onClose()
    }

    const handleErrorResolved = () => {
        setTimeout(() => {
            setIsSubmitting(false)
        }, 200)

    }

    useEffect(() => {
        console.log("isSubmitting", isSubmitting)
    }, [isSubmitting])

    return (
        <Dialog open={isOpen} onOpenChange={handleOpenChange} modal={false}>

            <DialogContent className="glass border-white/10 text-white max-w-md p-0 overflow-hidden gap-0">
                <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <DialogTitle className="text-h3">Add Note</DialogTitle>
                </DialogHeader>
                <div className="p-6 max-h-[70vh] overflow-y-auto">
                    <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl pointer-events-none" />

                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="relative z-10">
                            <div className="space-y-6">

                                <div className="grid grid-cols-2 gap-4">
                                    <FormField
                                        control={form.control}
                                        name="subject"
                                        render={({ field, fieldState }) => (
                                            <FormItem>
                                                <FormLabel className="text-caption font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                                                    Subject
                                                </FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl>
                                                        <SelectTrigger className="h-10 bg-white/5 border-white/10 text-gray-400 focus:bg-white/10 rounded-xl">
                                                            <SelectValue placeholder="Select subject" />
                                                        </SelectTrigger>
                                                    </FormControl>

                                                    <SelectContent className="bg-transparent border-none">
                                                        <GlassCard variant="board">
                                                            {subjects.map((subject) => (
                                                                <SelectItem key={subject.value} value={subject.value}>
                                                                    <span className="text-body text-white">{subject.label}</span>
                                                                </SelectItem>
                                                            ))}
                                                        </GlassCard>
                                                    </SelectContent>
                                                </Select>
                                                <FormErrorMessage
                                                    fieldState={fieldState}
                                                    formState={form.formState}
                                                    config={{ strategy: "onSubmit", submitAttempted: isSubmitting }}
                                                    onErrorResolved={handleErrorResolved}
                                                />
                                            </FormItem>
                                        )}
                                    />

                                    <FormField
                                        control={form.control}
                                        name="course_id"
                                        render={({ field, fieldState }) => (
                                            <FormItem>
                                                <FormLabel className="text-caption font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                                                    Course
                                                </FormLabel>
                                                <CoursesSelect
                                                    value={field.value}
                                                    onValueChange={field.onChange}
                                                />
                                                <FormErrorMessage
                                                    fieldState={fieldState}
                                                    formState={form.formState}
                                                    config={{ strategy: "onSubmit", submitAttempted: isSubmitting }}
                                                    onErrorResolved={handleErrorResolved}
                                                />
                                            </FormItem>
                                        )}
                                    />
                                </div>

                                <FormField
                                    control={form.control}
                                    name="title"
                                    render={({ field, fieldState }) => (
                                        <FormItem className="space-y-1 group">
                                            <FormLabel className="text-caption font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                                                Title
                                            </FormLabel>
                                            <div className="relative">
                                                <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                                    <BookOpen className="h-4 w-4" />
                                                </div>
                                                <FormControl>
                                                    <Input
                                                        placeholder="e.g. Calculus Midterm, History Essay..."
                                                        className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                                        {...field}
                                                    />
                                                </FormControl>
                                            </div>
                                            <FormErrorMessage
                                                fieldState={fieldState}
                                                formState={form.formState}
                                                config={{ strategy: "onSubmit", submitAttempted: isSubmitting }}
                                                onErrorResolved={handleErrorResolved}
                                            />
                                        </FormItem>
                                    )}
                                />

                                <div className="flex gap-4">
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        onClick={() => handleOpenChange(false)}
                                        className="w-1/4 h-11 text-gray-400 hover:text-white hover:bg-white/5 rounded-xl transition-colors"
                                    >
                                        <ArrowLeft className="h-4 w-4" /> Cancel
                                    </Button>
                                    <Button
                                        type="submit"
                                        variant="default"
                                        disabled={isSubmitting}
                                        className="w-3/4"
                                    >
                                        {isSubmitting ? (
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        ) : (
                                            <>
                                                Generate Note <Check className="h-4 w-4" strokeWidth={1.5} />
                                            </>
                                        )}
                                    </Button>
                                </div>


                            </div>
                        </form>
                    </Form>

                </div>



            </DialogContent>
            {/* Stream Modal - stays open after streaming completes */}
            {showStreamModal && streamNoteData && (
                <NoteStreamModal
                    isOpen={showStreamModal}
                    onClose={handleCloseStreamModal}
                    onStop={stopStream}
                    note={streamNoteData}
                    content={content}
                    isStreaming={isStreaming}
                    error={error}
                />
            )}

        </Dialog>
    )
}