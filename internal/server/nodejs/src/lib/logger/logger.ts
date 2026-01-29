/**
 * Component-based logging system for AI service
 * 
 * Design principles:
 * 1. All context fields are optional - automatically included if set
 * 2. No conditional field extraction - just spread the context
 * 3. Logging happens in middleware, not in handlers
 * 4. Handlers update context, middleware logs automatically
 */

import pino from 'pino';
import path from 'path';
import fs from 'fs';

// ============================================================================
// CONTEXT INTERFACES - All fields optional, auto-included if set
// ============================================================================

/**
 * Base context - always present
 */
export interface BaseContext {
  request_id: string;
  user_id?: number;
  component: string;
  error?: string;
  stack?: string;
}

/**
 * API component context
 */
export interface APIContext extends BaseContext {
  component: 'api';
  path?: string;
  method?: string;
  route?: string;
  status_code?: number;
  duration?: number;
  client_ip?: string;
  error_code?: string;
}

/**
 * AI component context
 */
export interface AIContext extends BaseContext {
  component: 'ai';
  assignment_id?: string;
  assignment_title?: string;
  model?: string;
  temperature?: number;
  max_tokens?: number;
  user_message?: string;
  user_message_length?: number;
  ai_response_length?: number;
  total_tokens?: number;
  prompt_tokens?: number;
  completion_tokens?: number;
  stream_chunks?: number;
  first_token_ms?: number;
  generation_ms?: number;
}

/**
 * RAG component context
 */
export interface RAGContext extends BaseContext {
  component: 'rag';
  assignment_id?: string;
  query?: string;
  query_length?: number;
  chunks?: string[];
  chunks_found?: number;
  chunks_returned?: number;
  collection?: string;
  qdrant_ms?: number;
  avg_score?: number;
  max_score?: number;
}

/**
 * Tool component context
 */
export interface ToolContext extends BaseContext {
  component: 'tool';
  assignment_id?: string;
  tool_name?: string;
  tool_input?: any;
  tool_duration?: number;
  tool_success?: boolean;
  tool_error?: string;
  chunks_retrieved?: number;
  search_results?: number;
}

export type LogContext = APIContext | AIContext | RAGContext | ToolContext;

// ============================================================================
// LOGGER INITIALIZATION
// ============================================================================

let logger: pino.Logger;
let fileLogger: pino.Logger;

export function initLogger(): void {
  const logDir = process.env.LOG_DIR || './logs';
  
  if (!fs.existsSync(logDir)) {
    fs.mkdirSync(logDir, { recursive: true });
  }

  const logLevel = parseLogLevel(process.env.LOG_LEVEL);
  const consoleEnabled = process.env.LOG_CONSOLE_ENABLED !== 'false';

  const fileLoggerOptions: pino.LoggerOptions = {
    level: 'info',
    formatters: {
      level: (label) => ({ level: label }),
    },
    timestamp: pino.stdTimeFunctions.isoTime,
  };

  const streams: pino.StreamEntry[] = [
    {
      level: 'info',
      stream: pino.destination({
        dest: path.join(logDir, 'app.log'),
        sync: false,
        mkdir: true,
      }),
    },
    {
      level: 'error',
      stream: pino.destination({
        dest: path.join(logDir, 'error.log'),
        sync: false,
        mkdir: true,
      }),
    },
  ];

  fileLogger = pino(fileLoggerOptions, pino.multistream(streams));

  if (consoleEnabled) {
    const isDevelopment = process.env.NODE_ENV !== 'production';

    const consoleOptions: pino.LoggerOptions = {
      level: logLevel,
      formatters: {
        level: (label) => ({ level: label.toUpperCase() }),
      },
      timestamp: pino.stdTimeFunctions.isoTime,
    };

    if (isDevelopment) {
      logger = pino({
        ...consoleOptions,
        transport: {
          target: 'pino-pretty',
          options: {
            colorize: true,
            translateTime: 'HH:MM:ss',
            ignore: 'pid,hostname',
            singleLine: true,
          },
        },
      });

    } else {
      logger = pino(consoleOptions);
    }
  } else {
    logger = fileLogger;
  }
}

function parseLogLevel(levelStr: string | undefined): string {
  if (!levelStr) return 'info';
  const normalized = levelStr.toLowerCase().trim();
  const validLevels = ['debug', 'info', 'warn', 'error', 'fatal'];
  return validLevels.includes(normalized) ? normalized : 'info';
}

// ============================================================================
// LOGGING FUNCTIONS - Simple, no conditionals
// ============================================================================

export function logDebug(ctx: LogContext, message: string, fields?: Record<string, any>): void {
  logger?.debug({ ...ctx, ...fields }, message);
}

export function logInfo(ctx: LogContext, message?: string, fields?: Record<string, any>): void {
  const msg = message || 'Request processed';
  logger?.info({ ...ctx, ...fields }, msg);
  fileLogger?.info({ ...ctx, ...fields }, msg);
}

export function logWarn(
  ctx: LogContext,
  message: string,
  error?: Error,
  fields?: Record<string, any>
): void {
  const allFields = { ...ctx, ...fields };
  
  if (error) {
    allFields.error = error.message;
    allFields.stack = error.stack;
  }

  logger?.warn(allFields, message);
  fileLogger?.warn(allFields, message);
}

export function logError(
  ctx: LogContext,
  message: string,
  error?: Error,
  fields?: Record<string, any>
): void {
  const allFields = { ...ctx, ...fields };
  
  if (error) {
    allFields.error = error.message;
    allFields.stack = error.stack;
  }

  logger?.error(allFields, message);
  fileLogger?.error(allFields, message);
}

export function logFatal(
  ctx: LogContext,
  message: string,
  error?: Error,
  fields?: Record<string, any>
): void {
  const allFields = { ...ctx, ...fields };
  
  if (error) {
    allFields.error = error.message;
    allFields.stack = error.stack;
  }

  logger?.fatal(allFields, message);
  fileLogger?.fatal(allFields, message);

  setTimeout(() => process.exit(1), 100);
}



export { logger, fileLogger };