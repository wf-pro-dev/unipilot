"use client"

import { useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { LoginForm } from "./login-form"
import { RegisterForm } from "./register-form"

interface AuthPageProps {
  onLoginSuccess: () => void
}

export default function AuthPage({ onLoginSuccess }: AuthPageProps) {
  const [activeTab, setActiveTab] = useState<"login" | "register">("login")

  return (
    <div className="min-h-screen w-full flex items-center justify-center p-4 relative overflow-hidden z-0">
      {/* Ambient Background - deep space vibe */}
    
      {/* Content Container */}
      <div className="w-full z-10 relative">
        <AnimatePresence mode="wait" initial={false}>
          {activeTab === "login" ? (
            <motion.div
              key="login"
              initial={{ opacity: 0, scale: 0.95, filter: "blur(10px)" }}
              animate={{ opacity: 1, scale: 1, filter: "blur(0px)" }}
              exit={{ opacity: 0, scale: 0.95, filter: "blur(10px)" }}
              transition={{ duration: 0.3, ease: "easeOut" }}
            >
              <LoginForm 
                onLoginSuccess={onLoginSuccess} 
                onRegisterClick={() => setActiveTab("register")} 
              />
            </motion.div>
          ) : (
            <motion.div
              key="register"
              initial={{ opacity: 0, scale: 0.95, filter: "blur(10px)" }}
              animate={{ opacity: 1, scale: 1, filter: "blur(0px)" }}
              exit={{ opacity: 0, scale: 0.95, filter: "blur(10px)" }}
              transition={{ duration: 0.3, ease: "easeOut" }}
            >
              <RegisterForm
                onRegisterSuccess={onLoginSuccess} 
                onLoginClick={() => setActiveTab("login")} 
              />
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  )
}
