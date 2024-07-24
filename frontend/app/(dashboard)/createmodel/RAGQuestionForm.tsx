import React, { useState } from 'react';

const RAGQuestionForm = () => {
  const [llmType, setLlmType] = useState<string|null>(null);
  const [sourceType, setSourceType] = useState('');
  const [llmGuardrail, setLlmGuardrail] = useState('');

  return (
    <div>
      {/* Begin source files */}
      <span className="block text-lg font-semibold mb-2">Choose source files</span>
      <p className="mb-4">Please upload the necessary files for the RAG model.</p>

      <div className="mb-4">
        <label htmlFor="sourceType" className="block text-sm font-medium text-gray-700">Select Source Type</label>
        <select
          id="sourceType"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={sourceType}
          onChange={(e)=>{setSourceType(e.target.value)}}
        >
          <option value="">-- Please choose an option --</option>
          <option value="s3">S3</option>
          <option value="local">Local File</option>
        </select>
      </div>


      {sourceType === 's3' && (
        <div className="mb-4">
          <label htmlFor="s3Url" className="block text-sm font-medium text-gray-700">S3 URL</label>
          <input
            type="text"
            id="s3Url"
            className="mt-1 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
            placeholder="Enter S3 URL"
          />
        </div>
      )}

      {sourceType === 'local' && (
        <div className="mb-4">
          <label htmlFor="localFile" className="block text-sm font-medium text-gray-700">Upload File</label>
          <input
            type="file"
            id="localFile"
            className="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
            multiple
          />
        </div>
      )}

      {/* End source files */}

      {/* Begin choose LLM */}

      <span className="block text-lg font-semibold mb-2">Choose your LLM</span>
      <div className="mb-4">
        <label htmlFor="llmType" className="block text-sm font-medium text-gray-700">Select LLM Type</label>
        <select
          id="llmType"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={llmType ? llmType : ''}
          onChange={(e) => setLlmType(e.target.value)}
        >
          <option value="">-- Please choose an option --</option>
          <option value="OpenAI">OpenAI</option>
          <option value="Llama">Llama</option>
        </select>
      </div>

      {/* End choose LLM */}

      {/* Begin choose LLM guardrail */}

      <span className="block text-lg font-semibold mb-2">Would you like an LLM guardrail?</span>
      <div className="mb-4">
        <label htmlFor="llmGuardrail" className="block text-sm font-medium text-gray-700">Yes or No</label>
        <select
          id="llmGuardrail"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={llmGuardrail}
          onChange={(e)=>setLlmGuardrail(e.target.value)}
        >
          <option value="">-- Please choose an option --</option>
          <option value="Yes">Yes</option>
          <option value="No">No</option>
        </select>
      </div>

      {/* End choose LLM guardrail */}

    </div>
  );
};

export default RAGQuestionForm;
