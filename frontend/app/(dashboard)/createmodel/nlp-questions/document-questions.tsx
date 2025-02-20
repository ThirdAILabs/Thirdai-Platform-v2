import React, { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  TextField,
  Paper,
  Button,
  Alert,
  CircularProgress,
  Divider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
} from '@mui/material';
import FolderOpenIcon from '@mui/icons-material/FolderOpen';
import DescriptionIcon from '@mui/icons-material/Description';
import FolderIcon from '@mui/icons-material/Folder';
import InfoIcon from '@mui/icons-material/Info';
import { useRouter } from 'next/navigation';
import { uploadDocument, trainNLPTextModel } from '@/lib/backend';

interface DocumentQuestionsProps {
  workflowNames: string[];
  onCreateModel?: (modelId: string) => void;
  stayOnPage?: boolean;
  appName?: string;
}

interface FolderStructure {
  [category: string]: string[]; // category name -> array of file names
}

const DocumentQuestions = ({
  workflowNames,
  onCreateModel,
  stayOnPage,
  appName,
}: DocumentQuestionsProps) => {
  const [modelName, setModelName] = useState(!appName ? '' : appName);
  const [selectedFolder, setSelectedFolder] = useState<FileList | null>(null);
  const [folderStructure, setFolderStructure] = useState<FolderStructure>({});
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [warningMessage, setWarningMessage] = useState('');

  const router = useRouter();

  const handleFolderSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (files && files.length > 0) {
      setSelectedFolder(files);
      setError('');
      setSuccess(false);

      // Process folder structure using webkitRelativePath
      const structure: FolderStructure = {};
      Array.from(files).forEach((file) => {
        const pathParts = file.webkitRelativePath.split('/');
        if (pathParts.length >= 3) {
          // Changed from >= 2 to >= 3 since we expect 3 parts
          const category = pathParts[1]; // Changed from pathParts[0] to pathParts[1]
          if (!structure[category]) {
            structure[category] = [];
          }
          structure[category].push(file.name);
        }
      });

      setFolderStructure(structure);

      // Validate folder structure
      if (Object.keys(structure).length < 2) {
        setError('At least 2 different categories (subfolders) are required');
        return;
      }

      try {
        const { upload_id } = await uploadDocument(files);
        if (!upload_id) {
          setError('Error validating folder structure');
        }
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : 'Error validating folder structure';
        setError(errorMessage);
      }
    }
  };

  const validateFolderStructure = (): boolean => {
    if (Object.keys(folderStructure).length === 0) {
      setError('Please select a folder containing your document categories');
      return false;
    }

    // Get actual categories (first level folders)
    const categories = new Set(
      Array.from(selectedFolder!)
        .map((file) => {
          const parts = file.webkitRelativePath.split('/');
          return parts.length >= 3 ? parts[1] : ''; // Changed from parts[0] to parts[1]
        })
        .filter(Boolean)
    );

    if (categories.size < 2) {
      setError('At least 2 different categories (subfolders) are required');
      return false;
    }

    // Check for minimum documents per category
    const categoryFiles = new Map<string, number>();
    Array.from(selectedFolder!).forEach((file) => {
      const pathParts = file.webkitRelativePath.split('/');
      if (pathParts.length >= 3) {
        // Changed from >= 2 to >= 3
        const category = pathParts[1]; // Changed from pathParts[0] to pathParts[1]
        categoryFiles.set(category, (categoryFiles.get(category) || 0) + 1);
      }
    });

    const insufficientCategories = Array.from(categoryFiles.entries())
      .filter(([_, count]) => count < 10)
      .map(([category]) => category);

    if (insufficientCategories.length > 0) {
      setError(
        `The following categories have fewer than 10 documents: ${insufficientCategories.join(', ')}`
      );
      return false;
    }

    // Check file extensions
    const validExtensions = ['.txt', '.doc', '.docx', '.pdf'];
    const invalidFiles: string[] = [];

    Array.from(selectedFolder!).forEach((file) => {
      const extension = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));
      if (!validExtensions.includes(extension)) {
        invalidFiles.push(file.name);
      }
    });

    if (invalidFiles.length > 0) {
      setError(
        `Invalid file types found: ${invalidFiles.join(', ')}. Only .txt, .doc, .docx, and .pdf files are supported.`
      );
      return false;
    }

    return true;
  };

  const handleUpload = async () => {
    if (!selectedFolder) {
      setError('Please select a folder first');
      return;
    }

    if (!modelName) {
      setError('Please enter an app name');
      return;
    }

    if (!validateFolderStructure()) {
      return;
    }

    setIsUploading(true);
    setError('');
    setSuccess(false);

    try {
      // First validate the folder structure
      const { upload_id } = await uploadDocument(selectedFolder);

      if (!upload_id) {
        setError('Error validating folder structure');
        return;
      }

      // Extract the categories from folder structure
      const categories = new Set(
        Array.from(selectedFolder!)
          .map((file) => {
            const parts = file.webkitRelativePath.split('/');
            return parts.length >= 3 ? parts[1] : '';
          })
          .filter(Boolean)
      );

      // If validation passes, proceed with training
      const trainingResult = await trainNLPTextModel({
        uploadId: upload_id,
        model_name: modelName,
        nTargetClasses: categories.size,
        doc_classification: true, // Make sure this is set to true
      });

      setSuccess(true);
      onCreateModel?.(trainingResult.model_id);

      if (!stayOnPage) {
        router.push('/');
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'An error occurred while training the model';
      setError(errorMessage);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div>
      {/* App name section */}
      <Box sx={{ mb: 4 }}>
        <span className="block text-lg font-semibold">App Name</span>
        <TextField
          className="text-md w-full"
          value={modelName}
          onChange={(e) => {
            const name = e.target.value;
            setModelName(name);
          }}
          onBlur={(e) => {
            const name = e.target.value;
            const regexPattern = /^[\w-]+$/;
            let warningMessage = '';

            if (name.includes(' ')) {
              warningMessage = 'The app name cannot contain spaces. Please remove the spaces.';
            } else if (name.includes('.')) {
              warningMessage =
                "The app name cannot contain periods ('.'). Please remove the periods.";
            } else if (!regexPattern.test(name)) {
              warningMessage =
                'The app name can only contain letters, numbers, underscores, and hyphens. Please modify the name.';
            } else if (workflowNames.includes(name)) {
              warningMessage =
                'An app with the same name already exists. Please choose a different name.';
            }
            setWarningMessage(warningMessage);
            setModelName(name);
          }}
          placeholder="Enter app name"
          style={{ marginTop: '10px' }}
          disabled={!!appName && !workflowNames.includes(modelName)}
        />
        {warningMessage && (
          <span style={{ color: 'red', marginTop: '10px' }}>{warningMessage}</span>
        )}
      </Box>

      <Divider sx={{ my: 4 }} />

      {/* Folder Upload Section */}
      <Box sx={{ width: '100%' }}>
        <Typography variant="h6" gutterBottom>
          Upload Your Document Classification Dataset
        </Typography>

        <Paper sx={{ p: 3, mb: 3 }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
            {/* Quick Overview */}
            <Box>
              <Typography variant="subtitle1" color="primary" gutterBottom>
                Dataset Requirements
              </Typography>
              <List dense>
                <ListItem>
                  <ListItemIcon>
                    <FolderIcon color="primary" />
                  </ListItemIcon>
                  <ListItemText
                    primary="Select a main folder containing category subfolders"
                    secondary="Each subfolder name represents a document type/category"
                  />
                </ListItem>
                <ListItem>
                  <ListItemIcon>
                    <DescriptionIcon color="primary" />
                  </ListItemIcon>
                  <ListItemText
                    primary="Place example documents in each category folder"
                    secondary="Minimum 10 documents per category (.txt, .doc, .docx, .pdf)"
                  />
                </ListItem>
                <ListItem>
                  <ListItemIcon>
                    <InfoIcon color="primary" />
                  </ListItemIcon>
                  <ListItemText
                    primary="Minimum 2 different categories required"
                    secondary="Example: invoices/, receipts/, contracts/"
                  />
                </ListItem>
              </List>
            </Box>

            {/* Visual Example */}
            <Box sx={{ bgcolor: '#f5f5f5', p: 2, borderRadius: 1 }}>
              <Typography variant="subtitle2" gutterBottom>
                Example Structure:
              </Typography>
              <Box sx={{ fontFamily: 'monospace', pl: 2, fontSize: '0.9rem' }}>
                📁 my-documents/
                <br />
                ┣ 📁 invoices/
                <br />
                ┃ ┣ 📄 invoice1.pdf
                <br />
                ┃ ┣ 📄 invoice2.pdf
                <br />
                ┃ ┗ 📄 invoice3.pdf
                <br />
                ┗ 📁 contracts/
                <br />
                &nbsp;&nbsp; ┣ 📄 contract1.pdf
                <br />
                &nbsp;&nbsp; ┣ 📄 contract2.pdf
                <br />
                &nbsp;&nbsp; ┗ 📄 contract3.pdf
              </Box>
            </Box>

            {/* Upload Button */}
            <Box>
              <Button
                variant="contained"
                component="label"
                startIcon={<FolderOpenIcon />}
                size="large"
                sx={{ mb: 2 }}
              >
                Select Folder
                <input
                  type="file"
                  hidden
                  {...{ webkitdirectory: '', directory: '' }}
                  onChange={handleFolderSelect}
                />
              </Button>

              {selectedFolder && (
                <Typography variant="body2" sx={{ mt: 1 }}>
                  Selected folder contains {selectedFolder.length} files
                </Typography>
              )}
            </Box>
          </Box>
        </Paper>

        {/* Add folder structure preview */}
        {Object.keys(folderStructure).length > 0 && (
          <Paper sx={{ p: 2, mb: 2, backgroundColor: '#f5f5f5' }}>
            <Typography variant="subtitle1" sx={{ mb: 1 }}>
              Selected Folder Structure:
            </Typography>
            {Object.entries(folderStructure).map(([category, files]) => (
              <Box key={category} sx={{ mb: 1 }}>
                <Typography variant="body2" sx={{ fontWeight: 'bold' }}>
                  {category}/ ({files.length} documents)
                </Typography>
              </Box>
            ))}
          </Paper>
        )}

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {success && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Document classification model training initiated successfully!
          </Alert>
        )}

        {selectedFolder && !success && (
          <Button
            variant="contained"
            onClick={handleUpload}
            disabled={isUploading || !!warningMessage}
            sx={{ mt: 2 }}
          >
            {isUploading ? 'Training Model...' : 'Train Model'}
          </Button>
        )}

        {isUploading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
            <CircularProgress size={24} sx={{ mr: 1 }} />
            <Typography>Training model...</Typography>
          </Box>
        )}
      </Box>
    </div>
  );
};

export default DocumentQuestions;
