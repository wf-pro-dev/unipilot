import { Condition, QdrantClient, ScoredPoint, Value } from '@qdrant/js-client-grpc';
import { generateEmbedding } from './embedding';
import dotenv from 'dotenv';
import { QdrantChunk } from '../types';

dotenv.config();

const qdrantClient = new QdrantClient({
    host: process.env.QDRANT_HOST || 'localhost',
    port: parseInt(process.env.QDRANT_PORT || '6334'),
});

/**
 * Performs Retrieval Augmented Generation (RAG) by finding relevant document chunks
 * from the Qdrant vector database. Returns raw chunks for the main model to reason over.
 * 
 * This function implements the retrieval part of RAG:
 * 1. Converts user query to embedding vector
 * 2. Searches assignment-specific collection for similar content
 * 3. Retrieves top 3 most relevant document chunks
 * 
 * The main model will synthesize the answer from these chunks, enabling multi-step reasoning.
 * 
 * @param {string} userQuery - The user's question or search query
 * @param {string} assignmentID - Assignment identifier for collection targeting
 * @returns {Promise<{chunks: Array<{text: string, score: number, index: number}>, question: string, retrieved: boolean, error?: boolean}>} Raw retrieved chunks with metadata
 * @throws {Error} When embedding generation or vector search fails
 */
const findRelevantContent = async ( userQuery: string, assignmentID: string): Promise<{
    chunks: Array<QdrantChunk>;
    question: string;
    retrieved: boolean;
    error?: boolean;
}> => {
    try {

        // Step 1: Convert user query to embedding vector for semantic search
        const vector = await generateEmbedding(userQuery);

        // Step 2: Validate embedding generation success
        if (!vector) {
            console.error("Failed to generate embedding for query:", userQuery);
            return {
                chunks: [],
                question: userQuery,
                retrieved: false,
                error: true,
            };
        }

        // Step 2: Generate filter for document IDs
        const filters = generateFilter([assignmentID]);

        // Step 3: Perform vector similarity search in assignment-specific collection
        const collection = `unipilot-qdrant-db-${assignmentID}`;
        const relevantContent = await qdrantClient.api('points').search({
            collectionName: collection,
            vector: vector,
            limit: BigInt(10), // Retrieve top 3 most similar chunks
            // Include chunk text and assignment metadata in results
            // filter: {
            //     should: filters,
            // },
            withPayload: {
                selectorOptions: {
                    case: "include",
                    value: {
                        fields: ["chunk_id","chunk_text", "assignment_id", "document_file_name"]
                    }
                }
            }
        });

        // Convert protobuf response to JSON
        const rawjsonResponse = relevantContent.toJson();

        if (!rawjsonResponse) {
            return {
                chunks: [],
                question: userQuery,
                retrieved: false,
            };
        }
        const jsonResponse: { result: ScoredPoint[] } = JSON.parse(JSON.stringify(rawjsonResponse));

        if (!jsonResponse || !jsonResponse.result || jsonResponse.result.length === 0) {
            return {
                chunks: [],
                question: userQuery,
                retrieved: false,
            };
        }


        const chunks = jsonResponse.result.map((point: ScoredPoint, index: number) => {
            const payload = point.payload || {};

            const chunkId = extractChunkId(payload);

            const chunkText = extractChunkText(payload);

            const documentFileName = extractDocumentFileName(payload);

            return {
                id: chunkId.toString(),
                text: chunkText,
                documentFileName: documentFileName,
                score: point.score || 0,
                index: index + 1
            } as QdrantChunk;
        }).filter((chunk) => chunk.text);


        if (chunks.length === 0) {
            return {
                chunks: [],
                question: userQuery,
                retrieved: false,
            };
        }

        // Return raw chunks for the main model to reason over
        // This enables multi-step reasoning and better answer synthesis
        return {
            chunks: chunks,
            question: userQuery,
            retrieved: true,
        };

    } catch (error) {
        return {
            chunks: [],
            question: userQuery,
            retrieved: false,
            error: true,
        };
    }
};

const extractChunkId = (payload: any) => {
    return payload['chunk_id'].stringValue as string;
}

const extractChunkText = (payload: any) => {
    return payload['chunk_text'].stringValue as string;
}

const extractDocumentFileName = (payload: any) => {
    return payload['document_file_name'].stringValue as string;
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