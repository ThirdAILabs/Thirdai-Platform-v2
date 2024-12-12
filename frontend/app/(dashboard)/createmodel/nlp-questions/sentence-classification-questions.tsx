import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  TextField,
  Divider,
} from '@mui/material';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import ModelTrainingIcon from '@mui/icons-material/ModelTraining';
import SyntheticClassification from './sentence-questions-synthetic/SyntheticClassification';
import CSVUpload from './sentence-questions-upload-file/CSVUpload';

const CREATION_METHODS = {
  // PRETRAINED: 'pretrained',
  UPLOAD_DATA: 'upload-data',
  SYNTHETIC: 'synthetic',
};

interface SCQQuestionsProps {
  question: string;
  answer: string;
  workflowNames: string[];
  onCreateModel?: (modelId: string) => void;
  stayOnPage?: boolean;
  appName?: string;
}

const SCQQuestions = ({
  question,
  answer,
  workflowNames,
  onCreateModel,
  stayOnPage,
  appName,
}: SCQQuestionsProps) => {
  const [creationMethod, setCreationMethod] = useState<string>('');
  const [modelName, setModelName] = useState(!appName ? '' : appName);
  const [warningMessage, setWarningMessage] = useState('');
  const router = useRouter();

  useEffect(() => {
    if (appName) {
      if (workflowNames.includes(appName)) {
        setWarningMessage(
          'An App with the same name has been created. Please choose a different name.'
        );
      } else {
        setWarningMessage('');
      }
    }
  }, [appName, workflowNames]);

  const renderCreationMethodSelection = () => (
    <Box sx={{ width: '100%', mb: 4 }}>
      <Typography variant="h6" sx={{ mb: 2 }}>
        Choose how you want to create your Text Classification model
      </Typography>

      <Grid container spacing={2}>
        {/* Pretrained Model Option */}
        {/* <Grid item xs={12} md={4}>
          <Card
            sx={{
              height: '100%',
              cursor: 'pointer',
              border:
                creationMethod === CREATION_METHODS.PRETRAINED
                  ? '2px solid #1976d2'
                  : '1px solid #e0e0e0',
              '&:hover': { borderColor: '#1976d2' },
            }}
            onClick={() => setCreationMethod(CREATION_METHODS.PRETRAINED)}
          >
            <CardContent
              sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
              }}
            >
              <ModelTrainingIcon sx={{ fontSize: 40, mb: 2, color: 'primary.main' }} />
              <Typography variant="h6" gutterBottom>
                Use Pretrained Model
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Quick start with our pretrained models for common tasks like sentiment analysis or
                content moderation
              </Typography>
            </CardContent>
          </Card>
        </Grid> */}

        {/* Upload Data Option */}
        <Grid item xs={12} md={4}>
          <Card
            sx={{
              height: '100%',
              cursor: 'pointer',
              border:
                creationMethod === CREATION_METHODS.UPLOAD_DATA
                  ? '2px solid #1976d2'
                  : '1px solid #e0e0e0',
              '&:hover': { borderColor: '#1976d2' },
            }}
            onClick={() => setCreationMethod(CREATION_METHODS.UPLOAD_DATA)}
          >
            <CardContent
              sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
              }}
            >
              <UploadFileIcon sx={{ fontSize: 40, mb: 2, color: 'primary.main' }} />
              <Typography variant="h6" gutterBottom>
                Upload Your Data
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Train with your own labeled dataset via a CSV uploaded from your computer or S3
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        {/* Synthetic Data Option */}
        <Grid item xs={12} md={4}>
          <Card
            sx={{
              height: '100%',
              cursor: 'pointer',
              border:
                creationMethod === CREATION_METHODS.SYNTHETIC
                  ? '2px solid #1976d2'
                  : '1px solid #e0e0e0',
              '&:hover': { borderColor: '#1976d2' },
            }}
            onClick={() => setCreationMethod(CREATION_METHODS.SYNTHETIC)}
          >
            <CardContent
              sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
              }}
            >
              <AutoAwesomeIcon sx={{ fontSize: 40, mb: 2, color: 'primary.main' }} />
              <Typography variant="h6" gutterBottom>
                Generate Training Data
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Provide examples and let AI generate training data for your custom classification
                needs
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );

  const renderSelectedMethod = () => {
    switch (creationMethod) {
      // case CREATION_METHODS.PRETRAINED:
      //   return (
      //     <Box sx={{ width: '100%' }}>
      //       Coming Soon
      //       {/* <Typography variant="h6" gutterBottom>
      //         Select a Pretrained Model
      //       </Typography>
      //       <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
      //         Choose from our collection of pretrained models for common text classification tasks.
      //       </Typography> */}
      //     </Box>
      //   );
      case CREATION_METHODS.UPLOAD_DATA:
        return (
          <Box sx={{ width: '100%' }}>
            <Typography variant="h6" gutterBottom>
              Upload Training Data
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Upload a CSV file containing your labeled training data. The file should include text
              samples and their corresponding labels.
            </Typography>
            <CSVUpload
              modelName={modelName}
              onSuccess={() => {
                if (!stayOnPage) {
                  router.push('/');
                }
                onCreateModel?.(modelName);
              }}
              onError={(errorMessage: string) => {
                console.error('Error training model:', errorMessage);
              }}
            />
          </Box>
        );
      case CREATION_METHODS.SYNTHETIC:
        return (
          <SyntheticClassification
            workflowNames={workflowNames}
            question={question}
            answer={answer}
            onModelCreated={onCreateModel}
            modelName={modelName}
          />
        );
      default:
        return null;
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

      {/* Method selection and content */}
      {renderCreationMethodSelection()}
      {creationMethod && <Divider sx={{ my: 4 }} />}
      {renderSelectedMethod()}
    </div>
  );
};

export default SCQQuestions;
