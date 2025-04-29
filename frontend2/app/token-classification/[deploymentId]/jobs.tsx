'use client';

import { useEffect, useState } from 'react';
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
  CircularProgress,
} from '@mui/material';
import { Card, CardContent } from '@mui/material';
import Link from 'next/link';
import { useParams } from 'next/navigation';

// Report interface to match the original
interface Report {
  name: string;
  report_id: string;
  status: string;
  submitted_at: string;
  updated_at: string;
  documents: Array<{
    path: string;
    location: string;
    source_id: string;
    options: Record<string, any>;
    metadata: Record<string, any>;
  }>;
  msg: string | null;
  content: {
    report_id: string;
    results: Array<Record<string, Array<{ text: string; tag: string }>>>;
  };
}

// Mock data matching the original
const mockReports: Report[] = [
  {
    name: "Medical Records Review",
    report_id: "tcr_123456789",
    status: "completed", 
    submitted_at: "2024-01-15T10:30:00Z",
    updated_at: "2024-01-15T10:35:00Z",
    documents: [
      {
        path: "medical_records.txt",
        location: "s3://bucket/medical_records.txt",
        source_id: "doc_123",
        options: {},
        metadata: {}
      }
    ],
    msg: null,
    content: {
      report_id: "tcr_123456789",
      results: [
        {
          "medical_records.txt": [
            { text: "123-45-6789", tag: "SSN" },
            { text: "01/15/1980", tag: "DOB" },
            { text: "555-123-4567", tag: "PHONE" },
            { text: "123 Main St, Apt 4B", tag: "ADDRESS" },
            { text: "robert.chen1982@gmail.com", tag: "EMAIL" }
          ]
        }
      ]
    }
  },
  {
    name: "Insurance Claims Processing",
    report_id: "tcr_987654321", 
    status: "completed",
    submitted_at: "2024-01-15T11:30:00Z",
    updated_at: "2024-01-15T11:35:00Z",
    documents: [
      {
        path: "insurance_claims.txt",
        location: "s3://bucket/insurance_claims.txt",
        source_id: "doc_456",
        options: {},
        metadata: {}
      }
    ],
    msg: null,
    content: {
      report_id: "tcr_987654321",
      results: [
        {
          "insurance_claims.txt": [
            { text: "CLM-123456", tag: "CLAIM_ID" },
            { text: "POL-789012", tag: "POLICY_NUMBER" },
            { text: "John Smith", tag: "NAME" }
          ]
        }
      ]
    }
  },
  {
    name: "Customer Support Chat Logs",
    report_id: "tcr_456789123",
    status: "completed",
    submitted_at: "2024-01-15T12:30:00Z",
    updated_at: "2024-01-15T12:35:00Z",
    documents: [
      {
        path: "chat_logs.txt",
        location: "s3://bucket/chat_logs.txt",
        source_id: "doc_789",
        options: {},
        metadata: {}
      }
    ],
    msg: null,
    content: {
      report_id: "tcr_456789123",
      results: [
        {
          "chat_logs.txt": [
            { text: "4832-5691-2748-1035", tag: "CREDIT_CARD" },
            { text: "09/27", tag: "EXPIRATION_DATE" },
            { text: "382", tag: "CVV" }
          ]
        }
      ]
    }
  },
];

