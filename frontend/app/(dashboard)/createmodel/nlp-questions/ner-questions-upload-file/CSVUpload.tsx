import React, { useState } from 'react';
import { 
  Box, 
  Button, 
  Typography, 
  Alert, 
  CircularProgress, 
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Chip,
  Stack
} from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import { trainTokenClassifierFromCSV, validateTokenClassifierCSV } from '@/lib/backend';

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
      const validationResult = await validateTokenClassifierCSV(file);
      
      if (!validationResult.valid) {
        setError(validationResult.message);
        return false;
      }
  
      // Show detected labels in confirmation dialog
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
      const errorMessage = error instanceof Error ? error.message : 'An error occurred while training the model';
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
            • File must have exactly two columns named 'source' and 'target'
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • Each row must have the same number of tokens in source and target
          </Typography>
          <Typography variant="body2" color="text.secondary">
            • Target column should contain labels for each token (use 'O' for non-entities)
          </Typography>
        </Box>

        <Button
          variant="contained"
          component="label"
          startIcon={<CloudUploadIcon />}
          sx={{ mb: 2 }}
        >
          Select CSV File
          <input
            type="file"
            hidden
            accept=".csv"
            onChange={handleFileSelect}
          />
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

      {/* Confirmation Dialog */}
      <Dialog open={showConfirmation} onClose={() => setShowConfirmation(false)}>
        <DialogTitle>Confirm Token Types</DialogTitle>
        <DialogContent>
          <Typography sx={{ mb: 2 }}>
            The following token types were detected in your CSV:
          </Typography>
          <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mb: 2 }}>
            {detectedLabels.map((label) => (
              <Chip key={label} label={label} color="primary" sx={{ mb: 1 }} />
            ))}
          </Stack>
          <Typography>
            Are you sure you want to proceed with training using these token types?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowConfirmation(false)}>Cancel</Button>
          <Button
            onClick={() => {
              setShowConfirmation(false);
              handleUpload();
            }}
            variant="contained"
          >
            Proceed with Training
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default CSVUpload;
