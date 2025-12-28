/**
 * @fileoverview AI Embedding Service for Document Vectorization
 *
 * This module provides text embedding functionality using Google's Gemini AI service.
 * It converts text documents into 768-dimensional vectors for semantic search and
 * retrieval-augmented generation (RAG) capabilities in the UniPilot system.
 *
 * @requires ai - AI SDK for embedding generation
 * @requires @ai-sdk/google - Google AI provider for Gemini models
 * @requires dotenv - Environment variable management
 */
/**
 * Generates semantic embeddings for text content using Google's Gemini AI.
 * Converts input text into 768-dimensional vectors suitable for semantic search,
 * document similarity, and retrieval-augmented generation (RAG) applications.
 *
 * @async
 * @function generateEmbedding
 * @param {string} value - The input text to convert into embeddings
 * @returns {Promise<number[]|null>} Promise resolving to 768-dimensional embedding vector or null on error
 *
 * @description
 * This function processes text through Google's Gemini embedding model to create
 * high-quality semantic vectors. The embeddings capture semantic meaning and can
 * be used for:
 * - Document similarity comparison
 * - Semantic search in vector databases (Qdrant)
 * - Clustering and classification tasks
 * - Retrieval-augmented generation (RAG) systems
 *
 * @example
 * // Generate embeddings for a document
 * const text = "This is a sample document about machine learning.";
 * const embedding = await generateEmbedding(text);
 * if (embedding) {
 *   console.log(`Generated ${embedding.length}-dimensional vector`);
 *   // Store in vector database for semantic search
 * }
 *
 * @throws {Error} Logs errors to console and returns null on failure
 *
 * @performance
 * - Processing time varies with text length (typically 100-500ms)
 * - Rate limits apply based on Google AI API quotas
 * - Recommended to batch process multiple documents
 *
 * @security
 * - Requires valid GEMINI_API_KEY environment variable
 * - Text content is sent to Google AI services for processing
 * - No sensitive data should be embedded without proper data handling policies
 */
declare const generateEmbedding: (value: string) => Promise<import("@ai-sdk/provider").EmbeddingModelV2Embedding | null>;
/**
 * Module exports for AI embedding functionality.
 *
 * @exports {Object} embedding - Embedding service module
 * @exports {Function} embedding.generateEmbedding - Main function for generating text embeddings
 *
 * @example
 * // Import and use the embedding service
 * import { generateEmbedding } from './lib/embedding.ts';
 *
 * async function processDocument(text) {
 *   const vector = await generateEmbedding(text);
 *   if (vector) {
 *     // Store vector in Qdrant database for semantic search
 *     await storeInVectorDB(vector);
 *   }
 * }
 */
export { generateEmbedding };
//# sourceMappingURL=embedding.d.ts.map