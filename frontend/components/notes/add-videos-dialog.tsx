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

            <DialogContent className="border-gray-700 glass">
                
                <div>
                    <label className="text-sm font-medium text-gray-300 mb-2 block">
                        YouTube URL
                    </label>
                    <Input
                        value={newVideoUrl}
                        onChange={(e) => setNewVideoUrl(e.target.value)}
                        placeholder="https://www.youtube.com/watch?v=..."
                        className="bg-gray-800 border-gray-600 text-white"
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                                handleAddVideo()
                            }
                        }}
                    />
                </div>

                <DialogDescription>
                    <p>Paste the YouTube video URL below to add it to your note.</p>
                </DialogDescription>

                <DialogFooter>
                    <Button variant="outline" onClick={handleClose}>Cancel</Button>
                    <Button onClick={handleAddVideo}>Add Video</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}