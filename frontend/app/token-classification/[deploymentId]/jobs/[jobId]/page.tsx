'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  Container,
  Box,
  Typography,
  Breadcrumbs,
  Link as MuiLink,
  IconButton,
  Stack,
  Button,
} from '@mui/material';
import { RefreshRounded, PauseRounded, StopRounded, ArrowBack } from '@mui/icons-material';
// import { getReportStatus, Report, getTagCounts, TagCount } from '@/lib/backend';
import Configuration from './configuration';
import Analytics from './analytics';
import Outputs from './outputs';
import { DatabaseTable } from './(database-table)/DatabaseTable';
import { loadMoreMockClassifiedTokenRecords, loadMoreMockObjectRecords, mockGroups, mockTags } from './mockdata';
import { AnalyticsDashboard } from '@/components/AnalyticsDashboard';
import ConfigurationCard from '@/components/ConfigurationCard';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

interface Token {
  text: string;
  tag: string;
}

interface HighlightColor {
  text: string;
  tag: string;
}

const SELECTING_COLOR = '#EFEFEF';
const SELECTED_COLOR = '#DFDFDF';

function Highlight({
  currentToken,
  tagColors,
}: {
  currentToken: Token;
  tagColors: Record<string, HighlightColor>;
}) {
  // Skip highlighting for "O" tags
  if (currentToken.tag == 'O') {
    return (
      <span>
        {currentToken.text}
        <span> </span>
      </span>
    );
  }

  return (
    <span
      style={{
        backgroundColor:
          currentToken.tag !== 'O'
            ? tagColors[currentToken.tag]?.text || 'transparent'
            : 'transparent',
        padding: '2px',
        borderRadius: '2px',
        userSelect: 'none',
        display: 'inline-flex',
        alignItems: 'center',
      }}
    >
      {currentToken.text}
      {currentToken.tag !== 'O' && (
        <span
          style={{
            backgroundColor: tagColors[currentToken.tag]?.tag,
            color: 'white',
            fontSize: '11px',
            fontWeight: 'bold',
            borderRadius: '2px',
            marginLeft: '4px',
            padding: '1px 3px',
          }}
        >
          {currentToken.tag}
        </span>
      )}
      <span> </span>
    </span>
  );
}

export default function JobDetail() {
  const params = useParams();
  const [lastUpdated, setLastUpdated] = useState(0);
  const [tabValue, setTabValue] = useState('analytics');

  return (
    <Container maxWidth="lg" sx={{ px: 4 }}>
      <Box sx={{ py: 3 }}>
        {/* Breadcrumbs */}
        <Stack direction="row" spacing={2} alignItems="center" mb={3}>
          <Breadcrumbs aria-label="breadcrumb">
            <MuiLink component={Link} href={`/token-classification/${params.deploymentId}/jobs`}>
              Jobs
            </MuiLink>
            <Typography color="text.primary">HIPAA 25</Typography>
          </Breadcrumbs>
        </Stack>

        {/* Title and Back Button */}
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between" mb={3}>
          <Typography variant="h5">HIPAA 25</Typography>
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
            <ConfigurationCard 
              sourceS3Config={{ name: '/path/to/bucket' }}
              sourceLocalConfig={{ name: '' }}
              saveS3Config={{ name: '/path/to/bucket' }}
              saveLocalConfig={{ name: 'local' }}
              selectedSource={'s3'}
              selectedSaveLocation={'s3'}
              initialGroups={[
                {
                  name: 'group1',
                  definition: 'group1 definition',
                },
              ]}
              jobStarted={true}
            /> 
          </TabsContent>
          
          <TabsContent value="analytics">
            <AnalyticsDashboard
              progress={40}
              tokensProcessed={1229000}
              latencyData={[
                { timestamp: '2024-03-10T12:00:00', latency: 5.6 },
                { timestamp: '2024-03-10T12:00:01', latency: 3.6 },
                { timestamp: '2024-03-10T12:00:02', latency: 2.6 },
                { timestamp: '2024-03-10T12:00:03', latency: 5.8 },
                { timestamp: '2024-03-10T12:00:04', latency: 5.7 },
                { timestamp: '2024-03-10T12:00:05', latency: 5.2 },
                { timestamp: '2024-03-10T12:00:06', latency: 5.1 },
                { timestamp: '2024-03-10T12:00:07', latency: 5.0 },
                { timestamp: '2024-03-10T12:00:08', latency: 4.9 },
                { timestamp: '2024-03-10T12:00:09', latency: 4.8 },
                { timestamp: '2024-03-10T12:00:10', latency: 4.7 },
                { timestamp: '2024-03-10T12:00:11', latency: 4.6 },
                { timestamp: '2024-03-10T12:00:12', latency: 4.5 },
                { timestamp: '2024-03-10T12:00:13', latency: 5.2 },
                { timestamp: '2024-03-10T12:00:14', latency: 5.3 },
                { timestamp: '2024-03-10T12:00:15', latency: 4.2 },
                { timestamp: '2024-03-10T12:00:16', latency: 4.1 },
                { timestamp: '2024-03-10T12:00:17', latency: 4.8 },
                { timestamp: '2024-03-10T12:00:18', latency: 3.9 },
                { timestamp: '2024-03-10T12:00:19', latency: 4.9 },
                { timestamp: '2024-03-10T12:00:20', latency: 5.0 },
                { timestamp: '2024-03-10T12:00:21', latency: 3.6 },
                { timestamp: '2024-03-10T12:00:22', latency: 5.1 },
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