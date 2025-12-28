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
declare const findRelevantContent: (userQuery: string, assignmentID: string) => Promise<{
    chunks: Array<{
        text: string;
        score: number;
        index: number;
    }>;
    question: string;
    retrieved: boolean;
    error?: boolean;
}>;
export { findRelevantContent };
//# sourceMappingURL=qdrant.d.ts.map