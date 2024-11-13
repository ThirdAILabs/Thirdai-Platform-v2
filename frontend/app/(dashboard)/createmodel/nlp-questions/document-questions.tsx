import React, { useEffect, useState } from 'react';
import { 
  Box, 
  Typography, 
  TextField,
  Paper,
  Button,
  Alert,
  CircularProgress,
  Divider
} from '@mui/material';
import FolderOpenIcon from '@mui/icons-material/FolderOpen';
import { useRouter } from 'next/navigation';

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
  appName 
}: DocumentQuestionsProps) => {
  const [modelName, setModelName] = useState(!appName ? '' : appName);
  const [selectedFolder, setSelectedFolder] = useState<FileList | null>(null);
  const [folderStructure, setFolderStructure] = useState<FolderStructure>({});
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [warningMessage, setWarningMessage] = useState('');
  
  const router = useRouter();

  const handleFolderSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (files && files.length > 0) {
      setSelectedFolder(files);
      setError('');
      setSuccess(false);
      
      // Process folder structure
      const structure: FolderStructure = {};
      Array.from(files).forEach((file) => {
        const path = file.webkitRelativePath;
        const [category] = path.split('/').filter(Boolean);
        
        if (!structure[category]) {
          structure[category] = [];
        }
        structure[category].push(file.name);
      });
      
      setFolderStructure(structure);
    }
  };

  const validateFolderStructure = (): boolean => {
    // Check if folder structure exists
    if (Object.keys(folderStructure).length === 0) {
      setError('Please select a folder containing your document categories');
      return false;
    }

    // Check for minimum number of categories
    if (Object.keys(folderStructure).length < 2) {
      setError('At least 2 different categories (subfolders) are required');
      return false;
    }

    // Check for minimum documents per category
    const insufficientCategories = Object.entries(folderStructure)
      .filter(([_, files]) => files.length < 10)
      .map(([category]) => category);

    if (insufficientCategories.length > 0) {
      setError(`The following categories have fewer than 10 documents: ${insufficientCategories.join(', ')}`);
      return false;
    }

    // Check file extensions
    const validExtensions = ['.txt', '.doc', '.docx', '.pdf'];
    const invalidFiles: string[] = [];

    Object.values(folderStructure).flat().forEach(filename => {
      const extension = filename.toLowerCase().substring(filename.lastIndexOf('.'));
      if (!validExtensions.includes(extension)) {
        invalidFiles.push(filename);
      }
    });

    if (invalidFiles.length > 0) {
      setError(`Invalid file types found: ${invalidFiles.join(', ')}. Only .txt, .doc, .docx, and .pdf files are supported.`);
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
      // Create FormData with all files
      const formData = new FormData();
      formData.append('modelName', modelName);
      Array.from(selectedFolder).forEach((file) => {
        formData.append('files', file, file.webkitRelativePath);
      });

      // Mock API call for now - replace with actual API endpoint
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      setSuccess(true);
      onCreateModel?.(modelName);
      
      if (!stayOnPage) {
        router.push('/');
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'An error occurred while training the model';
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
              warningMessage = "The app name cannot contain periods ('.'). Please remove the periods.";
            } else if (!regexPattern.test(name)) {
              warningMessage = 'The app name can only contain letters, numbers, underscores, and hyphens. Please modify the name.';
            } else if (workflowNames.includes(name)) {
              warningMessage = 'An app with the same name already exists. Please choose a different name.';
            }
            setWarningMessage(warningMessage);
            setModelName(name);
          }}
          placeholder="Enter app name"
          style={{ marginTop: '10px' }}
          disabled={!!appName && !workflowNames.includes(modelName)}
        />
        {warningMessage && <span style={{ color: 'red', marginTop: '10px' }}>{warningMessage}</span>}
      </Box>

      <Divider sx={{ my: 4 }} />

      {/* Folder Upload Section */}
      <Box sx={{ width: '100%' }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Upload your document classification dataset
        </Typography>

        <Paper sx={{ p: 3, mb: 3 }}>
          <Box sx={{ mb: 3 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              Folder Structure Requirements:
            </Typography>
            
            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle1" color="primary" sx={{ mb: 1 }}>
                Required Structure:
              </Typography>
              <Box sx={{ pl: 2 }}>
                <Typography variant="body2" color="text.secondary">
                  main-folder/
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ pl: 2 }}>
                  ├── category1/
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ pl: 4 }}>
                  ├── document1.txt
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ pl: 4 }}>
                  └── document2.txt
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ pl: 2 }}>
                  └── category2/
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ pl: 4 }}>
                  ├── document1.txt
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ pl: 4 }}>
                  └── document2.txt
                </Typography>
              </Box>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Typography variant="subtitle1" color="primary" sx={{ mb: 1 }}>
                Key Points:
              </Typography>
              <Typography variant="body2" color="text.secondary">
                • Folder names represent document categories/labels
              </Typography>
              <Typography variant="body2" color="text.secondary">
                • Each subfolder must contain example documents for that category
              </Typography>
              <Typography variant="body2" color="text.secondary">
                • At least 2 different categories (subfolders) required
              </Typography>
              <Typography variant="body2" color="text.secondary">
                • Minimum 10 documents per category recommended
              </Typography>
              <Typography variant="body2" color="text.secondary">
                • Supported file formats: .txt, .doc, .docx, .pdf
              </Typography>
            </Box>
          </Box>

          <Button
            variant="contained"
            component="label"
            startIcon={<FolderOpenIcon />}
            sx={{ mb: 2 }}
          >
            Select Folder
            <input
              type="file"
              hidden
              // Enable folder selection
              {...{ webkitdirectory: "", directory: "" }}
              onChange={handleFolderSelect}
            />
          </Button>

          {selectedFolder && (
            <Typography variant="body2">
              Selected folder contains {selectedFolder.length} files
            </Typography>
          )}
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