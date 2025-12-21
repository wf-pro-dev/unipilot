"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Loader2, Mail, User, Lock, GraduationCap, Globe, Calendar, BookOpen, UserPlus } from "lucide-react"
import { useRegister } from "@/hooks/use-auth"
import { motion } from "framer-motion"
import { toast } from "sonner"

interface RegisterFormProps {
    onRegisterSuccess?: () => void
}

// Common universities for the dropdown
const universities = [
    "Austin Community College",
    "Harvard University",
    "Stanford University",
    "MIT",
    "University of California, Berkeley",
    "University of Oxford",
    "University of Cambridge",
    "Yale University",
    "Princeton University",
    "Columbia University",
    "University of Chicago",
    "Other"
]

// Common languages
const languages = [
    { code: "en", name: "English" },
    { code: "es", name: "Spanish" },
    { code: "fr", name: "French" },
    { code: "de", name: "German" },
    { code: "it", name: "Italian" },
    { code: "pt", name: "Portuguese" },
    { code: "zh", name: "Chinese" },
    { code: "ja", name: "Japanese" },
    { code: "ko", name: "Korean" },
    { code: "ar", name: "Arabic" }
]

const semesters = [
    "Fall",
    "Spring",
    "Summer"
]

const currentYear = new Date().getFullYear()
const years = Array.from({ length: 5 }, (_, i) => (currentYear + i).toString())

