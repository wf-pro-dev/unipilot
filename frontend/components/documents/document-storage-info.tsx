"use client"

import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Progress } from "@/components/ui/progress"
import { Badge } from "@/components/ui/badge"
import { 
  HardDrive, 
  Info,
  RefreshCw
} from "lucide-react"
import { useUserStorageInfo } from "@/hooks/use-documents"

export function DocumentStorageInfo() {
  const { data: storageInfo, isLoading, refetch } = useUserStorageInfo()

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 B"
    const k = 1024
    const sizes = ["B", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
  }

  const getStoragePercentage = (): number => {
    if (!storageInfo) return 0
    const maxStorage = 2 * 1024 * 1024 * 1024 // 2GB in bytes
    return Math.min((storageInfo.total_size / maxStorage) * 100, 100)
  }

  const getStorageColor = (): string => {
    const percentage = getStoragePercentage()
    if (percentage >= 90) return "text-red-500"
    if (percentage >= 75) return "text-yellow-500"
    return "text-blue-500"
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button 
        variant="outline" 
        size="sm" 
        className="flex-1 bg-white/5 border-white/10 hover:bg-white/10 text-white flex justify-center h-10 transition-all" >
          <span className="text-sm text-white transition-colors">Storage</span>
          <HardDrive className="w-1 h-1 text-slate-500" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 glass border-white/10 bg-black/90 backdrop-blur-xl p-0 overflow-hidden" align="end">
        <div className="p-4 pb-3 border-b border-white/5 bg-white/5">
          {/* Header */}
          <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
              <HardDrive className="w-4 h-4 text-gray-400" />
              <h4 className="font-semibold text-white text-sm">Usage</h4>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => refetch()}
              disabled={isLoading}
              className="p-0 w-7 h-7 hover:bg-white/10 text-gray-400 hover:text-white rounded-full"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        <div className="p-4 space-y-5">
          {storageInfo ? (
            <>
              {/* Storage Bar */}
              <div className="space-y-2">
                <div className="flex justify-between items-center text-xs text-gray-400">
                  <span>Used Storage</span>
                  <span className="font-medium text-white">
                    {formatFileSize(storageInfo.total_size)} <span className="text-gray-500">/ 2 GB</span>
                  </span>
                </div>
                <Progress value={getStoragePercentage()} className="h-1.5 bg-white/10" indicatorClassName="bg-purple-500" />
                <div className={`text-[10px] text-right ${getStorageColor()} font-medium`}>
                  {Math.round(getStoragePercentage())}% used
                </div>
              </div>

              {/* Statistics */}
              <div className="grid grid-cols-2 gap-3">
                <div className="text-center p-3 rounded-xl bg-white/5 border border-white/5">
                  <div className="text-xl font-bold text-white mb-0.5">
                    {storageInfo.document_count}
                  </div>
                  <div className="text-[10px] uppercase tracking-wider text-gray-500 font-medium">Documents</div>
                </div>
                <div className="text-center p-3 rounded-xl bg-white/5 border border-white/5">
                  <div className="text-xl font-bold text-white mb-0.5">
                    {formatFileSize(storageInfo.total_size)}
                  </div>
                  <div className="text-[10px] uppercase tracking-wider text-gray-500 font-medium">Used</div>
                </div>
              </div>

              {/* Storage Levels */}
              <div className="space-y-2">
                {getStoragePercentage() >= 90 && (
                  <Badge variant="destructive" className="justify-center w-full py-1.5 font-normal">
                    Storage Almost Full
                  </Badge>
                )}
                {getStoragePercentage() >= 75 && getStoragePercentage() < 90 && (
                  <Badge variant="secondary" className="justify-center w-full text-yellow-400 bg-yellow-500/10 border border-yellow-500/20 py-1.5 font-normal">
                    Storage Getting Full
                  </Badge>
                )}
              </div>

              {/* Last Updated */}
              <div className="text-[10px] text-center text-gray-600 pt-2 border-t border-white/5">
                Last updated: {new Date(storageInfo.calculated_at).toLocaleString()}
              </div>
            </>
          ) : isLoading ? (
            <div className="flex justify-center items-center py-8">
              <RefreshCw className="mr-2 w-4 h-4 animate-spin text-purple-400" />
              <span className="text-xs text-gray-400">Loading info...</span>
            </div>
          ) : (
            <div className="flex justify-center items-center py-8">
              <Button variant="outline" onClick={() => refetch()} size="sm" className="bg-white/5 border-white/10 hover:bg-white/10 text-gray-300 hover:text-white">
                <Info className="mr-2 w-3.5 h-3.5" />
                Load Info
              </Button>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
} 