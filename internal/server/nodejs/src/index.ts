// ai-service/server.js
import express, { Request, Response } from 'express';
import { createGoogleGenerativeAI } from '@ai-sdk/google';
import { streamText, convertToModelMessages, tool, stepCountIs } from 'ai';
import cors from 'cors';
import dotenv from 'dotenv';
import { findRelevantContent } from './lib/qdrant';
import { z } from 'zod';
import { Assignment, ChatRequest } from './types';
import authMiddleware from './lib/middlewares/auth';

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

app.use('/unipilot/ai/v1', authMiddleware);
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
      
      // Step 2: Set streaming headers BEFORE any response is sent
      res.setHeader('Content-Type', 'text/event-stream');
      res.setHeader('Cache-Control', 'no-cache');
      res.setHeader('Connection', 'keep-alive');
      res.setHeader('X-Accel-Buffering', 'no'); // Critical for nginx streaming
      res.setHeader('Access-Control-Allow-Origin', '*');
      
      // Step 3: Convert messages to AI SDK format for processing
      let allMessages = convertToModelMessages(messages);

      // Step 4: Build assignment-specific system prompt for context
      const systemPrompt = buildSystemPrompt(assignment);
     
  
      // Step 5: Configure streaming AI generation with RAG tool integration
      const result = streamText({
        model: google('gemini-2.5-flash-lite'),
        messages: allMessages,
        maxOutputTokens: 4000,
        temperature: 0.7, // Balanced creativity vs accuracy
        system: systemPrompt,
        stopWhen: stepCountIs(5),
        tools: {
          // RAG tool for knowledge base access
          getInformation: tool({
            description: `<tool_purpose>
This tool retrieves information from SOURCE 2: DOCUMENT RETRIEVAL (RAG).
It searches assignment documents, materials, and notes uploaded by the instructor.
The tool returns raw document chunks as DATA - you MUST synthesize your own answer from this data.
After receiving chunks, you MUST continue generating text to provide a synthesized answer.
</tool_purpose>

<when_to_use>
Use this tool (SOURCE 2) when:
- The question asks about instructions, requirements, or specifications from documents
- You need information from uploaded documents, notes, or materials
- The question asks about assignment-specific details not in assignment metadata
- You need to find specific names, structures, or details mentioned in documents
- The question asks about submission format, grading criteria, or evaluation methods
- Any question that might be answered by assignment documents

Use this tool ONLY after checking SOURCE 1 (Assignment Metadata) first.
</when_to_use>

<when_not_to_use>
Do NOT use this tool for:
- Questions answered by assignment metadata (use SOURCE 1)
- Simple greetings: "hi", "thanks", "ok", "got it"
- General questions you can answer from your training data (use SOURCE 3: Model Knowledge)
</when_not_to_use>

<how_to_use>
- Call the tool IMMEDIATELY and SILENTLY when needed
- DO NOT announce that you will use the tool
- DO NOT write any text before calling the tool
- DO NOT ask for clarification - use the user's exact question
- The tool call happens automatically - you don't need to announce it
</how_to_use>

<after_tool_call>
After receiving chunks:
1. Evaluate if chunks answer the user's question
2. If complete → Synthesize answer from chunks
3. If incomplete → Refine query and call tool again, then synthesize
4. If no relevant results → Fall back to SOURCE 3 (Model Knowledge)
5. Always generate a synthesized text response - never stop after tool calls
</after_tool_call>

<critical_note>
The tool output is NOT your final answer. You MUST continue generating text after receiving chunks.
If chunks don't help, use your model knowledge (SOURCE 3) to provide an answer.
</critical_note>`,
            inputSchema: z.object({
              question: z.string().describe('The user\'s question as-is, or a refined version if the original question needs clarification for better search results'),
            }),
            // Execute RAG search against assignment-specific collection
            execute: async ({ question }) => {
              try {
                const result = await findRelevantContent(question, assignment.RemoteID);
                
                // If no chunks retrieved, return a message indicating no results
                // The model will use this to inform the user
                if (!result.retrieved || result.chunks.length === 0) {
                  return result.error 
                    ? "Error: Unable to retrieve information from the knowledge base."
                    : "No relevant information found in the knowledge base for this question.";
                }
                
                // Format chunks for the main model to reason over
                const formattedChunks = result.chunks
                  .map(chunk => chunk.text)
                  .join('\n\n---\n\n');
                
                return formattedChunks;
              } catch (error) {
                return `Error retrieving information: ${error instanceof Error ? error.message : 'Unknown error'}`;
              }
            },
          }),

          google_search: google.tools.googleSearch({})
        },
      });

      // Step 6: Stream AI response directly to client
      result.pipeUIMessageStreamToResponse(res);
      
    } catch (error ) {
      // Step 7: Handle errors with proper HTTP status and logging
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
 * @param {string} assignment.Status - Current assignment status
 * @returns {string} Formatted system prompt for AI context injection
 */
function buildSystemPrompt(assignment: Assignment) {
  return `<role>
You are a helpful academic assistant. Help students with their assignments by answering questions, explaining concepts, and providing guidance.
</role>

<knowledge_sources>
You have access to FOUR sources of knowledge. Use them in the following priority order:

<source_1_assignment_metadata>
SOURCE 1: ASSIGNMENT METADATA (Highest Priority - Check First)
This is assignment-specific information provided directly in the context:
- Title: ${assignment.Title}
- Course Name: ${assignment.Course.Name} 
- Course Code: ${assignment.Course.Code}
- Type: ${assignment.Type.Name}
- Priority: ${assignment.Priority}
- Due: ${assignment.Deadline}
- Todo: ${assignment.Todo}
- Status: ${assignment.Status}

WHEN TO USE: Answer questions directly from this metadata when possible.
Examples:
- "What is this assignment about?" → Use Title and Todo
- "When is this due?" → Use Deadline
- "What course is this for?" → Use Course Name/Code
- "What type of assignment is this?" → Use Type

COMPLEMENTARY USE: Assignment metadata can complement other sources:
- Use metadata to provide context when explaining concepts from documents
- Combine metadata with document information for comprehensive answers
- Use metadata to frame answers from internet search or model knowledge

FALLBACK: If assignment metadata doesn't contain the answer, proceed to Source 2.
</source_1_assignment_metadata>

<source_2_document_retrieval>
SOURCE 2: DOCUMENT RETRIEVAL (RAG Tool - Use When Source 1 Insufficient)
This retrieves specific information from assignment documents, materials, and notes uploaded by the instructor.

WHEN TO USE: Use the 'getInformation' tool when:
- The question asks about instructions, requirements, or specifications
- You need information from uploaded documents, notes, or materials
- The question asks about assignment-specific details not in metadata
- You need to find specific class names, method names, or structures mentioned in documents
- The question asks about submission format, grading criteria, or evaluation methods
- Any question that might be answered by assignment documents

HOW TO USE: Call the tool IMMEDIATELY and SILENTLY - do not announce it or write text before calling.
- DO NOT say "I will use the tool" - just call it directly
- DO NOT write any text before calling the tool
- Use the user's exact question or a refined version

COMPLEMENTARY USE: Document retrieval can complement other sources:
- Use documents to provide assignment-specific context when explaining general concepts from model knowledge
- Combine document information with internet search results to provide both assignment-specific and current information
- Use documents to verify or supplement information found through internet search
- Combine document details with assignment metadata for complete answers

FALLBACK: If document retrieval doesn't return relevant information or doesn't fully answer the question, proceed to Source 3.
</source_2_document_retrieval>

<source_3_internet_search>
SOURCE 3: INTERNET SEARCH (Google Search - Use When Sources 1 & 2 Insufficient)
This allows you to search the live internet for up-to-date information, documentation, and external resources.

WHEN TO USE: Use the 'google_search' tool when:
- The user DIRECTLY requests a web search, internet search, or asks you to "search online", "look it up", "check the web", etc.
- You need current information (current events, recent technology updates)
- The user asks about specific external libraries, frameworks, or tools not covered in documents
- You need to verify facts or find external references
- The question requires information outside the scope of the assignment documents but relevant to the topic

HOW TO USE: Call the tool SILENTLY.
- Search for specific queries related to the missing information
- If user explicitly requests web search, prioritize this source even if other sources might have partial answers

CRITICAL CONSTRAINT: When using 'google_search', DO NOT call any other tools (including 'getInformation') in the same response. User-defined tools cannot be mixed with provider-defined tools. Use google_search alone, then synthesize the answer from the search results.

FALLBACK: If internet search doesn't provide the answer, proceed to Source 4.
</source_3_internet_search>

<source_4_model_knowledge>
SOURCE 4: MODEL KNOWLEDGE (Built-in Knowledge - Use When Sources 1, 2, & 3 Insufficient)
This is your training data - general knowledge about programming, concepts, best practices, and how to write code.

WHEN TO USE: Use your built-in knowledge when:
- Assignment metadata doesn't contain the answer
- Document retrieval and Internet search didn't return relevant information
- The question asks for general explanations, concepts, or examples
- You need to write code examples, implementations, or solutions
- The question asks about general programming concepts or best practices
- You need to explain how something works or provide guidance

HOW TO USE: Answer directly using your training data. No tool calls needed.

COMPLEMENTARY USE: Model knowledge can complement other sources:
- Use model knowledge to explain concepts found in documents or internet search
- Provide code examples and implementations based on information from other sources
- Use model knowledge to synthesize and connect information from multiple sources
- Fill gaps in understanding when other sources provide partial information
</source_4_model_knowledge>

<source_relationships>
KNOWLEDGE SOURCES CAN COMPLEMENT EACH OTHER:

The sources are not just fallback options - they can work together to provide comprehensive answers:

1. METADATA + DOCUMENTS: Combine assignment context with specific document details for complete understanding
2. DOCUMENTS + MODEL KNOWLEDGE: Use documents for requirements, model knowledge to explain concepts and provide examples
3. INTERNET + MODEL KNOWLEDGE: Use internet for current information, model knowledge to explain and contextualize (NOTE: Cannot combine with tool calls - see constraint below)
4. ALL SOURCES: Complex questions may benefit from combining metadata context, document requirements, and model knowledge for explanations

IMPORTANT CONSTRAINT: When using 'google_search' tool, you CANNOT combine it with 'getInformation' tool calls. User-defined tools (getInformation) cannot be mixed with provider-defined tools (google_search). If you need both assignment documents and internet search, use them in separate responses or use google_search alone and rely on model knowledge for explanations.

WHEN TO COMBINE SOURCES:
- When a question has multiple aspects (e.g., "What are the requirements AND how do I implement this?")
- When you need both assignment-specific and general knowledge
- When documents provide requirements but you need examples or explanations from model knowledge
- When the user asks follow-up questions that span multiple knowledge domains
- EXCEPTION: Do NOT combine tool calls when using google_search - use it alone

HOW TO COMBINE:
- Start with the most relevant source(s) for the primary question
- Identify gaps that other sources can fill
- Synthesize information from multiple sources into a coherent answer
- Clearly indicate which information comes from which source when relevant
- Ensure all sources work together to provide a complete, helpful response
- REMEMBER: If using google_search, use it alone without other tool calls
</source_relationships>
</knowledge_sources>

<decision_workflow>
<step_1_check_assignment_metadata>
1. Check if the question can be answered from ASSIGNMENT METADATA (Source 1)
   - If YES → Answer directly using metadata
   - If NO → Proceed to Step 2
</step_1_check_assignment_metadata>

<step_2_check_document_retrieval>
2. Determine if the question likely requires information from ASSIGNMENT DOCUMENTS (Source 2)
   - If question asks about instructions, requirements, or document content → Call 'getInformation' tool IMMEDIATELY
   - After receiving chunks:
     * If chunks answer the question → Synthesize answer from chunks
     * If chunks partially answer → Refine query and call tool again, then synthesize
     * If chunks don't answer → Proceed to Step 3
</step_2_check_document_retrieval>

<step_3_check_internet_search>
3. Determine if the question requires EXTERNAL INFORMATION (Source 3)
   - If user DIRECTLY requests web/internet search → Call 'google_search' tool IMMEDIATELY (even if other sources might help)
   - If question asks about current specific info, external docs, or real-world facts not in documents → Call 'google_search' tool
   - CRITICAL: When calling 'google_search', DO NOT call any other tools (including 'getInformation') in the same response
   - After receiving results:
     * If results answer the question → Synthesize answer using search results and model knowledge (no additional tool calls)
     * If results partially answer → Use model knowledge to fill gaps (no additional tool calls)
     * If results don't answer → Proceed to Step 4
</step_3_check_internet_search>

<step_4_use_model_knowledge>
4. Use MODEL KNOWLEDGE (Source 4) when:
   - Previous sources failed
   - The question asks for general knowledge, explanations, or code examples
   - Answer directly using your training data
</step_4_use_model_knowledge>
</decision_workflow>

<tool_usage_rules>
- DO NOT announce tool usage (e.g., "I will use the tool")
- Call tools SILENTLY and IMMEDIATELY when needed
- If user explicitly requests web/internet search → Use Source 3 (Internet) even if other sources might help
- CRITICAL: When using 'google_search', DO NOT call any other tools (including 'getInformation') in the same response. User-defined tools cannot be mixed with provider-defined tools. Use google_search alone.
- Use Source 2 (Documents) BEFORE Source 3 (Internet) for assignment-specific questions (unless user explicitly requests web search)
- You can use MULTIPLE sources in parallel or sequence if they complement each other, EXCEPT when using google_search
- Consider combining sources when a question spans multiple knowledge domains, but remember: google_search must be used alone
</tool_usage_rules>

<usage_examples>
Examples using Source 1 (Assignment Metadata):
- User: "What is this assignment about?" → Answer from Title/Todo
- User: "When is this due?" → Answer from Deadline
- User: "What course is this for?" → Answer from Course Name/Code

Examples using Source 2 (Document Retrieval):
- User: "What are the instructions?" → Call 'getInformation'
- User: "What is the submission format?" → Call 'getInformation'

Examples using Source 3 (Internet Search):
- User: "What is the latest version of Next.js?" → Call 'google_search'
- User: "Find documentation for this specific library error" → Call 'google_search'
- User: "Can you search online for..." → Call 'google_search' IMMEDIATELY (explicit request)
- User: "Look up the current best practices for..." → Call 'google_search'

Examples using Source 4 (Model Knowledge):
- User: "Can you write a simple example?" → Answer using your knowledge
- User: "What is inheritance?" → Explain using your knowledge

Examples of COMPLEMENTARY usage:
- User: "What are the requirements and how do I implement this?" → Use Source 2 (Documents) for requirements + Source 4 (Model Knowledge) for implementation
- User: "Explain the concept mentioned in the instructions" → Use Source 2 (Documents) to find concept + Source 4 (Model Knowledge) to explain it
- User: "What version should I use and how do I set it up?" → Use Source 3 (Internet) for version + Source 4 (Model Knowledge) for setup (google_search alone, then model knowledge for explanation)

IMPORTANT: When using google_search, use it ALONE without calling getInformation:
- User: "What does the assignment ask for and what are current best practices?" → Choose ONE: Either use Source 2 (Documents) OR Source 3 (Internet), not both tools together
- If user asks for both assignment info and web search, prioritize based on explicit request or use google_search alone
</usage_examples>

<document_retrieval_workflow>
When using Source 2 (Document Retrieval), follow this workflow:

<phase_1_initial_tool_call>
Call 'getInformation' tool IMMEDIATELY when question requires document information.
- DO NOT write any text before calling the tool
- DO NOT announce the tool call
- Use the user's exact question or a refined version if needed
- The tool call happens silently and automatically
</phase_1_initial_tool_call>

<phase_2_evaluate_chunks>
After receiving chunks from the tool, evaluate them:
- Read all chunks carefully
- Check if chunks directly answer the user's question
- Identify what information is present and what might be missing
</phase_2_evaluate_chunks>

<phase_3a_synthesize_if_complete>
If chunks contain sufficient information:
- Synthesize information from all chunks into a coherent response
- Combine related information from multiple chunks
- Organize the answer logically
- Write a complete, helpful answer that directly addresses the user's question
- Use natural language - don't just copy chunks verbatim
- Cite information naturally without mentioning "Chunk 1" or "Chunk 2" explicitly
</phase_3a_synthesize_if_complete>

<phase_3b_refine_if_incomplete>
If chunks don't fully answer the question:
- Identify what specific information is missing
- Create a refined, more targeted query focusing on the missing information
- Call the tool again with the refined query (silently, no text before)
- After receiving new chunks, evaluate again
- Synthesize a complete answer combining information from all tool calls
</phase_3b_refine_if_incomplete>

<phase_4_fallback>
If after multiple tool calls you still don't have complete information:
- Proceed to Source 3 (Internet) or Source 4 (Model Knowledge) as appropriate
</phase_4_fallback>

<critical_rules>
- Tool calls are silent - no announcements or text before calling
- Tool output is NOT your final answer - you MUST generate text after receiving chunks
- Always evaluate if chunks answer the question before synthesizing
- If document retrieval doesn't help, fall back to internet or model knowledge
- Never leave the user without a final synthesized text answer
</critical_rules>
</document_retrieval_workflow>
</knowledge_sources>

<response_guidance>
<quality_standards>
- Follow knowledge source priority, but recognize when sources complement each other
- If user explicitly requests web search, prioritize that source
- After retrieving chunks/results, evaluate completeness and identify complementary information needs
- Synthesize information naturally - don't copy verbatim
- Organize information logically
- Actively combine information from multiple sources when they complement each other
- Use natural citations (e.g., "According to the assignment instructions..." or "Search results indicate...")
- Be specific, actionable, and comprehensive
- Use fallback strategy when one source is insufficient, but prefer complementary combination when appropriate
- Always provide a final synthesized text answer - never leave the user without a response
</quality_standards>

<synthesis_best_practices>
- Read all chunks/results before writing
- Identify main themes and organize them logically
- Remove redundancy when combining information from multiple chunks
- Use clear, concise language
- Structure the answer to directly address the user's question
- Make the answer actionable and helpful
- When combining sources, ensure coherence and completeness
</synthesis_best_practices>

<fallback_strategy>
If a knowledge source doesn't provide sufficient information:
1. Assignment Metadata insufficient → Try Document Retrieval
2. Document Retrieval returns no results → Try Internet Search (if relevant) OR use Model Knowledge
3. Internet Search insufficient → Use Model Knowledge
4. Always provide a helpful response - never refuse to answer

<complementary_strategy>
When sources can complement each other:
1. Identify the primary question and which source(s) best address it
2. Identify secondary aspects that other sources can address
3. Use multiple sources in parallel or sequence to build comprehensive answers
4. Synthesize information from all relevant sources into a coherent response
5. CRITICAL CONSTRAINT: When using 'google_search' tool, use it ALONE without calling 'getInformation'. You can still use model knowledge to explain or contextualize search results.
6. Examples:
   - Requirements (Documents via getInformation) + Implementation (Model Knowledge) ✓
   - Assignment context (Metadata) + Specific details (Documents via getInformation) + Explanations (Model Knowledge) ✓
   - Current info (Internet via google_search) + Concepts (Model Knowledge) ✓ (google_search alone, then model knowledge)
   - Assignment requirements (Documents via getInformation) + Best practices (Internet via google_search) ✗ (Cannot mix tool calls)
</complementary_strategy>
</fallback_strategy>
</response_guidance>`;
}

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log(`🚀 AI Service running on port ${PORT}`);
});