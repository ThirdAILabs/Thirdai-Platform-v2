import React, { useState } from 'react';
import { Box, Button, Typography, Alert, CircularProgress, Paper } from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import {
  trainNLPTextModel,
  trainTextClassifierWithCSV,
  uploadDocument,
  validateCSV,
} from '@/lib/backend';
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
  const [uploadId, setUploadId] = useState('');
  const validateAndProcessFile = async (file: File) => {
    try {
      const { upload_id } = await uploadDocument(file);

      if (!upload_id) {
        setError('Error uploading file');
        setSelectedFile(null);
        return false;
      }

      setUploadId(upload_id); //set the upload id to the state

      const type = 'text';
      const { labels } = await validateCSV({ upload_id, type });

      if (labels && labels?.length > 0) {
        setDetectedLabels(labels);
        setShowConfirmation(true);
      } else {
        setError('No valid token types found in the file');
        setSelectedFile(null);
        setUploadId('');
      }
      setShowConfirmation(true);
      return true;
    } catch (error) {
      setError('Error validating file format');
      setSelectedFile(null);
      setUploadId('');
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

    if (!detectedLabels.length) {
      setError('No valid labels detected');
      return;
    }

    setIsUploading(true);
    setError('');
    setSuccess(false);

    try {
      const response = await trainNLPTextModel({
        model_name: modelName,
        uploadId: uploadId,
        textColumn: 'source',
        labelColumn: 'target',
        nTargetClasses: detectedLabels.length,
      });

      if (response?.model_id) {
        setSuccess(true);
        onSuccess?.();
      } else {
        throw new Error('Failed to train model');
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
            • File must have exactly two columns named &apos;text&apos; and &apos;label&apos;
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