export function RegisterForm({ onRegisterSuccess }: RegisterFormProps) {
    const [username, setUsername] = useState("")
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [confirmPassword, setConfirmPassword] = useState("")
    const [university, setUniversity] = useState("")
    const [language, setLanguage] = useState("en")
    const [semester, setSemester] = useState("")
    const [year, setYear] = useState(currentYear.toString())
   
    const { mutate: register, isPending: isLoading } = useRegister()

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        // Validation
        if (password !== confirmPassword) {
            toast.error("Passwords do not match")
            return
        }

        if (password.length < 6) {
            toast.error("Password must be at least 6 characters long")
            return
        }

        if (!email.includes("@")) {
            toast.error("Please enter a valid email address")
            return
        }

        if (!semester || !year) {
            toast.error("Please select a semester and year")
            return
        }

        register({ username, email, password, university, language, semester, year }, {
            onSuccess: () => {
                onRegisterSuccess?.()
            },
            onError: (error) => {
                toast.error(error.message)
            }
        })

    }

    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5, ease: "easeOut" }}
            className="w-full max-w-xl mx-auto"
        >
            <Card className="glass border-white/10 text-white overflow-hidden p-0 shadow-2xl">
                <CardHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <div className="flex items-center gap-3 mb-1">
                        <div className="p-2 bg-blue-500/10 rounded-lg">
                            <UserPlus className="w-5 h-5 text-blue-400" />
                        </div>
                        <CardTitle className="text-xl font-semibold text-white">
                            Create Account
                        </CardTitle>
                    </div>
                    <CardDescription className="text-gray-400 text-sm">
                        Join Unipilot to organize your academic life
                    </CardDescription>
                </CardHeader>
                <CardContent className="p-6 space-y-6">
                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                            <div className="space-y-2">
                                <Label htmlFor="username" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Username</Label>
                                <div className="relative group">
                                    <User className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 transition-colors group-focus-within:text-blue-400" />
                                    <Input
                                        id="username"
                                        type="text"
                                        placeholder="jdoe"
                                        value={username}
                                        onChange={(e) => setUsername(e.target.value)}
                                        className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10 text-white placeholder:text-gray-500"
                                        required
                                        disabled={isLoading}
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="email" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Email</Label>
                                <div className="relative group">
                                    <Mail className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 transition-colors group-focus-within:text-blue-400" />
                                    <Input
                                        id="email"
                                        type="email"
                                        placeholder="john@example.com"
                                        value={email}
                                        onChange={(e) => setEmail(e.target.value)}
                                        className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10 text-white placeholder:text-gray-500"
                                        required
                                        disabled={isLoading}
                                    />
                                </div>
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                            <div className="space-y-2">
                                <Label htmlFor="password" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Password</Label>
                                <div className="relative group">
                                    <Lock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 transition-colors group-focus-within:text-blue-400" />
                                    <Input
                                        id="password"
                                        type="password"
                                        placeholder="••••••••"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10 text-white placeholder:text-gray-500"
                                        required
                                        disabled={isLoading}
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="confirmPassword" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Confirm Password</Label>
                                <div className="relative group">
                                    <Lock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 transition-colors group-focus-within:text-blue-400" />
                                    <Input
                                        id="confirmPassword"
                                        type="password"
                                        placeholder="••••••••"
                                        value={confirmPassword}
                                        onChange={(e) => setConfirmPassword(e.target.value)}
                                        className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10 text-white placeholder:text-gray-500"
                                        required
                                        disabled={isLoading}
                                        />
                                </div>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="university" className="text-gray-400 text-xs font-medium uppercase tracking-wider">University</Label>
                            <div className="relative group">
                                <GraduationCap className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 z-10 transition-colors group-focus-within:text-blue-400" />
                                <Select value={university} onValueChange={setUniversity} required disabled={isLoading}>
                                    <SelectTrigger className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10 text-white transition-all">
                                        <SelectValue placeholder="Select your university" />
                                    </SelectTrigger>
                                    <SelectContent className="glass border-white/10 text-white max-h-[200px]">
                                        {universities.map((uni) => (
                                            <SelectItem key={uni} value={uni} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                                {uni}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                            <div className="space-y-2">
                                <Label htmlFor="semester" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Semester</Label>
                                <div className="relative group">
                                    <BookOpen className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 z-10 transition-colors group-focus-within:text-blue-400" />
                                    <Select value={semester} onValueChange={setSemester} required disabled={isLoading}>
                                        <SelectTrigger className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10 text-white transition-all">
                                            <SelectValue placeholder="Sem" />
                                        </SelectTrigger>
                                        <SelectContent className="glass border-white/10 text-white">
                                            {semesters.map((sem) => (
                                                <SelectItem key={sem} value={sem} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                                    {sem}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="year" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Year</Label>
                                <div className="relative group">
                                    <Calendar className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 z-10 transition-colors group-focus-within:text-blue-400" />
                                    <Select value={year} onValueChange={setYear} required disabled={isLoading}>
                                        <SelectTrigger className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10 text-white transition-all">
                                            <SelectValue placeholder="Year" />
                                        </SelectTrigger>
                                        <SelectContent className="glass border-white/10 text-white max-h-[200px]">
                                            {years.map((y) => (
                                                <SelectItem key={y} value={y} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                                    {y}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="language" className="text-gray-400 text-xs font-medium uppercase tracking-wider">Language</Label>
                                <div className="relative group">
                                    <Globe className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500 z-10 transition-colors group-focus-within:text-blue-400" />
                                    <Select value={language} onValueChange={setLanguage} disabled={isLoading}>
                                        <SelectTrigger className="pl-10 bg-white/5 border-white/10 focus:border-blue-500 focus:ring-blue-500/20 h-10 text-white transition-all">
                                            <SelectValue placeholder="Lang" />
                                        </SelectTrigger>
                                        <SelectContent className="glass border-white/10 text-white max-h-[200px]">
                                            {languages.map((lang) => (
                                                <SelectItem key={lang.code} value={lang.code} className="focus:bg-white/10 focus:text-white cursor-pointer">
                                                    {lang.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        </div>

                        <Button 
                            type="submit" 
                            className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-6 shadow-[0_0_15px_rgba(37,99,235,0.3)] transition-all duration-300 transform hover:scale-[1.01]" 
                            disabled={isLoading || !university || !semester || !year}
                        >
                            {isLoading ? (
                                <>
                                    <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                                    Creating Account...
                                </>
                            ) : (
                                "Create Account"
                            )}
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </motion.div>
    )
}
