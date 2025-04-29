'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  Box,
  Button,
  Typography,
  TextField,
  Card,
  CardContent,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  FormHelperText,
  CircularProgress,
  Alert,
  Breadcrumbs,
  Link as MuiLink,
} from '@mui/material';

export default function NewJobPage() {
  const params = useParams();
  const router = useRouter();
  const deploymentId = params.deploymentId as string;
  
  const [jobName, setJobName] = useState('');
  const [fileType, setFileType] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [fileName, setFileName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  
  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      const selectedFile = event.target.files[0];
      setFile(selectedFile);
      setFileName(selectedFile.name);
    }
  };
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!jobName.trim()) {
      setError('Job name is required');
      return;
    }
    
    if (!fileType) {
      setError('File type is required');
      return;
    }
    
    if (!file) {
      setError('Please upload a file');
      return;
    }
    
    setError(null);
    setIsSubmitting(true);
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1500));
      setSuccess(true);
      
      // Redirect after success
      setTimeout(() => {
        router.push(`/token-classification/${deploymentId}/jobs`);
      }, 2000);
    } catch (err) {
      setError('Failed to create job. Please try again.');
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <div className="bg-muted min-h-screen">
      <header className="w-full p-4 bg-muted border-b">
        <div className="max-w-7xl mx-auto space-y-4">
          <Breadcrumbs>
            <MuiLink href="#" underline="hover" color="inherit">
              Token Classification
            </MuiLink>
            <MuiLink 
              href={`/token-classification/${deploymentId}`} 
              underline="hover" 
              color="inherit"
            >
              {deploymentId}
            </MuiLink>
            <MuiLink 
              href={`/token-classification/${deploymentId}/jobs`} 
              underline="hover" 
              color="inherit"
            >
              Jobs
            </MuiLink>
            <Typography color="text.primary">New Job</Typography>
          </Breadcrumbs>
          
          <Typography variant="h5" className="font-bold">
            Create New Classification Job
          </Typography>
        </div>
      </header>

      <main className="max-w-3xl mx-auto p-4">
        <Card>
          <CardContent>
            {success ? (
              <Alert severity="success" sx={{ mb: 2 }}>
                Job created successfully! Redirecting...
              </Alert>
            ) : (
              <form onSubmit={handleSubmit}>
                {error && (
                  <Alert severity="error" sx={{ mb: 3 }}>
                    {error}
                  </Alert>
                )}
                
                <Box sx={{ mb: 3 }}>
                  <TextField
                    label="Job Name"
                    variant="outlined"
                    fullWidth
                    value={jobName}
                    onChange={(e) => setJobName(e.target.value)}
                    required
                    disabled={isSubmitting}
                  />
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <FormControl fullWidth required>
                    <InputLabel>File Type</InputLabel>
                    <Select
                      value={fileType}
                      onChange={(e) => setFileType(e.target.value)}
                      label="File Type"
                      disabled={isSubmitting}
                    >
                      <MenuItem value="csv">CSV</MenuItem>
                      <MenuItem value="txt">TXT</MenuItem>
                      <MenuItem value="excel">Excel</MenuItem>
                      <MenuItem value="pdf">PDF</MenuItem>
                    </Select>
                    <FormHelperText>
                      Select the type of file you want to process
                    </FormHelperText>
                  </FormControl>
                </Box>
                
                <Box sx={{ mb: 4 }}>
                  <Typography variant="subtitle2" gutterBottom>
                    Upload File
                  </Typography>
                  <Box
                    sx={{
                      border: '2px dashed #ccc',
                      borderRadius: 1,
                      p: 3,
                      textAlign: 'center',
                      cursor: 'pointer',
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
                      onChange={handleFileChange}
                      disabled={isSubmitting}
                    />
                    
                    {fileName ? (
                      <Typography>{fileName}</Typography>
                    ) : (
                      <Typography color="text.secondary">
                        Click to select a file or drag and drop here
                      </Typography>
                    )}
                  </Box>
                </Box>
                
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Button
                    variant="outlined"
                    onClick={() => router.push(`/token-classification/${deploymentId}/jobs`)}
                    disabled={isSubmitting}
                  >
                    Cancel
                  </Button>
                  
                  <Button
                    type="submit"
                    variant="contained"
                    disabled={isSubmitting}
                    startIcon={isSubmitting ? <CircularProgress size={20} /> : null}
                  >
                    {isSubmitting ? 'Creating...' : 'Create Job'}
                  </Button>
                </Box>
              </form>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
} 