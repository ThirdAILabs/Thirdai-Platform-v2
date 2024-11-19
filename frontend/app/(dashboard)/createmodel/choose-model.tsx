'use client';

import { useState, useEffect } from 'react';
import EnterpriseSearchQuestions from './rag-questions/enterprise-search-questions';
import ChatbotQuestions from './rag-questions/chatbot-questions';
import NLPQuestions from './nlp-questions/nlp-questions';
import { fetchWorkflows, Workflow } from '@/lib/backend';
import { FormControl, InputLabel, Select, MenuItem, Typography, Box, Divider } from '@mui/material';

const USE_CASES = [
  {
    name: 'Enterprise Search',
    value: 'enterprise-search',
    description: 'Build a search engine that understands context and meaning in your documents',
  },
  {
    name: 'Chatbot',
    value: 'chatbot',
    description:
      'Create an AI assistant that can engage in conversations and answer questions using your data',
  },
  {
    name: 'NLP / Text Analytics',
    value: 'nlp-text-analytics',
    description: 'Extract insights, classify content, and analyze unstructured text data at scale',
  },
];

export default function ChooseProblem() {
  const [modelType, setModelType] = useState('');
  const [privateModels, setPrivateModels] = useState<Workflow[]>([]);
  const [workflows, setWorkflows] = useState<Workflow[]>([]);

  useEffect(() => {
    async function getModels() {
      try {
        const response = await fetchWorkflows();
        setPrivateModels(response);
        setWorkflows(response);
      } catch (err) {
        if (err instanceof Error) {
          console.log(err.message);
        } else {
          console.log('An unknown error occurred');
        }
      }
    }

    getModels();
  }, []);

  const workflowNames = workflows.map((workflow) => workflow.model_name);

  const handleChange = (event: any) => {
    setModelType(event.target.value);
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
      <Box>
        <Typography variant="h6" sx={{ mb: 1 }}>
          Use case
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Please select the app type based on your use case.
        </Typography>

        <FormControl fullWidth>
          <InputLabel>Select a use case</InputLabel>
          <Select value={modelType} label="Select a use case" onChange={handleChange}>
            {USE_CASES.map((useCase) => (
              <MenuItem key={useCase.value} value={useCase.value} sx={{ py: 2 }}>
                <Box>
                  <Typography variant="subtitle1">{useCase.name}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {useCase.description}
                  </Typography>
                </Box>
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Box>

      {modelType && (
        <Box sx={{ width: '100%' }}>
          <Divider sx={{ mb: 3 }} />
          {modelType === 'chatbot' && (
            <ChatbotQuestions 
              models={privateModels} 
              workflowNames={workflowNames}
            />
          )}
          {modelType === 'nlp-text-analytics' && <NLPQuestions workflowNames={workflowNames} />}
          {modelType === 'enterprise-search' && (
            <EnterpriseSearchQuestions models={privateModels} workflowNames={workflowNames} />
          )}
        </Box>
      )}
    </Box>
  );
}
