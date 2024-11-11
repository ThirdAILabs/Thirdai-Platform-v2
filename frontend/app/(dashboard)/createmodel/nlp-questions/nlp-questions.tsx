import React, { useState, ChangeEvent } from 'react';
import NERQuestions from './ner-questions';
import SCQQuestions from './sentence-classification-questions';
import { trainUDTWithCSV } from '@/lib/backend';
import { 
  Button, 
  Card, 
  CardContent,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
  Typography,
  Tabs,
  Tab,
  Box,
  Paper,
  Alert,
} from '@mui/material';

type ModelType = 'ner' | 'classification' | '';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

interface CSVValidationResult {
  success: boolean;
  error?: string;
}

interface NLPQuestionsProps {
  workflowNames: string[];
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

const NLPQuestions: React.FC<NLPQuestionsProps> = ({ workflowNames }) => {
  console.log('workflowNames', workflowNames)
  const [modelType, setModelType] = useState<ModelType>('');
  const [tabValue, setTabValue] = useState<number>(0);
  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [csvError, setCsvError] = useState<string>('');
  const [modelName, setModelName] = useState<string>('');
  const [warningMessage, setWarningMessage] = useState<string>('');
  const [isCreating, setIsCreating] = useState<boolean>(false);
  const [createSuccess, setCreateSuccess] = useState<boolean>(false);

  const validateModelName = (name: string): boolean => {
    if (workflowNames.includes(name)) {
      console.log('error naming')
      setWarningMessage('An app with the same name already exists. Please choose a different name.');
      return false;
    }

    if (!name.trim()) {
      setWarningMessage('Model name cannot be empty.');
      return false;
    }

    const isValid = /^[a-zA-Z0-9-_]+$/.test(name);
    if (!isValid) {
      setWarningMessage('Model name can only contain letters, numbers, underscores, and hyphens.');
      return false;
    }

    if (name.includes(' ')) {
      setWarningMessage('Model name cannot contain spaces.');
      return false;
    }

    if (name.includes('.')) {
      setWarningMessage('Model name cannot contain periods.');
      return false;
    }

    setWarningMessage('');
    return true;
  };

  const handleFileUpload = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    if (!file.name.endsWith('.csv')) {
      setCsvError('Please upload a CSV file');
      setCsvFile(null);
      return;
    }

    try {
      // Read and validate CSV format
      const reader = new FileReader();
      reader.onload = async (e: ProgressEvent<FileReader>) => {
        const text = e.target?.result as string;
        const isValid = await validateCSVFormat(text);
        
        if (!isValid.success) {
          setCsvError(isValid.error || 'Invalid CSV format');
          setCsvFile(null);
          return;
        }
        
        setCsvFile(file);
        setCsvError('');
      };
      reader.readAsText(file);
    } catch (error) {
      setCsvError(`Error reading file: ${error instanceof Error ? error.message : 'Unknown error'}`);
      setCsvFile(null);
    }
  };

  const validateCSVFormat = async (csvText: string): Promise<CSVValidationResult> => {
    const lines = csvText.split('\n').filter(line => line.trim());
    if (lines.length === 0) {
      return {
        success: false,
        error: 'CSV file is empty'
      };
    }

    const headers = lines[0].toLowerCase().split(',').map(h => h.trim());

    // For NER, we expect 'source' and 'target' columns
    if (!headers.includes('source') || !headers.includes('target')) {
      return {
        success: false,
        error: 'CSV must have "source" and "target" columns'
      };
    }

    // Validate that each row has matching token counts
    for (let i = 1; i < lines.length; i++) {
      const columns = lines[i].split(',');
      if (columns.length < 2) continue;

      const source = columns[0].trim();
      const target = columns[1].trim();
      const sourceTokens = source.split(' ').length;
      const targetTokens = target.split(' ').length;

      if (sourceTokens !== targetTokens) {
        return {
          success: false,
          error: `Row ${i + 1}: Number of tokens in source (${sourceTokens}) does not match target (${targetTokens})`
        };
      }
    }

    return { success: true };
  };

