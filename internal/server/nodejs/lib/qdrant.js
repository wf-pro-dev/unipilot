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
    url: 'http://localhost:6334',
    
});

const ListCollections = async () => {
    const collections = await qdrantClient.api('collections').list();
    return collections.collections;
};


const findRelevantContent = async (userQuery) => {
    try {
        const vector = await generateEmbedding(userQuery);
        
        // Check if embedding generation failed
        if (!vector) {
            console.error("Failed to generate embedding for query:", userQuery);
            return "Unable to generate embedding for the query.";
        }
        
        const collection = 'unipilot-qdrant-db-1';
        const relevantContent = await qdrantClient.api('points').search({
            collectionName: collection,
            vector: vector,
            limit: BigInt(3),
            // Correct PayloadSelector format to include only chunk_text
            withPayload: {
                selectorOptions: {
                    case: "include",
                    value: {
                        fields: ["chunk_text"]
                    }
                }
            }
        });
        
        const jsonResponse = relevantContent.toJson();
                
        if (!jsonResponse.result || jsonResponse.result.length === 0) {
            return "No relevant information found in the knowledge base.";
        }
        
        // Extract chunks with proper payload value extraction
        const chunks = jsonResponse.result.map((point, index) => {
            const payload = point.payload || {};
        
            const score = point.score || 0;
            
            return {
                text: payload.chunk_text.stringValue,
                score: score,
                index: index + 1
            };
        }).filter(chunk => chunk.text);
        
        if (chunks.length === 0) {
            return "Found relevant documents but no text content was available in the payload.";
        }

        console.log("chunks", chunks);
        
        // Generate assistance message using AI
        const chunksText = chunks.map(chunk => 
            `[Chunk ${chunk.index}, Relevance: ${chunk.score.toFixed(3)}]\n${chunk.text}`
        ).join('\n\n---\n\n');
        
        const assistancePrompt = `Based on the following relevant document chunks retrieved for the user's question "${userQuery}", provide a helpful and concise assistance message that directly addresses their question using the information from these chunks:\n\n${chunksText}\n\nProvide a clear, helpful response that answers the user's question using the information from the retrieved chunks.`;
        
        console.log("assistancePrompt", assistancePrompt);

        const { text: assistanceMessage } = await generateText({
            model: google('gemini-2.0-flash-lite'),
            prompt: assistancePrompt,
            maxTokens: 1000,
            temperature: 0.7,
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