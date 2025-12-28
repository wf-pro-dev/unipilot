"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
// ai-service/server.js
const express_1 = __importDefault(require("express"));
const google_1 = require("@ai-sdk/google");
const ai_1 = require("ai");
const cors_1 = __importDefault(require("cors"));
const dotenv_1 = __importDefault(require("dotenv"));
const qdrant_1 = require("./lib/qdrant");
const zod_1 = require("zod");
dotenv_1.default.config();
const google = (0, google_1.createGoogleGenerativeAI)({
    apiKey: process.env.GEMINI_API_KEY || '',
});
const app = (0, express_1.default)();
// Middleware
app.use((0, cors_1.default)({
    origin: '*', // or '*' for testing
    methods: ['GET', 'POST', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization', 'User-Agent', 'x-session-id'],
    credentials: true
}));
app.use(express_1.default.json({ limit: '10mb' }));
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
        // Step 2: Set streaming headers BEFORE any response is sent
        res.setHeader('Content-Type', 'text/event-stream');
        res.setHeader('Cache-Control', 'no-cache');
        res.setHeader('Connection', 'keep-alive');
        res.setHeader('X-Accel-Buffering', 'no'); // Critical for nginx streaming
        res.setHeader('Access-Control-Allow-Origin', '*');
        // Step 3: Convert messages to AI SDK format for processing
        let allMessages = (0, ai_1.convertToModelMessages)(messages);
        // Step 4: Build assignment-specific system prompt for context
        const systemPrompt = buildSystemPrompt(assignment);
        // Step 5: Configure streaming AI generation with RAG tool integration
        const result = (0, ai_1.streamText)({
            model: google('gemini-2.0-flash-lite'),
            messages: allMessages,
            maxOutputTokens: 4000,
            temperature: 0.7, // Balanced creativity vs accuracy
            system: systemPrompt,
            stopWhen: (0, ai_1.stepCountIs)(5),
            tools: {
                // RAG tool for knowledge base access
                getInformation: (0, ai_1.tool)({
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
                    inputSchema: zod_1.z.object({
                        question: zod_1.z.string().describe('The user\'s question as-is, or a refined version if the original question needs clarification for better search results'),
                    }),
                    // Execute RAG search against assignment-specific collection
                    execute: async ({ question }) => {
                        try {
                            const result = await (0, qdrant_1.findRelevantContent)(question, assignment.RemoteID);
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
                        }
                        catch (error) {
                            return `Error retrieving information: ${error instanceof Error ? error.message : 'Unknown error'}`;
                        }
                    },
                }),
            },
        });
        // Step 6: Stream AI response directly to client
        result.pipeUIMessageStreamToResponse(res);
    }
    catch (error) {
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
 * @param {string} assignment.StatusName - Current assignment status
 * @returns {string} Formatted system prompt for AI context injection
 */
function buildSystemPrompt(assignment) {
    return `<role>
You are a helpful academic assistant. Help students with their assignments by answering questions, explaining concepts, and providing guidance.
</role>

<knowledge_sources>
You have access to THREE sources of knowledge. Use them in the following priority order:

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
- Status: ${assignment.StatusName}

WHEN TO USE: Answer questions directly from this metadata when possible.
Examples:
- "What is this assignment about?" → Use Title and Todo
- "When is this due?" → Use Deadline
- "What course is this for?" → Use Course Name/Code
- "What type of assignment is this?" → Use Type

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

FALLBACK: If document retrieval doesn't return relevant information or doesn't fully answer the question, proceed to Source 3.
</source_2_document_retrieval>

<source_3_model_knowledge>
SOURCE 3: MODEL KNOWLEDGE (Built-in Knowledge - Use When Sources 1 & 2 Insufficient)
This is your training data - general knowledge about programming, concepts, best practices, and how to write code.

WHEN TO USE: Use your built-in knowledge when:
- Assignment metadata doesn't contain the answer
- Document retrieval doesn't return relevant information
- The question asks for general explanations, concepts, or examples
- You need to write code examples, implementations, or solutions
- The question asks about general programming concepts or best practices
- You need to explain how something works or provide guidance

HOW TO USE: Answer directly using your training data. No tool calls needed.
</source_3_model_knowledge>
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

<step_3_use_model_knowledge>
3. Use MODEL KNOWLEDGE (Source 3) when:
   - Assignment metadata doesn't answer the question
   - Document retrieval didn't return relevant information
   - The question asks for general knowledge, explanations, or code examples
   - Answer directly using your training data
</step_3_use_model_knowledge>
</decision_workflow>

<tool_usage_rules>
When using Source 2 (Document Retrieval):
- DO NOT announce that you will use the tool (e.g., "I will use the getInformation tool")
- DO NOT write any text before calling the tool - call it immediately and silently
- DO NOT ask for clarification before calling the tool
- DO NOT say "I don't have access" or "I need to use the tool" - just CALL IT SILENTLY
- DO NOT ask what the user wants to know - use their exact question
- Call the tool directly - the tool call happens automatically when you use it
</tool_usage_rules>

<usage_examples>
Examples using Source 1 (Assignment Metadata):
- User: "What is this assignment about?" → Answer from Title/Todo
- User: "When is this due?" → Answer from Deadline
- User: "What course is this for?" → Answer from Course Name/Code

Examples using Source 2 (Document Retrieval):
- User: "What are the instructions?" → Call tool immediately (no text before)
- User: "What are the requirements?" → Call tool immediately (no text before)
- User: "What is the submission format?" → Call tool immediately (no text before)

Examples using Source 3 (Model Knowledge):
- User: "Can you write a simple example?" → Answer using your knowledge
- User: "What is inheritance?" → Explain using your knowledge
- User: "How do I implement a method?" → Show example using your knowledge
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

<phase_4_fallback_to_model_knowledge>
If after multiple tool calls you still don't have complete information:
- Synthesize what you have from chunks
- Use your model knowledge (Source 3) to fill gaps or provide general guidance
- Note what information might be missing from documents
- Always provide a helpful response - never leave the user without an answer
</phase_4_fallback_to_model_knowledge>

<critical_rules>
- Tool calls are silent - no announcements or text before calling
- Tool output is NOT your final answer - you MUST generate text after receiving chunks
- Always evaluate if chunks answer the question before synthesizing
- If document retrieval doesn't help, fall back to model knowledge
- Never leave the user without a final synthesized text answer
</critical_rules>
</document_retrieval_workflow>
</knowledge_sources>

<response_guidance>
<quality_standards>
- Always follow the knowledge source priority: Assignment Metadata → Document Retrieval → Model Knowledge
- After retrieving chunks, evaluate completeness before synthesizing
- Synthesize information naturally - don't copy chunks verbatim
- Organize information logically
- Combine information from multiple sources seamlessly when needed
- Use natural citations (e.g., "According to the assignment instructions...")
- Be specific, actionable, and comprehensive
- If one source doesn't provide complete information, fall back to the next source
- Always provide a final synthesized text answer - never leave the user without a response
</quality_standards>

<synthesis_best_practices>
- Read all chunks before writing
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
2. Document Retrieval returns no results → Use Model Knowledge
3. Document Retrieval returns partial results → Synthesize from chunks + use Model Knowledge for gaps
4. Always provide a helpful response - never refuse to answer
</fallback_strategy>
</response_guidance>`;
}
const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
    console.log(`🚀 AI Service running on port ${PORT}`);
});
//# sourceMappingURL=index.js.map