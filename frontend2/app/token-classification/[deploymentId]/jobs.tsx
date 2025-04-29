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
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h6">Token Classification Jobs</Typography>
          <Link href={`/token-classification/${deploymentId}/jobs/new`} passHref>
            <Button variant="contained" color="primary">
              New Job
            </Button>
          </Link>
        </Box>

        {isLoading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
            <CircularProgress />
          </Box>
        ) : error ? (
          <Box sx={{ my: 2 }}>
            <Typography color="error">{error}</Typography>
          </Box>
        ) : (
          <TableContainer component={Paper} sx={{ boxShadow: 'none' }}>
            <Table>
              <TableHead sx={{ bgcolor: 'rgba(0, 0, 0, 0.04)' }}>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Report ID</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Submitted At</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {jobs.map((job) => (
                  <TableRow key={job.id}>
                    <TableCell>{job.name}</TableCell>
                    <TableCell>{job.reportId}</TableCell>
                    <TableCell>
                      <Chip 
                        label={job.status} 
                        color={getStatusColor(job.status) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>{formatDate(job.createdAt)}</TableCell>
                    <TableCell>
                      <Link href={`/token-classification/${deploymentId}/jobs/${job.reportId}`} passHref>
                        <Button size="small" color="primary">
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