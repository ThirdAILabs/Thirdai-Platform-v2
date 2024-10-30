// app/NLPQuestions.js
import React, { useState, useEffect } from 'react';
import NERQuestions from './ner-questions';
import SCQQuestions from './sentence-classification-questions';
import { Input } from '@/components/ui/input';
import { Button, TextField } from '@mui/material';
import { CardDescription } from '@/components/ui/card';
import { Divider } from '@mui/material';
import DropdownMenu from '@/components/ui/dropDownMenu';

// Predefined models
const CUSTOM_MODEL = 'Create custom model';
const CUSTOMER_SENTIMENT = 'Customer Review Sentiment';
const DOCUMENT_ENTITIES = 'Document Entity Extractor';

const predefinedModels = [
  { 
    name: CUSTOMER_SENTIMENT,
    type: 'Sentence classification',
    description: 'Automatically analyze customer reviews and feedback to determine sentiment and key themes'
  },
  {
    name: DOCUMENT_ENTITIES,
    type: 'Token classification',
    description: 'Extract important entities like people, organizations, dates, and locations from documents'
  }
];

interface NLPQuestionsProps {
  workflowNames: string[];
}

const NLPQuestions = ({ workflowNames }: NLPQuestionsProps) => {
  const [question, setQuestion] = useState('');
  const [loadingAnswer, setLoadingAnswer] = useState<boolean>(false);
  const [answer, setAnswer] = useState('');
  const [confirmedAnswer, setConfirmedAnswer] = useState<boolean>(false);
  const [hasOpenAIKey, setHasOpenAIKey] = useState<boolean>(false);
  const [selectedModel, setSelectedModel] = useState('');

  useEffect(() => {
    // Check if OpenAI key is configured
    const checkOpenAIKey = async () => {
      try {
        const response = await fetch('/endpoints/check-openai-key');
        const { hasKey } = await response.json();
        setHasOpenAIKey(hasKey);
      } catch (error) {
        console.error('Error checking OpenAI key:', error);
        setHasOpenAIKey(false);
      }
    };

    checkOpenAIKey();
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuestion(e.target.value);
  };

  const handleModelSelect = (model: string) => {
    setSelectedModel(model);
    if (model === CUSTOM_MODEL) {
      setAnswer('');
      setQuestion('');
    } else {
      const modelData = predefinedModels.find(m => m.name === model);
      if (modelData) {
        setAnswer(modelData.type);
        setQuestion(modelData.description);
      }
    }
  };

  const submit = async () => {
    if (!question) {
      console.error('Question is not valid');
      alert('Question is not valid');
      return;
    }
    if (loadingAnswer) {
      return;
    }

    setLoadingAnswer(true);

    try {
      const response = await fetch('/endpoints/which-nlp-use-case', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ question }),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const result = await response.json();
      setAnswer(result.answer);
    } catch (error) {
      console.error('Error during fetch:', error);
      alert('Error during fetch:' + error);
    } finally {
      setLoadingAnswer(false);
    }
  };

  // Create options array for dropdown
  const modelOptions = [
    ...(hasOpenAIKey ? [{ name: CUSTOM_MODEL }] : []),
    ...predefinedModels
  ];

  return (
    <div style={{ width: '100%' }}>
      {!confirmedAnswer && !answer && (
        <>
          <span className="block text-lg font-semibold">NLP task assistant</span>
          <CardDescription>
            Select a pre-built model or create a custom one for your specific needs.
          </CardDescription>
          <div style={{ marginTop: '10px', marginBottom: '20px' }}>
            <DropdownMenu
              title="Select model type"
              handleSelectedTeam={handleModelSelect}
              teams={modelOptions}
            />
          </div>

          {selectedModel === CUSTOM_MODEL && (
            <>
              <CardDescription>
                Say &quot;I want to analyze my customers&apos; reviews&quot;
              </CardDescription>
              <CardDescription>
                or &quot;I want to analyze the individual tokens within a report document&quot;
              </CardDescription>
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'row',
                  gap: '20px',
                  justifyContent: 'space-between',
                  margin: '20px 0',
                }}
              >
                <TextField
                  className="text-md w-full"
                  value={question}
                  onChange={handleInputChange}
                  placeholder="Describe your use case..."
                  onKeyDown={(e) => {
                    if (e.keyCode === 13 && e.shiftKey === false) {
                      e.preventDefault();
                      submit();
                    }
                  }}
                />
              </div>
              <Button
                onClick={submit}
                variant="contained"
                color={loadingAnswer ? 'success' : 'primary'}
                style={{ width: '100%' }}
              >
                {loadingAnswer ? 'Understanding your use case...' : 'Submit'}
              </Button>
            </>
          )}
          {selectedModel && selectedModel !== CUSTOM_MODEL && (
            <Button
              onClick={() => setConfirmedAnswer(true)}
              variant="contained"
              style={{ width: '100%' }}
            >
              Continue with Selected Model
            </Button>
          )}
        </>
      )}
      
      {!confirmedAnswer && answer && (
        <div style={{ marginTop: '20px' }}>
          <span className="block text-lg font-semibold" style={{ marginBottom: '10px' }}>
            Our recommendation
          </span>
          <CardDescription>{answer}</CardDescription>
          <div
            style={{
              width: '100%',
              marginTop: '20px',
              display: 'flex',
              flexDirection: 'row',
              justifyContent: 'space-between',
              gap: '10px',
            }}
          >
            <Button style={{ width: '100%' }} variant="outlined" onClick={() => {
              setAnswer('');
              setSelectedModel('');
            }}>
              Retry
            </Button>
            <Button style={{ width: '100%' }} onClick={() => setConfirmedAnswer(true)}>
              Continue
            </Button>
          </div>
        </div>
      )}
      {confirmedAnswer &&
        answer &&
        (answer.includes('Sentence classification') ? (
          <SCQQuestions workflowNames={workflowNames} question={question} answer={answer} />
        ) : answer.includes('Token classification') ? (
          <NERQuestions workflowNames={workflowNames} modelGoal={question} />
        ) : null)}
    </div>
  );
};

export default NLPQuestions;
