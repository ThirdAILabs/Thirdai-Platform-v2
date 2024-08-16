import Link from 'next/link';
import React, { useState, useEffect } from 'react';
import { SelectModel } from '@/lib/db';
import NERQuestions from './nlp-questions/ner-questions';
import SemanticSearchQuestions from './semantic-search-questions';
import { create_workflow, add_models_to_workflow } from '@/lib/backend';

const RAGQuestions = ({
  models,
}: {
  models: SelectModel[];
}) => {

  console.log('All models:', models);

  // Begin state variables & func for source
  const [ifUseExistingSS, setUseExistingSS] = useState<string|null>(null);
  const [existingSSmodels, setExistingSSmodels] = useState<SelectModel[]>([]);
  const [ssIdentifier, setSsIdentifier] = useState<string | null>(null);
  const [createdSS, setCreatedSS] = useState<boolean>(false);
  
  useEffect(() => {
    setExistingSSmodels(models.filter(model => model.type === 'ndb'));
  }, [models]);
  
  console.log('Existing Semantic Search Models:', existingSSmodels);
  
  // End state variables & func for source
  
  // Begin state variables & func for LLM guardrail
  
  const [ifUseLGR, setIfUseLGR] = useState('');
  const [ifUseExistingLGR, setIfUseExistingLGR] = useState<string|null>(null);
  const [existingNERModels, setExistingNERModels] = useState<SelectModel[]>([]);
  const [grIdentifier, setGrIdentifier] = useState<string | null>(null);
  const [createdGR, setCreatedGR] = useState<boolean>(false);

  useEffect(() => {
    setExistingNERModels(models.filter(model => model.type === 'udt'));
  }, [models]);

  console.log('Existing NER Models:', existingNERModels);

  // End state variables & func for LLM guardrail

  const [modelName, setModelName] = useState('')

  // Begin state variables & func for LLM

  const [llmType, setLlmType] = useState<string|null>(null);

  // End state variables & func for LLM

  const handleSubmit = async () => {
    const workflowName = `Workflow for ${modelName}`;
    const workflowTypeName = 'rag';
  
    try {
      // Step 1: Create the workflow
      const workflowResponse = await create_workflow({ name: workflowName, typeName: workflowTypeName });
      const workflowId = workflowResponse.data.workflow_id;
      console.log('Workflow created:', workflowId);
  
      // Step 2: Prepare the models to be added
      const modelIdentifiers = [];
      const components = [];
  
      // Find and add the semantic search model
      if (ssIdentifier) {
        const ssModel = existingSSmodels.find(model => `${model.username}/${model.model_name}` === ssIdentifier);
        if (ssModel) {
          modelIdentifiers.push(ssModel.model_id);
          components.push('search');
        } else {
          console.error(`Semantic search model with identifier ${ssIdentifier} not found.`);
        }
      }

      // Find and add the NER model if it exists
      if (grIdentifier) {
        const nerModel = existingNERModels.find(model => `${model.username}/${model.model_name}` === grIdentifier);
        if (nerModel) {
          modelIdentifiers.push(nerModel.model_id);
          components.push('nlp');
        } else {
          console.error(`NER model with identifier ${grIdentifier} not found.`);
        }
      }
  
      // Step 3: Add the models to the workflow
      if (modelIdentifiers.length > 0) {
        const addModelsResponse = await add_models_to_workflow({
          workflowId,
          modelIdentifiers,
          components,
        });
        console.log('Models added to workflow:', addModelsResponse);
      } else {
        console.error('No models to add to the workflow');
      }
    } catch (error) {
      console.error('Error during workflow creation or model addition:', error);
    }
  };

  return (
    <div>
      {/* Begin Semantic Search Model */}

            <div className="mb-4">
              <span className="block text-lg font-semibold mb-2">Retrieval App</span>
              <label htmlFor="useExistingSemanticSearch" className="block text-sm font-medium text-gray-700">Use an existing semantic search model?</label>
              <select
                id="useExistingSemanticSearch"
                className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                value={ifUseExistingSS ? ifUseExistingSS : ''}
                onChange={(e) => {
                  setUseExistingSS(e.target.value);
                  setCreatedSS(false);
                }}
              >
                <option value="">-- Please choose an option --</option>
                <option value="Yes">Yes</option>
                <option value="No">No, create a new one</option>
              </select>
            </div>

            {/* Begin existing Semantic Search Models Dropdown */}

            {ifUseExistingSS === 'Yes' && (
              <div className="mb-4">
                <label htmlFor="semanticSearchModels" className="block text-sm font-medium text-gray-700">
                  Choose from existing semantic search model(s)
                </label>
                <select
                  id="semanticSearchModels"
                  className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                  value={ssIdentifier || ''}
                  onChange={(e) => {
                    setSsIdentifier(e.target.value);
                  }}
                >
                  <option value="">-- Please choose a model --</option>
                  {existingSSmodels.map((model, index) => (
                    <option key={index} value={`${model.username}/${model.model_name}`}>
                      {`${model.username}/${model.model_name}`}
                    </option>
                  ))}
                </select>
              </div>
            )}

              {/* End existing Semantic Search Models Dropdown */}

              {/* Begin Create new Semantic Search Model */}

      {
        ifUseExistingSS === 'No' &&
        <>
          <div style={{visibility: createdSS ? 'hidden' : 'visible'}}>
            <SemanticSearchQuestions 
              onCreateModel={(username, modelName) => {
                // TODO: SOMEHOW GET CURRENT USERNAME
                console.log("username", username);
                setSsIdentifier(`${username}/${modelName}`);
                setCreatedSS(true);
              }}
              stayOnPage
              />
          </div>
          <div style={{visibility: createdSS ? 'visible' : 'hidden'}}>
            Semantic search model created.
          </div>
        </>
      }
              {/* End Create new Semantic Search Model */}

      {/* End Semantic Search Model */}




      {/* Begin choose LLM guardrail */}

            <span className="block text-lg font-semibold mb-2">LLM guardrail</span>
            <div className="mb-4">
              <label htmlFor="llmGuardrail" className="block text-sm font-medium text-gray-700">Would you like to add LLM guardrail?</label>
              <select
                id="llmGuardrail"
                className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                value={ifUseLGR}
                onChange={(e)=>{
                  if (e.target.value === "No") {
                    setGrIdentifier(null);
                  }
                  setIfUseLGR(e.target.value);
                  setCreatedGR(false);
                }}
              >
                <option value="">-- Please choose an option --</option>
                <option value="Yes">Yes</option>
                <option value="No">No</option>
              </select>
            </div>

            {/* Begin choose to use existing LLM guardrail */}

            {ifUseLGR === 'Yes' && (
              <div className="mb-4">
                <label htmlFor="useExistingGuardrail" className="block text-sm font-medium text-gray-700">Use an existing NER model for LLM guardrail?</label>
                <select
                  id="useExistingGuardrail"
                  className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                  value={ifUseExistingLGR ? ifUseExistingLGR : ''}
                  onChange={(e) => {
                    setIfUseExistingLGR(e.target.value);
                    setCreatedGR(false);
                  }}
                >
                  <option value="">-- Please choose an option --</option>
                  <option value="Yes">Yes</option>
                  <option value="No">No, create a new one</option>
                </select>
              </div>
            )}

            {/* End choose to use existing LLM guardrail */}

            {/* Begin creating a new NER model */}

            {ifUseLGR === 'Yes' && ifUseExistingLGR === 'No' && (
              <>
                <div style={{visibility: createdGR ? 'hidden' : 'visible'}}>
                  <NERQuestions 
                    onCreateModel={(username, modelName) => {
                      // TODO: SOMEHOW GET USERNAME
                      console.log("username", username);
                      setGrIdentifier(`${username}/${modelName}`);
                      setCreatedGR(true);
                    }} 
                    stayOnPage
                  />
                </div>
                <div style={{visibility: createdGR ? 'visible' : 'hidden'}}>
                  Guardrail model created.
                </div>
              </>
            )}

            {/* Begin creating a new NER model */}

            {/* Begin existing NER Models Dropdown */}
            {ifUseLGR === 'Yes' && ifUseExistingLGR === 'Yes' && (
              <div className="mb-4">
                <label htmlFor="nerModels" className="block text-sm font-medium text-gray-700">
                  Choose from existing NLP App(s)
                </label>
                <select
                  id="nerModels"
                  className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                  value={grIdentifier ? grIdentifier : ''}
                  onChange={(e) => setGrIdentifier(e.target.value)}
                >
                  <option value="">-- Please choose a model --</option>
                  {existingNERModels.map(model => (
                    <option key={model.id} value={`${model.username}/${model.model_name}`}>
                      {`${model.username}/${model.model_name}`}
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




      {/* Begin Model Name Input Field */}
            <span className="block text-lg font-semibold mb-2">Name your App</span>
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
                placeholder="Enter App name"
              />
            </div>
      {/* End Model Name Input Field */}



      {/* Begin create and deploy */}
            {
              (ssIdentifier) 
              && 
              (ifUseLGR === 'No' || grIdentifier) 
              && 
              llmType 
              && 
              modelName 
              &&
              <div className="flex justify-center">
                <Link href="/">
                <button
                  type="button"
                  className="mb-4 bg-blue-500 text-white px-4 py-2 rounded-md"
                  onClick={handleSubmit}
                >
                  {`${ifUseExistingSS === 'No' || (ifUseLGR === 'Yes' && ifUseExistingLGR === 'No') ? 'Create' : 'Create and Deploy'}`}
                </button>
                </Link>
              </div>
            }
      {/* End create and deploy */}

    </div>
  );
};

export default RAGQuestions;
