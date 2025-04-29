import React from 'react';
import { Card, CardContent } from '@mui/material';
import { Box, Typography, Grid } from '@mui/material';

interface LatencyDataPoint {
  timestamp: string;
  latency: number;
}

interface ClusterSpecs {
  cpus: number;
  vendorId: string;
  modelName: string;
  cpuMhz: number;
}

interface AnalyticsDashboardProps {
  progress: number;
  tokensProcessed: number;
  latencyData: LatencyDataPoint[];
  tokenTypes: string[];
  tokenCounts: Record<string, number>;
  clusterSpecs: ClusterSpecs;
}

export function AnalyticsDashboard({ 
  progress, 
  tokensProcessed,
  latencyData,
  tokenTypes,
  tokenCounts,
  clusterSpecs
}: AnalyticsDashboardProps) {
  // Calculate average latency for display
  const avgLatency = latencyData.reduce((acc, curr) => acc + curr.latency, 0) / latencyData.length;
  
  // Format token count for readability
  const formatTokenCount = (count: number) => {
    if (count >= 1000000) {
      return `${(count / 1000000).toFixed(1)}M`;
    } else if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}K`;
    }
    return count.toString();
  };

  return (
    <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
      <Typography variant="h6" gutterBottom>Analytics</Typography>
      
      <Grid container spacing={2} sx={{ mt: 2 }}>
        <Grid item xs={3}>
          <Card sx={{ border: '1px solid #e0e0e0', boxShadow: 'none', height: '100%' }}>
            <CardContent sx={{ textAlign: 'center' }}>
              <Typography variant="body2" color="text.secondary">Progress</Typography>
              <Typography variant="h4" sx={{ mt: 1 }}>{progress}%</Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={3}>
          <Card sx={{ border: '1px solid #e0e0e0', boxShadow: 'none', height: '100%' }}>
            <CardContent sx={{ textAlign: 'center' }}>
              <Typography variant="body2" color="text.secondary">Tokens Processed</Typography>
              <Typography variant="h4" sx={{ mt: 1 }}>{formatTokenCount(tokensProcessed)}</Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={3}>
          <Card sx={{ border: '1px solid #e0e0e0', boxShadow: 'none', height: '100%' }}>
            <CardContent sx={{ textAlign: 'center' }}>
              <Typography variant="body2" color="text.secondary">Live Latency</Typography>
              <Typography variant="h4" sx={{ mt: 1 }}>{avgLatency.toFixed(3)}ms</Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={3}>
          <Card sx={{ border: '1px solid #e0e0e0', boxShadow: 'none', height: '100%' }}>
            <CardContent>
              <Typography variant="body2" color="text.secondary">Cluster Specs</Typography>
              <Box sx={{ mt: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">CPUs:</Typography>
                  <Typography variant="body2">{clusterSpecs.cpus}</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">Vendor:</Typography>
                  <Typography variant="body2">{clusterSpecs.vendorId}</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">CPU:</Typography>
                  <Typography variant="body2">{clusterSpecs.modelName}</Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
      
      <Box sx={{ mt: 3 }}>
        <Typography variant="subtitle1" gutterBottom>Identified Tokens</Typography>
        <Card sx={{ border: '1px solid #e0e0e0', boxShadow: 'none' }}>
          <CardContent>
            {tokenTypes.map(tokenType => (
              <Box key={tokenType} sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">{tokenType}</Typography>
                <Typography variant="body2">{formatTokenCount(tokenCounts[tokenType])}</Typography>
              </Box>
            ))}
          </CardContent>
        </Card>
      </Box>
    </Box>
  );
} 