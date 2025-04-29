'use client';

import React from 'react';
import { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  CircularProgress,
} from '@mui/material';
import { useParams } from 'next/navigation';
import TrainingResults from './metrics/TrainingResults';
import ExamplesVisualizer from './metrics/ExamplesVisualizer';

const Dashboard = () => {
  const { deploymentId } = useParams();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    // Simulating loading for a short time
    const timer = setTimeout(() => {
      setIsLoading(false);
    }, 800);
    
    return () => clearTimeout(timer);
  }, []);
  
  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }
  
  if (error) {
    return (
      <Box sx={{ p: 2 }}>
        <Typography color="error">{error}</Typography>
      </Box>
    );
  }
  
  return (
    <Box sx={{ padding: '24px', backgroundColor: '#F5F7FA', minHeight: 'calc(100vh - 112px)' }}>
      <TrainingResults />
      <ExamplesVisualizer />
    </Box>
  );
};

export default Dashboard; 