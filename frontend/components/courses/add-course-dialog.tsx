"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Form, FormControl, FormField, FormItem, FormLabel } from "@/components/ui/form"
import { cn } from "@/lib/utils"
import { Calendar } from "@/components/ui/calendar"
import { 
  CalendarIcon, Plus, BookOpen, MapPin, Clock, 
  User, Mail, Palette, ArrowRight, ArrowLeft, Check, Loader2 
} from "lucide-react"
import { format } from "date-fns"
import { motion, AnimatePresence } from "framer-motion"
import { toast } from "sonner"
import { models } from "@/wailsjs/go/models"
import { 
  CourseValues,
  courseSchema
} from "./shema"

interface AddCourseDialogProps {
  onAdd: (course: models.LocalCourse) => void
}

const colors = [
  { name: "Blue", value: "bg-blue-500" },
  { name: "Green", value: "bg-green-500" },
  { name: "Purple", value: "bg-purple-500" },
  { name: "Red", value: "bg-red-500" },
  { name: "Orange", value: "bg-orange-500" },
  { name: "Pink", value: "bg-pink-500" },
]

const semesters = [
  { name: "FALL 2024", value: "FALL 2024" },
  { name: "SPRING 2025", value: "SPRING 2025" },
  { name: "SUMMER 2025", value: "SUMMER 2025" },
  { name: "FALL 2025", value: "FALL 2025" },
  { name: "SPRING 2026", value: "SPRING 2026" },
  { name: "SUMMER 2026", value: "SUMMER 2026" },
  { name: "FALL 2026", value: "FALL 2026" },
  { name: "SPRING 2027", value: "SPRING 2027" },
  { name: "SUMMER 2027", value: "SUMMER 2027" },
  { name: "FALL 2027", value: "FALL 2027" },
  { name: "SPRING 2028", value: "SPRING 2028" },
  { name: "SUMMER 2028", value: "SUMMER 2028" },
]

