"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.findRelevantContent = void 0;
const js_client_grpc_1 = require("@qdrant/js-client-grpc");
const embedding_1 = require("./embedding");
const dotenv_1 = __importDefault(require("dotenv"));
dotenv_1.default.config();
const qdrantClient = new js_client_grpc_1.QdrantClient({
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
const findRelevantContent = async (userQuery, assignmentID) => {
    try {
        // Step 1: Convert user query to embedding vector for semantic search
        const vector = await (0, embedding_1.generateEmbedding)(userQuery);
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
            limit: BigInt(3), // Retrieve top 3 most similar chunks
            // Include chunk text and assignment metadata in results
            // filter: {
            //     should: filters,
            // },
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
            return {
                chunks: [],
                question: userQuery,
                retrieved: false,
            };
        }
        const jsonResponse = JSON.parse(JSON.stringify(rawjsonResponse));
        if (!jsonResponse || !jsonResponse.result || jsonResponse.result.length === 0) {
            return {
                chunks: [],
                question: userQuery,
                retrieved: false,
            };
        }
        const chunks = jsonResponse.result.map((point, index) => {
            const payload = point.payload || {};
            const chunkText = extractChunkText(payload);
            console.log("chunkText", chunkText);
            return {
                text: chunkText,
                score: point.score || 0,
                index: index + 1
            };
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
            chunks: chunks.map(chunk => ({
                text: chunk.text,
                score: chunk.score,
                index: chunk.index
            })),
            question: userQuery,
            retrieved: true,
        };
    }
    catch (error) {
        console.error("findRelevantContent error", error);
        return {
            chunks: [],
            question: userQuery,
            retrieved: false,
            error: true,
        };
    }
};
exports.findRelevantContent = findRelevantContent;
const extractChunkText = (payload) => {
    return payload['chunk_text'].stringValue;
};
const generateFilter = (documentIDs) => {
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
        };
    });
};
//# sourceMappingURL=qdrant.js.map