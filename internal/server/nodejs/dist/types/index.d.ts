import { UIMessage } from "ai";
export interface Assignment {
    RemoteID: string;
    Title: string;
    Course: {
        Name: string;
        Code: string;
    };
    Type: {
        Name: string;
    };
    Priority: string;
    Deadline: string;
    Todo: string;
    Status: string;
}
export interface ChatRequest {
    messages: UIMessage[];
    assignment: Assignment;
}
export interface QdrantChunk {
    text: string;
    score: number;
    index: number;
}
export interface ProcessEnv {
    GEMINI_API_KEY?: string;
    QDRANT_HOST?: string;
    QDRANT_PORT?: string;
    PORT?: string;
}
//# sourceMappingURL=index.d.ts.map