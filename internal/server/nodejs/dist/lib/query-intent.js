"use strict";
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
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.classifyIntent = classifyIntent;
exports.rewriteQueryForIntent = rewriteQueryForIntent;
exports.processQuery = processQuery;
// @ts-ignore - stopword doesn't have TypeScript definitions
const stopword_1 = require("stopword");
const compromise_1 = __importDefault(require("compromise"));
const natural_1 = require("natural");
// Initialize tokenizer once for better performance
const tokenizer = new natural_1.WordTokenizer();
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
function classifyIntent(userQuery) {
    const query = userQuery.toLowerCase().trim();
    // Use compromise for better linguistic analysis
    const doc = (0, compromise_1.default)(query);
    // Requirements/Specifications patterns
    // "What are the requirements?", "What do I need to do?", "What should I implement?"
    if (doc.has('requirement') || doc.has('specification') || doc.has('criteria') ||
        doc.has('constraint') || doc.match('need to').found || doc.match('must do').found ||
        doc.match('what do').found || doc.match('what need').found ||
        /requirement|specification|criteria|constraint|need to|should.*implement|must.*do|what.*do|what.*need/i.test(query) ||
        /what.*required|what.*supposed|what.*expected/i.test(query)) {
        return 'requirements';
    }
    // Objectives/Goals patterns
    // "What are the goals?", "What is this assignment about?", "What are the learning objectives?"
    if (doc.has('goal') || doc.has('objective') || doc.has('aim') || doc.has('purpose') ||
        doc.has('target') || doc.match('learning outcome').found ||
        doc.match('assignment about').found || doc.match('this about').found ||
        /goal|objective|aim|purpose|target|learning outcome|what.*assignment.*about|what.*this.*about/i.test(query) ||
        /assignment.*goal|assignment.*objective|assignment.*purpose/i.test(query)) {
        return 'objectives';
    }
    // Code Structure patterns
    // "What class should I implement?", "What methods do I need?", "What is the class name?"
    if (doc.match('class').has('#Verb') || doc.match('method').has('#Verb') ||
        doc.match('function').has('#Verb') || doc.match('interface').has('#Verb') ||
        doc.match('what class').found || doc.match('which class').found ||
        doc.match('what method').found || doc.match('which method').found ||
        /class.*implement|class.*name|class.*should|class.*need|method.*implement|method.*name|method.*need/i.test(query) ||
        /function.*implement|function.*name|interface.*implement|what.*class|which.*class|what.*method|which.*method/i.test(query) ||
        /code.*structure|implementation.*class|structure.*code/i.test(query)) {
        return 'code_structure';
    }
    // Instructions/Steps patterns
    // "How do I start?", "What are the steps?", "How should I approach this?"
    if (doc.match('how start').found || doc.match('how begin').found ||
        doc.match('how do').found || doc.match('how should').found ||
        doc.match('what step').found || doc.has('process') || doc.has('procedure') ||
        doc.has('instruction') || doc.has('guide') || doc.has('approach') ||
        /how.*start|how.*begin|how.*do|how.*should|what.*step|step.*by|process|procedure/i.test(query) ||
        /instruction|guide|approach|way.*do|how.*approach|how.*tackle/i.test(query)) {
        return 'instructions';
    }
    // Concepts/Explanations patterns
    // "What is X?", "Explain Y", "What does Z mean?"
    if (doc.match('what is').found || doc.match('what does').found ||
        doc.match('what mean').found || doc.has('explain') || doc.has('define') ||
        doc.has('describe') || doc.has('concept') || doc.has('theory') ||
        /what.*is|what.*does|what.*mean|explain|define|describe|concept|theory|meaning/i.test(query) ||
        /tell.*about|can.*explain|help.*understand/i.test(query)) {
        return 'concepts';
    }
    // Submission patterns
    // "How should I submit?", "What format?", "What files do I need?"
    if (doc.has('submit') || doc.has('submission') || doc.has('format') ||
        doc.match('how submit').found || doc.match('where submit').found ||
        doc.match('what format').found || doc.has('deliverable') ||
        /submit|submission|format|file.*submit|how.*submit|where.*submit|what.*format/i.test(query) ||
        /deliverable|turn.*in|hand.*in/i.test(query)) {
        return 'submission';
    }
    // Grading patterns
    // "How will I be graded?", "What are the grading criteria?", "What is worth points?"
    if (doc.has('grade') || doc.has('point') || doc.has('score') || doc.has('mark') ||
        doc.has('evaluate') || doc.has('assess') || doc.has('rubric') ||
        doc.match('how grade').found || doc.match('what worth').found ||
        doc.match('grading criteria').found ||
        /grad|point|score|mark|evaluat|assess|rubric|criteria.*grad|worth.*point/i.test(query) ||
        /how.*grad|what.*worth|how.*score|grading.*criteria/i.test(query)) {
        return 'grading';
    }
    // Clarification patterns
    // "Can you clarify X?", "I don't understand Y", "What does this mean?"
    if (doc.has('clarify') || doc.match("don't understand").found || doc.match("dont understand").found ||
        doc.has('confuse') || doc.has('unclear') || doc.match('not sure').found ||
        doc.match('what mean').found || doc.match('help understand').found ||
        doc.match('can explain').found || doc.match('could explain').found ||
        /clarif|don.*understand|confus|unclear|not.*sure|what.*mean|help.*understand/i.test(query) ||
        /can.*explain|could.*explain|please.*explain/i.test(query)) {
        return 'clarification';
    }
    // Default to general if no specific pattern matches
    return 'general';
}
/**
 * Normalizes a query by removing stopwords, tokenizing, and stemming
 *
 * Uses stopword library for professional stopword removal and natural library
 * for tokenization and stemming to normalize word forms (e.g., "requirements" -> "requir")
 *
 * @param {string} query - The original query
 * @returns {string} Normalized query with stopwords removed and words stemmed
 */
