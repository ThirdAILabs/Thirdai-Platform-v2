import React, { useState } from 'react';
import { Box, Button, Typography, Alert, CircularProgress, Paper } from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { trainSentenceClassifierFromCSV, validateSentenceClassifierCSV } from '@/lib/backend';
import LabelConfirmationDialog from './LabelConfirmationDialog';

interface CSVUploadProps {
  modelName: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

const CSVUpload = ({ modelName, onSuccess, onError }: CSVUploadProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [detectedLabels, setDetectedLabels] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);

  const validateAndProcessFile = async (file: File) => {
    try {
      const validationResult = await validateSentenceClassifierCSV(file);

      if (!validationResult.valid) {
        setError(validationResult.message);
        return false;
      }

      setDetectedLabels(validationResult.labels || []);
      setShowConfirmation(true);
      return true;
    } catch (error) {
      setError('Error validating file format');
      return false;
    }
  };

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setError('');
      setSuccess(false);
      await validateAndProcessFile(file);
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
      const response = await trainSentenceClassifierFromCSV({
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
      <Paper sx={{ p: 3, mb: 2 }}>
        <Box sx={{ mb: 3 }}>
          <Typography variant="h6" sx={{ mb: 1 }}>
            CSV File Requirements:
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • File must have exactly two columns named 'text' and 'label'
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • Each row should contain a text sample and its corresponding label
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • Labels should be consistent across similar examples
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • At least two different labels must be present
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • Minimum of 10 examples for each label is recommended
          </Typography>
        </Box>

        <Button
          variant="contained"
          component="label"
          startIcon={<CloudUploadIcon />}
          sx={{ mb: 2 }}
        >
          Select CSV File
          <input type="file" hidden accept=".csv" onChange={handleFileSelect} />
        </Button>

        {selectedFile && (
          <Typography variant="body2" sx={{ mb: 2 }}>
            Selected file: {selectedFile.name}
          </Typography>
        )}
      </Paper>

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

      {/* Label Confirmation Dialog */}
      <LabelConfirmationDialog
        open={showConfirmation}
        onClose={() => setShowConfirmation(false)}
        onConfirm={() => {
          setShowConfirmation(false);
          handleUpload();
        }}
        labels={detectedLabels}
      />

      {isUploading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
          <CircularProgress size={24} sx={{ mr: 1 }} />
          <Typography>Training model...</Typography>
        </Box>
      )}
    </Box>
  );
};

export default CSVUpload;
