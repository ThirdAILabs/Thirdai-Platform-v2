import React, { useState } from 'react';
import { Box, Button, Typography, Alert, CircularProgress, Paper } from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { trainTokenClassifierFromCSV } from '@/lib/backend';

interface CSVUploadProps {
  modelName: string;
  onSuccess?: () => void;
  onError?: (error: string) => void; // Explicitly typing the error parameter as string
}

const CSVUpload = ({ modelName, onSuccess, onError }: CSVUploadProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState(false);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (!file.name.endsWith('.csv')) {
        setError('Please select a CSV file');
        return;
      }
      setSelectedFile(file);
      setError('');
      setSuccess(false);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('Please select a CSV file first');
      return;
    }

    setIsUploading(true);
    setError('');
    setSuccess(false);

    try {
      const response = await trainTokenClassifierFromCSV({
        modelName: modelName,
        file: selectedFile,
        testSplit: 0.1,
      });

      if (response.status === 'success') {
        setSuccess(true);
        onSuccess?.();
      } else {
        throw new Error(response.message || 'Failed to train model');
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'An error occurred while training the model';
      setError(errorMessage);
      onError?.(errorMessage);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Paper
        elevation={0}
        sx={{
          p: 3,
          border: '2px dashed #ccc',
          backgroundColor: '#fafafa',
          textAlign: 'center',
          mb: 2,
        }}
      >
        <input
          accept=".csv"
          style={{ display: 'none' }}
          id="csv-upload-button"
          type="file"
          onChange={handleFileSelect}
        />
        <label htmlFor="csv-upload-button">
          <Button
            component="span"
            variant="outlined"
            startIcon={<CloudUploadIcon />}
            sx={{ mb: 2 }}
          >
            Select CSV File
          </Button>
        </label>

        {selectedFile && (
          <Typography variant="body2" color="text.secondary">
            Selected file: {selectedFile.name}
          </Typography>
        )}
      </Paper>

      {selectedFile && (
        <Button
          variant="contained"
          onClick={handleUpload}
          disabled={isUploading}
          fullWidth
          sx={{ mb: 2 }}
        >
          {isUploading ? (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <CircularProgress size={20} color="inherit" />
              Training Model...
            </Box>
          ) : (
            'Train Model'
          )}
        </Button>
      )}

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Model training initiated successfully!
        </Alert>
      )}
    </Box>
  );
};

export default CSVUpload;
