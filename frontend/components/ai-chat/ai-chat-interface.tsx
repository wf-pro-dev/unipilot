'use client';

import { useChat } from '@ai-sdk/react';
import { useMemo, useState, useEffect } from 'react';
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
import { useRef } from 'react';

interface AIChatInterfaceProps {
  assignment: assignment.LocalAssignment;
}

export default function Chat({ assignment }: AIChatInterfaceProps) {
  const [input, setInput] = useState('');
  const { data: conversationHistory, isLoading: historyLoading } = useConversationHistory(assignment.ID);
  const { mutate: saveUIMessage } = useSaveUIMessage();
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messagesStartRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }
  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
  // Convert conversation history to UIMessage format
  const initialMessages = useMemo(() => {
    if (!conversationHistory) return [];

    return conversationHistory.map(message => ({
      id: message.id, // Ensure ID is string
      role: message.role,
      parts: message.parts,
      createdAt: message.created_at ? new Date(message.created_at) : new Date(),
    } as UIMessage));
  }, [conversationHistory]);

  console.log('Initial messages for useChat:', initialMessages);

  const { messages, sendMessage, setMessages } = useChat({
    // Start with empty array, we'll set messages after load
    messages: [],
    transport: new DefaultChatTransport({
      api: 'https://wwwill.dedyn.io/unipilot/ai/v1',
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
    scrollToBottom()
  }, [messages])

  // Set messages once conversation history is loaded
  useEffect(() => {
    if (initialMessages.length > 0 && messages.length === 0) {
      console.log('Setting initial messages from conversation history');
      setMessages(initialMessages);
    }
  }, [initialMessages, messages.length, setMessages]);


  const isMessageValid = useMemo(() => {
    return input.trim().length > 0 && input.trim().length < 1000;
  }, [input]);

  console.log('Current messages:', messages);
  console.log('Conversation history:', conversationHistory);

  if (historyLoading) {
    return <div>Loading conversation history...</div>;
  }

  return (
    <div className="flex w-full h-full justify-center">
      {messages.length > 0 && (
        <div className="flex flex-col max-w-xl py-32 gap-12 ">
          <div ref={messagesStartRef} />
          {messages.map(message => (
            <div key={message.id} className="whitespace-pre-wrap">

              <div className="flex space-x-2">
                <div className="mt-4">
                  {message.role === 'user' ? (
                    <Avatar className="w-8 h-8">
                      <AvatarImage src="/placeholder.svg?height=32&width=32" alt="User" />
                      <AvatarFallback>{message.role === 'user' ? 'U' : 'A'}</AvatarFallback>
                    </Avatar>
                  ) : (
                    <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600  rounded-full items-center justify-center flex">
                      <Bot className="w-4 h-4" />
                    </div>
                  )}
                </div>
                <div className="flex flex-col gap-2">
                  <span className="text-gray-300 text-xs">{message.role === 'user' ? 'User: ' : 'AI: '}</span>
                  {message.parts.map((part, i) => {
                    switch (part.type) {
                      case 'text':
                        switch (message.role) {
                          case 'user':
                            return (
                              <div key={`${message.id}-${i}`} className="glass p-4 rounded-lg">
                                <div>{part.text}</div>
                              </div>
                            );
                          case 'assistant':
                            return (
                              <StyledMarkdownRenderer
                                content={part.text}
                              />
                            );
                        }
                      case 'tool-getInformation':
                        return (
                          <div key={`${message.id}-${i}`}>
                            <StyledMarkdownRenderer content={part.output} />
                          </div>
                        );
                      case 'step-start':
                        return
                      default:
                        return <div key={`${message.id}-${i}`}>[Unsupported part type: {part.type}]</div>;
                    }
                  })}
                </div>
              </div>
              <div ref={messagesEndRef} />
            </div>
          ))}
        </div>
      )}

      {/* Rest of your JSX remains the same */}
      {messages.length === 0 && !historyLoading && (
        <div className="flex flex-col justify-self-center space-y-2 items-center min-h-[200px] justify-center">
          <h1 className="text-3xl text-center font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent">{assignment.Title}</h1>
          <span className="text-gray-300">Ask me anything about your assignment</span>
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-2 p-2 rounded-lg glass">
              <div className={`w-1.5 h-1.5 rounded-full ${assignment.Course.Color}`} />
              <p className="text-white text-xs">{assignment.Course.Code}</p>
            </div>
            <div className="flex items-center p-2 rounded-lg glass">
              <p className="text-white text-xs">{formatDeadline(assignment.Deadline)}</p>
            </div>
          </div>
        </div>
      )}

      <div className="fixed w-full h-screen">
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
          <div className="absolute bottom-0 left-1/2 w-full max-w-2xl -translate-x-1/2">
            <div className="absolute -right-14 -top-14 p-4">
              <Button
                type="button"
                className="aspect-square p-3 glass hover:bg-white/50 rounded-full items-center justify-center flex"
                onClick={scrollToTop}
              >
                <ArrowUp className="w-4 h-4 text-white" />
              </Button>
            </div>

            <div className="flex flex-col border border-gray-600 rounded-3xl overflow-hidden">
              <TextareaAutosize
                className="w-full p-4 glass min-h-6 focus:outline-none"
                value={input}
                placeholder="Say something..."
                onChange={e => setInput(e.currentTarget.value)}
                required
                minRows={1}
                maxRows={10}
                style={{
                  resize: 'none',
                  
                }}
              />
              <div className="flex py-2 px-4 items-center justify-between glass">
                <div>
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="w-8 h-8 items-center justify-center flex bg-transparent border-blue-600 hover:bg-blue-600/10"
                  >
                    <FileText className="h-4 w-4 text-blue-400" />
                  </Button>
                </div>
                <div>
                  <Button
                    type="submit"
                    disabled={!isMessageValid}
                    className="w-8 h-8 items-center justify-center flex bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 transition-all duration-300"
                  >
                    <Send className="h-4 w-4" color="white" />
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