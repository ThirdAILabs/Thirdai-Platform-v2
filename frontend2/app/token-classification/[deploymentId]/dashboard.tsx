'use client';

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

// Types for model metrics
interface ModelMetrics {
  precision: number;
  recall: number;
  f1Score: number;
  accuracy: number;
  requestsPerDay: number[];
  averageLatency: number;
}

// Mock API for fetching dashboard data
const useMonitoringAPI = () => {
  return {
    getModelMetrics: async (): Promise<ModelMetrics> => {
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 800));
      
      return {
        precision: 0.94,
        recall: 0.91,
        f1Score: 0.925,
        accuracy: 0.93,
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

export default function Dashboard() {
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
    <div>
      <Card sx={{ mb: 4 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Model Status
          </Typography>
          
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={4}>
              <Paper sx={{ p: 2, height: '100%' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="subtitle2">Status</Typography>
                  <Chip 
                    label={modelDetails.status}
                    color={modelDetails.status === 'active' ? 'success' : 'default'}
                    size="small"
                  />
                </Box>
                <Divider sx={{ my: 1 }} />
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="subtitle2">Version</Typography>
                  <Typography variant="body2">{modelDetails.version}</Typography>
                </Box>
                <Divider sx={{ my: 1 }} />
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="subtitle2">Deployed At</Typography>
                  <Typography variant="body2">{formatDate(modelDetails.deployedAt)}</Typography>
                </Box>
              </Paper>
            </Grid>
            
            <Grid item xs={12} md={8}>
              <Paper sx={{ p: 2, height: '100%' }}>
                <Typography variant="subtitle1" gutterBottom>
                  Endpoint Details
                </Typography>
                <Box sx={{ mt: 2 }}>
                  <Typography variant="subtitle2">URL</Typography>
                  <Typography variant="body2" sx={{ wordBreak: 'break-all', mt: 0.5 }}>
                    {modelDetails.endpoint}
                  </Typography>
                </Box>
                <Box sx={{ mt: 2 }}>
                  <Typography variant="subtitle2">Last Updated</Typography>
                  <Typography variant="body2" sx={{ mt: 0.5 }}>
                    {formatDate(modelDetails.lastUpdated)}
                  </Typography>
                </Box>
              </Paper>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
      
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Performance Metrics
          </Typography>
          
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} md={6}>
              <Paper sx={{ p: 2 }}>
                <Typography variant="subtitle1" gutterBottom>
                  Accuracy Metrics
                </Typography>
                
                <Box sx={{ mt: 2 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                    <Typography variant="body2">Precision</Typography>
                    <Typography variant="body2">{metrics?.precision.toFixed(3)}</Typography>
                  </Box>
                  <LinearProgress 
                    variant="determinate" 
                    value={metrics?.precision ? metrics.precision * 100 : 0} 
                    sx={{ mb: 2, height: 8, borderRadius: 1 }}
                  />
                  
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                    <Typography variant="body2">Recall</Typography>
                    <Typography variant="body2">{metrics?.recall.toFixed(3)}</Typography>
                  </Box>
                  <LinearProgress 
                    variant="determinate" 
                    value={metrics?.recall ? metrics.recall * 100 : 0} 
                    sx={{ mb: 2, height: 8, borderRadius: 1 }}
                  />
                  
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                    <Typography variant="body2">F1 Score</Typography>
                    <Typography variant="body2">{metrics?.f1Score.toFixed(3)}</Typography>
                  </Box>
                  <LinearProgress 
                    variant="determinate" 
                    value={metrics?.f1Score ? metrics.f1Score * 100 : 0} 
                    sx={{ mb: 2, height: 8, borderRadius: 1 }}
                  />
                </Box>
              </Paper>
            </Grid>
            
            <Grid item xs={12} md={6}>
              <Paper sx={{ p: 2 }}>
                <Typography variant="subtitle1" gutterBottom>
                  Performance
                </Typography>
                
                <Box sx={{ mt: 2 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                    <Typography variant="body2">Average Latency</Typography>
                    <Typography variant="body2">{metrics?.averageLatency} ms</Typography>
                  </Box>
                  <Box sx={{ mt: 3 }}>
                    <Typography variant="body2" gutterBottom>
                      Requests per day (last week)
                    </Typography>
                    <Box sx={{ display: 'flex', alignItems: 'flex-end', mt: 1, height: 60 }}>
                      {metrics?.requestsPerDay.map((requests, index) => (
                        <Box 
                          key={index}
                          sx={{
                            height: `${(requests / 200) * 100}%`,
                            width: `${100 / metrics.requestsPerDay.length}%`,
                            backgroundColor: 'primary.main',
                            mx: 0.5,
                            borderTopLeftRadius: 2,
                            borderTopRightRadius: 2,
                          }}
                        />
                      ))}
                    </Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 0.5 }}>
                      <Typography variant="caption">7 days ago</Typography>
                      <Typography variant="caption">Today</Typography>
                    </Box>
                  </Box>
                </Box>
              </Paper>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    </div>
  );
} 