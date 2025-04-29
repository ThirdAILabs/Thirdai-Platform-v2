'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  Box,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  CircularProgress,
} from '@mui/material';
import { Card, CardContent } from '@mui/material';

// Mock data and API until we implement the real one
interface Job {
  id: string;
  name: string;
  status: 'completed' | 'running' | 'failed' | 'pending';
  createdAt: string;
  reportId: string;
}

// Mock backend service
const useTokenClassificationJobsAPI = () => {
  return {
    getJobs: async (): Promise<Job[]> => {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500));
      
      return [
        {
          id: 'job1',
          name: 'Medical Records Review',
          status: 'completed',
          createdAt: '2024-04-15T10:30:00Z',
          reportId: 'tcr_123456789',
        },
        {
          id: 'job2',
          name: 'Insurance Claims Processing',
          status: 'completed',
          createdAt: '2024-04-15T11:30:00Z',
          reportId: 'tcr_987654321',
        },
        {
          id: 'job3',
          name: 'Customer Support Chat Logs',
          status: 'running',
          createdAt: '2024-04-15T12:30:00Z',
          reportId: 'tcr_456789123',
        },
      ];
    }
  };
};

export default function Jobs() {
  const { deploymentId } = useParams();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const jobsAPI = useTokenClassificationJobsAPI();

  useEffect(() => {
    const fetchJobs = async () => {
      try {
        setIsLoading(true);
        setError(null);
        const fetchedJobs = await jobsAPI.getJobs();
        setJobs(fetchedJobs);
      } catch (err) {
        setError('Failed to fetch jobs');
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchJobs();
  }, []);

  // Format date as a readable string
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  // Get status chip color based on job status
  const getStatusColor = (status: Job['status']) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'running':
        return 'info';
      case 'failed':
        return 'error';
      case 'pending':
        return 'warning';
      default:
        return 'default';
    }
  };

  return (
    <Card sx={{ backgroundColor: '#ffffff', boxShadow: '0 1px 3px rgba(0,0,0,0.1)' }}>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h6" sx={{ fontWeight: 500, fontSize: '1.1rem' }}>
            Jobs
          </Typography>
          <Link href={`/token-classification/${deploymentId}/jobs/new`} passHref>
            <Button 
              variant="contained" 
              color="primary"
              sx={{
                backgroundColor: '#1a73e8',
                textTransform: 'none',
                boxShadow: 'none',
                '&:hover': {
                  backgroundColor: '#1765cc',
                  boxShadow: 'none'
                }
              }}
            >
              New Job
            </Button>
          </Link>
        </Box>

        {isLoading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
            <CircularProgress size={24} sx={{ color: '#1a73e8' }} />
          </Box>
        ) : error ? (
          <Box sx={{ my: 2 }}>
            <Typography color="error">{error}</Typography>
          </Box>
        ) : (
          <TableContainer 
            component={Paper} 
            sx={{ 
              boxShadow: 'none',
              border: '1px solid #eee',
              borderRadius: '4px',
              overflow: 'hidden'
            }}
          >
            <Table>
              <TableHead sx={{ bgcolor: '#f5f5f5' }}>
                <TableRow>
                  <TableCell sx={{ color: '#5f6368', fontWeight: 500, fontSize: '0.8rem' }}>Name</TableCell>
                  <TableCell sx={{ color: '#5f6368', fontWeight: 500, fontSize: '0.8rem' }}>Report ID</TableCell>
                  <TableCell sx={{ color: '#5f6368', fontWeight: 500, fontSize: '0.8rem' }}>Status</TableCell>
                  <TableCell sx={{ color: '#5f6368', fontWeight: 500, fontSize: '0.8rem' }}>Submitted At</TableCell>
                  <TableCell sx={{ color: '#5f6368', fontWeight: 500, fontSize: '0.8rem' }}>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {jobs.map((job) => (
                  <TableRow key={job.id} sx={{ '&:hover': { backgroundColor: '#f8f9fa' } }}>
                    <TableCell sx={{ fontSize: '0.9rem' }}>{job.name}</TableCell>
                    <TableCell sx={{ fontSize: '0.9rem', color: '#5f6368' }}>{job.reportId}</TableCell>
                    <TableCell>
                      <Chip 
                        label={job.status} 
                        color={getStatusColor(job.status) as any}
                        size="small"
                        sx={{ 
                          height: '22px', 
                          fontSize: '0.75rem',
                          textTransform: 'capitalize'
                        }}
                      />
                    </TableCell>
                    <TableCell sx={{ fontSize: '0.9rem', color: '#5f6368' }}>{formatDate(job.createdAt)}</TableCell>
                    <TableCell>
                      <Link href={`/token-classification/${deploymentId}/jobs/${job.reportId}`} passHref>
                        <Button 
                          size="small" 
                          sx={{ 
                            color: '#1a73e8', 
                            textTransform: 'none',
                            '&:hover': {
                              backgroundColor: 'transparent',
                              textDecoration: 'underline'
                            }
                          }}
                        >
                          View
                        </Button>
                      </Link>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </CardContent>
    </Card>
  );
} 