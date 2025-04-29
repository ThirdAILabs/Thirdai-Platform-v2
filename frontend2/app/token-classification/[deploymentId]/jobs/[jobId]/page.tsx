'use client';

import { useState } from 'react';
import { useParams } from 'next/navigation';
import { Container, Box, Typography, Breadcrumbs, Link as MuiLink, IconButton, Stack, Button, Tab, Tabs } from '@mui/material';
import { RefreshRounded, PauseRounded, StopRounded, ArrowBack } from '@mui/icons-material';
import Link from 'next/link';
import { AnalyticsDashboard } from '@/components/AnalyticsDashboard';
import ConfigurationCard from '@/components/ConfigurationCard';
import { DatabaseTable } from './(database-table)/DatabaseTable';

// Mock data for database table
const mockGroups = ['Reject', 'Sensitive', 'Safe'];
const mockTags = ['NAME', 'VIN', 'ORG', 'ID', 'SSN', 'ADDRESS', 'EMAIL'];

const loadMoreMockObjectRecords = () => {
  return Promise.resolve([
    {
      taggedTokens: [
        ['My', 'O'],
        ['name', 'O'],
        ['is', 'O'],
        ['John', 'NAME'],
        ['Smith', 'NAME'],
        ['and', 'O'],
        ['my', 'O'],
        ['social', 'O'],
        ['is', 'O'],
        ['123-45-6789', 'SSN'],
      ],
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      taggedTokens: [
        ['Jane', 'NAME'],
        ['Doe', 'NAME'],
        ['at', 'O'],
        ['123', 'ADDRESS'],
        ['Main', 'ADDRESS'],
        ['St', 'ADDRESS'],
        ['with', 'O'],
        ['vehicle', 'O'],
        ['1HGCM82633A004352', 'VIN'],
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
          <ConfigurationCard 
            sourceS3Config={{ name: 's3://thirdai-dev/customer-calls/2025/' }}
            sourceLocalConfig={{ name: '' }}
            saveS3Config={{ name: 'thirdai-dev/sensitive/customer-calls/2025/' }}
            saveLocalConfig={{ name: 'local' }}
            selectedSource={null}
            selectedSaveLocation={'s3'}
            initialGroups={[
              {
                name: 'Reject',
                definition: 'COUNT(tags) > 5',
              },
              {
                name: 'Sensitive',
                definition: 'COUNT(tags) > 0',
              },
              {
                name: 'Safe',
                definition: 'COUNT(tags) = 0',
              },
            ]}
            jobStarted={false}
          />
        )}

        {tabValue === 'analytics' && (
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
        )}

        {tabValue === 'output' && (
          <DatabaseTable 
            loadMoreObjectRecords={loadMoreMockObjectRecords}
            loadMoreClassifiedTokenRecords={loadMoreMockClassifiedTokenRecords}
            groups={mockGroups}
            tags={mockTags}
          />
        )}
      </Box>
    </Container>
  );
} 