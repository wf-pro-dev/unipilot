import { useState, useEffect, useCallback } from 'react'

interface NetworkStatus {
  online: boolean
  timestamp: number
  lastChecked: number
}

export function useNetworkStatus() {
  const [networkStatus, setNetworkStatus] = useState<NetworkStatus>({
    online: true,
    timestamp: Date.now(),
    lastChecked: Date.now()
  })
  const [isChecking, setIsChecking] = useState(false)

  const checkNetworkStatus = useCallback(async () => {
    if (isChecking) return

    setIsChecking(true)
    try {
      // First check browser's online status
      const browserOnline = navigator.onLine
      // Then check our backend network status
      const backendStatus = await window.go.main.App.GetNetworkStatus()
      
      setNetworkStatus({
        online: browserOnline && backendStatus.online,
        timestamp: backendStatus.timestamp * 1000, // Convert to milliseconds
        lastChecked: Date.now()
      })
    } catch (error) {
      console.error('Failed to check network status:', error)
      // If backend check fails, fall back to browser status
      setNetworkStatus(prev => ({
        ...prev,
        online: navigator.onLine,
        lastChecked: Date.now()
      }))
    } finally {
      setIsChecking(false)
    }
  }, [isChecking])

  const performSync = useCallback(async () => {
    try {
      console.log('Performing sync after coming back online...')
      await window.go.main.App.Sync()
      console.log('Sync completed successfully')
    } catch (error) {
      console.error('Failed to sync after coming back online:', error)
    }
  }, [])

  useEffect(() => {
    // Initial check
    checkNetworkStatus()

    // Set up periodic checks every 30 seconds
    const interval = setInterval(checkNetworkStatus, 30000)

    // Listen to browser online/offline events
    const handleOnline = async () => {
      setNetworkStatus(prev => ({
        ...prev,
        online: true,
        lastChecked: Date.now()
      }))
      
      // Re-check with backend after browser comes online
      setTimeout(async () => {
        await checkNetworkStatus()
        
        // If we're confirmed online, perform sync
        const currentStatus = await window.go.main.App.GetNetworkStatus()
        console.log("[Network Status] Current status:", currentStatus.online)
        if (currentStatus.online) {
          await performSync()
        }
      }, 1000)
    }

    const handleOffline = () => {
      console.log("[Network Status] Browser offline")
      setNetworkStatus(prev => ({
        ...prev,
        online: false,
        lastChecked: Date.now()
      }))
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      clearInterval(interval)
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [checkNetworkStatus, performSync])

  return {
    isOnline: networkStatus.online,
    lastChecked: networkStatus.lastChecked,
    isChecking,
    checkNetworkStatus,
    performSync
  }
}
