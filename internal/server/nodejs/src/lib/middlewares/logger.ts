/**
 * Middleware with automatic component logging
 * 
 * Design:
 * 1. Handlers update context with relevant fields
 * 2. Middleware logs automatically at appropriate times
 * 3. Context is passed through function calls
 * 4. No manual logging calls in handlers
 */

import { Request, Response, NextFunction } from 'express';
import { v4 as uuidv4 } from 'uuid';
import {
  logInfo,
  logWarn,
  logError,
  APIContext,
  AIContext,
  RAGContext,
  ToolContext,
  logDebug,
} from '../logger/logger';

// ============================================================================
// REQUEST EXTENSION
// ============================================================================

declare global {
  namespace Express {
    interface Request {
      // Base tracking
      requestId: string;
      startTime: number;
      userId?: number;

      // Component contexts - handlers update these
      apiContext: APIContext;
      aiContext?: AIContext;
      ragContext?: RAGContext;
      toolContext?: ToolContext;
    }
  }
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

function getClientIP(req: Request): string {
  const forwarded = req.get('X-Forwarded-For');
  if (forwarded) {
    const ips = forwarded.split(',');
    if (ips.length > 0) return ips[0].trim();
  }
  const realIP = req.get('X-Real-IP');
  if (realIP) return realIP;
  return req.ip || req.socket.remoteAddress || 'unknown';
}

// ============================================================================
// API MIDDLEWARE - Logs HTTP requests automatically
// ============================================================================

export function loggerMiddleware(
  req: Request,
  res: Response,
  next: NextFunction
): void {
  const startTime = Date.now();
  const requestId = uuidv4();

  // Initialize base tracking
  req.requestId = requestId;
  req.startTime = startTime;

  // Initialize API context
  req.apiContext = {
    request_id: requestId,
    component: 'api',
    path: req.path,
    method: req.method,
    client_ip: getClientIP(req),
  };

  // Log after response completes
  res.on('finish', () => {
    const duration = Date.now() - startTime;

    // Update context with final values
    req.apiContext.duration = duration;
    req.apiContext.status_code = res.statusCode;
    req.apiContext.route = req.route?.path || req.path;
    req.apiContext.user_id = req.userId;

    // Automatic logging based on performance/status
    if (duration > 1000) {
      logWarn(req.apiContext, `Slow request: ${duration}ms`);
    } else {
      logInfo(req.apiContext);
    }






  });

  next();
}


/**
 * Create AI callbacks for Vercel AI SDK streamText
 * Returns onChunk and onFinish callbacks that log automatically
 * 
 * Usage:
 * const aiCallbacks = createAICallbacks(req, assignment, userMessage, {
 *   model: 'gemini-2.5-flash-lite',
 *   temperature: 0.7,
 *   maxTokens: 4000
 * });
 * 
 * const result = streamText({
 *   model: google('gemini-2.5-flash-lite'),
 *   messages,
 *   onChunk: aiCallbacks.onChunk,
 *   onFinish: aiCallbacks.onFinish,
 * });
 */
export function createAICallbacks(
  req: Request,
  assignment: any | undefined,
  userMessage: string,
  config: {
    model: string;
    temperature?: number;
    maxTokens?: number;
  }
) {
  const startTime = Date.now();
  let firstTokenTime: number | undefined;
  let streamChunks = 0;
  let responseLength = 0;

  // Initialize AI context
  req.aiContext = {
    request_id: req.requestId,
    user_id: req.userId,
    component: 'ai',
    assignment_id: assignment?.RemoteID,
    assignment_title: assignment?.Title,
    model: config.model,
    temperature: config.temperature,
    max_tokens: config.maxTokens,
    user_message: userMessage,
    user_message_length: userMessage.length,
  };

  return {
    // Called by Vercel AI SDK for each chunk
    onChunk({ chunk }: { chunk: any }) {
      // Track first token time
      if (!firstTokenTime) {
        firstTokenTime = Date.now();
        req.aiContext!.first_token_ms = firstTokenTime - startTime;
      }

      // Track chunks and response length
      streamChunks++;
      if (chunk.type === 'text-delta' && chunk.textDelta) {
        responseLength += chunk.textDelta.length;
      }
    },

    // Called by Vercel AI SDK when streaming finishes
    onFinish({ usage }: { usage?: any }) {
      const duration = Date.now() - startTime;

      // Update context with final values
      req.aiContext!.ai_response_length = responseLength;
      req.aiContext!.stream_chunks = streamChunks;
      req.aiContext!.generation_ms = duration;

      // Add token usage if available
      if (usage) {
        req.aiContext!.prompt_tokens = usage.promptTokens;
        req.aiContext!.completion_tokens = usage.completionTokens;
        req.aiContext!.total_tokens = usage.totalTokens;
      }

      logInfo(req.aiContext!, `AI generation executed: ${req.aiContext!.model}`);

    },
  };
}

// ============================================================================
// RAG COMPONENT WRAPPER - Logs queries automatically
// ============================================================================

/**
 * Wrapper for RAG queries that logs automatically
 * Usage: const result = await withRAGLogging(req, assignmentId, async (ctx) => {
 *   return await findRelevantContent(query, ctx);
 * });
 */
export async function withRAGLogging<T>(
  req: Request,
  assignmentId: string,
  query: string,
  fn: () => Promise<T & { chunks?: any[] }>
): Promise<T> {
  const startTime = Date.now();

  // Initialize RAG context
  const ragContext: RAGContext = {
    request_id: req.requestId,
    user_id: req.userId,
    component: 'rag',
    assignment_id: assignmentId,
    query: query,
    query_length: query.length,
    collection: `assignment_${assignmentId}`,
  };

  try {
    // Execute the RAG query
    const result = await fn();

    const duration = Date.now() - startTime;

    // Update context with results
    ragContext.qdrant_ms = duration;
    if (result.chunks) {
      ragContext.chunks_found = result.chunks.length;
      ragContext.chunks_returned = result.chunks.length;

      const ids: string[] = result.chunks.map((c: any) => c.id);
      ragContext.chunks = ids;

      // Calculate scores if available
      const scores = result.chunks
        .map((c: any) => c.score)
        .filter((s: number) => s !== undefined);

      if (scores.length > 0) {
        ragContext.max_score = Math.max(...scores);
        ragContext.avg_score = scores.reduce((a: number, b: number) => a + b, 0) / scores.length;
      }
    }


    if (ragContext.qdrant_ms! > 1000) {
      logWarn(ragContext, `Slow RAG query: ${ragContext.collection}`);
    } else {
      logInfo(ragContext, `RAG query executed: ${ragContext.collection}`);
    }



    return result;
  } catch (error) {
    const duration = Date.now() - startTime;
    ragContext.qdrant_ms = duration;
    ragContext.chunks_found = 0;

    // Automatic error logging
    logError(ragContext, 'RAG query failed', error as Error);
    throw error;
  }
}

// ============================================================================
// TOOL COMPONENT WRAPPER - Logs tool calls automatically
// ============================================================================

/**
 * Wrapper for tool execution that logs automatically
 * Usage: const result = await withToolLogging(req, 'getInformation', assignmentId, input, async (ctx) => {
 *   return await executeToolLogic(ctx);
 * });
 */
export async function withToolLogging<T>(
  req: Request,
  toolName: string,
  assignmentId: string | undefined,
  input: any,
  fn: () => Promise<T>
): Promise<T> {
  const startTime = Date.now();

  // Initialize tool context
  const toolContext: ToolContext = {
    request_id: req.requestId,
    user_id: req.userId,
    component: 'tool',
    tool_name: toolName,
    assignment_id: assignmentId,
    tool_input: input,
  };

  try {
    // Execute the tool
    const result = await fn();

    const duration = Date.now() - startTime;

    // Update context with results
    toolContext.tool_duration = duration;
    toolContext.tool_success = true;


    if (toolContext.tool_duration! > 1000) {
      logWarn(toolContext, `Slow tool execution: ${toolContext.tool_name}`);
    } else {
      logInfo(toolContext, `Tool executed: ${toolContext.tool_name}`);
    }

    return result;
  } catch (error) {
    const duration = Date.now() - startTime;

    // Update context with error info
    toolContext.tool_duration = duration;
    toolContext.tool_success = false;
    toolContext.tool_error = error instanceof Error ? error.message : 'Unknown error';

    // Automatic error logging
    logError(toolContext, `Tool failed: ${toolName}`, error as Error);
    throw error;
  }
}