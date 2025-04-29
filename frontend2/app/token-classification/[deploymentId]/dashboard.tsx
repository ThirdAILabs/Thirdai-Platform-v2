'use client';

import React from 'react';
import { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Grid,
  Paper,
  Divider,
  LinearProgress,
  CircularProgress,
  Chip,
} from '@mui/material';
import { Card, CardContent } from '@mui/material';
import { useParams } from 'next/navigation';
import TrainingResults from './metrics/TrainingResults';

// Types for model metrics
interface ModelMetrics {
  precision?: number;
  recall?: number;
  f1Score?: number;
  accuracy?: number;
  requestsPerDay: number[];
  averageLatency: number;
}

// Mock API for fetching dashboard data
const useMonitoringAPI = () => {
  return {
    getModelMetrics: async (): Promise<ModelMetrics> => {
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 800));
      
      // Simulate a scenario where some metrics might be undefined
      const simulateDataAvailability = Math.random() > 0.3;
      
      return {
        precision: simulateDataAvailability ? 0.94 : undefined,
        recall: simulateDataAvailability ? 0.91 : undefined,
        f1Score: simulateDataAvailability ? 0.925 : undefined,
        accuracy: simulateDataAvailability ? 0.93 : undefined,
        requestsPerDay: [145, 156, 162, 170, 182, 167, 190],
        averageLatency: 120, // ms
      };
    },
    
    getModelDetails: async () => {
      await new Promise(resolve => setTimeout(resolve, 500));
      
      return {
        version: 'v2.1.3',
        deployedAt: '2024-04-15T10:30:00Z',
        status: 'active',
        endpoint: 'https://api.thirdai.com/token-classification',
        lastUpdated: '2024-04-28T14:25:00Z',
      };
    }
  };
};

const Dashboard = () => {
  const { deploymentId } = useParams();
  const [metrics, setMetrics] = useState<ModelMetrics | null>(null);
  const [modelDetails, setModelDetails] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const monitoringAPI = useMonitoringAPI();
  
  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setIsLoading(true);
        setError(null);
        
        const [metricsData, detailsData] = await Promise.all([
          monitoringAPI.getModelMetrics(),
          monitoringAPI.getModelDetails()
        ]);
        
        setMetrics(metricsData);
        setModelDetails(detailsData);
      } catch (err) {
        setError('Failed to fetch dashboard data');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchDashboardData();
  }, []);
  
  // Format date to readable string
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };
  
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
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card sx={{ backgroundColor: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.12)', height: '100%' }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 500, mb: 2 }}>
                Model Status
              </Typography>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
                <Chip 
                  label={modelDetails.status} 
                  color={modelDetails.status === 'active' ? 'success' : 'default'} 
                  size="small" 
                  sx={{ fontWeight: 500, borderRadius: '4px' }} 
                />
                <Typography variant="body2" color="text.secondary">
                  Last updated: {formatDate(modelDetails.lastUpdated)}
                </Typography>
              </Box>
              
              <Paper elevation={0} sx={{ p: 2, border: '1px solid #e0e0e0', borderRadius: '4px', mb: 2 }}>
                <Typography variant="subtitle2" sx={{ mb: 1 }}>
                  Model Information
                </Typography>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="body2" color="text.secondary">Base Model</Typography>
                    <Typography variant="body2" fontWeight={500}>BERT-large</Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="body2" color="text.secondary">Fine-tuned</Typography>
                    <Typography variant="body2" fontWeight={500}>Yes</Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="body2" color="text.secondary">Deployment Date</Typography>
                    <Typography variant="body2" fontWeight={500}>{formatDate(modelDetails.deployedAt)}</Typography>
                  </Box>
                </Box>
              </Paper>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <Card sx={{ backgroundColor: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.12)', height: '100%' }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 500, mb: 2 }}>
                Performance Metrics
              </Typography>
              
              {(!metrics?.precision && !metrics?.recall && !metrics?.f1Score) ? (
                <Box sx={{ p: 2, textAlign: 'center', border: '1px dashed #ccc', borderRadius: '4px', mb: 3 }}>
                  <Typography variant="body2" color="text.secondary">
                    Metrics data is currently unavailable. Please check back later.
                  </Typography>
                </Box>
              ) : (
                <>
                  <Box sx={{ mb: 3 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2">Precision</Typography>
                      <Typography variant="body2" fontWeight={500}>{metrics?.precision !== undefined ? (metrics.precision * 100).toFixed(1) : '0.0'}%</Typography>
                    </Box>
                    <LinearProgress 
                      variant="determinate" 
                      value={metrics?.precision !== undefined ? metrics.precision * 100 : 0} 
                      sx={{ 
                        height: 8, 
                        borderRadius: 4,
                        backgroundColor: '#e1f5fe',
                        '& .MuiLinearProgress-bar': {
                          backgroundColor: '#4A90E2',
                        }
                      }} 
                    />
                  </Box>
                  
                  <Box sx={{ mb: 3 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2">Recall</Typography>
                      <Typography variant="body2" fontWeight={500}>{metrics?.recall !== undefined ? (metrics.recall * 100).toFixed(1) : '0.0'}%</Typography>
                    </Box>
                    <LinearProgress 
                      variant="determinate" 
                      value={metrics?.recall !== undefined ? metrics.recall * 100 : 0}
                      sx={{ 
                        height: 8, 
                        borderRadius: 4,
                        backgroundColor: '#e8f5e9',
                        '& .MuiLinearProgress-bar': {
                          backgroundColor: '#50C878',
                        }
                      }} 
                    />
                  </Box>
                  
                  <Box sx={{ mb: 3 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2">F1 Score</Typography>
                      <Typography variant="body2" fontWeight={500}>{metrics?.f1Score !== undefined ? (metrics.f1Score * 100).toFixed(1) : '0.0'}%</Typography>
                    </Box>
                    <LinearProgress 
                      variant="determinate" 
                      value={metrics?.f1Score !== undefined ? metrics.f1Score * 100 : 0}
                      sx={{ 
                        height: 8, 
                        borderRadius: 4,
                        backgroundColor: '#fff3e0',
                        '& .MuiLinearProgress-bar': {
                          backgroundColor: '#FF8C00',
                        }
                      }} 
                    />
                  </Box>
                </>
              )}
              
              <Paper elevation={0} sx={{ p: 2, border: '1px solid #e0e0e0', borderRadius: '4px' }}>
                <Typography variant="subtitle2" sx={{ mb: 1 }}>
                  Requests per day
                </Typography>
                <Typography variant="h4" sx={{ fontWeight: 500, color: '#333' }}>
                  2,547
                </Typography>
                <Typography variant="body2" color="success.main" sx={{ display: 'flex', alignItems: 'center' }}>
                  ↑ 12% from last week
                </Typography>
              </Paper>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12}>
          <TrainingResults />
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard; 