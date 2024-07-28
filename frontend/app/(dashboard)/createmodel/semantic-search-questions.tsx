import React, { useState } from 'react';

const SemanticSearchQuestions = () => {
    // Begin state variables & func for source
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

    return (
      <div>
        {
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

        <div className="flex justify-center">
          <button
            type="button"
            className="mb-4 bg-blue-500 text-white px-4 py-2 rounded-md"
          >
            Create and Deploy
          </button>
        </div>
      </div>
    );
};

export default SemanticSearchQuestions;
