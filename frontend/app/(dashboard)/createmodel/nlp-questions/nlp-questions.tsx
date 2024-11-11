import React, { useState, ChangeEvent } from 'react';
import NERQuestions from './ner-questions';
import SCQQuestions from './sentence-classification-questions';
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
      {value === index && (
        <Box sx={{ p: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

const NLPQuestions: React.FC<NLPQuestionsProps> = ({ workflowNames }) => {
  const [modelType, setModelType] = useState<ModelType>('');
  const [tabValue, setTabValue] = useState<number>(0);
  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [csvError, setCsvError] = useState<string>('');

  const handleFileUpload = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    if (!file.name.endsWith('.csv')) {
      setCsvError('Please upload a CSV file');
      return;
    }

    try {
      const reader = new FileReader();
      reader.onload = async (e: ProgressEvent<FileReader>) => {
        const text = e.target?.result as string;
        const isValid = await validateCSVFormat(text, modelType);
        
        if (!isValid.success) {
          setCsvError(isValid.error || 'Invalid CSV format');
          return;
        }
        
        setCsvFile(file);
        setCsvError('');
      };
      reader.readAsText(file);
    } catch (error) {
      setCsvError(`Error reading file: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const validateCSVFormat = async (csvText: string, type: ModelType): Promise<CSVValidationResult> => {
    const lines = csvText.split('\n').filter(line => line.trim());
    if (lines.length === 0) {
      return {
        success: false,
        error: 'CSV file is empty'
      };
    }

    const headers = lines[0].toLowerCase().split(',').map(h => h.trim());

    if (type === 'ner') {
      if (headers.length !== 2 || !headers.includes('text') || !headers.includes('labels')) {
        return {
          success: false,
          error: 'NER CSV must have "text" and "labels" columns'
        };
      }

      // Validate NER JSON format in labels column
      try {
        for (let i = 1; i < lines.length; i++) {
          const columns = lines[i].split(',');
          if (columns.length !== 2) continue;
          const labels = JSON.parse(columns[1]);
          if (!Array.isArray(labels) || !labels.every(label => 
            typeof label === 'object' && 
            'text' in label && 
            'label' in label
          )) {
            return {
              success: false,
              error: `Invalid label format in row ${i + 1}`
            };
          }
        }
      } catch (error) {
        return {
          success: false,
          error: 'Invalid JSON format in labels column'
        };
      }
    } else if (type === 'classification') {
      if (headers.length !== 2 || !headers.includes('text') || !headers.includes('label')) {
        return {
          success: false,
          error: 'Classification CSV must have "text" and "label" columns'
        };
      }
    }

    return { success: true };
  };

  const handleModelTypeChange = (event: SelectChangeEvent<string>) => {
    setModelType(event.target.value as ModelType);
    // Reset state when model type changes
    setCsvFile(null);
    setCsvError('');
    setTabValue(0);
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleCreateModel = async () => {
    if (!csvFile) return;

    try {
      const formData = new FormData();
      formData.append('file', csvFile);
      formData.append('modelType', modelType);

      const response = await fetch('/api/create-model', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to create model');
      }

      const data = await response.json();
      console.log('Model created:', data);
      
      // Handle success (e.g., redirect or show success message)
    } catch (error) {
      console.error('Error creating model:', error);
      setCsvError(`Error creating model: ${error instanceof Error ? error.message : 'Unknown error'}`);
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
                  aria-label="model creation method tabs"
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
                  <input
                    type="file"
                    accept=".csv"
                    onChange={handleFileUpload}
                    style={{ 
                      width: '100%',
                      padding: '10px',
                      marginBottom: '10px',
                      border: '1px solid #ccc',
                      borderRadius: '4px'
                    }}
                  />
                  
                  {csvError && (
                    <Typography color="error" variant="body2" sx={{ mt: 1 }}>
                      {csvError}
                    </Typography>
                  )}
                  
                  {csvFile && !csvError && (
                    <Typography color="success.main" variant="body2" sx={{ mt: 1 }}>
                      File "{csvFile.name}" ready to process
                    </Typography>
                  )}

                  <Paper sx={{ p: 2, mt: 3, bgcolor: 'grey.50' }}>
                    <Typography variant="subtitle2" gutterBottom>
                      {modelType === 'ner' ? 'CSV format for NER:' : 'CSV format for Classification:'}
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
                      {modelType === 'ner' ? 
                        'text,labels\n"John works at Apple","[{\\"text\\":\\"John\\",\\"label\\":\\"PERSON\\"},{\\"text\\":\\"Apple\\",\\"label\\":\\"ORG\\"}]"' :
                        'text,label\n"Great product!","POSITIVE"\n"Poor service","NEGATIVE"'
                      }
                    </Box>
                  </Paper>

                  {csvFile && !csvError && (
                    <Button 
                      variant="contained"
                      fullWidth
                      sx={{ mt: 3 }}
                      onClick={handleCreateModel}
                    >
                      Create Model
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