export default function Jobs() {
  const params = useParams();
  const [reports, setReports] = useState<Report[]>(mockReports);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Mock API implementation would go here
  // useEffect(() => {
  //   const fetchReports = async () => {
  //     try {
  //       const deploymentId = params.deploymentId as string;
  //       // const reportsData = await listReports(deploymentId);
  //       // setReports(reportsData);
  //     } catch (err) {
  //       setError('Failed to fetch reports');
  //       console.error('Error fetching reports:', err);
  //     } finally {
  //       setLoading(false);
  //     }
  //   };
  //
  //   fetchReports();
  // }, [params.deploymentId]);

  if (loading) {
    return (
      <Card sx={{ boxShadow: '0 1px 3px rgba(0,0,0,0.1)', bgcolor: 'white' }}>
        <CardContent sx={{ p: 3 }}>
          <Box sx={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center', 
            mb: 3 
          }}>
            <Typography variant="h6" sx={{ fontWeight: 600, fontSize: '1.125rem' }}>
              Jobs
            </Typography>
            <Link href={`/token-classification/${params.deploymentId}/jobs/new`} passHref>
              <Button 
                variant="contained" 
                color="primary"
                sx={{
                  bgcolor: '#1976d2',
                  '&:hover': {
                    bgcolor: '#1565c0',
                  },
                  textTransform: 'none',
                  fontWeight: 500
                }}
              >
                New
              </Button>
            </Link>
          </Box>
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
            <CircularProgress size={24} />
          </Box>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card sx={{ boxShadow: '0 1px 3px rgba(0,0,0,0.1)', bgcolor: 'white' }}>
        <CardContent sx={{ p: 3 }}>
          <Box sx={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center', 
            mb: 3 
          }}>
            <Typography variant="h6" sx={{ fontWeight: 600, fontSize: '1.125rem' }}>
              Jobs
            </Typography>
            <Link href={`/token-classification/${params.deploymentId}/jobs/new`} passHref>
              <Button 
                variant="contained" 
                color="primary"
                sx={{
                  bgcolor: '#1976d2',
                  '&:hover': {
                    bgcolor: '#1565c0',
                  },
                  textTransform: 'none',
                  fontWeight: 500
                }}
              >
                New
              </Button>
            </Link>
          </Box>
          <Typography sx={{ textAlign: 'center', py: 2, color: 'error.main' }}>
            {error}
          </Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card sx={{ boxShadow: '0 1px 3px rgba(0,0,0,0.1)', bgcolor: 'white' }}>
      <CardContent sx={{ p: 3 }}>
        <Box sx={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          mb: 3 
        }}>
          <Typography variant="h6" sx={{ fontWeight: 600, fontSize: '1.125rem' }}>
            Jobs
          </Typography>
          <Link href={`/token-classification/${params.deploymentId}/jobs/new`} passHref>
            <Button 
              variant="contained" 
              color="primary"
              sx={{
                bgcolor: '#1976d2',
                '&:hover': {
                  bgcolor: '#1565c0',
                },
                textTransform: 'none',
                fontWeight: 500
              }}
            >
              New
            </Button>
          </Link>
        </Box>

        <TableContainer 
          component={Paper} 
          sx={{ 
            boxShadow: 'none', 
            border: '1px solid rgba(0, 0, 0, 0.12)', 
            borderRadius: '0.375rem',
            overflow: 'hidden'
          }}
        >
          <Table>
            <TableHead>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Report ID</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Status</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Submitted At</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Updated At</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Action</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {reports && reports.length > 0 ? (
                reports.map((report, index) => (
                  <TableRow
                    key={report.report_id}
                    sx={{
                      bgcolor: index % 2 === 0 ? 'white' : '#f9fafb'
                    }}
                  >
                    <TableCell>{report.name}</TableCell>
                    <TableCell>{report.report_id}</TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                        <Box 
                          sx={{ 
                            flex: 1, 
                            height: '8px', 
                            bgcolor: '#f1f5f9', 
                            borderRadius: '9999px',
                            overflow: 'hidden'
                          }}
                        >
                          <Box
                            sx={{
                              height: '100%',
                              width: report.status === 'completed' ? '100%' : '50%',
                              bgcolor: report.status === 'completed' ? '#4caf50' : '#1976d2',
                              borderRadius: '9999px',
                              transition: 'all 0.2s'
                            }}
                          />
                        </Box>
                        <Typography 
                          variant="body2" 
                          sx={{ 
                            color: 'text.secondary', 
                            whiteSpace: 'nowrap',
                            fontSize: '0.875rem'
                          }}
                        >
                          {report.status}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>{new Date(report.submitted_at).toLocaleString()}</TableCell>
                    <TableCell>{new Date(report.updated_at).toLocaleString()}</TableCell>
                    <TableCell>
                      <Link
                        href={`/token-classification/${params.deploymentId}/jobs/${report.report_id}`}
                        style={{
                          color: '#1976d2',
                          textDecoration: 'none'
                        }}
                      >
                        <Typography 
                          sx={{ 
                            '&:hover': { 
                              textDecoration: 'underline' 
                            } 
                          }}
                        >
                          View
                        </Typography>
                      </Link>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell 
                    colSpan={6} 
                    sx={{ 
                      textAlign: 'center', 
                      py: 2,
                      color: 'text.secondary'
                    }}
                  >
                    No reports found. Create a new report to get started.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </CardContent>
    </Card>
  );
} 