function normalizeQuery(query) {
    // Tokenize the query
    const tokens = tokenizer.tokenize(query.toLowerCase()) || [];
    // Remove stopwords using the stopword library
    const cleaned = (0, stopword_1.removeStopwords)(tokens);
    // Stem words to normalize variations (e.g., "requirements" -> "requir", "required" -> "requir")
    const stemmed = cleaned.map((word) => {
        // Only stem if word is longer than 3 chars (short words might be important)
        if (word.length > 3) {
            return natural_1.PorterStemmer.stem(word);
        }
        return word;
    });
    // Filter and clean
    return stemmed
        .filter((word) => {
        // Keep words that are:
        // - Longer than 2 characters
        // - Contain alphanumeric characters
        return word.length > 2 && /[a-z0-9]/.test(word);
    })
        .join(' ');
}
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
function rewriteQueryForIntent(userQuery, intent, assignmentContext = '') {
    // Normalize the base query
    const normalized = normalizeQuery(userQuery);
    // Intent-specific rewriting patterns based on how educational documents are structured
    switch (intent) {
        case 'requirements':
            // Educational docs use: "requirements", "specifications", "must", "should", "criteria"
            return `assignment requirements specifications criteria constraints must should ${normalized} ${assignmentContext}`;
        case 'objectives':
            // Educational docs use: "learning objectives", "goals", "aims", "outcomes"
            return `assignment learning objectives goals aims outcomes purpose targets ${normalized} ${assignmentContext}`;
        case 'code_structure':
            // Educational docs use: "class definition", "method signature", "interface", "implementation"
            return `assignment code structure class definition method implementation function interface ${normalized} ${assignmentContext}`;
        case 'instructions':
            // Educational docs use: "instructions", "steps", "guidelines", "procedures", "how to"
            return `assignment instructions steps procedures guidelines how to approach ${normalized} ${assignmentContext}`;
        case 'concepts':
            // Educational docs use: "concept", "explanation", "definition", "description"
            return `assignment concept explanation definition description theory ${normalized} ${assignmentContext}`;
        case 'submission':
            // Educational docs use: "submission", "format", "deliverables", "files"
            return `assignment submission format deliverables files turn in hand in ${normalized} ${assignmentContext}`;
        case 'grading':
            // Educational docs use: "grading", "rubric", "criteria", "points", "evaluation"
            return `assignment grading rubric criteria points evaluation assessment ${normalized} ${assignmentContext}`;
        case 'clarification':
            // For clarification, keep original query but add context terms
            return `assignment explanation clarification ${normalized} ${assignmentContext}`;
        case 'general':
        default:
            // For general queries, add assignment context
            return `${normalized} assignment ${assignmentContext}`;
    }
}
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
function processQuery(userQuery, assignmentContext = '') {
    const intent = classifyIntent(userQuery);
    const rewrittenQuery = rewriteQueryForIntent(userQuery, intent, assignmentContext);
    return {
        intent,
        rewrittenQuery
    };
}
//# sourceMappingURL=query-intent.js.map