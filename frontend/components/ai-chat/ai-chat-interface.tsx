'use client';

import { useChat } from '@ai-sdk/react';
import { useMemo, useState, useEffect, useRef } from 'react';
import { Button } from '../ui/button';
import { ArrowUp, FileText, Sparkles, Paperclip, ListTree, Scale, CalendarClock } from 'lucide-react';
import { assignment } from '@/wailsjs/go/models';
import { formatDeadline } from '@/lib/date-utils';
import TextareaAutosize from 'react-textarea-autosize';
import { DefaultChatTransport, generateId } from 'ai';
import { useConversationHistory, useSaveUIMessage } from '@/hooks/use-aimessages';
import { UIMessage } from '@ai-sdk/react';
import { Avatar, AvatarImage, AvatarFallback } from '../ui/avatar';
import { StyledMarkdownRenderer } from '../notes/markdown-renderer';
import { GlassCard } from '../ui/glass-card';
import { format } from 'date-fns';
import { parseDeadline } from '@/lib/date-utils';
import { cn } from '@/lib/utils';
import { DocumentUploadDialog } from '../documents/document-upload-dialog';
import { useUploadDocumentRAG } from '@/hooks/use-documents';
import { toast } from 'sonner';
import { document } from '@/wailsjs/go/models';
import { useGetAuthToken } from '@/hooks/use-auth';
import { useAuthContext } from '../provider/auth-provider';

interface AIChatInterfaceProps {
  assignment: assignment.LocalAssignment;
}

