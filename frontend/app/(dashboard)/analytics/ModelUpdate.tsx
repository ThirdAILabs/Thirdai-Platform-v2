import React, { useState } from 'react';
import { Button, Alert } from '@mui/material';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Upload } from 'lucide-react';
import { retrainTokenClassifier, trainUDTWithCSV } from '@/lib/backend';
import RecentSamples from './samples';

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

  // States for polling method
  const [isPollingUpdating, setIsPollingUpdating] = useState(false);
  const [pollingError, setPollingError] = useState('');
  const [pollingSuccess, setPollingSuccess] = useState(false);

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

  const handleUploadUpdate = async () => {
    if (!selectedFile) {
      setUploadError('Please select a CSV file first');
      return;
    }
  
    setIsUploadUpdating(true);
    setUploadError('');
    setUploadSuccess(false);
  
    try {
      const response = await trainUDTWithCSV({ 
        model_name: `${modelName}-temp`,
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
      {/* Polled Data Section with Recent Samples */}
      <Card>
        <CardHeader>
          <CardTitle>Update Model with Recent Data</CardTitle>
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
              {isPollingUpdating ? 'Initiating Update...' : pollingSuccess ? 'Update Initiated!' : 'Update Model with Polled Data'}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* CSV Upload Section */}
      <Card>
        <CardHeader>
          <CardTitle>Update Model via CSV Upload</CardTitle>
          <CardDescription>Upload a CSV file containing new feedback data to update the model</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
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