"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Loader2 } from "lucide-react"
import { motion } from "framer-motion"
import { useLogin } from "@/hooks/use-auth"
import { toast } from "sonner"

interface LoginFormProps {
  onLoginSuccess?: () => void
}

export function LoginForm({ onLoginSuccess }: LoginFormProps) {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const { mutate: login, isPending: isLoading } = useLogin()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    login({ username, password }, {
      onSuccess: () => {
        onLoginSuccess?.()
      },
      onError: (error) => {
        toast.error(String(error))
      }
    })
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.93 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.6, ease: "easeInOut" }}
      className="w-full flex items-center justify-center min-h-screen px-6"
    > 
      <Card className="glass w-full max-w-md mx-auto shadow-xl border-white/10">
        <CardHeader className="space-y-3 pb-8">
          <CardTitle className="text-h2 text-center bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            Welcome to Unipilot
          </CardTitle>
          <CardDescription className="text-center text-body-small text-gray-400">
            Enter your credentials to access your account
          </CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="username" className="text-body-small font-medium text-white">Username</Label>
              <Input
                id="username"
                type="text"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                disabled={isLoading}
                className="bg-background/50 border-gray-600/50 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 transition-smooth h-11"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password" className="text-body-small font-medium text-white">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={isLoading}
                className="bg-background/50 border-gray-600/50 focus:border-blue-500/50 focus:ring-2 focus:ring-blue-500/20 transition-smooth h-11"
              />
            </div>
          

            <Button type="submit" className="w-full h-11 transition-smooth hover:shadow-lg" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Logging in...
                </>
              ) : (
                "Login"
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </motion.div>
  )
} 