export function AddCourseDialog({ onAdd }: AddCourseDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState(1)
  const [step1Attempted, setStep1Attempted] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<CourseValues>({
    resolver: zodResolver(courseSchema),
    mode: "onChange",
    defaultValues: {
      name: "",
      code: "",
      credits: "3",
      semester: "",
      schedule: "",
      location: "",
      startDate: undefined,
      endDate: undefined,
      instructor: "",
      instructor_email: "",
      color: "bg-blue-500",
    },
  })

  const handleNext = async () => {
    if (step === 1) {
      setStep1Attempted(true)
      const step1Valid = await form.trigger(["name", "code", "credits", "semester"])
      if (step1Valid) setStep(2)
    } else if (step === 2) {
      const step2Valid = await form.trigger(["schedule", "location", "startDate", "endDate"])
      if (step2Valid) setStep(3)
    }
  }

  const handleBack = () => {
    if (step > 1) setStep(step - 1)
  }

  const onSubmit = async (data: CourseValues) => {
    setIsSubmitting(true)
    try {
      const courseData: models.LocalCourse = {
        Name: data.name,
        Code: data.code,
        Color: data.color,
        Semester: data.semester,
        Schedule: data.schedule,
        Credits: parseInt(data.credits),
        Location: data.location,
        Instructor: data.instructor,
        InstructorEmail: data.instructor_email,
        StartDate: data.startDate,
        EndDate: data.endDate,
      } as models.LocalCourse

      onAdd(courseData)
      toast.success("Course added successfully")
      handleOpenChange(false)
    } catch (error) {
      toast.error("Failed to add course")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (!isOpen) {
      setStep(1)
      setStep1Attempted(false)
      form.reset()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button variant="default" className="text-body bg-white/10 hover:bg-white/20 border-white/20">
          <Plus className="mr-2 w-4 h-4" />
          Add Course
        </Button>
      </DialogTrigger>
      <DialogContent className="glass border-white/10 text-white max-w-lg p-0 overflow-hidden gap-0">
        <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
          <div className="flex items-center justify-between">
            <DialogTitle className="text-h3">Add New Course</DialogTitle>
            <div className="flex gap-1.5">
              <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 1 ? "bg-white" : "bg-white/20")} />
              <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 2 ? "bg-white" : "bg-white/20")} />
              <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 3 ? "bg-white" : "bg-white/20")} />
            </div>
          </div>
          <p className="text-sm text-gray-400 mt-1">
            {step === 1 && "Basic course information"}
            {step === 2 && "Schedule and location"}
            {step === 3 && "Instructor details"}
          </p>
        </DialogHeader>
        
        <div className="p-6 max-h-[70vh] overflow-y-auto">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)}>
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
                    <div className="grid grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="name"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Course Name
                            </FormLabel>
                            <div className="relative">
                              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                <BookOpen className="h-4 w-4" />
                              </div>
                              <FormControl>
                                <Input
                                  placeholder="Data Structures"
                                  className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                  {...field}
                                />
                              </FormControl>
                            </div>
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="code"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Course Code
                            </FormLabel>
                            <div className="relative">
                              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                <BookOpen className="h-4 w-4" />
                              </div>
                              <FormControl>
                                <Input
                                  placeholder="CS 101"
                                  className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                  {...field}
                                />
                              </FormControl>
                            </div>
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="credits"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Credits
                            </FormLabel>
                            <div className="relative">
                              <FormControl>
                                <Input
                                  type="number"
                                  placeholder="3"
                                  className="h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                  min={1}
                                  max={4}
                                  {...field}
                                />
                              </FormControl>
                            </div>
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="semester"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Semester
                            </FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger className="h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl">
                                  <SelectValue placeholder="Select semester" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                {semesters.map((sem) => (
                                  <SelectItem key={sem.value} value={sem.value} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                    {sem.name}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
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
                    <FormField
                      control={form.control}
                      name="schedule"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Schedule
                          </FormLabel>
                          <div className="relative">
                            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                              <Clock className="h-4 w-4" />
                            </div>
                            <FormControl>
                              <Input
                                placeholder="M, T, W 9:00 AM - 10:30 AM / Async"
                                className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                {...field}
                              />
                            </FormControl>
                          </div>
                          {fieldState.error && (
                            <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                          )}
                          <p className="text-xs text-gray-400 mt-1 ml-1">
                            Format: Days (M, T, W, Th, F, Sa, Su) followed by time
                          </p>
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={form.control}
                      name="location"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Location
                          </FormLabel>
                          <div className="relative">
                            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                              <MapPin className="h-4 w-4" />
                            </div>
                            <FormControl>
                              <Input
                                placeholder="Building 1, Room 101 / Online"
                                className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                {...field}
                              />
                            </FormControl>
                          </div>
                          {fieldState.error && (
                            <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                          )}
                        </FormItem>
                      )}
                    />

                    <div className="grid grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="startDate"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 ml-1">
                              Start Date
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
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="endDate"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 ml-1">
                              End Date
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
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />
                    </div>

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
                        type="button"
                        onClick={handleNext}
                        className="w-3/4 h-11 bg-white text-black hover:bg-gray-200 rounded-xl font-medium transition-all hover:scale-[1.01] active:scale-[0.99]"
                      >
                        Next Step <ArrowRight className="ml-2 h-4 w-4 opacity-50" />
                      </Button>
                    </div>
                  </motion.div>
                )}

                {step === 3 && (
                  <motion.div
                    key="step3"
                    initial={{ x: 20, opacity: 0 }}
                    animate={{ x: 0, opacity: 1 }}
                    exit={{ x: 20, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="space-y-6"
                  >
                    <div className="grid grid-cols-2 gap-4">
                      <FormField
                        control={form.control}
                        name="instructor"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Instructor
                            </FormLabel>
                            <div className="relative">
                              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                <User className="h-4 w-4" />
                              </div>
                              <FormControl>
                                <Input
                                  placeholder="Dr. Smith"
                                  className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                  {...field}
                                />
                              </FormControl>
                            </div>
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name="instructor_email"
                        render={({ field, fieldState }) => (
                          <FormItem className="space-y-1 group">
                            <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                              Email
                            </FormLabel>
                            <div className="relative">
                              <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                <Mail className="h-4 w-4" />
                              </div>
                              <FormControl>
                                <Input
                                  placeholder="smith@example.com"
                                  className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl"
                                  {...field}
                                />
                              </FormControl>
                            </div>
                            {fieldState.error && (
                              <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                            )}
                          </FormItem>
                        )}
                      />
                    </div>

                    <FormField
                      control={form.control}
                      name="color"
                      render={({ field, fieldState }) => (
                        <FormItem className="space-y-1 group">
                          <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">
                            Color
                          </FormLabel>
                          <div className="relative">
                            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                              <Palette className="h-4 w-4" />
                            </div>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 rounded-xl">
                                  <SelectValue placeholder="Select color" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                {colors.map((color) => (
                                  <SelectItem key={color.value} value={color.value} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                    <div className="flex items-center gap-2">
                                      <div className={`w-4 h-4 rounded-full ${color.value}`} />
                                      <span>{color.name}</span>
                                    </div>
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </div>
                          {fieldState.error && (
                            <p className="text-xs text-red-400 mt-1 ml-1">{fieldState.error.message}</p>
                          )}
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
                            Add Course <Check className="ml-2 h-4 w-4 opacity-50" />
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