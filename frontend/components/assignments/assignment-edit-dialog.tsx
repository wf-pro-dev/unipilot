"use client"

import type React from "react"

import { useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { format, isSameDay } from "date-fns"
import { ArrowLeft, ArrowRight, BookOpen, CalendarIcon, Check, ClipboardList, Flag, Link, Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"
import { models } from "@/wailsjs/go/models"
import { LogInfo } from "@/wailsjs/runtime/runtime"
import { Textarea } from "../ui/textarea"
import { CoursesSelect } from "../courses/courses-select"
import { useAssignment } from "@/hooks/use-assignments"
import { assignmentSchema, AssignmentValues } from "./schema"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { Form, FormField, FormItem, FormLabel, FormControl } from "../ui/form"
import { FormErrorMessage } from "../auth/form-error-message"
import { AnimatePresence } from "framer-motion"
import { motion } from "framer-motion"
import { useUpdateAssignment } from "@/hooks/use-assignments"


const types = [
  { value: "HW", label: "HW", icon: "📝", color: "text-blue-400" },
  { value: "Group project", label: "Group Project", icon: "👥", color: "text-yellow-400" },
  { value: "Exam", label: "Exam", icon: "📚", color: "text-red-400" },
  { value: "Quiz", label: "Quiz", icon: "❓", color: "text-orange-400" },
  { value: "Lab", label: "Lab", icon: "🔬", color: "text-green-400" },
]

const priorities = [
  { value: "low", label: "Low", color: "text-green-400" },
  { value: "medium", label: "Medium", color: "text-accent-amber-500" },
  { value: "high", label: "High", color: "text-red-400" },
]

const statuses = [
  { value: "Not started", label: "Not started", color: "text-gray-400" },
  { value: "In progress", label: "In progress", color: "text-blue-400" },
  { value: "Done", label: "Done", color: "text-green-400" },
]

interface AssignmentEditDialogProps {
  assignmentId: number
  isOpen: boolean
  onClose: () => void
}

export function AssignmentEditDialog({ assignmentId, isOpen, onClose }: AssignmentEditDialogProps) {

  const { data: assignment } = useAssignment(assignmentId)

  if (!assignment) { return null }

  const updateMutation = useUpdateAssignment()

  const [step, setStep] = useState(1)
  const [step1Attempted, setStep1Attempted] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [deadline, setDeadline] = useState<Date>(new Date(assignment.Deadline) || new Date())
  const [selectedCourse, setSelectedCourse] = useState<models.LocalCourse | undefined>({ Code: assignment.CourseCode, ID: assignment.CourseID } as models.LocalCourse)
  const form = useForm<AssignmentValues>({
    resolver: zodResolver(assignmentSchema),
    mode: "onChange",
    defaultValues: {
      title: assignment.Title,
      course_code: assignment.CourseCode,
      course_id: assignment.CourseID,
      remote_course_id: assignment.RemoteCourseID,
      type: assignment.Type,
      status: assignment.Status,
      deadline: new Date(assignment.Deadline),
      priority: assignment.Priority,
      todo: assignment.Todo,
      link: assignment.Link,
    },
  })

  const key_to_column = {
    title: "Title",
    course_code: "CourseCode",
    type: "Type",
    status: "Status",
    priority: "Priority",
    todo: "Todo",
    link: "Link",
  }

  const handleNext = async () => {
    if (step === 1) {
      setStep1Attempted(true)
      const step1Valid = await form.trigger(["title", "course_code", "course_id", "remote_course_id", "type", "deadline"])
      if (step1Valid) setStep(2)
    }
  }

  const handleBack = () => {
    if (step > 1) setStep(step - 1)
  }
  const handleCourseChange = (course: models.LocalCourse | undefined) => {
    if (!course) return
    setSelectedCourse(course)
    form.setValue("course_id", course.ID)
    form.setValue("remote_course_id", course.RemoteID)
  }


  const onSubmit = async (data: AssignmentValues) => {
    setIsSubmitting(true)
    onClose()


    for (const [key, value] of Object.entries(data)) {

      var column = key_to_column[key as keyof typeof key_to_column] as keyof models.LocalAssignment
      if (!column) { continue }

      if (value !== assignment[column]) {

        const message = `Changes to ${column} value: ${value}`
        LogInfo(message)
        updateMutation.mutate({ assignment, column: key, value: value as string })

        // If the course code changes, update the course id and remote course id
        if (column === "CourseCode") {
          updateMutation.mutate({ assignment, column: "CourseID", value: data.course_id.toString() })
          if (data.remote_course_id) {
            updateMutation.mutate({ assignment, column: "RemoteCourseID", value: data.remote_course_id.toString() })
          }
        }
      }
      else {
        const message = `No changes to ${column}`
        LogInfo(message)
      }
    }

    if (!isSameDay(form.getValues("deadline"), new Date(assignment.Deadline))) {
      updateMutation.mutate({ assignment, column: "deadline", value: format(form.getValues("deadline"), "yyyy-MM-dd HH:mm:ssxxx") })
    }
    setDeadline(new Date())


    setIsSubmitting(false)
  }


  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="glass border-white/10 text-white max-w-md p-0 overflow-hidden gap-0">
        <DialogHeader className="px-6 pt-8 pb-4 border-b border-white/5 bg-white/5">
          <div className="flex items-center justify-between">
            <DialogTitle className="text-h3">Edit Assignment</DialogTitle>
            <div className="flex gap-1.5">
              <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 1 ? "bg-white" : "bg-white/20")} />
              <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 2 ? "bg-white" : "bg-white/20")} />
            </div>
          </div>
          <p className="text-sm text-gray-400 mt-1">
            {step === 1 ? "Basic assignment details" : "Additional information"}
          </p>
        </DialogHeader>

        <div className="p-6 max-h-[70vh] overflow-y-auto">
          <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl pointer-events-none" />

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="relative z-10">
              <AnimatePresence mode="wait">
                {step === 1 && (
                  <motion.div
                    key="step1"
                    initial={{ x: -20, opacity: 0 }}
                    animate={{ x: 0, opacity: 1 }}
                    exit={{ x: -20, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="space-y-6"
                  >
                    <FormField
                      control={form.control}
                      name="title"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Assignment Title
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
                            config={{ strategy: "onSubmit" }}
                          />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form.control}
                      name="course_code"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Course
                          </FormLabel>
                          <CoursesSelect
                            value={field.value}
                            onValueChange={field.onChange}
                            onCourseChange={handleCourseChange}
                            selectedCourse={selectedCourse}
                          />
                          <FormErrorMessage
                            fieldState={fieldState}
                            formState={form.formState}
                            config={{ strategy: "onSubmit" }}
                          />
                        </FormItem>
                      )}
                    />

                    <div className="grid grid-cols-2 gap-4">


                      <FormField
                        control={form.control}
                        name="type"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Type
                            </FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger className="h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl">
                                  <SelectValue placeholder="Select type" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                {types.map((type) => (
                                  <SelectItem key={type.value} value={type.value} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                    <div className="flex items-center gap-2">
                                      <span className={type.color}>{type.icon}</span>
                                      <span>{type.label}</span>
                                    </div>
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            <FormErrorMessage
                              fieldState={fieldState}
                              formState={form.formState}
                              config={{ strategy: "onSubmit" }}
                            />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="deadline"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 ml-1">
                              Due Date
                            </FormLabel>
                            <Popover>
                              <PopoverTrigger asChild>
                                <Button
                                  variant="outline"
                                  className={cn(
                                    "w-full justify-start text-left font-normal h-11 bg-white/5 border-white/10 hover:bg-white/10 text-white rounded-xl",
                                    !field.value && "text-gray-400"
                                  )}
                                >
                                  <CalendarIcon className="mr-2 h-4 w-4" />
                                  {field.value ? format(field.value, "MMM do, yyyy") : "Pick a date"}
                                </Button>
                              </PopoverTrigger>
                              <PopoverContent className="w-auto p-0 border-white/10 bg-black/90 backdrop-blur-xl">
                                <Calendar
                                  mode="single"
                                  selected={field.value}
                                  onSelect={field.onChange}
                                  initialFocus
                                  className="text-white"
                                />
                              </PopoverContent>
                            </Popover>
                            <FormErrorMessage
                              fieldState={fieldState}
                              formState={form.formState}
                              config={{ strategy: "onTouched" }}
                            />
                          </FormItem>
                        )}
                      />

                    </div>

                    <Button
                      type="button"
                      onClick={handleNext}
                      className="w-full h-11 bg-white text-black hover:bg-gray-200 rounded-xl font-medium transition-all hover:scale-[1.01] active:scale-[0.99]"
                    >
                      Next Step <ArrowRight className="ml-2 h-4 w-4 opacity-50" />
                    </Button>
                  </motion.div>
                )}

                {step === 2 && (
                  <motion.div
                    key="step2"
                    initial={{ x: 20, opacity: 0 }}
                    animate={{ x: 0, opacity: 1 }}
                    exit={{ x: 20, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="space-y-6"
                  >
                    <div className="grid grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="priority"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Priority
                            </FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger className="h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl">
                                  <SelectValue placeholder="Select priority" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                {priorities.map((priority) => (
                                  <SelectItem key={priority.value} value={priority.value} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                    <div className="flex items-center gap-2">
                                      <span className={priority.color}>●</span>
                                      <span>{priority.label}</span>
                                    </div>
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            <FormErrorMessage
                              fieldState={fieldState}
                              formState={form.formState}
                              config={{ strategy: "onSubmit" }}
                            />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="status"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Status
                            </FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger className="h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl">
                                  <SelectValue placeholder="Select status" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                {statuses.map((status) => (
                                  <SelectItem key={status.value} value={status.value} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                    <div className="flex items-center gap-2">
                                      <span className={status.color}>●</span>
                                      <span>{status.label}</span>
                                    </div>
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            <FormErrorMessage
                              fieldState={fieldState}
                              formState={form.formState}
                              config={{ strategy: "onSubmit" }}
                            />
                          </FormItem>
                        )}
                      />
                    </div>

                    <FormField
                      control={form.control}
                      name="link"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Link (Optional)
                          </FormLabel>
                          <div className="relative">
                            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                              <Link className="h-4 w-4" />
                            </div>
                            <FormControl>
                              <Input
                                placeholder="https://canvas.university.edu/..."
                                className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl font-mono text-xs"
                                {...field}
                              />
                            </FormControl>
                          </div>
                          <FormErrorMessage
                            fieldState={fieldState}
                            formState={form.formState}
                            config={{ strategy: "onSubmit" }}
                          />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form.control}
                      name="todo"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Notes & Todos (Optional)
                          </FormLabel>
                          <div className="relative">
                            <div className="absolute left-3 top-3 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                              <ClipboardList className="h-4 w-4" />
                            </div>
                            <FormControl>
                              <Textarea
                                placeholder="Add specific details, requirements, or sub-tasks..."
                                className="pl-10 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl min-h-[100px] resize-none"
                                {...field}
                              />
                            </FormControl>
                          </div>
                          <FormErrorMessage
                            fieldState={fieldState}
                            formState={form.formState}
                            config={{ strategy: "onSubmit" }}
                          />
                        </FormItem>
                      )}
                    />

                    <div className="flex gap-4 pt-2">
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={handleBack}
                        className="w-1/4 h-11 text-gray-400 hover:text-white hover:bg-white/5 rounded-xl transition-colors"
                      >
                        <ArrowLeft className="mr-2 h-4 w-4" /> Back
                      </Button>
                      <Button
                        type="submit"
                        className="w-3/4 h-11 bg-white text-black hover:bg-gray-200 rounded-xl font-medium transition-all hover:scale-[1.01] active:scale-[0.99]"
                        disabled={isSubmitting}
                      >
                        {isSubmitting ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <>
                            Save Changes <Check className="ml-2 h-4 w-4 opacity-50" />
                          </>
                        )}
                      </Button>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </form>
          </Form>
        </div>
      </DialogContent>
    </Dialog>
  )
}
