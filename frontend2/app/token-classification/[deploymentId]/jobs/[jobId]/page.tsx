'use client';

import { useState } from 'react';
import React from 'react';
import { useParams } from 'next/navigation';
import { 
  Container, Box, Typography, Breadcrumbs, Link as MuiLink, IconButton, 
  Stack, Button, Paper, Grid, Chip, Card, CardContent, CardHeader
} from '@mui/material';
import { RefreshRounded, PauseRounded, StopRounded, ArrowBack, Edit } from '@mui/icons-material';
import Link from 'next/link';
import { AnalyticsDashboard } from '@/components/AnalyticsDashboard';
import { DatabaseTable } from './(database-table)/DatabaseTable';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

// Mock data for database table
const mockGroups = ['Reject', 'Sensitive', 'Safe'];
const mockTags = ['NAME', 'VIN', 'ORG', 'ID', 'SSN', 'ADDRESS', 'EMAIL', 'PHONE', 'POLICY_ID', 'MED_REC_NO', 'LICENSE', 'EMPLOYER', 'USERNAME', 'URL', 'IP_ADDR', 'ZIP_CODE', 'ACCOUNT', 'INS_PROV', 'PROCEDURE', 'DATE', 'NATIONALITY', 'SERIAL_NO', 'CRED_CARD_NUM', 'CVV'];

