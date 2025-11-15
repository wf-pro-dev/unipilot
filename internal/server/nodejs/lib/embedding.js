const { embed } = require('ai');
const { createGoogleGenerativeAI } = require('@ai-sdk/google');
const dotenv = require('dotenv');

dotenv.config();

const google = createGoogleGenerativeAI({
    apiKey: process.env.GEMINI_API_KEY,
});

const embeddingModel =  google.textEmbeddingModel('gemini-embedding-001');;

const generateEmbedding = async (value) => {
  try {
    const input = value.replaceAll('\\n', ' ');
    console.log("input", input);
    const { embedding, usage } = await embed({
      model: embeddingModel,
      value: input,
      providerOptions: {
        google: {
          outputDimensionality: 768, // Changed from 'dimensions' to 'outputDimensionality'
        }
        
      }
    });
    return embedding;
    
  } catch (error) {
      console.error("generateEmbedding error", error);
      return null;
    }
};

module.exports = {
    generateEmbedding
};