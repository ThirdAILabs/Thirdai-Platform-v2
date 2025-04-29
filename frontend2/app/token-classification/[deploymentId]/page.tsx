'use client';

import { useState } from 'react';
import { useParams, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Box, Container, Typography, Tabs, Tab, Button } from '@mui/material';

// Mock data for jobs
const mockJobs = [
  {
    name: "Medical Records Review",
    reportId: "tcr_123456789",
    status: "completed", 
    submittedAt: "2024-01-15T10:30:00Z",
  },
  {
    name: "Insurance Claims Processing",
    reportId: "tcr_987654321", 
    status: "completed",
    submittedAt: "2024-01-15T11:30:00Z",
  },
  {
    name: "Customer Support Chat Logs",
    reportId: "tcr_456789123",
    status: "completed",
    submittedAt: "2024-01-15T12:30:00Z",
  },
];

export default function TokenClassificationPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const [tabValue, setTabValue] = useState(searchParams.get('tab') || 'testing');

  const handleTabChange = (_event: React.SyntheticEvent, newValue: string) => {
    setTabValue(newValue);
  };

  return (
    <Container maxWidth="lg" sx={{ px: 4 }}>
      <Box sx={{ py: 3 }}>
        <Typography variant="h5" component="h1" gutterBottom>
          Token Classification - {params.deploymentId}
        </Typography>

        <Tabs 
          value={tabValue} 
          onChange={handleTabChange}
          aria-label="token classification tabs"
          sx={{ mb: 3 }}
        >
          <Tab label="Monitoring" value="monitoring" />
          <Tab label="Testing" value="testing" />
          <Tab label="Jobs" value="jobs" />
        </Tabs>

        {tabValue === 'monitoring' && (
          <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
            <Typography variant="h6" gutterBottom>Monitoring Dashboard</Typography>
            
            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle1">Model Status</Typography>
              <Box sx={{ mt: 2, p: 2, border: '1px solid #e0e0e0', borderRadius: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Status:</Typography>
                  <Typography variant="body2" sx={{ color: 'green' }}>Active</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Deployment URL:</Typography>
                  <Typography variant="body2">https://api.thirdai.com/token-classification</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Last updated:</Typography>
                  <Typography variant="body2">April 28, 2024</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">Model version:</Typography>
                  <Typography variant="body2">v2.1.3</Typography>
                </Box>
              </Box>
            </Box>

            <Box sx={{ mt: 3 }}>
              <Typography variant="subtitle1">Performance Metrics</Typography>
              <Box sx={{ mt: 2, p: 2, border: '1px solid #e0e0e0', borderRadius: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Precision:</Typography>
                  <Typography variant="body2">0.97</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                  <Typography variant="body2">Recall:</Typography>
                  <Typography variant="body2">0.92</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2">F1 Score:</Typography>
                  <Typography variant="body2">0.94</Typography>
                </Box>
              </Box>
            </Box>
          </Box>
        )}

        {tabValue === 'testing' && (
          <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
            <Typography variant="h6" gutterBottom>Testing Interface</Typography>
            <Typography paragraph>Test token classification on your text inputs.</Typography>
            
            <Box component="form" sx={{ mt: 2 }}>
              <textarea 
                className="w-full p-2 border border-gray-300 rounded h-32"
                placeholder="Enter text to classify tokens..."
              />
              <Button 
                variant="contained" 
                color="primary" 
                sx={{ mt: 2 }}
              >
                Classify Tokens
              </Button>
            </Box>
            
            <Box sx={{ mt: 4 }}>
              <Typography variant="subtitle1" gutterBottom>Results will appear here</Typography>
              <Box sx={{ p: 3, bgcolor: '#f5f5f5', borderRadius: 1, minHeight: '100px' }}>
                <Typography variant="body2" color="text.secondary">
                  No results to display. Enter text and click "Classify Tokens" to see results.
                </Typography>
              </Box>
            </Box>
          </Box>
        )}

        {tabValue === 'jobs' && (
          <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
              <Typography variant="h6">Jobs</Typography>
              <Link 
                href={`/token-classification/${params.deploymentId}/jobs/new`}
                style={{ textDecoration: 'none' }}
              >
                <Button variant="contained" color="primary">
                  New Job
                </Button>
              </Link>
            </Box>

            <div className="overflow-hidden border border-gray-200 rounded-md">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-100">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Report ID</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Submitted At</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {mockJobs.map((job) => (
                    <tr key={job.reportId}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{job.name}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{job.reportId}</td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                          {job.status}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(job.submittedAt).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-blue-500">
                        <Link href={`/token-classification/${params.deploymentId}/jobs/${job.reportId}`}>
                          View
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Box>
        )}
      </Box>
    </Container>
  );
} 