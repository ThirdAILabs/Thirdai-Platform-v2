import React, { useState } from 'react';
import {
  Box,
  Button,
  Typography,
  Alert,
  Paper,
  CircularProgress,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import { trainTokenClassifierWithCSV, validateCSV, uploadDocument } from '@/lib/backend';
import { set } from 'lodash';

interface CSVUploadProps {
  modelName: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
  workflowNames?: string[];
}

interface TokenTypeDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  tokenTypes: string[];
}

const TokenTypeDialog = ({ open, onClose, onConfirm, tokenTypes }: TokenTypeDialogProps) => (
  <Dialog open={open} onClose={onClose}>
    <DialogTitle>Confirm Token Types</DialogTitle>
    <DialogContent>
      <Typography variant="body1" sx={{ mb: 2 }}>
        The following token types were detected in your data:
      </Typography>
      <Box component="ul" sx={{ pl: 2 }}>
        {tokenTypes.map((type) => (
          <Typography component="li" key={type}>
            {type}
          </Typography>
        ))}
      </Box>
    </DialogContent>
    <DialogActions>
      <Button onClick={onClose} color="inherit">
        Cancel
      </Button>
      <Button onClick={onConfirm} variant="contained">
        Confirm & Train
      </Button>
    </DialogActions>
  </Dialog>
);

const CSVUpload = ({ modelName, onSuccess, onError, workflowNames = [] }: CSVUploadProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [warningMessage, setWarningMessage] = useState('');
  const [detectedLabels, setDetectedLabels] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [UploadId, setUploadId] = useState<string>('');
  const validateModelName = (name: string) => {
    if (workflowNames.includes(name)) {
      setWarningMessage('A model with this name already exists. Please choose a different name.');
      return false;
    }

    const isValid = /^[a-zA-Z0-9-_]+$/.test(name);
    const isNotEmpty = name.trim().length > 0;

    if (!isValid && isNotEmpty) {
      setWarningMessage('Model name can only contain letters, numbers, underscores, and hyphens.');
      return false;
    }

    if (name.includes(' ')) {
      setWarningMessage('Model name cannot contain spaces.');
      return false;
    }

    setWarningMessage('');
    return isValid && isNotEmpty;
  };

  const handleFileInput = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.type !== 'text/csv') {
      setError('Please upload a CSV file');
      setSelectedFile(null);
      return;
    }

    const MAX_FILE_SIZE = 500 * 1024 * 1024; // 500MB
    if (file.size > MAX_FILE_SIZE) {
      setError('File size must be less than 500MB');
      setSelectedFile(null);
      return;
    }

    setSelectedFile(file);
    setError('');
    setIsValidating(true);

    try {
      const { upload_id } = await uploadDocument(file);

      if (!upload_id) {
        setError('Error uploading file');
        setSelectedFile(null);
        return;
      }

      setUploadId(upload_id); //set the upload id to the state

      const type = 'token';
      const { labels } = await validateCSV({ upload_id, type });
      if (labels && labels?.length > 0) {
        setDetectedLabels(labels);
        setShowConfirmation(true);
      } else {
        setError('No valid token types found in the file');
        setSelectedFile(null);
        setUploadId('');
      }
    } catch (error) {
      setError('Error validating file format');
      setSelectedFile(null);
      setUploadId('');
    } finally {
      setIsValidating(false);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('Please select a CSV file first');
      return;
    }

    if (!validateModelName(modelName)) {
      setError('Please enter a valid model name');
      return;
    }

    if (!detectedLabels.length) {
      setError('No valid token types detected');
      return;
    }

    setIsUploading(true);
    setError('');
    setSuccess(false);

    try {
      const response = await trainTokenClassifierWithCSV({
        model_name: modelName,
        model_options: {
          target_labels: detectedLabels,
          source_column: 'source',
          target_column: 'target',
          default_tag: 'O',
        },
        data: {
          supervised_files: [
            {
              path: UploadId,
              location: 'upload',
            },
          ],
        },
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
    <Paper sx={{ p: 3, width: '100%' }}>
      <Box sx={{ mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Typography variant="h5" component="h2" sx={{ mr: 1 }}>
            Train Token Classification Model
          </Typography>
          <Tooltip title="Upload a CSV file to train a new token classification model">
            <HelpOutlineIcon color="action" fontSize="small" />
          </Tooltip>
        </Box>

        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          CSV File Requirements:
        </Typography>
        <Box component="ul" sx={{ pl: 2, mb: 3 }}>
          <Typography component="li" variant="body2" color="text.secondary">
            • File must have exactly two columns named &apos;source&apos; and &apos;target&apos;
          </Typography>
          <Typography component="li" variant="body2" color="text.secondary">
            • Each row must have the same number of tokens in source and target
          </Typography>
          <Typography component="li" variant="body2" color="text.secondary">
            • Target column should contain labels for each token (use &apos;O&apos; for
            non-entities)
          </Typography>
          <Typography component="li" variant="body2" color="text.secondary">
            • At least one token type other than &apos;O&apos; must be present
          </Typography>
        </Box>
      </Box>

      <Box
        sx={{
          border: '2px dashed',
          borderColor: 'divider',
          borderRadius: 1,
          p: 3,
          textAlign: 'center',
          cursor: 'pointer',
          mb: 3,
          '&:hover': {
            borderColor: 'primary.main',
          },
        }}
        onClick={() => document.getElementById('file-input')?.click()}
      >
        <input
          type="file"
          id="file-input"
          style={{ display: 'none' }}
          accept=".csv"
          onChange={handleFileInput}
        />
        <CloudUploadIcon sx={{ mb: 1, color: 'action.active' }} />
        <Typography color={selectedFile ? 'success.main' : 'text.secondary'}>
          {selectedFile ? `Selected: ${selectedFile.name}` : 'Click to select a CSV file'}
        </Typography>
      </Box>

      {isValidating && (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2, mb: 2 }}>
          <CircularProgress size={24} sx={{ mr: 1 }} />
          <Typography>Validating file...</Typography>
        </Box>
      )}

      {warningMessage && (
        <Alert severity="warning" sx={{ mb: 2 }}>
          {warningMessage}
        </Alert>
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

      <TokenTypeDialog
        open={showConfirmation}
        onClose={() => setShowConfirmation(false)}
        onConfirm={() => {
          setShowConfirmation(false);
          handleUpload();
        }}
        tokenTypes={detectedLabels}
      />

      {isUploading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
          <CircularProgress size={24} sx={{ mr: 1 }} />
          <Typography>Training model...</Typography>
        </Box>
      )}
    </Paper>
  );
};

export default CSVUpload;
