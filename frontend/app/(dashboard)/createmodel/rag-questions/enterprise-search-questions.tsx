import React, { useState, useEffect } from 'react';
import { Workflow } from '@/lib/backend';
import { CardDescription } from '@/components/ui/card';
import { Button, TextField } from '@mui/material';
import { useRouter } from 'next/navigation';
import DropdownMenu from '@/components/ui/dropDownMenu';
import { create_enterprise_search_workflow, EnterpriseSearchOptions } from '@/lib/backend';
import SemanticSearchQuestions from '../semantic-search-questions';
import NERQuestions from '../nlp-questions/ner-questions';

interface EnterpriseSearchQuestionsProps {
  models: Workflow[];
  workflowNames: string[];
}

enum LlmProvider {
  OpenAI = 'openai',
  OnPrem = 'on-prem',
  SelfHosted = 'self-hosted',
  None = 'none',
}

const EnterpriseSearchQuestions: React.FC<EnterpriseSearchQuestionsProps> = ({ models, workflowNames }) => {
  const [currentStep, setCurrentStep] = useState(0);

  // Knowledge base state
  const [ifUseExistingSS, setUseExistingSS] = useState<string | null>(null);
  const [existingSSmodels, setExistingSSmodels] = useState<Workflow[]>([]);
  const [ssIdentifier, setSsIdentifier] = useState<string | null>(null);
  const [ssModelId, setSsModelId] = useState<string | null>(null);
  const [createdSS, setCreatedSS] = useState<boolean>(false);

  useEffect(() => {
    setExistingSSmodels(models.filter((model) => model.type === 'ndb'));
  }, [models]);

  // LLM Guardrail state
  const [ifUseLGR, setIfUseLGR] = useState('');
  const [ifUseExistingLGR, setIfUseExistingLGR] = useState<string | null>(null);
  const [existingNERModels, setExistingNERModels] = useState<Workflow[]>([]);
  const [grIdentifier, setGrIdentifier] = useState<string | null>(null);
  const [grModelId, setGrModelId] = useState<string | null>(null);
  const [createdGR, setCreatedGR] = useState<boolean>(false);

  useEffect(() => {
    setExistingNERModels(
      models.filter((model) => model.type === 'udt' && model.sub_type === 'token')
    );
  }, [models]);

  const [modelName, setModelName] = useState('');
  const [llmType, setLlmType] = useState<LlmProvider | null>(null);
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [warningMessage, setWarningMessage] = useState('');

  const handleSubmit = async () => {
    setIsLoading(true);

    try {
      let options: EnterpriseSearchOptions = {
        retrieval_id: ssModelId || '',
        guardrail_id: grModelId || '',
        llm_provider: '',
        default_mode: 'search',
      };

      if (llmType && llmType !== LlmProvider.None) {
        switch (llmType) {
          case LlmProvider.OpenAI:
            options.llm_provider = 'openai';
            break;
          case LlmProvider.OnPrem:
            options.llm_provider = 'on-prem';
            break;
          case LlmProvider.SelfHosted:
            options.llm_provider = 'self-host';
            break;
        }
      }

      options = Object.fromEntries(
        Object.entries(options).filter(([_, v]) => v !== undefined && v !== '')
      ) as EnterpriseSearchOptions;

      const workflowResponse = await create_enterprise_search_workflow({
        workflow_name: modelName,
        options,
      });
      console.log('Workflow created:', workflowResponse.data.model_id);
      router.push('/');
    } catch (error) {
      console.error('Error during workflow creation:', error);
      alert('Error during workflow creation: ' + error);
      setIsLoading(false);
    }
  };

  const modelDropDownList = existingSSmodels.map((model) => ({
    id: model.model_id,
    name: model.username + '/' + model.model_name,
  }));

  const grDropDownList = existingNERModels.map((model) => ({
    id: model.model_id,
    name: model.username + '/' + model.model_name,
  }));

  const handleSSIdentifier = (ssID: string) => {
    setSsIdentifier(ssID);
    const ssModel = existingSSmodels.find(
      (model) => `${model.username}/${model.model_name}` === ssID
    );
    if (ssModel) {
      setSsModelId(ssModel.model_id);
    }
  };

  const handleGrIdentifier = (grID: string) => {
    setGrIdentifier(grID);
    const grModel = existingNERModels.find(
      (model) => `${model.username}/${model.model_name}` === grID
    );
    if (grModel) {
      setGrModelId(grModel.model_id);
    }
  };

  const steps = [
    {
      title: 'App Name',
      content: (
        <div>
          <TextField
            className="text-md w-full"
            value={modelName}
            onChange={(e) => {
              const name = e.target.value;
              const regexPattern = /^[\w-]+$/;
              let warningMessage = '';

              if (name.includes(' ')) {
                warningMessage = 'The app name cannot contain spaces. Please remove the spaces.';
              }
              else if (name.includes('.')) {
                warningMessage = "The app name cannot contain periods ('.'). Please remove the periods.";
              }
              else if (!regexPattern.test(name)) {
                warningMessage = 'The app name can only contain letters, numbers, underscores, and hyphens. Please modify the name.';
              }
              else if (workflowNames.includes(name)) {
                warningMessage = 'An app with the same name already exists. Please choose a different name.';
              }

              setWarningMessage(warningMessage);
              setModelName(name);
            }}
            placeholder="Enter app name"
            style={{ marginTop: '10px' }}
          />
          {warningMessage && (
            <span style={{ color: 'red', marginTop: '10px' }}>{warningMessage}</span>
          )}
        </div>
      ),
    },
    {
      title: 'Knowledge base',
      content: (
        <div>
          {!createdSS && (
            <>
              <CardDescription>Use an existing knowledge base?</CardDescription>
              <div style={{ display: 'flex', flexDirection: 'row', gap: '10px', marginTop: '10px' }}>
                <Button
                  variant={ifUseExistingSS === 'Yes' ? 'contained' : 'outlined'}
                  onClick={() => {
                    setUseExistingSS('Yes');
                    setCreatedSS(false);
                  }}
                >
                  Yes
                </Button>
                <Button
                  variant={ifUseExistingSS === 'No' ? 'contained' : 'outlined'}
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
            <div className="mb-4 mt-2">
              <CardDescription>Choose from an existing Knowledge Base</CardDescription>
              <div className="mt-2">
                <DropdownMenu
                  title="Choose a Knowledge Base  "
                  handleSelectedTeam={handleSSIdentifier}
                  teams={modelDropDownList}
                />
              </div>
            </div>
          )}

          {ifUseExistingSS === 'No' && (
            <>
              {createdSS ? (
                <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg shadow-sm">
                  <div className="flex items-center justify-center space-x-3">
                    <span className="text-lg font-medium text-blue-800">Knowledge base queued</span>
                  </div>
                  <div className="mt-2 text-sm text-blue-600 text-center">
                    You may continue to the next step while this processes
                  </div>
                </div>
              ) : (
                <div>
                  <SemanticSearchQuestions
                    models={models}
                    workflowNames={workflowNames}
                    onCreateModel={(modelID) => {
                      setSsModelId(modelID);
                      setCreatedSS(true);
                    }}
                    stayOnPage={true}
                    appName={`${modelName}-KB`}
                  />
                </div>
              )}
            </>
          )}
        </div>
      ),
    },
    {
      title: 'LLM',
      content: (
        <div>
          <div>
            <span className="block text-lg font-semibold" style={{ marginTop: '20px' }}>
              LLM (Optional)
            </span>
            <div>
              <CardDescription>Choose an LLM for generating answers</CardDescription>
              <div style={{ display: 'flex', flexDirection: 'row', gap: '10px', marginTop: '10px' }}>
                <Button
                  variant={llmType === LlmProvider.OpenAI ? 'contained' : 'outlined'}
                  onClick={() => setLlmType(LlmProvider.OpenAI)}
                >
                  OpenAI
                </Button>
                <Button
                  variant={llmType === LlmProvider.OnPrem ? 'contained' : 'outlined'}
                  onClick={() => setLlmType(LlmProvider.OnPrem)}
                >
                  On-prem
                </Button>
                <Button
                  variant={llmType === LlmProvider.SelfHosted ? 'contained' : 'outlined'}
                  onClick={() => setLlmType(LlmProvider.SelfHosted)}
                >
                  Self-host
                </Button>
                <Button
                  variant={llmType === LlmProvider.None ? 'contained' : 'outlined'}
                  onClick={() => setLlmType(LlmProvider.None)}
                >
                  None
                </Button>
              </div>
            </div>
          </div>

          <div>
            <span className="block text-lg font-semibold" style={{ marginTop: '20px' }}>
              LLM Guardrail (Optional)
            </span>
            {!createdGR && (
              <>
                <CardDescription>
                  Would you like to redact PII (Personally Identifiable Information) from your
                  references?
                </CardDescription>
                <div style={{ display: 'flex', flexDirection: 'row', gap: '10px', marginTop: '10px' }}>
                  <Button
                    variant={ifUseLGR === 'Yes' ? 'contained' : 'outlined'}
                    onClick={() => {
                      setIfUseLGR('Yes');
                      setCreatedGR(false);
                    }}
                  >
                    Yes
                  </Button>
                  <Button
                    variant={ifUseLGR === 'No' ? 'contained' : 'outlined'}
                    onClick={() => {
                      setGrIdentifier(null);
                      setIfUseLGR('No');
                      setCreatedGR(false);
                    }}
                  >
                    No
                  </Button>
                </div>

                {ifUseLGR === 'Yes' && (
                  <>
                    <div style={{ marginTop: '20px' }}>
                      <CardDescription>
                        Use an existing NER model to reduce PII from your reference?
                      </CardDescription>
                      <div style={{ display: 'flex', flexDirection: 'row', gap: '10px', marginTop: '10px' }}>
                        <Button
                          variant={ifUseExistingLGR === 'Yes' ? 'contained' : 'outlined'}
                          onClick={() => {
                            setIfUseExistingLGR('Yes');
                            setCreatedGR(false);
                          }}
                        >
                          Yes
                        </Button>
                        <Button
                          variant={ifUseExistingLGR === 'No' ? 'contained' : 'outlined'}
                          onClick={() => {
                            setIfUseExistingLGR('No');
                            setCreatedGR(false);
                          }}
                        >
                          No
                        </Button>
                      </div>
                    </div>

                    {ifUseExistingLGR === 'Yes' && (
                      <div style={{ marginTop: '20px' }}>
                        <div className="mb-4 mt-2">
                          <CardDescription>Choose from existing NLP App(s)</CardDescription>
                          <div className="mt-2">
                            <DropdownMenu
                              title=" Please choose a model  "
                              handleSelectedTeam={handleGrIdentifier}
                              teams={grDropDownList}
                            />
                          </div>
                        </div>
                      </div>
                    )}

                    {ifUseExistingLGR === 'No' && (
                      <>
                        {createdGR ? (
                          <div>Guardrail model created.</div>
                        ) : (
                          <div>
                            <NERQuestions
                              workflowNames={workflowNames}
                              modelGoal="Model to detect sensitive PII"
                              onCreateModel={(modelID) => {
                                setGrModelId(modelID);
                                setCreatedGR(true);
                              }}
                              stayOnPage={true}
                              appName={`${modelName}-NER`}
                            />
                          </div>
                        )}
                      </>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        </div>
      ),
    },
  ];

  const missingRequirements = [];
  if (!modelName) missingRequirements.push('App Name is not specified (Step 1)');
  if (!ssModelId) missingRequirements.push('Retrieval app is not specified (Step 2)');

  const errorMessage = missingRequirements.length > 0 && (
    <div>
      {`Please go back and specify the following:`}
      <br />
      {missingRequirements.map((requirement, index) => (
        <span key={index}>
          {'• '}
          {requirement}
          <br />
        </span>
      ))}
    </div>
  );

  return (
    <div>
      <div
        className="mb-4"
        style={{
            display: 'flex',
            flexWrap: 'wrap',
            alignItems: 'flex-start',
            rowGap: '15px',
            columnGap: '15px',
          }}
        >
          {steps.map((step, index) => (
            <Button
              key={index}
              variant={index === currentStep ? 'contained' : 'outlined'}
              onClick={() => setCurrentStep(index)}
              style={{
                marginBottom: '10px',
                minWidth: '140px',
                height: '40px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                textTransform: 'none',
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                padding: '0 16px',
              }}
            >
              {step.title}
            </Button>
          ))}
        </div>
  
        {/* Step Content */}
        <div>{steps[currentStep].content}</div>
  
        {/* Step Controls */}
        <div style={{ marginTop: '50px', display: 'flex', justifyContent: 'space-between' }}>
          {currentStep > 0 ? (
            <Button onClick={() => setCurrentStep(currentStep - 1)}>Previous</Button>
          ) : (
            <div></div>
          )}
  
          {currentStep < steps.length - 1 ? (
            <Button onClick={() => setCurrentStep(currentStep + 1)}>Next</Button>
          ) : (
            <>
              {ssModelId && modelName ? (
                <div>
                  <Button
                    onClick={handleSubmit}
                    style={{ width: '100%' }}
                    disabled={isLoading || !(ssModelId && modelName)}
                  >
                    {isLoading ? (
                      <div className="flex items-center justify-center">
                        <div className="animate-spin rounded-full h-4 w-4 border-t-2 border-b-2 border-blue-500 mr-2"></div>
                        <span>Creating...</span>
                      </div>
                    ) : (
                      'Create'
                    )}
                  </Button>
                </div>
              ) : (
                <div style={{ color: 'red' }}>{errorMessage}</div>
              )}
            </>
          )}
        </div>
      </div>
    );
  };

export default EnterpriseSearchQuestions;
