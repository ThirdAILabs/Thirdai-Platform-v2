import React, { useState } from 'react';
import NERQuestions from './ner-questions';
import SCQQuestions from './sentence-classification-questions';
import { 
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Typography,
  Box,
  Divider
} from '@mui/material';
import { CardDescription } from '@/components/ui/card';

interface NLPQuestionsProps {
  workflowNames: string[];
}

const NLP_TYPES = [
  {
    value: 'text-classification',
    label: 'Text Classification',
    description: 'Analyze sentiment and classify text content',
  },
  {
    value: 'text-extraction',
    label: 'Text Extraction',
    description: 'Extract entities, PII, or HIPAA information',
  },
  {
    value: 'document-classification',
    label: 'Document Classification',
    description: 'Categorize entire documents',
  },
];

const NLPQuestions = ({ workflowNames }: NLPQuestionsProps) => {
  const [selectedType, setSelectedType] = useState<string>('');

  const handleChange = (event: any) => {
    setSelectedType(event.target.value);
  };

  const renderSelectedComponent = () => {
    switch(selectedType) {
      case 'text-classification':
        return (
          <SCQQuestions
            workflowNames={workflowNames}
            question="Text Classification Task"
            answer="Sentence classification selected"
          />
        );
      case 'text-extraction':
        return (
          <NERQuestions
            workflowNames={workflowNames}
            modelGoal="Text Extraction Task"
          />
        );
      case 'document-classification':
        return (
          <Box>
            <Typography variant="h6" sx={{ mb: 1 }}>
              Document Classification
            </Typography>
            <CardDescription>
              Document classification component to be implemented
            </CardDescription>
          </Box>
        );
      default:
        return (
          <Box sx={{ mt: 2 }}>
            <CardDescription>
              Please select an NLP task type to begin
            </CardDescription>
          </Box>
        );
    }
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Typography variant="h6" sx={{ mb: 1 }}>
        NLP Task Assistant
      </Typography>
      
      <FormControl fullWidth sx={{ mb: 3 }}>
        <InputLabel>Select an NLP task type</InputLabel>
        <Select
          value={selectedType}
          label="Select an NLP task type"
          onChange={handleChange}
        >
          {NLP_TYPES.map((type) => (
            <MenuItem key={type.value} value={type.value}>
              <Box>
                <Typography variant="subtitle1">
                  {type.label}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {type.description}
                </Typography>
              </Box>
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      {selectedType && <Divider sx={{ mb: 3 }} />}
      
      {renderSelectedComponent()}
    </Box>
  );
};

export default NLPQuestions;