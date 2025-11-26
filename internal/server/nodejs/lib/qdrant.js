const { QdrantClient } = require('@qdrant/js-client-grpc');
const { generateEmbedding } = require('./embedding');
const { createGoogleGenerativeAI } = require('@ai-sdk/google');
const { generateText } = require('ai');
const dotenv = require('dotenv');

dotenv.config();

const google = createGoogleGenerativeAI({
    apiKey: process.env.GEMINI_API_KEY,
});

const qdrantClient = new QdrantClient({
    host: process.env.QDRANT_HOST || 'localhost',
    port: parseInt(process.env.QDRANT_PORT || '6334'),
});

/**
 * Retrieves all available collections from the Qdrant vector database.
 * Used for administrative purposes and collection management.
 * 
 * @returns {Promise<Array>} Array of collection objects containing metadata about each collection
 * @throws {Error} When Qdrant client connection fails or API call errors
 */
const ListCollections = async () => {
    const collections = await qdrantClient.api('collections').list();
    return collections.collections;
};


/**
 * Performs Retrieval Augmented Generation (RAG) by finding relevant document chunks
 * from the Qdrant vector database and generating an AI-powered response.
 * 
 * This function implements the core RAG workflow:
 * 1. Converts user query to embedding vector
 * 2. Searches assignment-specific collection for similar content
 * 3. Retrieves top 3 most relevant document chunks
 * 4. Uses Google Gemini to generate contextual response
 * 
 * @param {string} userQuery - The user's question or search query
 * @param {string} assignmentID - Assignment identifier for collection targeting
 * @returns {Promise<string>} AI-generated response based on relevant document chunks
 * @throws {Error} When embedding generation, vector search, or AI generation fails
 */
const findRelevantContent = async (userQuery, assignmentID) => {
    try {
        console.log("a", userQuery, assignmentID);
        
        // Step 1: Convert user query to embedding vector for semantic search
        const vector = await generateEmbedding(userQuery);
        
        // Step 2: Validate embedding generation success
        if (!vector) {
            console.error("Failed to generate embedding for query:", userQuery);
            return "Unable to generate embedding for the query.";
        }
        
        // Step 3: Perform vector similarity search in assignment-specific collection
        const collection = `unipilot-qdrant-db-${assignmentID}`;
        const relevantContent = await qdrantClient.api('points').search({
            collectionName: collection,
            vector: vector,
            limit: BigInt(3), // Retrieve top 3 most similar chunks
            // Include chunk text and assignment metadata in results
            withPayload: {
                selectorOptions: {
                    case: "include",
                    value: {
                        fields: ["chunk_text", "assignment_id"]
                    }
                }
            }
        });
        
        // Step 4: Parse Qdrant response and validate results
        const jsonResponse = relevantContent.toJson();
                
        if (!jsonResponse.result || jsonResponse.result.length === 0) {
            return "No relevant information found in the knowledge base.";
        }
        
        // Step 5: Extract and structure document chunks with relevance scores
        const chunks = jsonResponse.result.map((point, index) => {

            console.log("point", point);
            const payload = point.payload || {};
        
            const score = point.score || 0;
            
            return {
                text: payload.chunk_text.stringValue,
                score: score,
                index: index + 1
            };
        }).filter(chunk => chunk.text); // Filter out chunks without text content

        
        
        if (chunks.length === 0) {
            return "Found relevant documents but no text content was available in the payload.";
        }

        // Step 6: Format chunks for AI context with relevance indicators
        const chunksText = chunks.map(chunk => 
            `[Chunk ${chunk.index}, Relevance: ${chunk.score.toFixed(3)}]\n${chunk.text}`
        ).join('\n\n---\n\n');
        
        // Step 7: Create AI prompt with retrieved context
        const assistancePrompt = `Based on the following relevant document chunks retrieved for the user's question "${userQuery}", provide a helpful and concise assistance message that directly addresses their question using the information from these chunks:\n\n${chunksText}\n\nProvide a clear, helpful response that answers the user's question using the information from the retrieved chunks.`;
        

        // Step 8: Generate contextual response using Google Gemini
        const { text: assistanceMessage } = await generateText({
            model: google('gemini-2.0-flash-lite'),
            prompt: assistancePrompt,
            maxTokens: 1000, // Limit response length for conciseness
            temperature: 0.7, // Balanced creativity vs accuracy
        });
        
        return assistanceMessage;
        
    } catch (error) {
        console.error("findRelevantContent error", error);
        return `Error retrieving information: ${error.message}`;
    }
};

module.exports = {
    ListCollections,
    findRelevantContent
};