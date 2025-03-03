import React, { useState } from 'react';
import { Button, TextField, Typography, Stepper, Step, StepLabel, Box, Chip, IconButton } from '@mui/material';
import { useRouter } from 'next/navigation';
import { CardDescription } from '@/components/ui/card';
import { create_knowledge_extraction } from '@/lib/backend';
import AddIcon from '@mui/icons-material/Add';

interface KnowledgeExtractionQuestionsProps {
  workflowNames: string[];
}

enum LlmProvider {
  OpenAI = 'openai',
  OnPrem = 'on-prem',
  SelfHosted = 'self-hosted',
}

// Define an interface for question and keywords
interface QuestionWithKeywords {
  question: string;
  keywords: string[];
}

const KnowledgeExtractionQuestions: React.FC<KnowledgeExtractionQuestionsProps> = ({
  workflowNames,
}) => {
  const [currentStep, setCurrentStep] = useState(0);
  const [modelName, setModelName] = useState('');
  const [questions, setQuestions] = useState<QuestionWithKeywords[]>([{ question: '', keywords: [] }]);
  const [llmType, setLlmType] = useState<LlmProvider | null>(null);
  const [warningMessage, setWarningMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [currentKeyword, setCurrentKeyword] = useState('');
  const router = useRouter();

  const validateAppName = (name: string): string => {
    if (!name) return 'App name is required.';
    if (name.includes(' ')) return 'The app name cannot contain spaces.';
    if (name.includes('.')) return 'The app name cannot contain periods.';
    if (!/^[\w-]+$/.test(name))
      return 'The app name can only contain letters, numbers, underscores, and hyphens.';
    if (workflowNames.includes(name)) return 'An app with this name already exists.';
    return '';
  };

  const handleAddQuestion = () => {
    setQuestions([...questions, { question: '', keywords: [] }]);
  };

  const handleQuestionChange = (index: number, value: string) => {
    const newQuestions = [...questions];
    newQuestions[index].question = value;
    setQuestions(newQuestions);
  };

  const handleKeywordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setCurrentKeyword(event.target.value);
  };

  const handleAddKeyword = (index: number) => {
    if (currentKeyword.trim()) {
      const newQuestions = [...questions];
      if (!newQuestions[index].keywords.includes(currentKeyword.trim())) {
        newQuestions[index].keywords.push(currentKeyword.trim());
        setQuestions(newQuestions);
      }
      setCurrentKeyword('');
    }
  };

  const handleKeywordKeyPress = (event: React.KeyboardEvent, index: number) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      handleAddKeyword(index);
    }
  };

  const handleDeleteKeyword = (questionIndex: number, keywordIndex: number) => {
    const newQuestions = [...questions];
    newQuestions[questionIndex].keywords.splice(keywordIndex, 1);
    setQuestions(newQuestions);
  };

  const handleDeleteQuestion = (index: number) => {
    const newQuestions = questions.filter((_, i) => i !== index);
    setQuestions(newQuestions);
  };

  const handleSubmit = async () => {
    setIsLoading(true);
    try {
      // Check if we have a valid LLM provider
      if (!llmType) {
        throw new Error("Please select an LLM provider");
      }
  
      const params = {
        model_name: modelName,
        questions: questions,
        llm_provider: llmType.toLowerCase(), // Keep as lowercase to match the interface
        advanced_indexing: true,
        rerank: true,
        generate_answers: true,
      };

      await create_knowledge_extraction(params);
      router.push('/');
    } catch (error) {
      console.error('Error during workflow creation:', error);
      alert(error instanceof Error ? error.message : 'Failed to create workflow');
      setIsLoading(false);
    }
  };

  const steps = [
    {
      title: 'App Name',
      content: (
        <div className="mt-5">
          <TextField
            className="text-md w-full"
            value={modelName}
            onChange={(e) => {
              const name = e.target.value;
              const warning = validateAppName(name);
              setWarningMessage(warning);
              setModelName(name);
            }}
            placeholder="Enter app name"
          />
          {warningMessage && (
            <span style={{ color: 'red', marginTop: '10px' }}>{warningMessage}</span>
          )}
        </div>
      ),
    },
    {
      title: 'Questions',
      content: (
        <div className="mt-5">
          <CardDescription>Define questions and relevant keywords to extract answers from documents</CardDescription>
          {questions.map((questionItem, index) => (
            <div key={index} className="mt-6 p-4 border rounded-md">
              <div className="flex gap-4">
                <TextField
                  className="text-md flex-grow"
                  value={questionItem.question}
                  onChange={(e) => handleQuestionChange(index, e.target.value)}
                  placeholder={`Question ${index + 1}`}
                  fullWidth
                />
                {questions.length > 1 && (
                  <Button
                    variant="outlined"
                    color="error"
                    onClick={() => handleDeleteQuestion(index)}
                  >
                    Delete
                  </Button>
                )}
              </div>
              
              <div className="mt-3">
                <Typography variant="subtitle2" className="mb-2">Keywords (optional)</Typography>
                <div className="flex flex-wrap gap-2 mb-2">
                  {questionItem.keywords.map((keyword, keywordIndex) => (
                    <Chip
                      key={keywordIndex}
                      label={keyword}
                      onDelete={() => handleDeleteKeyword(index, keywordIndex)}
                      size="small"
                    />
                  ))}
                </div>
                <div className="flex gap-2 items-center">
                  <TextField
                    size="small"
                    value={currentKeyword}
                    onChange={handleKeywordChange}
                    onKeyPress={(e) => handleKeywordKeyPress(e, index)}
                    placeholder="Add keyword"
                    className="flex-grow"
                  />
                  <IconButton 
                    color="primary" 
                    onClick={() => handleAddKeyword(index)}
                    size="small"
                  >
                    <AddIcon />
                  </IconButton>
                </div>
              </div>
            </div>
          ))}
          <Button onClick={handleAddQuestion} className="mt-4" startIcon={<AddIcon />}>
            Add Question
          </Button>
        </div>
      ),
    },
    {
      title: 'LLM',
      content: (
        <div className="mt-5">
          <CardDescription>Choose an LLM for extracting answers</CardDescription>
          <div className="flex gap-4 mt-4">
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
          </div>
        </div>
      ),
    },
  ];

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const isFormValid = (): boolean => {
    const validations = [
      !isLoading,
      !!modelName,
      !!llmType, // This ensures LLM type is selected
      questions.every((questionItem) => !!questionItem.question.trim()),
    ];

    return validations.every(Boolean);
  };

  return (
    <div>
      <Box sx={{ width: '100%' }}>
        <Stepper activeStep={currentStep}>
          {steps.map((step) => (
            <Step key={step.title}>
              <StepLabel>{step.title}</StepLabel>
            </Step>
          ))}
        </Stepper>
      </Box>

      <div className="mt-8">{steps[currentStep].content}</div>

      <div className="flex justify-end gap-4 mt-8">
        {currentStep > 0 && <Button onClick={handlePrevious}>Previous</Button>}
        {currentStep < steps.length - 1 ? (
          <Button
            onClick={handleNext}
            disabled={currentStep === 0 && (!modelName || !!warningMessage)}
          >
            Next
          </Button>
        ) : (
          <Button onClick={handleSubmit} disabled={!isFormValid()}>
            {isLoading ? 'Creating...' : 'Create'}
          </Button>
        )}
      </div>
    </div>
  );
};

export default KnowledgeExtractionQuestions;