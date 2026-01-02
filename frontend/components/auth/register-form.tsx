"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import {
    Loader2, Mail, User, Lock, GraduationCap, Globe, Calendar, BookOpen,
    ArrowRight, ArrowLeft, Check, Sparkles
} from "lucide-react"
import { motion, AnimatePresence } from "framer-motion"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
} from "@/components/ui/form"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"

import { useRegister } from "@/hooks/use-auth"
import { registerSchema, RegisterValues } from "./schema"
import { cn } from "@/lib/utils"
import { FormErrorMessage } from "./form-error-message"
import { useFormErrorDisplay } from "./use-form-error-display"

interface RegisterFormProps {
    onRegisterSuccess?: () => void
    onLoginClick?: () => void
    className?: string
}

// Data constants
const universities = [
    "Austin Community College", "Harvard University", "Stanford University", "MIT",
    "University of California, Berkeley", "University of Oxford", "University of Cambridge",
    "Yale University", "Princeton University", "Columbia University", "University of Chicago", "Other"
]

const languages = [
    { code: "en", name: "English" }, { code: "es", name: "Spanish" },
    { code: "fr", name: "French" }, { code: "de", name: "German" },
    { code: "it", name: "Italian" }, { code: "pt", name: "Portuguese" },
    { code: "zh", name: "Chinese" }, { code: "ja", name: "Japanese" },
    { code: "ko", name: "Korean" }, { code: "ar", name: "Arabic" }
]

const semesters = ["Fall", "Spring", "Summer"]
const currentYear = new Date().getFullYear()
const years = Array.from({ length: 5 }, (_, i) => (currentYear + i).toString())

