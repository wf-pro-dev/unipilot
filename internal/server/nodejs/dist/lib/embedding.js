"use strict";
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
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.generateEmbedding = void 0;
const ai_1 = require("ai");
const google_1 = require("@ai-sdk/google");
const dotenv_1 = __importDefault(require("dotenv"));
// Load environment variables for API configuration
dotenv_1.default.config();
/**
 * Google AI client instance configured with API key from environment variables.
 * Provides access to Google's Gemini AI models for text processing and embedding generation.
 *
 * @type {GoogleGenerativeAI}
 */
const google = (0, google_1.createGoogleGenerativeAI)({
    apiKey: process.env.GEMINI_API_KEY,
});
/**
 * Gemini text embedding model instance for generating 768-dimensional vectors.
 * Uses the 'gemini-embedding-001' model optimized for semantic similarity tasks.
 *
 * @type {TextEmbeddingModel}
 */
const embeddingModel = google.textEmbeddingModel('gemini-embedding-001');
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
const generateEmbedding = async (value) => {
    try {
        // Step 1: Input preprocessing and normalization
        // Replace escaped newlines with spaces to prevent embedding fragmentation
        const input = value.replaceAll('\\n', ' ');
        console.log("input", input);
        // Step 2: AI model invocation with Google Gemini
        // Call Google's embedding API with 768-dimensional output for Qdrant compatibility
        const { embedding, usage } = await (0, ai_1.embed)({
            model: embeddingModel,
            value: input,
            providerOptions: {
                google: {
                    // 768 dimensions matches Qdrant vector database configuration
                    outputDimensionality: 768,
                }
            }
        });
        // Step 3: Return successful embedding result
        // Embedding is a float32 array representing semantic meaning
        return embedding;
    }
    catch (error) {
        // Step 4: Error handling and monitoring
        // Log failures for API quota monitoring and debugging
        console.error("generateEmbedding error", error);
        // Return null to allow calling code to handle embedding failures gracefully
        return null;
    }
};
exports.generateEmbedding = generateEmbedding;
//# sourceMappingURL=embedding.js.map