export default function Chat({ assignment }: AIChatInterfaceProps) {
  const [input, setInput] = useState('');
  const { data: conversationHistory, isLoading: historyLoading } = useConversationHistory(assignment.ID);
  const { mutate: saveUIMessage } = useSaveUIMessage();
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messagesStartRef = useRef<HTMLDivElement>(null)
  const previousAssignmentIdRef = useRef<number | null>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const { mutate: documentRAGMutation } = useUploadDocumentRAG()
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadType, setUploadType] = useState<"support" | "submission">("support")

  
 const { token } = useAuthContext()


  const scrollToBottom = () => {
    requestAnimationFrame(() => {
      messagesEndRef.current?.scrollIntoView({ behavior: 'instant' })
    })
  }

  const handleUploadComplete = () => {
    setUploadDialogOpen(false)
    // The hooks will automatically refetch and update the UI
  }

  const handleUpload = (type: "support" | "submission") => {
    setUploadType(type)
    setUploadDialogOpen(true)
  }

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: 'instant' })
  }

  const handleAddDocumentToContext = (doc: document.LocalDocument) => {

    documentRAGMutation(doc, {
      onSuccess: () => {
        toast.success(doc.FileName + " added to context")
        // ✅ No need to set state - query will update via invalidation
      },
      onError: () => {
        toast.error(doc.FileName + " failed to add to RAG")
      }
    })

  }

  // Convert conversation history to UIMessage format
  const initialMessages = useMemo(() => {
    if (!conversationHistory) return [];

    return conversationHistory.map(message => ({
      id: message.ID, // Ensure ID is string
      role: message.Role,
      parts: message.Parts as any,
      createdAt: message.CreatedAt ? new Date(message.CreatedAt) : new Date(),
    } as UIMessage));
  }, [conversationHistory]);


  const { messages, sendMessage, setMessages } = useChat({
    // Start with empty array, we'll set messages after load
    messages: [],
    transport: new DefaultChatTransport({
      api: 'https://wwwill.xyz/unipilot/ai/v1',
      prepareSendMessagesRequest: ({ id, messages, trigger, messageId }) => {
        console.log("prepareSendMessagesRequest", token)
        return {
          headers: {
            'X-Session-ID': id,
            'Authorization': `Bearer ${token}`
          },
          body: {
            messages: messages.slice(-10), // Only send last 10 messages
            trigger,
            messageId,
            assignment
          },
        };
      },
    }),
    onFinish: (message) => {
      console.log('Message finished:', message);
      saveUIMessage({
        assignmentID: assignment.ID,
        vercelMessage: message.message,
      });
    },
    onError: (error) => {
      console.error('Chat error:', error);
    },
  });

  useEffect(() => {
    if (messages.length > 0) {
      const timeoutId = setTimeout(() => {
        scrollToBottom()
      }, 50)
      return () => clearTimeout(timeoutId)
    }
  }, [messages.length])

  // Reset messages when assignment changes
  useEffect(() => {
    const assignmentChanged = previousAssignmentIdRef.current !== null &&
      previousAssignmentIdRef.current !== assignment.ID;

    if (assignmentChanged) {
      console.log('Assignment changed, clearing messages');
      setMessages([]);
    }

    previousAssignmentIdRef.current = assignment.ID;
  }, [assignment.ID, setMessages]);

  // Set messages once conversation history is loaded
  useEffect(() => {
    if (initialMessages.length > 0 && messages.length === 0) {
      console.log('Setting initial messages from conversation history');
      setMessages(initialMessages);
    }
  }, [initialMessages, messages.length, setMessages, scrollToBottom]);


  const isMessageValid = useMemo(() => {
    return input.trim().length > 0 && input.trim().length < 1000;
  }, [input]);

  const handleSuggestionClick = (suggestion: string) => {
    setInput(suggestion);
    if (inputRef.current) {
      inputRef.current.focus();
    }
  };

  const handleFormSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isMessageValid) return;

    const id = generateId();
    saveUIMessage({
      assignmentID: assignment.ID,
      vercelMessage: {
        id: id,
        role: 'user',
        parts: [{ type: 'text', text: input, state: 'done' }],
      } as UIMessage,
    });
    sendMessage({ text: input });
    setInput('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleFormSubmit(e);
    }
  };


  if (historyLoading) {
    return (
      <div className="flex w-full h-full justify-center items-center">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 bg-white/5 rounded-full flex items-center justify-center animate-pulse">
            <Sparkles className="w-6 h-6 text-white/50" />
          </div>
          <div className="text-muted-foreground text-sm">Loading chat history...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex w-full h-full justify-center relative">
      {messages.length > 0 && (
        <div className="flex flex-col max-w-3xl pt-24 pb-48 gap-6 w-full px-4">
          <div ref={messagesStartRef} />
          {messages.map((message) => (
            <div key={message.id} className="w-full group animate-in fade-in slide-in-from-bottom-2 duration-300">
              <div className={`flex w-full gap-4 ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
                <div className="mt-1 flex-shrink-0">
                  {message.role === 'user' ? (
                    <Avatar className="w-8 h-8 ring-2 ring-primary/20 shadow-lg">
                      <AvatarImage src="/placeholder.svg?height=32&width=32" alt="User" />
                      <AvatarFallback className="bg-primary text-primary-foreground text-xs">U</AvatarFallback>
                    </Avatar>
                  ) : (
                    <div className="w-8 h-8 bg-gradient-to-br from-indigo-500 to-violet-500 rounded-xl items-center justify-center flex shadow-lg shadow-indigo-500/20">
                      <Sparkles className="w-4 h-4 text-white" />
                    </div>
                  )}
                </div>

                <div className={`flex flex-col gap-1 min-w-0 w-full ${message.role === 'user' ? 'items-end' : 'items-start'}`}>
                  <div className={`flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
                    <span className="text-xs font-medium text-foreground">{message.role === 'user' ? 'You' : 'Unipilot AI'}</span>
                    <span className="text-[10px] text-muted-foreground">{format(parseDeadline((message as any).createdAt), "H:mm") || new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  </div>

                  {message.parts.map((part, i) => {
                    switch (part.type) {
                      case 'text':
                        switch (message.role) {
                          case 'user':
                            return (
                              <GlassCard key={`${message.id}-${i}`} variant="board">
                                <div className="text-sm leading-relaxed">{part.text}</div>
                              </GlassCard>
                            );
                          case 'assistant':
                            return (
                              <div key={`${message.id}-${i}`} className="text-sm leading-relaxed max-w-full">
                                <StyledMarkdownRenderer content={part.text} />
                              </div>
                            );
                        }
                      case 'tool-getInformation':
                        return (
                          <div />
                        );
                      case 'step-start':
                        return
                      default:
                        return null;
                    }
                  })}
                </div>
              </div>
            </div>
          ))}
          <div ref={messagesEndRef} />
        </div>
      )}

      {messages.length === 0 && !historyLoading && (
        <div className="flex flex-col w-full h-full items-center justify-center max-w-2xl px-6 pb-32 gap-8 animate-in fade-in zoom-in-95 duration-500">

          <div className="relative group cursor-default">
            <div className="absolute -inset-4 bg-gradient-to-r from-indigo-500/20 to-purple-500/20 rounded-full blur-xl opacity-50 group-hover:opacity-75 transition duration-1000 group-hover:duration-500 animate-pulse"></div>
            <div className="relative h-20 w-20 bg-gradient-to-br from-indigo-500 to-violet-600 rounded-2xl flex items-center justify-center shadow-2xl shadow-indigo-500/20 ring-1 ring-white/20">
              <Sparkles className="w-10 h-10 text-white" />
            </div>
          </div>

          <div className="text-center space-y-3 max-w-md">
            <h1 className="text-3xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-b from-white to-white/60">
              {assignment.Title}
            </h1>
            <p className="text-muted-foreground text-sm">
              I've analyzed your assignment. Select a topic below or ask me anything to get started.
            </p>
          </div>

          <div className="flex flex-wrap gap-2 justify-center">
            <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-white/5 border border-white/5 text-xs text-muted-foreground">
              <div className={`w-1.5 h-1.5 rounded-full ${assignment.Course.Color}`} />
              <span>{assignment.Course.Code}</span>
            </div>
            <div className="flex items-center px-3 py-1 rounded-full bg-white/5 border border-white/5 text-xs text-muted-foreground">
              <span>Due {formatDeadline(assignment.Deadline)}</span>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 w-full">
            {[
              {
                text: "Summarize key requirements",
                icon: FileText,
                desc: "Get a quick overview"
              },
              {
                text: "Create a study outline",
                icon: ListTree,
                desc: "Structure your work"
              },
              {
                text: "Explain grading criteria",
                icon: Scale,
                desc: "Understand how you're scored"
              },
              {
                text: "List important deadlines",
                icon: CalendarClock,
                desc: "Never miss a date"
              }
            ].map((suggestion, idx) => (
              <GlassCard
                key={suggestion.text}
                variant="outline"
                onClick={() => handleSuggestionClick(suggestion.text)}
                className="p-5 flex items-start gap-4 group/card overflow-hidden relative"
                style={{ animationDelay: `${idx * 100}ms` }}
              >
                <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/5 to-purple-500/5 opacity-0 group-hover/card:opacity-100 transition-opacity duration-500" />

                <div className="p-3 rounded-xl bg-white/5 border border-white/10 group-hover/card:bg-indigo-500/20 group-hover/card:border-indigo-500/30 transition-all duration-300 shadow-lg shadow-black/20">
                  <suggestion.icon className="w-5 h-5 text-gray-400 group-hover/card:text-indigo-400 transition-colors" />
                </div>

                <div className="flex flex-col gap-1 z-10">
                  <span className="text-sm font-semibold text-gray-200 group-hover/card:text-white transition-colors tracking-tight">
                    {suggestion.text}
                  </span>
                  <span className="text-xs text-muted-foreground group-hover/card:text-gray-400 transition-colors">
                    {suggestion.desc}
                  </span>
                </div>

                <div className="absolute right-4 top-1/2 -translate-y-1/2 opacity-0 group-hover/card:opacity-100 transition-all duration-300 transform translate-x-4 group-hover/card:translate-x-0">
                  <ArrowUp className="w-4 h-4 text-indigo-400 rotate-90" />
                </div>
              </GlassCard>
            ))}
          </div>
        </div>
      )}

      <div className="fixed bottom-0 flex justify-center w-full  p-4 z-40">
        <form
          onSubmit={handleFormSubmit}
          className="max-w-3xl mx-auto w-full relative"
        >
          {messages.length > 5 && (
            <div className="absolute -top-12 right-0">
              <Button
                type="button"
                size="sm"
                variant="ghost"
                className="h-8 w-8 rounded-full bg-white/5 border border-white/10 hover:bg-white/10 backdrop-blur-md"
                onClick={scrollToTop}
              >
                <ArrowUp className="w-4 h-4 text-muted-foreground" />
              </Button>
            </div>
          )}

          <div className={`relative group transition-all duration-300 ${input.trim().length > 0 ? 'scale-[1.01]' : ''}`}>
            <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500/30 to-purple-500/30 rounded-3xl opacity-0 group-focus-within:opacity-100 transition duration-500 blur-md"></div>
            <div className="relative flex flex-col rounded-3xl bg-white/5 backdrop-blur-xl border border-white/10 shadow-2xl shadow-black/40 overflow-hidden focus-within:border-white/20 transition-colors">
              <TextareaAutosize
                ref={inputRef}
                className="w-full p-4 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none resize-none min-h-[50px] max-h-[200px]"
                value={input}
                placeholder="Ask anything..."
                onChange={e => setInput(e.currentTarget.value)}
                onKeyDown={handleKeyDown}
                required
                minRows={1}
                maxRows={10}
              />
              <div className="flex items-center justify-between p-4 bg-white/5">
                <div className="flex items-center gap-1">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="w-8 h-8 text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg transition-colors"
                    onClick={() => handleUpload("support")}
                  >

                    <Paperclip className="h-4 w-4" />
                  </Button>

                </div>
                <Button
                  type="submit"
                  disabled={!isMessageValid}
                  size="icon"
                  className={cn(
                    "p-1 aspect-square rounded-full transition-all duration-300 shadow-lg",
                    isMessageValid
                      ? "bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-500/25 w-10"
                      : "bg-white/5 text-muted-foreground cursor-not-allowed"
                  )}
                >
                  {isMessageValid ? <ArrowUp className="h-4 w-4" /> : <div className="h-1.5 w-1.5 rounded-full bg-white/20" />}
                </Button>
              </div>
            </div>
          </div>

          <div className="text-center mt-2">
            <p className="text-[10px] text-muted-foreground/60">AI can make mistakes. Verify important information.</p>
          </div>
        </form>
      </div>
      <DocumentUploadDialog
        isOpen={uploadDialogOpen}
        onClose={() => setUploadDialogOpen(false)}
        onUploadComplete={handleUploadComplete}
        assignmentId={assignment.ID}
        remoteAssignmentId={assignment.RemoteID}
        documentType={uploadType}
        onSuccess={handleAddDocumentToContext}
      />
    </div>
  );
}
