import React, { useState, useEffect } from 'react';
import {
  TextField,
  Button,
  Paper,
  Alert,
  Typography,
  CircularProgress,
  IconButton,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

interface Endpoint {
  id: string;
  name: string;
  endpoint: string;
}

interface FormData {
  name: string;
  endpoint: string;
  apiKey: string;
}

interface APIResponse {
  success: boolean;
  message?: string;
  endpoints?: Endpoint[];
}

export default function SelfHostLLM() {
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [formData, setFormData] = useState<FormData>({
    name: '',
    endpoint: '',
    apiKey: '',
  });

  useEffect(() => {
    fetchEndpoints();
  }, []);

  const fetchEndpoints = async (): Promise<void> => {
    try {
      const response = await fetch('/endpoints/get_self_host_endpoints');
      const data: APIResponse = await response.json();
      setEndpoints(data.endpoints || []);
    } catch (error) {
      setError('Failed to fetch endpoints');
    }
  };

  const validateEndpoint = async (): Promise<boolean> => {
    try {
      setLoading(true);
      const response = await fetch('/endpoints/test_endpoint', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      const data: APIResponse = await response.json();
      if (!data.success) throw new Error(data.message);

      return true;
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to validate endpoint');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!formData.name || !formData.endpoint || !formData.apiKey) {
      setError('All fields are required');
      return;
    }

    const isValid = await validateEndpoint();
    if (!isValid) return;

    try {
      const response = await fetch('/endpoints/add_self_host_endpoint', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      const data: APIResponse = await response.json();
      if (!data.success) throw new Error(data.message);

      setSuccess('Endpoint added successfully');
      setFormData({ name: '', endpoint: '', apiKey: '' });
      fetchEndpoints();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to add endpoint');
    }
  };

  const handleDelete = async (endpointId: string): Promise<void> => {
    try {
      const response = await fetch('/endpoints/delete_self_host_endpoint', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: endpointId }),
      });

      const data: APIResponse = await response.json();
      if (!data.success) throw new Error(data.message);

      setSuccess('Endpoint deleted successfully');
      fetchEndpoints();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to delete endpoint');
    }
  };

  const handleInputChange = (field: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({ ...prev, [field]: e.target.value }));
  };

  return (
    <div>
      <Paper elevation={1} sx={{ p: 3, mb: 3 }}>
        <Alert severity="info" sx={{ mb: 3 }}>
          Ensure your self-hosted API endpoint is compatible with OpenAI API streaming/generate
          endpoint format: https://platform.openai.com/docs/api-reference/making-requests
        </Alert>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: '16px' }}>
            <Typography variant="subtitle1">Endpoint Name</Typography>
            <TextField
              fullWidth
              value={formData.name}
              onChange={handleInputChange('name')}
              placeholder="Enter a unique identifier"
              margin="dense"
            />
          </div>

          <div style={{ marginBottom: '16px' }}>
            <Typography variant="subtitle1">API Endpoint</Typography>
            <TextField
              fullWidth
              value={formData.endpoint}
              onChange={handleInputChange('endpoint')}
              placeholder="https://your-api-endpoint.com/v1/chat/completions"
              margin="dense"
            />
          </div>

          <div style={{ marginBottom: '16px' }}>
            <Typography variant="subtitle1">API Key</Typography>
            <TextField
              fullWidth
              type="password"
              value={formData.apiKey}
              onChange={handleInputChange('apiKey')}
              placeholder="Enter your API key"
              margin="dense"
            />
          </div>

          <Button
            type="submit"
            variant="contained"
            fullWidth
            disabled={loading}
            startIcon={loading && <CircularProgress size={20} />}
          >
            {loading ? 'Testing...' : 'Add Endpoint'}
          </Button>
        </form>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}
        {success && (
          <Alert severity="success" sx={{ mt: 2 }}>
            {success}
          </Alert>
        )}
      </Paper>

      {endpoints.length > 0 && (
        <Paper elevation={1} sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Existing Endpoints
          </Typography>
          <List>
            {endpoints.map((endpoint) => (
              <ListItem key={endpoint.id} divider>
                <ListItemText primary={endpoint.name} secondary={endpoint.endpoint} />
                <ListItemSecondaryAction>
                  <IconButton
                    edge="end"
                    aria-label="delete"
                    onClick={() => handleDelete(endpoint.id)}
                  >
                    <DeleteIcon />
                  </IconButton>
                </ListItemSecondaryAction>
              </ListItem>
            ))}
          </List>
        </Paper>
      )}
    </div>
  );
}
