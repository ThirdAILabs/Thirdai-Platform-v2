'use client'

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button'; // Import the necessary UI components
import { Input } from '@/components/ui/input'; // Import the necessary UI components
import { CardDescription } from '@/components/ui/card';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useRouter } from 'next/navigation'; // Correct router for /app directory
import { Workflow, getWorkflowDetails, WorkflowDetailsResponse } from "@/lib/backend"

export default function EditWorkflowPage({ params }: { params: { workflow_id: string } }) {
  const { workflow_id } = params;
  const [workflowDetails, setWorkflowDetails] = useState<WorkflowDetailsResponse['data'] | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [currentStep, setCurrentStep] = useState(0);
  const [ssModelId, setSsModelId] = useState<string | null>(null);
  const [grModelId, setGrModelId] = useState<string | null>(null);
  const [llmType, setLlmType] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [ifUseExistingSS, setUseExistingSS] = useState<string | null>(null);
  const [createdSS, setCreatedSS] = useState(false);
  const [ifUseLGR, setIfUseLGR] = useState<string | null>(null);
  const [createdGR, setCreatedGR] = useState(false);
  const [workflowNames, setWorkflowNames] = useState<string[]>([]);

  useEffect(() => {
    async function fetchWorkflowDetails() {
      try {
        const { data } = await getWorkflowDetails(workflow_id); // Call the new API function and extract the data
        setWorkflowDetails(data); // Set only the data property
        setLoading(false);
      } catch (error) {
        setError('Failed to load workflow details');
        setLoading(false);
      }
    }

    if (workflow_id) {
      fetchWorkflowDetails();
    }
  }, [workflow_id]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  if (!workflowDetails) {
    return <div>No workflow details found</div>;
  }

  // Workflow steps
  const steps = [
    {
      title: 'App Name',
      content: (
        <div>
          <span className="block text-lg font-semibold">App Name</span>
          {/* Show the workflow name, but make it non-editable */}
          <Input
            className="text-md"
            value={workflowDetails.name} // Access the name from workflowDetails
            readOnly
            style={{ marginTop: '10px' }}
          />
        </div>
      ),
    },
    {
      title: 'Retrieval App',
      content: (
        <div>
          <span className="block text-lg font-semibold">Retrieval App</span>
          {!createdSS && (
            <>
              <CardDescription>Use an existing retrieval app?</CardDescription>
              <div style={{ display: 'flex', flexDirection: 'row', gap: '10px', marginTop: '10px' }}>
                <Button
                  variant={ifUseExistingSS ? (ifUseExistingSS === 'Yes' ? 'secondary' : 'outline') : 'default'}
                  onClick={() => {
                    setUseExistingSS('Yes');
                    setCreatedSS(false);
                  }}
                >
                  Yes
                </Button>
                <Button
                  variant={ifUseExistingSS ? (ifUseExistingSS === 'No' ? 'secondary' : 'outline') : 'default'}
                  onClick={() => {
                    setUseExistingSS('No');
                    setCreatedSS(false);
                  }}
                >
                  No, create a new one
                </Button>
              </div>
            </>
          )}

          {ifUseExistingSS === 'Yes' && (
            <div className="mb-4">
              <CardDescription>Choose from existing semantic search model(s)</CardDescription>
              <select
                id="semanticSearchModels"
                className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                value={ssModelId || ''}
                onChange={(e) => {
                  const ssID = e.target.value;
                  setSsModelId(ssID);
                  // Logic to find the chosen semantic search model by ID and update the state
                }}
              >
                <option value="">-- Please choose a model --</option>
                {/* Replace this with actual models */}
                {['model1', 'model2'].map((model, index) => (
                  <option key={index} value={model}>
                    {model}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
      ),
    },
    {
      title: 'LLM Guardrail',
      content: (
        <div>
          <span className="block text-lg font-semibold" style={{ marginTop: '20px' }}>
            LLM Guardrail
          </span>
          {/* Guardrail logic here */}
        </div>
      ),
    },
    {
      title: 'Chat',
      content: (
        <div>
          <span className="block text-lg font-semibold" style={{ marginTop: '20px' }}>
            Chat
          </span>
          <div>
            <CardDescription>Choose an LLM option</CardDescription>
            <div style={{ display: 'flex', flexDirection: 'row', gap: '10px', marginTop: '10px' }}>
              <Button
                variant={llmType ? (llmType === 'OpenAI' ? 'secondary' : 'outline') : 'default'}
                onClick={() => setLlmType('OpenAI')}
              >
                OpenAI
              </Button>
              <Button
                variant={llmType ? (llmType === 'On-prem' ? 'secondary' : 'outline') : 'default'}
                onClick={() => setLlmType('On-prem')}
              >
                On-prem
              </Button>
              <Button
                variant={llmType ? (llmType === 'Self-host' ? 'secondary' : 'outline') : 'default'}
                onClick={() => setLlmType('Self-host')}
              >
                Self-host
              </Button>
            </div>
          </div>
        </div>
      ),
    },
  ];

  return (
    <div>
      {/* Step Navigation */}
      <div className="mb-4">
        {steps.map((step, index) => (
          <Button
            key={index}
            variant={index === currentStep ? 'secondary' : 'outline'}
            onClick={() => setCurrentStep(index)}
            style={{ marginRight: '10px' }}
          >
            {step.title}
          </Button>
        ))}
      </div>

      {/* Step Content */}
      <div>{steps[currentStep].content}</div>

      {/* Step Controls */}
      <div style={{ marginTop: '50px', display: 'flex', justifyContent: 'space-between' }}>
        {/* Previous Button */}
        {currentStep > 0 ? (
          <Button onClick={() => setCurrentStep(currentStep - 1)}>Previous</Button>
        ) : (
          <div></div>
        )}

        {/* Next Button or Create/Deploy Button */}
        {currentStep < steps.length - 1 ? (
          <Button onClick={() => setCurrentStep(currentStep + 1)}>Next</Button>
        ) : (
          <>
            {ssModelId && llmType ? (
              <Button onClick={() => console.log('Submit workflow')}>
                {isLoading ? 'Creating...' : 'Create'}
              </Button>
            ) : (
              <div style={{ color: 'red' }}>Please complete all required fields</div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

