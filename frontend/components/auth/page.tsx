import { useState } from "react"
import { LoginForm } from "./login-form"
import { RegisterForm } from "./register-form"
import { Button } from "../ui/button"

export default function AuthPage() {
  const [LoginMode, setLoginMode] = useState<"login" | "register">("login")
  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden">
      {/* Video Background */}
      <video
        autoPlay
        muted
        loop
        className="absolute inset-0 w-full h-full object-cover z-0"
      >
        <source src="/start_vid.mp4" type="video/mp4" />
      </video>

      {/* Overlay for better readability */}
      <div className="absolute inset-0 bg-black/50 z-10" />
      <div className="absolute inset-0 backdrop-blur-lg z-10" />

      {/* Content */}
      <div className="relative z-20">

        {LoginMode === "login" && (
          <LoginForm onLoginSuccess={() => window.location.reload()} />
        )}
        {LoginMode === "register" && (
          <RegisterForm onRegisterSuccess={() => window.location.reload()} />
        )}
      </div>

      <div className="absolute bottom-0 left-0 right-0 p-4 text-center text-white z-20">
        {LoginMode === "login" && (
          <div className="flex flex-row items-center justify-center">
            <p className="text-sm">Don't have an account ? </p>
            <Button variant="link" onClick={() => setLoginMode("register")}>Register</Button>
          </div>
        )}

        {LoginMode === "register" && (
          <div className="flex flex-row items-center justify-center">
            <p className="text-sm">Already have an account ? </p>
            <Button variant="link" onClick={() => setLoginMode("login")}>Login</Button>
          </div>
        )}
      </div>
    </div>
  )
}