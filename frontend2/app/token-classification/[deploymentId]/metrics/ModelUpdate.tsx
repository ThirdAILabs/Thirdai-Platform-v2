'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  TextField,
  Button,
  Card,
  CardContent,
  Tooltip,
  Alert,
  Chip,
  Paper,
  Divider,
  IconButton
} from '@mui/material';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowDropUpIcon from '@mui/icons-material/ArrowDropUp';
import RecentSamples from './RecentSamples';

// Mock functions to simulate API calls
// These would be replaced with actual API calls in a real implementation
const getTrainReport = async (modelId: string) => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 800));
  
  return {
    data: {
      metrics: {
        before: { precision: 0.85, recall: 0.82, f1: 0.83 },
        after: { precision: 0.91, recall: 0.89, f1: 0.90 }
      }
    }
  };
};

const getLabels = async () => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 600));
  
  return ['PERSON', 'ORGANIZATION', 'LOCATION', 'DATE', 'O', 'PHONE', 'EMAIL', 'ADDRESS', 'PRODUCT', 'EVENT', 'TIME', 'MONEY', 'PERCENT', 'QUANTITY'];
};

const trainUDTWithCSV = async ({ model_name, file, base_model_identifier, test_split }: any) => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 1500));
  
  return {
    status: 'success',
    message: 'Training initiated successfully',
    model_id: 'model_123456'
  };
};

const retrainTokenClassifier = async ({ model_name, base_model_id }: any) => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 1500));
  
  return {
    status: 'success',
    message: 'Retraining initiated successfully',
    model_id: 'model_789012'
  };
};

interface ModelUpdateProps {
  username?: string;
  modelName?: string;
  deploymentUrl?: string;
  workflowNames?: string[];
  deployStatus?: string;
  modelId?: string;
}

