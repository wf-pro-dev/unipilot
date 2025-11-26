// ai-service/server.js
const express = require('express');
const { createGoogleGenerativeAI } = require('@ai-sdk/google');
const { streamText, convertToModelMessages, tool, stepCountIs } = require('ai');
const cors = require('cors');
const dotenv = require('dotenv');
const { findRelevantContent } = require('./lib/qdrant');
const { z } = require('zod');

dotenv.config();

const google = createGoogleGenerativeAI({
    apiKey: process.env.GEMINI_API_KEY,
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
app.get('/health', (req, res) => {
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
app.post('/unipilot/ai/v1', async (req, res) => {
    
    try {
      // Step 1: Extract and validate request payload
      const { messages, assignment } = req.body;
      
      // Step 2: Convert messages to AI SDK format for processing
      let allMessages = convertToModelMessages(messages);

      // Step 3: Build assignment-specific system prompt for context
      let systemPrompt = '';
      if (assignment) {
        console.log("assignment", assignment);
        systemPrompt = buildSystemPrompt(assignment);
      }
  
      // Step 4: Configure streaming AI generation with RAG tool integration
      const result = streamText({
        model: google('gemini-2.0-flash-lite'),
        messages: allMessages,
        maxTokens: 4000, // Limit response length for performance
        temperature: 0.7, // Balanced creativity vs accuracy
        system: systemPrompt,
        tools: {
          // RAG tool for knowledge base access
          getInformation: tool({
            description: `Use this tool to get information from your knowledge base to answer every user question.`,
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
      
    } catch (error) {
      // Step 6: Handle errors with proper HTTP status and logging
      console.error('AI chat error:', error);
      if (!res.headersSent) {
        res.status(500).json({ 
          error: 'AI service unavailable',
          details: error.message 
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
function buildSystemPrompt(assignment) {
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

Provide specific, actionable advice.`;
}

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`🚀 AI Service running on port ${PORT}`);
});