/**
 * @fileoverview Query Intent Classification and Rewriting for Educational RAG
 *
 * This module classifies student queries about assignments into intent categories
 * and rewrites them to better match how information is structured in educational documents.
 *
 * Uses NLP libraries for improved text processing:
 * - stopword: Professional stopword removal
 * - compromise: Advanced pattern matching with linguistic analysis
 * - natural: Tokenization and stemming for better word matching
 *
 * Common student questions patterns:
 * - Requirements: "What are the requirements?", "What do I need to do?"
 * - Objectives: "What are the goals?", "What is this assignment about?"
 * - Code Structure: "What class should I implement?", "What methods do I need?"
 * - Instructions: "How do I start?", "What are the steps?"
 * - Concepts: "What is X?", "Explain Y"
 * - Submission: "How should I submit?", "What format?"
 * - Grading: "How will I be graded?", "What are the criteria?"
 * - Clarification: "Can you clarify X?", "I don't understand Y"
 */
/**
 * Query intent types based on common student questions about assignments
 */
export type QueryIntent = 'requirements' | 'objectives' | 'code_structure' | 'instructions' | 'concepts' | 'submission' | 'grading' | 'clarification' | 'general';
/**
 * Classifies a user query into an intent category using NLP-enhanced pattern matching
 *
 * Uses compromise library for better linguistic understanding and natural for stemming
 * to catch word variations (e.g., "requirement" vs "required" vs "require")
 *
 * @param {string} userQuery - The student's question
 * @returns {QueryIntent} The classified intent
 *
 * @example
 * classifyIntent("What are the requirements?") // returns 'requirements'
 * classifyIntent("What class should I implement?") // returns 'code_structure'
 */
export declare function classifyIntent(userQuery: string): QueryIntent;
/**
 * Rewrites a query based on its intent to better match educational document structure
 *
 * Educational documents typically structure information as:
 * - Requirements/Specifications sections
 * - Learning Objectives/Goals sections
 * - Code Structure/Implementation sections
 * - Instructions/Steps sections
 * - Concept Explanations
 * - Submission Guidelines
 * - Grading Rubrics
 *
 * @param {string} userQuery - The original user query
 * @param {QueryIntent} intent - The classified intent
 * @param {string} assignmentContext - Assignment title or context for better matching
 * @returns {string} Rewritten query optimized for semantic search
 *
 * @example
 * rewriteQuery("What are the goals?", 'objectives', "CS101 Assignment")
 * // returns "assignment learning objectives goals aims CS101 Assignment"
 */
export declare function rewriteQueryForIntent(userQuery: string, intent: QueryIntent, assignmentContext?: string): string;
/**
 * Main function to process a user query: classify intent and rewrite for better retrieval
 *
 * @param {string} userQuery - The student's question
 * @param {string} assignmentContext - Assignment title or ID for context
 * @returns {{intent: QueryIntent, rewrittenQuery: string}} Intent classification and rewritten query
 *
 * @example
 * processQuery("What are the requirements?", "CS101 Assignment")
 * // returns { intent: 'requirements', rewrittenQuery: 'assignment requirements...' }
 */
export declare function processQuery(userQuery: string, assignmentContext?: string): {
    intent: QueryIntent;
    rewrittenQuery: string;
};
//# sourceMappingURL=query-intent.d.ts.map