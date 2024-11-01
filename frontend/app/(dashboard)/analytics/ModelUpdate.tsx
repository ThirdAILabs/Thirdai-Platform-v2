import React, { useState, useEffect } from 'react';
import { 
  TextField,
  Typography,
  Box,
  Tooltip,
  Button, 
  Alert
} from '@mui/material';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Upload, HelpCircle } from 'lucide-react';
import { retrainTokenClassifier, trainUDTWithCSV, getTrainReport } from '@/lib/backend';
import RecentSamples from './samples';
import { TrainingResults } from './MetricsChart';
import type { TrainReportData } from '@/lib/backend';

interface ModelUpdateProps {
  username: string;
  modelName: string;
  deploymentUrl: string;
}

export default function ModelUpdate({ username, modelName, deploymentUrl }: ModelUpdateProps) {
  // States for CSV upload
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploadUpdating, setIsUploadUpdating] = useState(false);
  const [uploadError, setUploadError] = useState('');
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const [newModelName, setNewModelName] = useState(``);

  // States for polling method
  const [isPollingUpdating, setIsPollingUpdating] = useState(false);
  const [pollingError, setPollingError] = useState('');
  const [pollingSuccess, setPollingSuccess] = useState(false);

  // States for training report
  const [trainReport, setTrainReport] = useState<TrainReportData | null>(null);
  const [isLoadingReport, setIsLoadingReport] = useState(true);
  const [reportError, setReportError] = useState('');

  // Fetch initial training report - Using mock data for now
  useEffect(() => {
    const fetchInitialReport = async () => {
      try {
        setIsLoadingReport(true);
        setReportError('');
        const response = await getTrainReport(`${username}/${modelName}`);
        setTrainReport(response.data); // Extract just the data field
      } catch (error) {
        setReportError(error instanceof Error ? error.message : 'Failed to fetch training report');
      } finally {
        setIsLoadingReport(false);
      }
    };

    fetchInitialReport();
  }, [username, modelName]);

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      if (file.type === 'text/csv') {
        setSelectedFile(file);
        setUploadError('');
      } else {
        setUploadError('Please upload a CSV file');
        setSelectedFile(null);
      }
    }
  };

  const validateModelName = (name: string) => {
    // Add naming validation rules here
    const isValid = /^[a-zA-Z0-9-_]+$/.test(name);
    const isNotEmpty = name.trim().length > 0;
    return isValid && isNotEmpty;
  };

  const handleUploadUpdate = async () => {
    if (!selectedFile) {
      setUploadError('Please select a CSV file first');
      return;
    }
    if (!validateModelName(newModelName)) {
      setUploadError('Please enter a valid model name (alphanumeric characters, hyphens, and underscores only)');
      return;
    }
  
    setIsUploadUpdating(true);
    setUploadError('');
    setUploadSuccess(false);
  
    try {
      const response = await trainUDTWithCSV({ 
        model_name: newModelName,
        file: selectedFile,
        base_model_identifier: `${username}/${modelName}`,
        test_split: 0.1
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
    setIsPollingUpdating(true);
    setPollingError('');
    setPollingSuccess(false);

    try {
      const response = await retrainTokenClassifier({ 
        model_name: modelName
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

  return (
    <div className="space-y-6 w-full px-8">
      {/* Training Report Section */}
            {isLoadingReport ? (
        <Card>
          <CardContent>
            <div className="text-center py-8">Loading training report...</div>
          </CardContent>
        </Card>
      ) : reportError ? (
        <>
        </>
        // <Card>
        //   <CardContent>
        //     <Alert severity="error" sx={{ mb: 2 }}>
        //       {reportError}
        //     </Alert>
        //   </CardContent>
        // </Card>
      ) : trainReport && (
        <TrainingResults report={trainReport} />
      )}

      {/* Polled Data Section with Recent Samples */}
      <Card>
        <CardHeader>
          <CardTitle>Update Model with Recent User Feedback</CardTitle>
          <CardDescription>View and use recent labeled samples to update the model</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Recent Samples Component */}
          <div className="mb-6">
            <RecentSamples deploymentUrl={deploymentUrl} />
          </div>

          {/* Update Button Section */}
          <div className="space-y-4">
            {pollingError && (
              <Alert 
                severity="error" 
                sx={{ mb: 2 }}
              >
                {pollingError}
              </Alert>
            )}

            {pollingSuccess && (
              <Alert 
                severity="success" 
                sx={{ mb: 2 }}
              >
                Update process initiated successfully with polled data.
              </Alert>
            )}

            <Button
              onClick={handlePollingUpdate}
              disabled={isPollingUpdating}
              variant="contained"
              color={pollingSuccess ? 'success' : 'primary'}
              fullWidth
            >
              {isPollingUpdating ? 'Initiating Update...' : pollingSuccess ? 'Update Initiated!' : 'Update Model with User Feedback'}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* CSV Upload Section */}
      <Card>
          <CardHeader>
            <CardTitle>Update Model with your own data</CardTitle>
            <CardDescription>
                Upload a CSV file with token-level annotations. Your CSV file should follow these requirements:<br/><br/>
                • Two columns: 'source' and 'target'<br/>
                • Source column: Contains full text<br/>
                • Target column: Space-separated labels matching each word/token from source<br/>
                • IMPORTANT: Number of tokens in source (split by space) MUST match number of labels in target<br/><br/>
                Example (6 tokens each):<br/>
                Source: "The borrower name is John Smith"<br/>
                Target: "O O O O NAME NAME"<br/><br/>
                Let's count tokens:<br/>
                Source: The(1) borrower(2) name(3) is(4) John(5) Smith(6)<br/>
                Target: O(1) O(2) O(3) O(4) NAME(5) NAME(6)<br/><br/>
            </CardDescription>
          </CardHeader>
          <CardContent>
          <div className="space-y-4">
            <Box sx={{ mb: 4 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <Typography variant="h6" component="h3" sx={{ mr: 1 }}>
                  Name Your Updated Model
                </Typography>
                <Tooltip title="Use alphanumeric characters, hyphens, and underscores only. This will be the identifier for your updated model.">
                  <HelpCircle size={20} />
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
                helperText="Example: my-model-v2 or updated_model_123"
                error={!!uploadError && uploadError.includes('model name')}
                sx={{ mt: 1 }}
              />
            </Box>
            <div 
              className="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer hover:border-blue-500 transition-colors"
              onClick={() => document.getElementById('file-input')?.click()}
            >
              <input
                type="file"
                id="file-input"
                className="hidden"
                accept=".csv"
                onChange={handleFileInput}
              />
              <Upload className="mx-auto mb-2 text-gray-400" size={24} />
              {selectedFile ? (
                <p className="text-green-600">Selected: {selectedFile.name}</p>
              ) : (
                <p className="text-gray-600">Click to select a CSV file</p>
              )}
            </div>

            {uploadError && (
              <Alert 
                severity="error" 
                sx={{ mb: 2 }}
              >
                {uploadError}
              </Alert>
            )}

            {uploadSuccess && (
              <Alert 
                severity="success" 
                sx={{ mb: 2 }}
              >
                Update process initiated successfully with uploaded CSV.
              </Alert>
            )}

            <Button
              onClick={handleUploadUpdate}
              disabled={isUploadUpdating || !selectedFile}
              variant="contained"
              color={uploadSuccess ? 'success' : 'primary'}
              fullWidth
            >
              {isUploadUpdating ? 'Initiating Update...' : uploadSuccess ? 'Update Initiated!' : 'Update Model with CSV'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}