  const handleModelTypeChange = (event: SelectChangeEvent<string>) => {
    setModelType(event.target.value as ModelType);
    setCsvFile(null);
    setCsvError('');
    setTabValue(0);
    setModelName('');
    setWarningMessage('');
    setCreateSuccess(false);
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleCreateModel = async () => {
    if (!csvFile) {
      setCsvError('Please select a CSV file first');
      return;
    }

    if (!validateModelName(modelName)) {
      return;
    }

    setIsCreating(true);
    setCsvError('');
    setCreateSuccess(false);

    try {
      const response = await trainUDTWithCSV({
        model_name: modelName,
        file: csvFile,
        base_model_identifier: '', // Empty for new model creation
        test_split: 0.1,
      });

      if (response.status === 'success') {
        setCreateSuccess(true);
      } else {
        throw new Error(response.message || 'Failed to create model');
      }
    } catch (error) {
      setCsvError(`Error creating model: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h5" component="div" gutterBottom>
          Create NLP Model
        </Typography>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          Choose your model type and data source to get started
        </Typography>

        <Box sx={{ mt: 3 }}>
          <FormControl fullWidth>
            <InputLabel>Model Type</InputLabel>
            <Select
              value={modelType}
              label="Model Type"
              onChange={handleModelTypeChange}
            >
              <MenuItem value="ner">Named Entity Recognition (NER)</MenuItem>
              <MenuItem value="classification">Sentence Classification</MenuItem>
            </Select>
          </FormControl>

          {modelType && (
            <Box sx={{ mt: 4 }}>
              <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                <Tabs 
                  value={tabValue} 
                  onChange={handleTabChange}
                >
                  <Tab label="Generate Synthetic Data" />
                  <Tab label="Upload CSV" />
                </Tabs>
              </Box>

              <TabPanel value={tabValue} index={0}>
                {modelType === 'ner' && (
                  <NERQuestions 
                    workflowNames={workflowNames} 
                    modelGoal=""
                  />
                )}
                {modelType === 'classification' && (
                  <SCQQuestions 
                    workflowNames={workflowNames}
                    question=""
                    answer=""
                  />
                )}
              </TabPanel>

              <TabPanel value={tabValue} index={1}>
                <Box sx={{ mt: 2 }}>
                  <FormControl fullWidth sx={{ mb: 3 }}>
                    <InputLabel>Model Name</InputLabel>
                    <input
                      type="text"
                      value={modelName}
                      onChange={(e) => setModelName(e.target.value)}
                      className="w-full px-3 py-2 border rounded"
                      placeholder="Enter model name"
                    />
                    {warningMessage && (
                      <Typography color="error" variant="caption" sx={{ mt: 1 }}>
                        {warningMessage}
                      </Typography>
                    )}
                  </FormControl>

                  <input
                    type="file"
                    accept=".csv"
                    onChange={handleFileUpload}
                    className="w-full px-3 py-2 border rounded mb-3"
                  />
                  
                  {csvError && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                      {csvError}
                    </Alert>
                  )}
                  
                  {csvFile && !csvError && (
                    <Alert severity="success" sx={{ mb: 2 }}>
                      File "{csvFile.name}" ready to process
                    </Alert>
                  )}

                  <Paper sx={{ p: 2, mt: 3, bgcolor: 'grey.50' }}>
                    <Typography variant="subtitle2" gutterBottom>
                      CSV Format Requirements:
                    </Typography>
                    <Box 
                      component="pre" 
                      sx={{ 
                        p: 2, 
                        bgcolor: 'background.paper',
                        borderRadius: 1,
                        overflow: 'auto'
                      }}
                    >
                      source,target
John works at Apple,PERSON O O O ORG
Name is Jane Smith,O O PERSON PERSON
                    </Box>
                  </Paper>

                  {(csvFile && !csvError) && (
                    <Button 
                      variant="contained"
                      fullWidth
                      sx={{ mt: 3 }}
                      onClick={handleCreateModel}
                      disabled={isCreating}
                    >
                      {isCreating ? 'Creating Model...' : createSuccess ? 'Model Created!' : 'Create Model'}
                    </Button>
                  )}
                </Box>
              </TabPanel>
            </Box>
          )}
        </Box>
      </CardContent>
    </Card>
  );
};

export default NLPQuestions;
