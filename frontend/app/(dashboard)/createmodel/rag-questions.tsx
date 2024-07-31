import Link from 'next/link';
import React, { useState, useEffect } from 'react';
import { SelectModel } from '@/lib/db';
import NERQuestions from './nlp-questions/ner-questions';
import SemanticSearchQuestions from './semantic-search-questions';

const RAGQuestions = ({
  models,
}: {
  models: SelectModel[];
}) => {

  console.log('All models:', models);

  // Begin state variables & func for source
  const [semanticSearchModels, setSemanticSearchModels] = useState<SelectModel[]>([]);
  const [semanticSearchModelToUse, setSemanticSearchModelToUse] = useState<string|null>(null);
  const [ifUseExistingSemanticSearch, setUseExistingSemanticSearch] = useState<string|null>(null);
  const [sources, setSources] = useState<Array<{ type: string, value: string }>>([]);

  useEffect(() => {
    setSemanticSearchModels(models.filter(model => model.type === 'semantic search model'));
  }, [models]);

  console.log('Filtered Semantic Search Models:', semanticSearchModels);

  // End state variables & func for source

  // Begin state variables & func for LLM guardrail

  const [llmGuardrail, setLlmGuardrail] = useState('');
  const [nerModels, setNerModels] = useState<SelectModel[]>([]);
  const [ifUseExistingGuardrail, setIfUseExistingGuardrail] = useState<string|null>(null);
  const [nerModelToUse, setNerModelToUse] = useState<string|null>(null);

  useEffect(() => {
    setNerModels(models.filter(model => model.type === 'ner model'));
  }, [models]);

  console.log('Filtered NER Models:', nerModels);

  // End state variables & func for LLM guardrail

  const [modelName, setModelName] = useState('')

  // Begin state variables & func for LLM

  const [llmType, setLlmType] = useState<string|null>(null);

  // End state variables & func for LLM

  return (
    <div>
      {/* Begin source files */}
      <div className="mb-4">
        <span className="block text-lg font-semibold mb-2">Search Model</span>
        <label htmlFor="useExistingSemanticSearch" className="block text-sm font-medium text-gray-700">Use an existing semantic search model?</label>
        <select
          id="useExistingSemanticSearch"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={ifUseExistingSemanticSearch ? ifUseExistingSemanticSearch : ''}
          onChange={(e) => setUseExistingSemanticSearch(e.target.value)}
        >
          <option value="">-- Please choose an option --</option>
          <option value="Yes">Yes</option>
          <option value="No">No, create a new one</option>
        </select>
      </div>

      {/* Begin existing Semantic Search Models Dropdown */}

      {ifUseExistingSemanticSearch === 'Yes' && (
        <div className="mb-4">
          <label htmlFor="semanticSearchModels" className="block text-sm font-medium text-gray-700">
            Choose from existing semantic search model(s)
          </label>
          <select
            id="semanticSearchModels"
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
            value={semanticSearchModelToUse || ''}
            onChange={(e) => setSemanticSearchModelToUse(e.target.value)}
          >
            <option value="">-- Please choose a model --</option>
            {semanticSearchModels.map(model => (
              <option key={model.id} value={model.id}>
                {`${model.model_name}`}
              </option>
            ))}
          </select>
        </div>
      )}

      {/* End existing Semantic Search Models Dropdown */}

      {
        ifUseExistingSemanticSearch === 'No' &&
        <SemanticSearchQuestions/>
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
          <label htmlFor="useExistingGuardrail" className="block text-sm font-medium text-gray-700">Use an existing NER model for LLM guardrail?</label>
          <select
            id="useExistingGuardrail"
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
            value={ifUseExistingGuardrail ? ifUseExistingGuardrail : ''}
            onChange={(e) => setIfUseExistingGuardrail(e.target.value)}
          >
            <option value="">-- Please choose an option --</option>
            <option value="Yes">Yes</option>
            <option value="No">No, create a new one</option>
          </select>
        </div>
      )}

      {/* End choose to use existing LLM guardrail */}

      {/* Begin creating a new NER model */}
      {llmGuardrail === 'Yes' && ifUseExistingGuardrail === 'No' && (
        <NERQuestions />
      )}
      {/* Begin creating a new NER model */}

      {/* Begin existing NER Models Dropdown */}
      {llmGuardrail === 'Yes' && ifUseExistingGuardrail === 'Yes' && (
        <div className="mb-4">
          <label htmlFor="nerModels" className="block text-sm font-medium text-gray-700">
            Choose from existing NER Model(s)
          </label>
          <select
            id="nerModels"
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
            value={nerModelToUse || ''}
            onChange={(e) => setNerModelToUse(e.target.value)}
          >
            <option value="">-- Please choose a model --</option>
            {nerModels.map(model => (
              <option key={model.id} value={model.id}>
                {`${model.model_name}`}
              </option>
            ))}
          </select>
        </div>
      )}

      {/* End existing NER Models Dropdown */}

      {/* End choose LLM guardrail */}

      {/* Begin chat interface */}
      <span className="block text-lg font-semibold mb-2">Chat</span>
      <div className="mb-4">
        <label htmlFor="llmType" className="block text-sm font-medium text-gray-700">Choose an LLM option</label>
        <select
          id="llmType"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={llmType ? llmType : ''}
          onChange={(e) => setLlmType(e.target.value)}
        >
          <option value="">-- Please choose an option --</option>
          <option value="OpenAI">OpenAI</option>
          <option value="Llama">Llama</option>
          <option value="Self-host">Self-host</option>
        </select>
      </div>

      {/* End chat interface */}

      {/* Add Model Name Input Field */}
      <span className="block text-lg font-semibold mb-2">Name your model</span>
      <div className="mb-4">
        <label htmlFor="modelName" className="block text-sm font-medium text-gray-700">
          Model Name
        </label>
        <input
          type="text"
          id="modelName"
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          value={modelName || ''}
          onChange={(e) => setModelName(e.target.value)}
          placeholder="Enter model name"
        />
      </div>


      {/* Button to create and deploy */}
      {
        (semanticSearchModelToUse || sources) && (llmGuardrail === 'No' || (nerModelToUse || ifUseExistingGuardrail === 'No')) && llmType && modelName &&
        <div className="flex justify-center">
          <Link href="/">
          <button
            type="button"
            className="mb-4 bg-blue-500 text-white px-4 py-2 rounded-md"
            // onClick={async () => {

            //   let link = '', description = '';
            //   let model_stats: 'training' | 'active' | 'inactive' | 'archived' = 'training';

            //   if (llmGuardrail === 'No') {
            //     link = 'http://40.86.17.143/search?id=25fa3653-7fff-3366-ab44-532696fc6ae1&useGuardrail=false'
            //     description = 'This is an RAG model.'
            //   } else {
            //     link = 'http://40.86.17.143/search?id=25fa3653-7fff-3366-ab44-532696fc6ae1&useGuardrail=true'
            //     description = 'This is an RAG model trained by composing a semantic search model and an NER model as LLM guardrail model.'
            //   }

            //   if (ifUseExistingSemanticSearch === 'No' || (llmGuardrail === 'Yes' && ifUseExistingGuardrail === 'No')) {
            //     model_stats = 'training'
            //   } else {
            //     model_stats = 'active'
            //   }

            //   const modelData: Omit<SelectModel, 'id'> = {
            //     imageUrl: '/thirdai-small.png',
            //     name: modelName,
            //     status: model_stats,
            //     trainedAt: new Date(), // Use current date and time
            //     description: description,
            //     deployEndpointUrl: link,
            //     onDiskSizeKb: (300 * 1024).toString(),  // 300 MB converted to KB as string
            //     ramSizeKb: (300 * 1024 * 2).toString(),  // 300 * 2 MB converted to KB as string
            //     numberParameters: 51203077,
            //     rlhfCounts: 0,
            //     type: 'rag model'
            //   };

            //   try {
            //     const response = await fetch('/api/insertModel', {
            //       method: 'POST',
            //       headers: {
            //         'Content-Type': 'application/json'
            //       },
            //       body: JSON.stringify(modelData)
            //     });
          
            //     if (response.ok) {
            //       const result = await response.json();
            //       console.log('Model inserted:', result);

            //       if (ifUseExistingSemanticSearch === 'No' || (llmGuardrail === 'Yes' && ifUseExistingGuardrail === 'No')) {
            //         // if training is going on, don't redirect
            //       } else {
            //         window.open(link, '_blank');
            //       }
            //     } else {
            //       const error = await response.json();
            //       console.error('Failed to insert model:', error);
            //     }
            //   } catch (error) {
            //     console.error('Error inserting model:', error);
            //   }
            // }}
          >
            {`${ifUseExistingSemanticSearch === 'No' || (llmGuardrail === 'Yes' && ifUseExistingGuardrail === 'No') ? 'Create' : 'Create and Deploy'}`}
          </button>
          </Link>
        </div>
      }
    </div>
  );
};

export default RAGQuestions;
