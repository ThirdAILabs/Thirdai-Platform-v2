import React, { useState } from 'react';

const RAGQuestionForm = () => {
  // Begin state variables & func for source
  const [useExistingSemanticSearch, setUseExistingSemanticSearch] = useState<string|null>(null);
  const [sources, setSources] = useState<Array<{ type: string, value: string }>>([]);
  const [newSourceType, setNewSourceType] = useState<string>('');
  const [newSourceValue, setNewSourceValue] = useState<string>('');

  const handleAddSource = () => {
    if (newSourceType && newSourceValue) {
      setSources([...sources, { type: newSourceType, value: newSourceValue }]);
      setNewSourceType('');
      setNewSourceValue('');
    }
  };

  const handleSourceTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setNewSourceType(e.target.value);
  };

  const handleSourceValueChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewSourceValue(e.target.value);
  };

  const handleDeleteSource = (index: number) => {
    const updatedSources = sources.filter((_, i) => i !== index);
    setSources(updatedSources);
  };

  // End state variables & func for source

  // Begin state variables & func for LLM guardrail

  const [llmGuardrail, setLlmGuardrail] = useState('');
  const [useExistingGuardrail, setUseExistingGuardrail] = useState<string|null>(null);

  // End state variables & func for LLM guardrail

  // Begin state variables & func for LLM

  const [llmType, setLlmType] = useState<string|null>(null);
  const [chatInput, setChatInput] = useState('');
  const [chatHistory, setChatHistory] = useState<Array<{ sender: string, message: string }>>([]);

  const handleSendMessage = () => {
    if (chatInput.trim()) {
      setChatHistory([...chatHistory, { sender: 'User', message: chatInput }]);
      // Simulate LLM response for demo purposes
      setChatHistory([...chatHistory, { sender: 'User', message: chatInput }, { sender: 'LLM', message: `LLM response to "${chatInput}"` }]);
      setChatInput('');
    }
  };

  // End state variables & func for LLM

  return (
    <div>
      {/* Begin source files */}
      <div className="mb-4">
        <span className="block text-lg font-semibold mb-2">Search system</span>
        <label htmlFor="useExistingSemanticSearch" className="block text-sm font-medium text-gray-700">Use Existing Semantic Search?</label>
        <select
          id="useExistingSemanticSearch"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={useExistingSemanticSearch ? useExistingSemanticSearch : ''}
          onChange={(e) => setUseExistingSemanticSearch(e.target.value)}
        >
          <option value="">-- Please choose an option --</option>
          <option value="Yes">Yes</option>
          <option value="No">No, create a new one</option>
        </select>
      </div>

      {
        useExistingSemanticSearch === 'No' &&
        <>
          <span className="block text-lg font-semibold mb-2">Choose source files</span>
          <p className="mb-4">Please upload the necessary files for the RAG model.</p>

          <div className="mb-4">
            <label htmlFor="newSourceType" className="block text-sm font-medium text-gray-700">Select Source Type</label>
            <select
              id="newSourceType"
              className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
              value={newSourceType}
              onChange={handleSourceTypeChange}
            >
              <option value="">-- Please choose an option --</option>
              <option value="s3">S3</option>
              <option value="local">Local File</option>
            </select>
          </div>

          {newSourceType === 's3' && (
            <div className="mb-4">
              <label htmlFor="newSourceValue" className="block text-sm font-medium text-gray-700">S3 URL</label>
              <input
                type="text"
                id="newSourceValue"
                className="mt-1 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
                placeholder="Enter S3 URL"
                value={newSourceValue}
                onChange={handleSourceValueChange}
              />
            </div>
          )}

          {newSourceType === 'local' && (
            <div className="mb-4">
              <label htmlFor="newSourceValue" className="block text-sm font-medium text-gray-700">Upload File</label>
              <input
                type="file"
                id="newSourceValue"
                className="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
                onChange={(e) => {
                  if (e.target.files) {
                    setNewSourceValue(e.target.files[0].name);
                  }
                }}
                multiple
              />
            </div>
          )}

          <button
            type="button"
            className="mb-4 bg-blue-500 text-white px-4 py-2 rounded-md"
            onClick={handleAddSource}
          >
            Add Source
          </button>

          {/* Begin Display added sources */}
          <div>
            <h3 className="text-lg font-semibold mb-2">Added Sources</h3>
            <ul>
              {sources.map((source, index) => (
                <li key={index} className="mb-2 flex items-center justify-between">
                  <span className="inline-block px-4 py-2 rounded-md bg-gray-200 text-black">
                    <strong>{source.type === 's3' ? 'S3 URL' : 'Local File'}:</strong> {source.value}
                  </span>
                  <button
                    type="button"
                    className="ml-4 bg-red-500 text-white px-2 py-1 rounded-md"
                    onClick={() => handleDeleteSource(index)}
                  >
                    Delete
                  </button>
                </li>
              ))}
            </ul>
          </div>
          {/* End Display added sources */}
        </>
      }
      {/* End source files */}

      {/* Begin choose LLM guardrail */}

      <span className="block text-lg font-semibold mb-2">LLM guardrail</span>
      <div className="mb-4">
        <label htmlFor="llmGuardrail" className="block text-sm font-medium text-gray-700">Would you like to add LLM guardrail?</label>
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

      {/* Begin choose to use existing LLM guardrail */}

      {llmGuardrail === 'Yes' && (
        <div className="mb-4">
          <label htmlFor="useExistingGuardrail" className="block text-sm font-medium text-gray-700">Use Existing Guardrail?</label>
          <select
            id="useExistingGuardrail"
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
            value={useExistingGuardrail ? useExistingGuardrail : ''}
            onChange={(e) => setUseExistingGuardrail(e.target.value)}
          >
            <option value="">-- Please choose an option --</option>
            <option value="Yes">Yes</option>
            <option value="No">No, create a new one</option>
          </select>
        </div>
      )}

      {/* End choose to use existing LLM guardrail */}

      {/* End choose LLM guardrail */}

      {/* Begin chat interface */}
      <span className="block text-lg font-semibold mb-2">Chat with LLM</span>
      <div className="mb-4">
        <label htmlFor="llmType" className="block text-sm font-medium text-gray-700">Choose your LLM</label>
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

        {
          llmType
          &&
          <>
            <div className="border border-gray-300 rounded-md p-4 h-64 overflow-y-scroll mt-4">          
              {chatHistory.map((chat, index) => (
                <div key={index} className={`mb-2 ${chat.sender === 'User' ? 'text-right' : 'text-left'}`}>
                  <span className={`inline-block px-4 py-2 rounded-md ${chat.sender === 'User' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-black'}`}>
                    <strong>{chat.sender}:</strong> {chat.message}
                  </span>
                </div>
              ))}
            </div>
            <div className="mt-4 flex">
              <input
                type="text"
                className="flex-grow border border-gray-300 rounded-l-md p-2"
                placeholder="Type your message..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
              />
              <button
                className="bg-blue-500 text-white px-4 py-2 rounded-r-md"
                onClick={handleSendMessage}
              >
                Send
              </button>
            </div>
          </>
        }
      </div>

      {/* End chat interface */}
    </div>
  );
};

export default RAGQuestionForm;
