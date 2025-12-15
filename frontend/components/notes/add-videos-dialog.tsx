import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { note } from "@/wailsjs/go/models"
import { Input } from "../ui/input"
import { useState } from "react"

interface AddVideosDialogProps {
    isOpen: boolean
    onClose: () => void
    note: note.LocalNote
    onAddVideo: (note: note.LocalNote, video: string) => void
}

export function AddVideosDialog({ isOpen, onClose, note, onAddVideo }: AddVideosDialogProps) {

    const [newVideoUrl, setNewVideoUrl] = useState("")


    const handleClose = () => {
        setNewVideoUrl("")
        onClose()
    }

    const handleAddVideo = () => {
        if (newVideoUrl.trim()) {
            onAddVideo(note, newVideoUrl)
            handleClose()
        }
    }


    return (
        <Dialog open={isOpen} onOpenChange={handleClose}>

            <DialogContent className="glass border-white/10 text-white max-w-md p-0 overflow-hidden gap-0">
                <DialogHeader className="p-6 pb-4 border-b border-white/5 bg-white/5">
                    <DialogTitle className="text-xl font-semibold">Add Video</DialogTitle>
                </DialogHeader>
                
                <div className="p-6 space-y-6">
                    <div className="space-y-2">
                        <label className="text-xs font-medium text-gray-400 uppercase tracking-wider block">
                            YouTube URL
                        </label>
                        <Input
                            value={newVideoUrl}
                            onChange={(e) => setNewVideoUrl(e.target.value)}
                            placeholder="https://www.youtube.com/watch?v=..."
                            className="bg-white/5 border-white/10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all h-10 text-white placeholder:text-gray-500"
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                    handleAddVideo()
                                }
                            }}
                        />
                    </div>

                    <DialogDescription className="text-gray-400 text-sm bg-white/5 p-3 rounded-lg border border-white/5">
                        Paste the YouTube video URL above to add it to your note. Supports standard URLs, Short URLs, and Embed URLs.
                    </DialogDescription>

                    <DialogFooter className="gap-3 pt-2">
                        <Button 
                            variant="outline" 
                            onClick={handleClose}
                            className="flex-1 border-white/10 bg-transparent hover:bg-white/5 text-gray-300 hover:text-white"
                        >
                            Cancel
                        </Button>
                        <Button 
                            onClick={handleAddVideo}
                            className="flex-1 bg-blue-600 hover:bg-blue-500 text-white shadow-[0_0_15px_rgba(37,99,235,0.2)]"
                        >
                            Add Video
                        </Button>
                    </DialogFooter>
                </div>
            </DialogContent>
        </Dialog>
    )
}