const ModelUpdate: React.FC<ModelUpdateProps> = ({
  username = 'user',
  modelName = 'token-classifier',
  deploymentUrl = 'https://api.example.com/token-classification',
  workflowNames = [],
  deployStatus = 'complete',
  modelId = 'model_123'
}) => {
  // States for CSV upload
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploadUpdating, setIsUploadUpdating] = useState(false);
  const [uploadError, setUploadError] = useState('');
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const [newModelName, setNewModelName] = useState(``);
  const [warningMessage, setWarningMessage] = useState('');

  // States for polling method
  const [isPollingUpdating, setIsPollingUpdating] = useState(false);
  const [pollingError, setPollingError] = useState('');
  const [pollingSuccess, setPollingSuccess] = useState(false);

  // States for training report
  const [trainReport, setTrainReport] = useState<any | null>(null);
  const [isLoadingReport, setIsLoadingReport] = useState(true);
  const [reportError, setReportError] = useState('');

  // New states for button cooldown
  const [uploadButtonDisabled, setUploadButtonDisabled] = useState(false);
  const [pollingButtonDisabled, setPollingButtonDisabled] = useState(false);

  // State for toggling tags list
  const [numTagDisplay, setNumTagDisplay] = useState<number>(5);
  const [tags, setTags] = useState<string[]>([]);

  // New state for polling model name
  const [pollingModelName, setPollingModelName] = useState(``);
  const [pollingWarningMessage, setPollingWarningMessage] = useState('');

  // File input reference
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  // Effect to validate model name on each change
  useEffect(() => {
    validateModelName(newModelName);
  }, [newModelName, workflowNames]);

  // Effect to validate polling model name
  useEffect(() => {
    validatePollingModelName(pollingModelName);
  }, [pollingModelName, workflowNames]);

  // Timer effect for upload button cooldown
  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (uploadButtonDisabled) {
      timer = setTimeout(() => {
        setUploadButtonDisabled(false);
      }, 3000);
    }
    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [uploadButtonDisabled]);

  // Timer effect for polling button cooldown
  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (pollingButtonDisabled) {
      timer = setTimeout(() => {
        setPollingButtonDisabled(false);
      }, 3000);
    }
    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [pollingButtonDisabled]);

  // Fetch initial training report
  useEffect(() => {
    const fetchInitialReport = async () => {
      try {
        setIsLoadingReport(true);
        setReportError('');
        const response = await getTrainReport(modelId);
        setTrainReport(response.data);
      } catch (error) {
        setReportError(error instanceof Error ? error.message : 'Failed to fetch training report');
      } finally {
        setIsLoadingReport(false);
      }
    };

    fetchInitialReport();
  }, [modelId]);

  // Fetch labels/tags
  useEffect(() => {
    const fetchTags = async () => {
      try {
        const labels = await getLabels();
        const filteredLabels = labels.filter((label) => label !== 'O');
        setTags(filteredLabels);
      } catch (error) {
        console.error('Error fetching labels:', error);
      }
    };
    
    fetchTags();
  }, []);

  const MAX_FILE_SIZE = 500 * 1024 * 1024; // 500MB in bytes

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      // Check file type
      if (file.type !== 'text/csv') {
        setUploadError('Please upload a CSV file');
        setSelectedFile(null);
        return;
      }

      // Check file size
      if (file.size > MAX_FILE_SIZE) {
        setUploadError('File size must be less than 500MB');
        setSelectedFile(null);
        return;
      }

      setSelectedFile(file);
      setUploadError('');
    }
  };

  const validateModelName = (name: string) => {
    // Check if name exists in workflowNames
    if (workflowNames.includes(name)) {
      setWarningMessage(
        'An app with the same name already exists. Please choose a different name.'
      );
      return false;
    }

    // Check for valid characters (alphanumeric, hyphens, and underscores)
    const isValid = /^[a-zA-Z0-9-_]+$/.test(name);
    const isNotEmpty = name.trim().length > 0;

    if (!isValid && isNotEmpty) {
      setWarningMessage(
        'The app name can only contain letters, numbers, underscores, and hyphens. Please modify the name.'
      );
      return false;
    }

    if (name.includes(' ')) {
      setWarningMessage('The app name cannot contain spaces. Please remove the spaces.');
      return false;
    }

    if (name.includes('.')) {
      setWarningMessage("The app name cannot contain periods ('.'). Please remove the periods.");
      return false;
    }

    setWarningMessage('');
    return isValid && isNotEmpty;
  };

  const validatePollingModelName = (name: string) => {
    if (workflowNames.includes(name)) {
      setPollingWarningMessage(
        'An app with the same name already exists. Please choose a different name.'
      );
      return false;
    }
    const isValid = /^[a-zA-Z0-9-_]+$/.test(name);
    const isNotEmpty = name.trim().length > 0;
    if (!isValid && isNotEmpty) {
      setPollingWarningMessage(
        'The app name can only contain letters, numbers, underscores, and hyphens.'
      );
      return false;
    }
    if (name.includes(' ')) {
      setPollingWarningMessage('The app name cannot contain spaces.');
      return false;
    }
    if (name.includes('.')) {
      setPollingWarningMessage("The app name cannot contain periods ('.').");
      return false;
    }
    setPollingWarningMessage('');
    return isValid && isNotEmpty;
  };

  const handleUploadUpdate = async () => {
    if (!selectedFile) {
      setUploadError('Please select a CSV file first');
      return;
    }
    if (!validateModelName(newModelName)) {
      setUploadError(
        'Please enter a valid model name (alphanumeric characters, hyphens, and underscores only)'
      );
      return;
    }

    setIsUploadUpdating(true);
    setUploadError('');
    setUploadSuccess(false);
    setUploadButtonDisabled(true);

    try {
      const response = await trainUDTWithCSV({
        model_name: newModelName,
        file: selectedFile,
        base_model_identifier: `${username}/${modelName}`,
        test_split: 0.1,
      });

      if (response.status === 'success') {
        setUploadSuccess(true);
      } else {
        throw new Error(response.message || 'Failed to initiate update');
      }
    } catch (error) {
      setUploadError(
        error instanceof Error
          ? error.message
          : 'An error occurred while initiating the model update'
      );
    } finally {
      setIsUploadUpdating(false);
    }
  };

  const handlePollingUpdate = async () => {
    if (!validatePollingModelName(pollingModelName)) {
      setPollingError('Please enter a valid model name for the new model.');
      return;
    }

    // Check deploy status before proceeding
    if (deployStatus !== 'complete') {
      setPollingError('Model must be fully deployed before updating with user feedback.');
      return;
    }

    setIsPollingUpdating(true);
    setPollingError('');
    setPollingSuccess(false);
    setPollingButtonDisabled(true);
    
    try {
      const response = await retrainTokenClassifier({
        model_name: pollingModelName,
        base_model_id: modelId,
      });
      
      if (response.status === 'success') {
        setPollingSuccess(true);
      } else {
        throw new Error(response.message || 'Failed to initiate update');
      }
    } catch (error) {
      setPollingError(
        error instanceof Error
          ? error.message
          : 'An error occurred while initiating the model update'
      );
    } finally {
      setIsPollingUpdating(false);
    }
  };

  const handleTagDisplayMore = () => {
    setNumTagDisplay(tags.length);
  };

  const handleTagDisplayLess = () => {
    setNumTagDisplay(5);
  };

  const openFileSelector = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  return (
    <Box sx={{ mb: 4 }}>
      {/* CSV Upload Section */}
      <Card sx={{ mb: 4, backgroundColor: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}>
        <CardContent>
          <Typography variant="h6" sx={{ fontWeight: 500, mb: 1 }}>
            Update Model with your own data
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            Upload a CSV file with token-level annotations. Your CSV file should follow these requirements:
            <Box component="ul" sx={{ pl: 2, mt: 1 }}>
              <Box component="li">Two columns: 'source' and 'target'</Box>
              <Box component="li">Source column: Contains full text</Box>
              <Box component="li">Target column: Space-separated labels matching each word/token from source</Box>
              <Box component="li" sx={{ fontWeight: 'bold' }}>
                IMPORTANT: Number of tokens in source (split by space) MUST match number of labels in target
              </Box>
            </Box>
            <Box sx={{ mt: 2 }}>
              Example (6 tokens each):<br />
              Source: "The borrower name is John Smith"<br />
              Target: "O O O O NAME NAME"
            </Box>
            
            <Box sx={{ mt: 2, display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1 }}>
              <Typography variant="body2">Tags Used for Training:</Typography>
              <Paper 
                variant="outlined" 
                sx={{ 
                  p: 1, 
                  display: 'flex', 
                  flexWrap: 'wrap', 
                  gap: 1, 
                  alignItems: 'center',
                  maxWidth: 600,
                  flex: 1
                }}
              >
                {tags.slice(0, numTagDisplay).map((tag, index) => (
                  <Chip 
                    key={`${index}-${tag}`} 
                    label={tag} 
                    size="small" 
                    sx={{ 
                      bgcolor: 'rgba(0, 0, 0, 0.05)', 
                      borderRadius: '4px' 
                    }} 
                  />
                ))}
                {tags.length > 5 && (
                  <Button
                    variant="outlined"
                    size="small"
                    onClick={numTagDisplay === 5 ? handleTagDisplayMore : handleTagDisplayLess}
                    endIcon={numTagDisplay === 5 ? <ArrowDropDownIcon /> : <ArrowDropUpIcon />}
                    sx={{ ml: 1 }}
                  >
                    {numTagDisplay === 5 ? 'Expand' : 'Collapse'}
                  </Button>
                )}
              </Paper>
            </Box>
          </Typography>
          
          <Box sx={{ mb: 3, mt: 4 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Typography variant="subtitle1" sx={{ mr: 1, fontWeight: 500 }}>
                Name Your Updated Model
              </Typography>
              <Tooltip title="Use alphanumeric characters, hyphens, and underscores only. This will be the identifier for your updated model.">
                <HelpOutlineIcon fontSize="small" color="action" />
              </Tooltip>
            </Box>
            <TextField
              fullWidth
              id="model-name"
              label="New Model Name"
              variant="outlined"
              value={newModelName}
              onChange={(e) => setNewModelName(e.target.value)}
              placeholder="Enter new model name"
              helperText={warningMessage || 'Example: my-model-v2 or updated_model_123'}
              error={!!warningMessage}
              size="small"
            />
          </Box>
          
          <Box 
            sx={{ 
              border: '2px dashed',
              borderColor: selectedFile ? 'primary.main' : 'divider',
              borderRadius: 1,
              p: 3,
              textAlign: 'center',
              cursor: 'pointer',
              mb: 3,
              transition: 'border-color 0.2s',
              '&:hover': {
                borderColor: 'primary.main'
              }
            }}
            onClick={openFileSelector}
          >
            <input
              type="file"
              ref={fileInputRef}
              hidden
              accept=".csv"
              onChange={handleFileInput}
            />
            <CloudUploadIcon color="action" sx={{ fontSize: 40, mb: 1 }} />
            {selectedFile ? (
              <Typography color="primary">Selected: {selectedFile.name}</Typography>
            ) : (
              <Typography color="textSecondary">Click to select a CSV file</Typography>
            )}
          </Box>
          
          {uploadError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {uploadError}
            </Alert>
          )}
          
          {uploadSuccess && (
            <Alert severity="success" sx={{ mb: 2 }}>
              Update process initiated successfully with uploaded CSV.
            </Alert>
          )}
          
          <Button
            onClick={handleUploadUpdate}
            disabled={isUploadUpdating || !selectedFile || uploadButtonDisabled}
            variant="contained"
            color={uploadSuccess ? 'success' : 'primary'}
            fullWidth
          >
            {isUploadUpdating
              ? 'Initiating Update...'
              : uploadSuccess
                ? 'Update Initiated!'
                : 'Update Model with CSV'}
          </Button>
        </CardContent>
      </Card>
      
      {/* Polled Data Section with Recent Samples */}
      <Card sx={{ mb: 4, backgroundColor: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}>
        <CardContent>
          <Typography variant="h6" sx={{ fontWeight: 500, mb: 1 }}>
            Update Model with Recent User Feedback
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            View and use recent labeled samples to update the model
          </Typography>
          
          <Box sx={{ mb: 4 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <Typography variant="subtitle1" sx={{ mr: 1, fontWeight: 500 }}>
                Name Your New Model
              </Typography>
              <Tooltip title="Use alphanumeric characters, hyphens, and underscores only. This will be the identifier for your updated model.">
                <HelpOutlineIcon fontSize="small" color="action" />
              </Tooltip>
            </Box>
            <TextField
              fullWidth
              id="polling-model-name"
              label="New Model Name"
              variant="outlined"
              value={pollingModelName}
              onChange={(e) => setPollingModelName(e.target.value)}
              placeholder="Enter new model name"
              helperText={pollingWarningMessage || 'Example: my-model-v2 or updated_model_123'}
              error={!!pollingWarningMessage}
              size="small"
            />
          </Box>
          
          <Box sx={{ mb: 4 }}>
            <RecentSamples deploymentUrl={deploymentUrl} />
          </Box>
          
          {pollingError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {pollingError}
            </Alert>
          )}
          
          {pollingSuccess && (
            <Alert severity="success" sx={{ mb: 2 }}>
              Update process initiated successfully with polled data.
            </Alert>
          )}
          
          <Button
            onClick={handlePollingUpdate}
            disabled={isPollingUpdating || pollingButtonDisabled || deployStatus !== 'complete'}
            variant="contained"
            color={pollingSuccess ? 'success' : 'primary'}
            fullWidth
          >
            {isPollingUpdating
              ? 'Initiating Update...'
              : pollingSuccess
                ? 'Update Initiated!'
                : deployStatus !== 'complete'
                  ? "Model Must Be Deployed First (refresh page once it's deployed)"
                  : 'Update Model with User Feedback'}
          </Button>
        </CardContent>
      </Card>
    </Box>
  );
};

export default ModelUpdate; 