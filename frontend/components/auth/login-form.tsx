"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2, Lock, User, ArrowRight } from "lucide-react"
import { toast } from "sonner"
import { motion } from "framer-motion"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { GlassCard } from "@/components/ui/glass-card"
import { useLogin } from "@/hooks/use-auth"
import { loginSchema, LoginValues } from "./schema"
import { cn } from "@/lib/utils"

interface LoginFormProps {
  onLoginSuccess?: () => void
  onRegisterClick?: () => void
  className?: string
}

export function LoginForm({ onLoginSuccess, onRegisterClick, className }: LoginFormProps) {
  const { mutate: login, isPending: isLoading } = useLogin()

  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  })

  const onSubmit = (data: LoginValues) => {
    login(data, {
      onSuccess: () => {
        onLoginSuccess?.()
      },
      onError: (error) => {
        toast.error(String(error))
      },
    })
  }

  return (
    <div className={cn("space-y-2 w-full mx-auto flex flex-col items-center justify-center", className)}>

      <div className="p-8 space-y-8 min-w-lg bg-white/5 border-white/30 shadow-lg shadow-black/20  relative rounded-2xl overflow-hidden">
        {/* Shine effect on hover */}
        <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-white/5 to-transparent z-0 rounded-2xl" />


        <div className="space-y-2 text-center z-10">
          <h1 className="text-2xl font-semibold tracking-tight text-white">Welcome back</h1>
          <p className="text-sm text-gray-400">
            Enter your credentials to access your workspace
          </p>
        </div>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6 z-10">
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem className="space-y-1 group">
                    <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">Username</FormLabel>
                    <div className="relative">
                      <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                        <User className="h-4 w-4" />
                      </div>
                      <FormControl>
                        <Input
                          placeholder="jdoe"
                          className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 transition-all rounded-xl"
                          disabled={isLoading}
                          {...field}
                        />
                      </FormControl>
                    </div>
                    <FormMessage className="text-xs text-red-400 ml-1" />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem className="space-y-1 group">
                    <FormLabel className="text-xs font-medium uppercase tracking-wider text-gray-400 group-focus-within:text-white ml-1 transition-colors duration-300">Password</FormLabel>
                    <div className="relative">
                      <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-white transition-colors duration-300">
                        <Lock className="h-4 w-4" />
                      </div>
                      <FormControl>
                        <Input
                          type="password"
                          placeholder="••••••••"
                          className="pl-10 h-11 bg-white/5 border-white/10 text-white placeholder:text-gray-500 focus:bg-white/10 transition-all rounded-xl"
                          disabled={isLoading}
                          {...field}
                        />
                      </FormControl>
                    </div>
                    <FormMessage className="text-xs text-red-400 ml-1" />
                  </FormItem>
                )}
              />
            </div>

            <Button
              type="submit"
              className="w-full h-11 bg-white text-black hover:bg-gray-200 rounded-xl font-medium transition-all hover:scale-[1.01] active:scale-[0.99]"
              disabled={isLoading}
            >
              {isLoading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <span className="flex items-center gap-2">
                  Login <ArrowRight className="h-4 w-4 opacity-50" />
                </span>
              )}
            </Button>
          </form>
        </Form>

      </div>

      <div className="text-center">
        <p className="text-sm text-gray-500">
          New to Unipilot?{" "}
          <button
            onClick={onRegisterClick}
            className="text-white hover:text-blue-400 font-medium transition-colors hover:underline underline-offset-4 decoration-blue-500/30"
          >
            Create an account
          </button>
        </p>
      </div>


    </div>
  )
}
