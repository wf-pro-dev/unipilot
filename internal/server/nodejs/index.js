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

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', service: 'ai-chat' });
});
// Main AI chat endpoint
app.post('/unipilot/ai/v1', async (req, res) => {
    
    try {
      const { messages, assignment } = req.body;
      
      let allMessages = convertToModelMessages(messages);

      let systemPrompt = '';
      if (assignment) {
        systemPrompt = buildSystemPrompt(assignment);
      }
  
      const result = streamText({
        model: google('gemini-2.0-flash-lite'),
        messages: allMessages,
        stopWhen: stepCountIs(5),
        maxTokens: 4000,
        temperature: 0.7,
        system: systemPrompt,
        tools: {
          getInformation: tool({
            description: `get information from your knowledge base to answer questions.`,
            inputSchema: z.object({
              question: z.string().describe('the users question'),
            }),
            execute: async ({ question }) => findRelevantContent(question),
          }),
        },
      });

      // This is the correct method from the AI SDK
      result.pipeUIMessageStreamToResponse(res);
      
    } catch (error) {
      console.error('AI chat error:', error);
      if (!res.headersSent) {
        res.status(500).json({ 
          error: 'AI service unavailable',
          details: error.message 
        });
      }
    }
});

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