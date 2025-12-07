import { Condition, QdrantClient, ScoredPoint, Value } from '@qdrant/js-client-grpc';
import { generateEmbedding } from './embedding';
import { createGoogleGenerativeAI } from '@ai-sdk/google';
import { generateText } from 'ai';
import dotenv from 'dotenv';
import { QdrantChunk } from '../types';

dotenv.config();

const google = createGoogleGenerativeAI({
    apiKey: process.env.GEMINI_API_KEY,
});

const qdrantClient = new QdrantClient({
    host: process.env.QDRANT_HOST || 'localhost',
    port: parseInt(process.env.QDRANT_PORT || '6334'),
});

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
const findRelevantContent = async (userQuery: string, assignmentID: string) => {
    try {

        // Step 1: Convert user query to embedding vector for semantic search
        const vector = await generateEmbedding(userQuery);

        // Step 2: Validate embedding generation success
        if (!vector) {
            console.error("Failed to generate embedding for query:", userQuery);
            return "Unable to generate embedding for the query.";
        }

        // Step 2: Generate filter for document IDs
        const filters = generateFilter([assignmentID]);

        // Step 3: Perform vector similarity search in assignment-specific collection
        const collection = `unipilot-qdrant-db-${assignmentID}`;
        const relevantContent = await qdrantClient.api('points').search({
            collectionName: collection,
            vector: vector,
            limit: BigInt(3), // Retrieve top 3 most similar chunks
            // Include chunk text and assignment metadata in results
            filter: {
                should: filters,
            },
            withPayload: {
                selectorOptions: {
                    case: "include",
                    value: {
                        fields: ["chunk_text", "assignment_id"]
                    }
                }
            }
        });

        // Convert protobuf response to JSON
        const rawjsonResponse = relevantContent.toJson();

        if (!rawjsonResponse) {
            return "No relevant information found in the knowledge base.";
        }
        const jsonResponse: { result: ScoredPoint[] } = JSON.parse(JSON.stringify(rawjsonResponse));

        if (!jsonResponse || !jsonResponse.result || jsonResponse.result.length === 0) {
            return "No relevant information found in the knowledge base.";
        }

        const chunks = jsonResponse.result.map((point: ScoredPoint, index: number) => {
            const payload = point.payload || {};


            const chunkText = extractChunkText(payload);
            console.log("chunkText", chunkText);

            return {
                text: chunkText,
                score: point.score || 0,
                index: index + 1
            } as QdrantChunk;
        }).filter((chunk) => chunk.text);


        if (chunks.length === 0) {
            return "Found relevant documents but no text content was available in the payload.";
        }

        // Step 6: Format chunks for AI context with relevance indicators
        const chunksText = chunks.map((chunk: QdrantChunk) =>
            `[Chunk ${chunk.index}, Relevance: ${chunk.score.toFixed(3)}]\n${chunk.text}`
        ).join('\n\n---\n\n');

        // Step 7: Create AI prompt with retrieved context
        const assistancePrompt = `Based on the following relevant document chunks retrieved for the user's question "${userQuery}", provide a helpful and concise assistance message that directly addresses their question using the information from these chunks:\n\n${chunksText}\n\nProvide a clear, helpful response that answers the user's question using the information from the retrieved chunks.`;


        // Step 8: Generate contextual response using Google Gemini
        const { text: assistanceMessage } = await generateText({
            model: google('gemini-2.0-flash-lite'),
            prompt: assistancePrompt,
            maxOutputTokens: 1000, // Limit response length for conciseness
            temperature: 0.7, // Balanced creativity vs accuracy
        });

        return assistanceMessage;

    } catch (error) {
        console.error("findRelevantContent error", error);
        return `Error retrieving information: ${error instanceof Error ? error.message : 'Unknown error'}`;
    }
};

const extractChunkText = (payload: any) => {
    return payload['chunk_text'].stringValue as string;
}

const generateFilter = (documentIDs: string[]) => {
    return documentIDs.map((documentID) => {
        return {

            conditionOneOf: {
                case: 'field',
                value: {
                    key: 'document_id',
                    match: {
                        matchValue: { case: 'keyword', value: documentID },
                    },
                },
            },
        } as Condition;
    }
    )

}

export { findRelevantContent };