const loadMoreMockObjectRecords = () => {
  return Promise.resolve([
    {
      taggedTokens: [
        ['My', 'O'] as [string, string],
        ['name', 'O'] as [string, string],
        ['is', 'O'] as [string, string],
        ['John', 'NAME'] as [string, string],
        ['Smith', 'NAME'] as [string, string],
        ['and', 'O'] as [string, string],
        ['my', 'O'] as [string, string],
        ['social', 'O'] as [string, string],
        ['is', 'O'] as [string, string],
        ['123-45-6789', 'SSN'] as [string, string],
      ],
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      taggedTokens: [
        ['Jane', 'NAME'] as [string, string],
        ['Doe', 'NAME'] as [string, string],
        ['at', 'O'] as [string, string],
        ['123', 'ADDRESS'] as [string, string],
        ['Main', 'ADDRESS'] as [string, string],
        ['St', 'ADDRESS'] as [string, string],
        ['with', 'O'] as [string, string],
        ['vehicle', 'O'] as [string, string],
        ['1HGCM82633A004352', 'VIN'] as [string, string],
      ],
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
  ]);
};

const loadMoreMockClassifiedTokenRecords = () => {
  return Promise.resolve([
    {
      token: 'John',
      tag: 'NAME',
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      token: 'Smith',
      tag: 'NAME',
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      token: '123-45-6789',
      tag: 'SSN',
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      token: 'Jane',
      tag: 'NAME',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
    {
      token: 'Doe',
      tag: 'NAME',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
    {
      token: '123 Main St',
      tag: 'ADDRESS',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
    {
      token: '1HGCM82633A004352',
      tag: 'VIN',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
  ]);
};

// Source card component
interface SourceCardProps {
  title: string;
  description: string;
  isSelected?: boolean;
  onClick: () => void;
}

const SourceCard: React.FC<SourceCardProps> = ({ title, description, isSelected = false, onClick }) => (
  <Card 
    sx={{ 
      width: '100%', 
      cursor: 'pointer',
      border: isSelected ? '2px solid #1976d2' : '1px solid #e0e0e0',
      position: 'relative'
    }}
    onClick={onClick}
  >
    <CardContent>
      <Typography variant="h6" component="div">{title}</Typography>
      <Typography variant="body2" color="text.secondary">{description}</Typography>
      {isSelected && (
        <Box sx={{ position: 'absolute', top: 10, right: 10 }}>
          <IconButton size="small">
            <Edit fontSize="small" />
          </IconButton>
        </Box>
      )}
    </CardContent>
  </Card>
);

export default function JobDetail() {
  const params = useParams();
  const [lastUpdated, setLastUpdated] = useState(0);
  const [tabValue, setTabValue] = useState('configuration');
  const [selectedSource, setSelectedSource] = useState('s3');
  const [selectedSaveLocation, setSelectedSaveLocation] = useState('s3');

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
        <Tabs value={tabValue} onValueChange={setTabValue}>
          <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between" mb={3}>
            <TabsList>
              <TabsTrigger value="configuration">Configuration</TabsTrigger>
              <TabsTrigger value="analytics">Analytics</TabsTrigger>
              <TabsTrigger value="output">Output</TabsTrigger>
            </TabsList>

            <Stack direction="row" spacing={2} alignItems="center">
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
          </Stack>

          <TabsContent value="configuration">
            <Box>
              {/* Source section */}
              <Typography variant="h6" sx={{ mb: 2 }}>Source</Typography>
              <Grid container spacing={2} sx={{ mb: 4 }}>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="S3 Bucket" 
                    description="s3://thirdai-dev/customer-calls/2025/"
                    isSelected={selectedSource === 's3'}
                    onClick={() => setSelectedSource('s3')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="Local Storage" 
                    description="Configure now"
                    isSelected={selectedSource === 'local'}
                    onClick={() => setSelectedSource('local')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="More options" 
                    description="coming soon"
                    isSelected={selectedSource === 'more'}
                    onClick={() => setSelectedSource('more')}
                  />
                </Grid>
              </Grid>
              
              {/* Tags section */}
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
                <Typography variant="h6">Tags</Typography>
                <Chip label="Select All" variant="outlined" />
              </Stack>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 4 }}>
                {mockTags.map(tag => (
                  <Chip 
                    key={tag} 
                    label={tag} 
                    sx={{ 
                      bgcolor: '#4285f4',
                      color: 'white',
                      borderRadius: '4px',
                      fontWeight: 500,
                      mb: 1
                    }} 
                  />
                ))}
              </Box>
              
              {/* Groups section */}
              <Typography variant="h6" sx={{ mb: 2 }}>Groups</Typography>
              <Grid container spacing={2} sx={{ mb: 4 }}>
                <Grid item xs={12} sm={4}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">Reject</Typography>
                      <Typography variant="body2" fontFamily="monospace" sx={{ mt: 1 }}>
                        COUNT(tags) &gt; 5
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">Sensitive</Typography>
                      <Typography variant="body2" fontFamily="monospace" sx={{ mt: 1 }}>
                        COUNT(tags) &gt; 0
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6">Safe</Typography>
                      <Typography variant="body2" fontFamily="monospace" sx={{ mt: 1 }}>
                        COUNT(tags) = 0
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Card sx={{ 
                    height: '100%', 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    border: '1px dashed #ccc'
                  }}>
                    <CardContent>
                      <Typography variant="body1" textAlign="center">
                        Define new group
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
              
              {/* Save Groups To section */}
              <Typography variant="h6" sx={{ mb: 2 }}>Save Groups To</Typography>
              <Grid container spacing={2}>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="S3 Bucket" 
                    description="thirdai-dev/sensitive/customer-calls/2025/"
                    isSelected={selectedSaveLocation === 's3'}
                    onClick={() => setSelectedSaveLocation('s3')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="Local Storage" 
                    description="local"
                    isSelected={selectedSaveLocation === 'local'}
                    onClick={() => setSelectedSaveLocation('local')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="No storage location" 
                    description="You can still save groups"
                    isSelected={selectedSaveLocation === 'none'}
                    onClick={() => setSelectedSaveLocation('none')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <SourceCard 
                    title="More options" 
                    description="coming soon"
                    isSelected={selectedSaveLocation === 'more'}
                    onClick={() => setSelectedSaveLocation('more')}
                  />
                </Grid>
              </Grid>
            </Box>
          </TabsContent>

          <TabsContent value="analytics">
            <AnalyticsDashboard
              progress={40}
              tokensProcessed={1229000}
              latencyData={[
                { timestamp: '2024-03-10T12:00:00', latency: 0.096 },
                { timestamp: '2024-03-10T12:00:01', latency: 0.09 },
                { timestamp: '2024-03-10T12:00:02', latency: 0.082 },
                { timestamp: '2024-03-10T12:00:03', latency: 0.101 },
                { timestamp: '2024-03-10T12:00:04', latency: 0.098 },
                { timestamp: '2024-03-10T12:00:05', latency: 0.095 },
                { timestamp: '2024-03-10T12:00:06', latency: 0.097 },
                { timestamp: '2024-03-10T12:00:07', latency: 0.099 },
                { timestamp: '2024-03-10T12:00:08', latency: 0.094 },
                { timestamp: '2024-03-10T12:00:09', latency: 0.093 },
                { timestamp: '2024-03-10T12:00:10', latency: 0.088 }, 
                { timestamp: '2024-03-10T12:00:11', latency: 0.082 },
                { timestamp: '2024-03-10T12:00:12', latency: 0.079 },
                { timestamp: '2024-03-10T12:00:13', latency: 0.087 },
                { timestamp: '2024-03-10T12:00:14', latency: 0.083 },
                { timestamp: '2024-03-10T12:00:15', latency: 0.084 },
                { timestamp: '2024-03-10T12:00:16', latency: 0.086 },
                { timestamp: '2024-03-10T12:00:17', latency: 0.083 },
                { timestamp: '2024-03-10T12:00:18', latency: 0.089 },
                { timestamp: '2024-03-10T12:00:19', latency: 0.091 },
                { timestamp: '2024-03-10T12:00:20', latency: 0.083 },
                { timestamp: '2024-03-10T12:00:21', latency: 0.092 },
                { timestamp: '2024-03-10T12:00:22', latency: 0.094 },
              ]}
              tokenTypes={['NAME', 'VIN', 'ORG', 'ID', 'SSN', 'ADDRESS', 'EMAIL']}
              tokenCounts={{
                'NAME': 21200000,
                'VIN': 19800000,
                'ORG': 13300000,
                'ID': 13300000,
                'SSN': 13300000,
                'ADDRESS': 5600000,
                'EMAIL': 3800000
              }}
              clusterSpecs={{
                cpus: 48,
                vendorId: 'GenuineIntel',
                modelName: 'Intel Xeon E5-2680',
                cpuMhz: 1197.408
              }}
            />
          </TabsContent>

          <TabsContent value="output">
            <DatabaseTable 
              loadMoreObjectRecords={loadMoreMockObjectRecords}
              loadMoreClassifiedTokenRecords={loadMoreMockClassifiedTokenRecords}
              groups={mockGroups}
              tags={mockTags}
            />
          </TabsContent>
        </Tabs>
      </Box>
    </Container>
  );
} 