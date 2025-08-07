"use client"

import { LoginForm } from "@/components/auth/login-form"
import { useRouter } from "next/navigation"

export default function LoginPage() {
  const router = useRouter()

  const handleLoginSuccess = () => {
    // Redirect to dashboard after successful login
    router.push("/")
  }

  return <LoginForm onLoginSuccess={handleLoginSuccess} />
} 