export function RegisterForm({ onRegisterSuccess, onLoginClick, className }: RegisterFormProps) {
    const [step, setStep] = useState(1)
    const [step1Attempted, setStep1Attempted] = useState(false)
    const { mutate: register, isPending: isLoading } = useRegister()

    const form = useForm<RegisterValues>({
        resolver: zodResolver(registerSchema),
        mode: "onChange",
        defaultValues: {
            username: "",
            email: "",
            password: "",
            confirmPassword: "",
            university: "",
            semester: "",
            year: currentYear.toString(),
            language: "en",
        },
        
    })

    // Centralized error display control
    useFormErrorDisplay(form, step, step1Attempted)

    const handleNext = async () => {
        setStep1Attempted(true)
        const step1Valid = await form.trigger(["username", "email", "password", "confirmPassword"])
        if (step1Valid) setStep(2)
    }

    const handleBack = () => setStep(1)

    const onSubmit = (data: RegisterValues) => {
        register(data, {
            onSuccess: () => onRegisterSuccess?.(),
            onError: (error) => toast.error(error.message),
        })
    }

    return (
        <div className={cn("space-y-2 w-full mx-auto flex flex-col items-center justify-center ", className)}>

            <div className="p-8 space-y-8 w-1/3 bg-white/5 border border-white/15 shadow-lg shadow-black/20 relative rounded-2xl overflow-hidden z-10">
                {/* Shine effect on hover */}
                <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl pointer-events-none" />

                <div className="flex items-center justify-between mb-2 z-10">
                    <div>
                        <h1 className="text-2xl font-semibold tracking-tight text-white">
                            Create Account
                        </h1>
                        <p className="text-sm text-gray-400 mt-1">
                            {step === 1 ? "Start your academic journey" : "Tell us about your studies"}
                        </p>
                    </div>

                    {/* Minimalist Step Indicator */}
                    <div className="flex gap-1.5">
                        <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 1 ? "bg-white" : "bg-white/20")} />
                        <div className={cn("h-1.5 w-8 rounded-full transition-all duration-300", step === 2 ? "bg-white" : "bg-white/20")} />
                    </div>
                </div>


                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="relative z-10">
                        <AnimatePresence mode="wait" >
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
                                        name="username"
                                        render={({ field, fieldState }) => (
                                            <FormItem className="space-y-1 group">
                                                <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">Username</FormLabel>
                                                <div className="relative">
                                                    <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                                        <User className="h-4 w-4" />
                                                    </div>
                                                    <FormControl>
                                                        <Input placeholder="jdoe" className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10  rounded-xl" {...field} />
                                                    </FormControl>
                                                </div>
                                                <FormErrorMessage
                                                    fieldState={fieldState}
                                                    formState={form.formState}
                                                    config={{ strategy: "onStepAttempt", stepAttempted: step1Attempted }}
                                                />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="email"
                                        render={({ field, fieldState }) => (
                                            <FormItem className="space-y-1 group">
                                                <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">Email</FormLabel>
                                                <div className="relative">
                                                    <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                                        <Mail className="h-4 w-4" />
                                                    </div>
                                                    <FormControl>
                                                        <Input placeholder="john@example.com" className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10  rounded-xl" {...field} />
                                                    </FormControl>
                                                </div>
                                                <FormErrorMessage
                                                    fieldState={fieldState}
                                                    formState={form.formState}
                                                    config={{ strategy: "onStepAttempt", stepAttempted: step1Attempted }}
                                                />
                                            </FormItem>
                                        )}
                                    />



                                    <FormField
                                        control={form.control}
                                        name="password"
                                        render={({ field, fieldState }) => (
                                            <FormItem className="space-y-1 group">
                                                <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">Password</FormLabel>
                                                <div className="relative">
                                                    <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                                        <Lock className="h-4 w-4" />
                                                    </div>
                                                    <FormControl>
                                                        <Input type="password" placeholder="••••••••" className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10  rounded-xl" {...field} />
                                                    </FormControl>
                                                </div>
                                                <FormErrorMessage
                                                    fieldState={fieldState}
                                                    formState={form.formState}
                                                    config={{ strategy: "onStepAttempt", stepAttempted: step1Attempted }}
                                                />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="confirmPassword"
                                        render={({ field, fieldState }) => (
                                            <FormItem className="space-y-1 group">
                                                <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">Confirm</FormLabel>
                                                <div className="relative">
                                                    <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                                                        <Lock className="h-4 w-4" />
                                                    </div>
                                                    <FormControl>
                                                        <Input type="password" placeholder="••••••••" className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10  rounded-xl" {...field} />
                                                    </FormControl>
                                                </div>
                                                <FormErrorMessage
                                                    fieldState={fieldState}
                                                    formState={form.formState}
                                                    config={{ strategy: "onStepAttempt", stepAttempted: step1Attempted }}
                                                />
                                            </FormItem>
                                        )}
                                    />


                                    <div>
                                        <Button
                                            type="button"
                                            onClick={handleNext}
                                            className="w-full h-11 bg-white text-black hover:bg-gray-200 rounded-xl font-medium transition-all hover:scale-[1.01] active:scale-[0.99]"
                                        >
                                            Next Step <ArrowRight className="ml-2 h-4 w-4 opacity-50" />
                                        </Button>
                                    </div>
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


                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                                        <FormField
                                            control={form.control}
                                            name="university"
                                            render={({ field, fieldState }) => (
                                                <FormItem className="space-y-1 group">
                                                    <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-hover:text-white ml-1 transition-colors duration-300">University</FormLabel>
                                                    <div className="relative group">
                                                        <GraduationCap className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 group-hover:text-white z-10 transition-colors duration-300" />
                                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                            <FormControl>
                                                                <SelectTrigger className="pl-10 h-11 bg-white/5 border-white/10 text-gray-500 focus:ring-0 rounded-xl">
                                                                    <SelectValue placeholder="University" />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                                                {universities.map((uni) => (
                                                                    <SelectItem key={uni} value={uni} className="focus:bg-white/10 focus:text-white cursor-pointer">{uni}</SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    </div>
                                                    <FormErrorMessage
                                                        fieldState={fieldState}
                                                        formState={form.formState}
                                                        config={{ strategy: "onTouched" }}
                                                    />
                                                </FormItem>
                                            )}
                                        />
                                        <FormField
                                            control={form.control}
                                            name="semester"
                                            render={({ field, fieldState }) => (
                                                <FormItem className="space-y-1 group">
                                                    <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-hover:text-white ml-1 transition-colors duration-300">Semester</FormLabel>
                                                    <div className="relative group">
                                                        <BookOpen className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 group-hover:text-white z-10 transition-all duration-300" />
                                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                            <FormControl>
                                                                <SelectTrigger className="pl-10 h-11 bg-white/5 border-white/10 text-gray-500 focus:ring-0 rounded-xl">
                                                                    <SelectValue placeholder="Semester" />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                                                {semesters.map((sem) => (
                                                                    <SelectItem key={sem} value={sem} className="focus:bg-white/10 focus:text-white cursor-pointer">{sem}</SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    </div>
                                                    <FormErrorMessage
                                                        fieldState={fieldState}
                                                        formState={form.formState}
                                                        config={{ strategy: "onTouched" }}
                                                    />
                                                </FormItem>
                                            )}
                                        />

                                        <FormField
                                            control={form.control}
                                            name="year"
                                            render={({ field, fieldState }) => (
                                                <FormItem className="space-y-1 group">
                                                    <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-hover:text-white ml-1 transition-colors duration-300">Year</FormLabel>
                                                    <div className="relative group">
                                                        <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 group-hover:text-white z-10 transition-all duration-300" />
                                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                            <FormControl>
                                                                <SelectTrigger className="pl-10 h-11 bg-white/5 border-white/10 text-white focus:ring-0 rounded-xl">
                                                                    <SelectValue placeholder="Year" />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                                                {years.map((y) => (
                                                                    <SelectItem key={y} value={y} className="focus:bg-white/10 focus:text-white cursor-pointer">{y}</SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    </div>
                                                    <FormErrorMessage
                                                        fieldState={fieldState}
                                                        formState={form.formState}
                                                        config={{ strategy: "onTouched" }}
                                                    />
                                                </FormItem>
                                            )}
                                        />

                                        <FormField
                                            control={form.control}
                                            name="language"
                                            render={({ field, fieldState }) => (
                                                <FormItem className="space-y-1 group">
                                                    <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-hover:text-white ml-1 transition-colors duration-300">Language</FormLabel>
                                                    <div className="relative group">
                                                        <Globe className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 group-hover:text-white z-10 transition-colors duration-300" />
                                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                            <FormControl>
                                                                <SelectTrigger className="pl-10 h-11 bg-white/5 border-white/10 text-white focus:ring-0 rounded-xl">
                                                                    <SelectValue placeholder="Language" />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectContent className="bg-black/90 border-white/10 text-white backdrop-blur-xl">
                                                                {languages.map((lang) => (
                                                                    <SelectItem key={lang.code} value={lang.code} className="focus:bg-white/10 focus:text-white cursor-pointer">{lang.name}</SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    </div>
                                                    <FormErrorMessage
                                                        fieldState={fieldState}
                                                        formState={form.formState}
                                                        config={{ strategy: "onTouched" }}
                                                    />
                                                </FormItem>
                                            )}
                                        />
                                    </div>

                                    <div className="flex gap-4 pt-4">
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
                                            disabled={isLoading}
                                        >
                                            {isLoading ? (
                                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            ) : (
                                                <>
                                                    Create Account <Check className="ml-2 h-4 w-4 opacity-50" />
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
            <div className="text-center">
                <p className="text-sm text-gray-500">
                    Already have an account?{" "}
                    <button
                        onClick={onLoginClick}
                        className="text-white hover:text-blue-400 font-medium transition-colors hover:underline underline-offset-4 decoration-blue-500/30"
                    >
                        Sign in
                    </button>
                </p>
            </div>
        </div>
    )
}
