'use client';

import React from 'react';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

export default function ConfigurePage() {
  const searchParams = useSearchParams();
  const [cacheEnabled, setCacheEnabled] = useState(false);
  const [rerankEnabled, setRerankEnabled] = useState(false);
  const [llmEndpoint, setLlmEndpoint] = useState('');
  const workflowId = searchParams.get('id');
  const hasLLM = searchParams.get('llm') === 'true';
  
  const endpoints = ['endpoint1', 'endpoint2', 'endpoint3'];

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const response = await fetch(`/api/config/${workflowId}`);
        const data = await response.json();
        setCacheEnabled(data.cacheEnabled);
        setRerankEnabled(data.rerankEnabled);
        if (hasLLM && data.llmEndpoint) setLlmEndpoint(data.llmEndpoint);
      } catch (error) {
        console.error('Error fetching config:', error);
      }
    };
    
    fetchConfig();
  }, [workflowId, hasLLM]);

  const handleSave = async () => {
    try {
      await fetch(`/api/config/${workflowId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          cacheEnabled,
          rerankEnabled,
          ...(hasLLM && { llmEndpoint })
        })
      });
    } catch (error) {
      console.error('Error saving config:', error);
    }
  };

  return (
    <Box sx={{ maxWidth: 600, margin: '0 auto', padding: 4 }}>
      <Typography variant="h4" gutterBottom>
        Configure Application
      </Typography>
      
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
        <FormControlLabel
          control={
            <Switch
              checked={cacheEnabled}
              onChange={(e) => setCacheEnabled(e.target.checked)}
            />
          }
          label="LLM Generation Cache"
        />

        <FormControlLabel
          control={
            <Switch
              checked={rerankEnabled}
              onChange={(e) => setRerankEnabled(e.target.checked)}
            />
          }
          label="Search Rerank"
        />

        {hasLLM && (
          <FormControl fullWidth>
            <InputLabel>LLM Endpoint</InputLabel>
            <Select
              value={llmEndpoint}
              label="LLM Endpoint"
              onChange={(e) => setLlmEndpoint(e.target.value)}
            >
              {endpoints.map(endpoint => (
                <MenuItem key={endpoint} value={endpoint}>
                  {endpoint}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        )}

        <Button
          variant="contained"
          onClick={handleSave}
          sx={{ mt: 2 }}
        >
          Save Changes
        </Button>
      </Box>
    </Box>
  );
}