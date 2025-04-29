'use client';

import { useState } from 'react';
import { useParams } from 'next/navigation';
import { Container, Box, Typography, Breadcrumbs, Link as MuiLink, IconButton, Stack, Button, Tab, Tabs } from '@mui/material';
import { RefreshRounded, PauseRounded, StopRounded, ArrowBack } from '@mui/icons-material';
import Link from 'next/link';

export default function JobDetail() {
  const params = useParams();
  const [lastUpdated, setLastUpdated] = useState(0);
  const [tabValue, setTabValue] = useState('analytics');

  const handleTabChange = (_event: React.SyntheticEvent, newValue: string) => {
    setTabValue(newValue);
  };

  return (
    <Container maxWidth="lg" sx={{ px: 4 }}>
      <Box sx={{ py: 3 }}>
        {/* Breadcrumbs */}
        <Stack direction="row" spacing={2} alignItems="center" mb={3}>
          <Breadcrumbs aria-label="breadcrumb">
            <MuiLink component={Link} href={`/token-classification/${params.deploymentId}/jobs`}>
              Jobs
            </MuiLink>
            <Typography color="text.primary">Customer Calls</Typography>
          </Breadcrumbs>
        </Stack>

        {/* Title and Back Button */}
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between" mb={3}>
          <Typography variant="h5">Customer Calls</Typography>
          <Button
            variant="outlined"
            startIcon={<ArrowBack />}
            component={Link}
            href={`/token-classification/${params.deploymentId}?tab=jobs`}
          >
            Back to Jobs
          </Button>
        </Stack>

        {/* Tabs, Controls and Content */}
        <Tabs 
          value={tabValue} 
          onChange={handleTabChange}
          aria-label="job detail tabs"
          sx={{ mb: 3 }}
        >
          <Tab label="Configuration" value="configuration" />
          <Tab label="Analytics" value="analytics" />
          <Tab label="Output" value="output" />
        </Tabs>

        <Stack direction="row" spacing={2} alignItems="center" justifyContent="flex-end" mb={3}>
          <Typography variant="body2" color="text.secondary">
            Last updated: {lastUpdated} seconds ago
          </Typography>
          <IconButton onClick={() => setLastUpdated(0)} size="small">
            <RefreshRounded />
          </IconButton>
          <IconButton size="small">
            <PauseRounded />
          </IconButton>
          <IconButton size="small">
            <StopRounded />
          </IconButton>
        </Stack>

        {tabValue === 'configuration' && (
          <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
            <Typography variant="h6" gutterBottom>Configuration</Typography>
            
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle1">Source</Typography>
              <Box sx={{ bgcolor: '#f5f5f5', p: 2, borderRadius: 1, mt: 1 }}>
                <Typography>s3://thirdai-dev/customer-calls/2025/</Typography>
              </Box>
            </Box>
            
            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle1">Save Location</Typography>
              <Box sx={{ bgcolor: '#f5f5f5', p: 2, borderRadius: 1, mt: 1 }}>
                <Typography>thirdai-dev/sensitive/customer-calls/2025/</Typography>
              </Box>
            </Box>
            
            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle1">Tags</Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
                {['NAME', 'VIN', 'ORG', 'ID', 'SSN', 'ADDRESS', 'EMAIL'].map(tag => (
                  <Box key={tag} sx={{ bgcolor: '#e3f2fd', color: '#1976d2', px: 2, py: 0.5, borderRadius: 1 }}>
                    {tag}
                  </Box>
                ))}
              </Box>
            </Box>
            
            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle1">Groups</Typography>
              <Box sx={{ mt: 1 }}>
                <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, mb: 2 }}>
                  <Box sx={{ p: 2, borderBottom: '1px solid #e0e0e0' }}>
                    <Typography variant="subtitle2">Reject</Typography>
                  </Box>
                  <Box sx={{ p: 2 }}>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>COUNT(tags) &gt; 5</Typography>
                  </Box>
                </Box>
                
                <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, mb: 2 }}>
                  <Box sx={{ p: 2, borderBottom: '1px solid #e0e0e0' }}>
                    <Typography variant="subtitle2">Sensitive</Typography>
                  </Box>
                  <Box sx={{ p: 2 }}>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>COUNT(tags) &gt; 0</Typography>
                  </Box>
                </Box>
                
                <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1 }}>
                  <Box sx={{ p: 2, borderBottom: '1px solid #e0e0e0' }}>
                    <Typography variant="subtitle2">Safe</Typography>
                  </Box>
                  <Box sx={{ p: 2 }}>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>COUNT(tags) = 0</Typography>
                  </Box>
                </Box>
              </Box>
            </Box>
          </Box>
        )}

        {tabValue === 'analytics' && (
          <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
            <Typography variant="h6" gutterBottom>Analytics</Typography>
            
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 2, mt: 2 }}>
              <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, p: 2, textAlign: 'center' }}>
                <Typography variant="body2" color="text.secondary">Progress</Typography>
                <Typography variant="h4" sx={{ mt: 1 }}>40%</Typography>
              </Box>
              
              <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, p: 2, textAlign: 'center' }}>
                <Typography variant="body2" color="text.secondary">Tokens Processed</Typography>
                <Typography variant="h4" sx={{ mt: 1 }}>1.2M</Typography>
              </Box>
              
              <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, p: 2, textAlign: 'center' }}>
                <Typography variant="body2" color="text.secondary">Live Latency</Typography>
                <Typography variant="h4" sx={{ mt: 1 }}>0.093ms</Typography>
              </Box>
              
              <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, p: 2 }}>
                <Typography variant="body2" color="text.secondary">Cluster Specs</Typography>
                <Box sx={{ mt: 1 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="body2">CPUs:</Typography>
                    <Typography variant="body2">48</Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="body2">Vendor:</Typography>
                    <Typography variant="body2">Intel</Typography>
                  </Box>
                </Box>
              </Box>
            </Box>
            
            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle1" gutterBottom>Identified Tokens</Typography>
              <Box sx={{ border: '1px solid #e0e0e0', borderRadius: 1, p: 2, mt: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">NAME</Typography>
                  <Typography variant="body2">21.2M</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">VIN</Typography>
                  <Typography variant="body2">19.8M</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">ORG</Typography>
                  <Typography variant="body2">13.3M</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">SSN</Typography>
                  <Typography variant="body2">13.3M</Typography>
                </Box>
              </Box>
            </Box>
          </Box>
        )}

        {tabValue === 'output' && (
          <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
            <Typography variant="h6" gutterBottom>Output</Typography>
            
            <Box sx={{ mt: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                <Box>
                  <Typography variant="body2" color="text.secondary">Showing results for:</Typography>
                  <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                    <Button variant="outlined" size="small">All Groups</Button>
                    <Button variant="outlined" size="small">All Tags</Button>
                  </Box>
                </Box>
                <Box>
                  <Button variant="contained" color="primary" size="small">
                    Download Results
                  </Button>
                </Box>
              </Box>
              
              <div className="overflow-hidden border border-gray-200 rounded-md">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-100">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Document</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Group</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tags</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Preview</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    <tr>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">call_transcript_1.txt</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">Sensitive</td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex flex-wrap gap-1">
                          <span className="px-2 py-1 text-xs rounded bg-blue-100 text-blue-800">NAME</span>
                          <span className="px-2 py-1 text-xs rounded bg-green-100 text-green-800">SSN</span>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        My name is <span className="bg-blue-100">John Smith</span> and my social is <span className="bg-green-100">123-45-6789</span>...
                      </td>
                    </tr>
                    <tr>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">call_transcript_2.txt</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">Reject</td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex flex-wrap gap-1">
                          <span className="px-2 py-1 text-xs rounded bg-blue-100 text-blue-800">NAME</span>
                          <span className="px-2 py-1 text-xs rounded bg-purple-100 text-purple-800">ADDRESS</span>
                          <span className="px-2 py-1 text-xs rounded bg-red-100 text-red-800">VIN</span>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        <span className="bg-blue-100">Jane Doe</span> at <span className="bg-purple-100">123 Main St</span> with vehicle <span className="bg-red-100">1HGCM82633A004352</span>...
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </Box>
          </Box>
        )}
      </Box>
    </Container>
  );
} 