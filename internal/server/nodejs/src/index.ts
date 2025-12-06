// ai-service/server.js
import express, { Request, Response } from 'express';
import { createGoogleGenerativeAI } from '@ai-sdk/google';
import { streamText, convertToModelMessages, tool } from 'ai';
import cors from 'cors';
import dotenv from 'dotenv';
import { findRelevantContent } from './lib/qdrant';
import { z } from 'zod';
import { Assignment, ChatRequest } from './types';

dotenv.config();

const google = createGoogleGenerativeAI({
    apiKey: process.env.GEMINI_API_KEY || '',
});

const app = express();
// Middleware
app.use(cors({
    origin: '*', // or '*' for testing
    methods: ['GET', 'POST', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization','User-Agent','x-session-id'],
    credentials: true
}));

app.use(express.json({ limit: '10mb' }));

/**
 * Health check endpoint for service monitoring and load balancer health checks.
 * Returns service status and identification for operational monitoring.
 * 
 * @route GET /health
 * @returns {Object} JSON response with service status and identification
 */
app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'healthy', service: 'ai-chat' });
});
/**
 * Main AI chat endpoint providing streaming conversational AI with RAG capabilities.
 * Integrates Google Gemini AI with Qdrant vector database for context-aware responses.
 * 
 * Features:
 * - Streaming text generation for real-time responses
 * - RAG integration for assignment-specific knowledge retrieval
 * - Tool-based information access from vector database
 * - Assignment context injection for personalized assistance
 * 
 * @route POST /unipilot/ai/v1
 * @param {Object} req.body - Request payload
 * @param {Array} req.body.messages - Conversation history in AI SDK format
 * @param {Object} req.body.assignment - Assignment context object
 * @param {string} req.body.assignment.RemoteID - Assignment ID for RAG collection targeting
 * @param {string} req.body.assignment.Title - Assignment title for context
 * @param {Object} req.body.assignment.Course - Course information
 * @param {Object} req.body.assignment.Type - Assignment type information
 * @returns {Stream} Streaming AI response with tool integration
 * @throws {500} When AI service is unavailable or processing fails
 */
app.post('/unipilot/ai/v1', async (req: Request, res: Response) => {
    
    try {
      // Step 1: Extract and validate request payload
      const { messages, assignment }: ChatRequest = req.body;
      
      // Step 2: Convert messages to AI SDK format for processing
      let allMessages = convertToModelMessages(messages);

      // Step 3: Build assignment-specific system prompt for context
      const systemPrompt = buildSystemPrompt(assignment);
     
  
      // Step 4: Configure streaming AI generation with RAG tool integration
      const result = streamText({
        model: google('gemini-2.0-flash-lite'),
        messages: allMessages,
        maxOutputTokens: 4000,
        temperature: 0.7, // Balanced creativity vs accuracy
        system: systemPrompt,
        tools: {
          // RAG tool for knowledge base access
          getInformation: tool({
            description: `Use this tool EVERY TIME the user asks a question that requires information from the knowledge base/documents. 
            Call this tool for each new question, even if you've called it before in this conversation. 
            Previous tool calls do not provide information for new questions - always call the tool when a question needs knowledge base information.`,
            inputSchema: z.object({
              question: z.string().describe('the users question'),
            }),
            // Execute RAG search against assignment-specific collection
            execute: async ({ question }) => findRelevantContent(question, assignment.RemoteID),
          }),
        },
      });

      // Step 5: Stream AI response directly to client
      result.pipeUIMessageStreamToResponse(res);
      
    } catch (error ) {
      // Step 6: Handle errors with proper HTTP status and logging
      console.error('AI chat error:', error);
      if (!res.headersSent) {
        res.status(500).json({ 
          error: 'AI service unavailable',
          details: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    }
});

/**
 * Builds a contextual system prompt for AI assistant based on assignment details.
 * Creates personalized academic assistance context using assignment metadata.
 * 
 * The system prompt provides the AI with:
 * - Assignment-specific context for relevant responses
 * - Course information for subject-appropriate guidance
 * - Assignment status and priority for urgency awareness
 * - Clear instruction for actionable academic advice
 * 
 * @param {Object} assignment - Assignment object containing context information
 * @param {string} assignment.Title - Assignment title
 * @param {Object} assignment.Course - Course information object
 * @param {string} assignment.Course.Name - Course name
 * @param {string} assignment.Course.Code - Course code
 * @param {Object} assignment.Type - Assignment type information
 * @param {string} assignment.Type.Name - Type name (essay, project, etc.)
 * @param {string} assignment.Priority - Assignment priority level
 * @param {string} assignment.Deadline - Assignment due date
 * @param {string} assignment.Todo - Assignment todo/description
 * @param {string} assignment.StatusName - Current assignment status
 * @returns {string} Formatted system prompt for AI context injection
 */
function buildSystemPrompt(assignment: Assignment) {
  return `You are a helpful academic assistant. Help with this assignment:

ASSIGNMENT CONTEXT:
- Title: ${assignment.Title}
- Course Name: ${assignment.Course.Name} 
- Course Code: ${assignment.Course.Code}
- Type: ${assignment.Type.Name}
- Priority: ${assignment.Priority}
- Due: ${assignment.Deadline}
- Todo: ${assignment.Todo}
- Status: ${assignment.StatusName}

CRITICAL INSTRUCTION FOR KNOWLEDGE BASE ACCESS:
You have access to a 'getInformation' tool that searches the knowledge base (documents, materials, notes) for this specific assignment.
1. You MUST call the 'getInformation' tool EVERY TIME the user asks a question that might require information from the assignment documents or knowledge base.
2. Do NOT rely on your internal knowledge if the answer could be in the documents.
3. Treat every user message independently regarding tool use. Even if you called the tool in previous turns, you MUST call it again for a new question if it requires knowledge base information.
4. Previous tool calls in the history do NOT provide information for the current question. Always fetch fresh information.
5. Example triggers: "What is this assignment about?", "Summarize the reading", "What are the requirements?", "Explain the concept from the notes".

Provide specific, actionable advice based on the information retrieved.`;
}

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`🚀 AI Service running on port ${PORT}`);
});