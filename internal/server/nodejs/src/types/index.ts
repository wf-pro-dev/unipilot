import { UIMessage } from "ai";

// Assignment types
export interface Assignment {
    ID: string;
    Title: string;
    Course: {
      Name: string;
      Code: string;
    };
    Type: string;
    Priority: string;
    Deadline: string;
    Todo: string;
    Status: string;
  }
  
  // Express request/response types
  export interface ChatRequest {
    messages: UIMessage[];
    assignment: Assignment;
  }
  
  // Qdrant response types
  export interface QdrantChunk {
    id: string;
    text: string;
    documentFileName: string;
    score: number;
    index: number;
  }
  
  // Environment variables
  export interface ProcessEnv {
    GEMINI_API_KEY?: string;
    QDRANT_HOST?: string;
    QDRANT_PORT?: string;
    PORT?: string;
  }