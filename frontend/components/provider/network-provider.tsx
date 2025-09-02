"use client"

import React, { createContext, useContext, ReactNode } from 'react'
import { useNetworkStatus } from '@/hooks/use-network-status'

interface NetworkContextType {
  isOnline: boolean
  isChecking: boolean
  lastChecked: number
  checkNetworkStatus: () => Promise<void>
}

const NetworkContext = createContext<NetworkContextType | undefined>(undefined)

export function NetworkProvider({ children }: { children: ReactNode }) {
  const networkStatus = useNetworkStatus()

  return (
    <NetworkContext.Provider value={networkStatus}>
      {children}
    </NetworkContext.Provider>
  )
}

export function useNetworkContext() {
  const context = useContext(NetworkContext)
  if (context === undefined) {
    throw new Error('useNetworkContext must be used within a NetworkProvider')
  }
  return context
}
