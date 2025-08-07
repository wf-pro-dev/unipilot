"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Loader2, Mail, User, Lock, GraduationCap, Globe } from "lucide-react"
import { useAuth } from "@/hooks/use-auth"
import { motion } from "framer-motion"

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

export function RegisterForm({ onRegisterSuccess }: RegisterFormProps) {
    const [username, setUsername] = useState("")
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [confirmPassword, setConfirmPassword] = useState("")
    const [university, setUniversity] = useState("")
    const [language, setLanguage] = useState("en")
    const [error, setError] = useState("")
    const { register, isLoading } = useAuth()

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError("")

        // Validation
        if (password !== confirmPassword) {
            setError("Passwords do not match")
            return
        }

        if (password.length < 6) {
            setError("Password must be at least 6 characters long")
            return
        }

        if (!email.includes("@")) {
            setError("Please enter a valid email address")
            return
        }

        const result = await register(username, email, password, university, language)

        if (result.success) {
            onRegisterSuccess?.()
        } else {
            setError(result.error || "Registration failed")
        }
    }

    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.93  }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, ease: "easeInOut" }}
            className="w-full"
        >
            <Card className="glass w-full min-w-xl">
                <CardHeader className="space-y-1">
                    <CardTitle className="text-2xl font-bold text-center">Join Unipilot</CardTitle>
                    <CardDescription className="text-center">
                        Create your account to get started
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="username">Username</Label>
                            <div className="relative">
                                <User className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                                <Input
                                    id="username"
                                    type="text"
                                    placeholder="Enter your username"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    className="pl-10 glass"
                                    required
                                    disabled={isLoading}
                                />
                            </div>
                        </div>


                        <div className="space-y-2">
                            <Label htmlFor="email">Email</Label>
                            <div className="relative">
                                <Mail className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                                <Input
                                    id="email"
                                    type="email"
                                    placeholder="Enter your email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    className="pl-10 glass"
                                    required
                                    disabled={isLoading}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="password">Password</Label>
                            <div className="relative">
                                <Lock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                                <Input
                                    id="password"
                                    type="password"
                                    placeholder="Enter your password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="pl-10 glass"
                                    required
                                    disabled={isLoading}
                                />
                            </div>
                        </div>


                        <div className="space-y-2">
                            <Label htmlFor="confirmPassword">Confirm Password</Label>
                            <div className="relative">
                                <Lock className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                                <Input
                                    id="confirmPassword"
                                    type="password"
                                    placeholder="Confirm your password"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    className="pl-10 glass"
                                    required
                                    disabled={isLoading}
                                    />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">

                            <div className="space-y-2">
                                <Label htmlFor="university">University</Label>
                                <div className="relative">
                                    <GraduationCap className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground z-10" />
                                    <Select value={university} onValueChange={setUniversity} required disabled={isLoading}>
                                        <SelectTrigger className="pl-10 glass">
                                            <SelectValue placeholder="Select your university" />
                                        </SelectTrigger>
                                        <SelectContent className="glass">
                                            {universities.map((uni) => (
                                                <SelectItem key={uni} value={uni}>
                                                    {uni}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="language">Preferred Language</Label>
                                <div className="relative">
                                    <Globe className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground z-10" />
                                    <Select value={language} onValueChange={setLanguage} disabled={isLoading}>
                                        <SelectTrigger className="pl-10 glass">
                                            <SelectValue placeholder="Select your language" />
                                        </SelectTrigger>
                                        <SelectContent className="glass">
                                            {languages.map((lang) => (
                                                <SelectItem key={lang.code} value={lang.code}>
                                                    {lang.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        </div>
                        {error && (
                            <Alert variant="destructive">
                                <AlertDescription>{error}</AlertDescription>
                            </Alert>
                        )}

                        <Button type="submit" className="w-full" disabled={isLoading || !university}>
                            {isLoading ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
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