'use client';

import { useChat } from '@ai-sdk/react';
import { useMemo, useState, useEffect, useRef } from 'react';
import { Button } from '../ui/button';
import { ArrowUp, Bot, FileText, Send } from 'lucide-react';
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

  const scrollToBottom = () => {
    // Use requestAnimationFrame to ensure DOM is painted before scrolling
    requestAnimationFrame(() => {
      // Use a small timeout to ensure layout is stable
  
        messagesEndRef.current?.scrollIntoView({ behavior: 'instant' })

    })
  }
  
  const scrollToTop = () => {
    // Cancel any ongoing smooth scrolls first
    window.scrollTo({ top: 0, behavior: 'instant' })
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
        return {
          headers: {
            'X-Session-ID': id,
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
    // Only scroll if messages exist and are rendered
    if (messages.length > 0) {
      // Delay scroll to ensure DOM is fully rendered
      const timeoutId = setTimeout(() => {
        scrollToBottom()
      }, 50) // Small delay to let React finish rendering

      return () => clearTimeout(timeoutId)
    }
  }, [messages.length]) // Only depend on length to avoid excessive calls


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


  


  if (historyLoading) {
    return (
      <div className="flex w-full h-full justify-center items-center">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 bg-white/5 rounded-full flex items-center justify-center animate-pulse">
            <Bot className="w-6 h-6 text-white/50" />
          </div>
          <div className="text-muted-foreground text-sm">Loading chat history...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex w-full h-full justify-center">
      {messages.length > 0 && (
        <div className="flex flex-col max-w-3xl pt-32 pb-60 gap-8 w-full">
          <div ref={messagesStartRef} />
          {messages.map((message) => (
            <div key={message.id} className="w-full group animate-in fade-in slide-in-from-bottom-4 duration-500">
              <div className={`flex w-full gap-4 ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
                <div className="mt-1 flex-shrink-0">
                  {message.role === 'user' ? (
                    <Avatar className="w-8 h-8 ring-2 ring-primary/20 shadow-lg">
                      <AvatarImage src="/placeholder.svg?height=32&width=32" alt="User" />
                      <AvatarFallback className="bg-primary text-primary-foreground text-xs">U</AvatarFallback>
                    </Avatar>
                  ) : (
                    <div className="w-8 h-8 bg-gradient-to-br from-indigo-500 to-violet-500 rounded-xl items-center justify-center flex shadow-lg shadow-indigo-500/20">
                      <Bot className="w-5 h-5 text-white" />
                    </div>
                  )}
                </div>

                <div className={`flex flex-col gap-2 min-w-0 w-full ${message.role === 'user' ? 'items-end' : 'items-start'}`}>
                  <div className={`flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
                    <span className="text-caption font-medium text-foreground">{message.role === 'user' ? 'You' : 'Unipilot AI'}</span>
                    <span className="text-caption text-muted-foreground">{format(parseDeadline(message.createdAt as any), "EEEE H:mm") || new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  </div>

                  {message.parts.map((part, i) => {
                    switch (part.type) {
                      case 'text':
                        switch (message.role) {
                          case 'user':
                            return (
                              <GlassCard key={`${message.id}-${i}`} className="bg-white/5 border-white/5 shadow-lg shadow-black/40 max-w-[80%] px-5 py-3.5 rounded-2xl rounded-tr-sm">
                                <div className="text-body-small">{part.text}</div>
                              </GlassCard>
                            );
                          case 'assistant':
                            return (
                              <div key={`${message.id}-${i}`} className="p-4 my-2 max-w-full border-l-2 border-primary/40">
                                <StyledMarkdownRenderer content={part.text} />
                              </div>
                            );
                        }
                      case 'tool-getInformation':
                        return (
                          <div key={`${message.id}-${i}`} className="p-4 my-2 max-w-full border-l-2 border-primary/40">
                            <StyledMarkdownRenderer content={(part as any).output} />
                          </div>
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
        <div className="flex flex-col justify-self-center space-y-8 items-center min-h-[200px] justify-center max-w-2xl px-4">
          <div className="text-center space-y-4">
            <div className="w-16 h-16 bg-gradient-to-br from-indigo-500 to-violet-500 rounded-2xl items-center justify-center flex shadow-2xl shadow-indigo-500/30 mx-auto mb-6">
              <Bot className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-h2 font-bold tracking-tight">{assignment.Title}</h1>
            <p className="text-body text-muted-foreground max-w-md mx-auto">
              I've analyzed your assignment materials. How can I help you get started?
            </p>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full glass border border-white/5">
              <div className={`w-2 h-2 rounded-full ${assignment.Course.Color}`} />
              <span className="text-caption font-medium">{assignment.Course.Code}</span>
            </div>
            <div className="flex items-center px-3 py-1.5 rounded-full glass border border-white/5">
              <span className="text-caption font-medium">{formatDeadline(assignment.Deadline)}</span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 w-full mt-4">
            {["Summarize key requirements", "Create an outline", "Explain grading criteria", "List important deadlines"].map((suggestion) => (
              <button
                key={suggestion}
                onClick={() => {
                  setInput(suggestion);
                  // Optional: auto-send logic could go here
                }}
                className="glass p-4 rounded-xl text-left hover:bg-white/5 hover:scale-[1.02] transition-all border border-white/5 group"
              >
                <p className="text-body-small font-medium group-hover:text-primary transition-colors">{suggestion}</p>
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="fixed bottom-8 w-full px-4">
        <form
          onSubmit={e => {
            e.preventDefault();
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
          }}
        >
          <div className="relative w-full max-w-3xl mx-auto">
            <div className="absolute -right-16 bottom-0 pb-2">
              <Button
                type="button"
                className="h-10 w-10 rounded-full glass border border-white/10 hover:bg-white/10 transition-smooth"
                onClick={scrollToTop}
              >
                <ArrowUp className="w-4 h-4 text-muted-foreground" />
              </Button>
            </div>

            <div className="flex flex-col rounded-3xl glass border border-white/10 shadow-2xl shadow-black/20 overflow-hidden backdrop-blur-xl">
              <TextareaAutosize
                className="w-full p-4 bg-transparent text-body placeholder:text-muted-foreground focus:outline-none resize-none"
                value={input}
                placeholder="Ask a question..."
                onChange={e => setInput(e.currentTarget.value)}
                required
                minRows={1}
                maxRows={10}
              />
              <div className="flex py-3 px-4 items-center justify-between border-t border-white/5 bg-white/5">
                <div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="w-8 h-8 hover:bg-white/10 text-muted-foreground hover:text-primary transition-smooth"
                  >
                    <FileText className="h-4 w-4" />
                  </Button>
                </div>
                <div>
                  <Button
                    type="submit"
                    disabled={!isMessageValid}
                    size="icon"
                    className="w-8 h-8 rounded-full bg-primary hover:bg-primary/90 transition-smooth disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <Send className="h-4 w-4 text-primary-foreground" />
                  </Button>
                </div>
              </div>

            </div>
          </div>
        </form>
      </div>
    </div>
  );
}