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
  Tabs,
  Tab,
  Paper,
  IconButton,
  Stack,
} from '@mui/material';
import { RefreshRounded, PauseRounded, StopRounded } from '@mui/icons-material';
import { getReportStatus } from '@/lib/backend';
import Configuration from './configuration';
import Analytics from './analytics';
import Outputs from './outputs';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`tabpanel-${index}`}
      aria-labelledby={`tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

export default function JobDetail() {
  const params = useParams();
  const [lastUpdated, setLastUpdated] = useState(0);
  const [tabValue, setTabValue] = useState(0);

  // Fetch report status when component mounts
  useEffect(() => {
    const fetchReportStatus = async () => {
      try {
        const deploymentId = params.deploymentId as string;
        const jobId = params.jobId as string;
        const report = await getReportStatus(deploymentId, jobId);
        console.log('Report Status:', report);
      } catch (error) {
        console.error('Error fetching report status:', error);
      }
    };

    fetchReportStatus();
  }, [params.deploymentId, params.jobId]);

  // Increment last updated counter every second
  useEffect(() => {
    const interval = setInterval(() => {
      setLastUpdated(prev => prev + 1);
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'grey.100' }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', bgcolor: 'background.paper', p: 2 }}>
        <Container maxWidth="lg">
          <Stack spacing={2}>
            <Breadcrumbs>
              <Link href="/" passHref>
                <MuiLink underline="hover" color="inherit">
                  Home
                </MuiLink>
              </Link>
              <Link href={`/token-classification/${params.deploymentId}`} passHref>
                <MuiLink underline="hover" color="inherit">
                  HIPAA 25
                </MuiLink>
              </Link>
              <Link href={`/token-classification/${params.deploymentId}`} passHref>
                <MuiLink underline="hover" color="inherit">
                  Jobs
                </MuiLink>
              </Link>
              <Typography color="text.primary">Daniel Docs 1</Typography>
            </Breadcrumbs>

            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h4" component="h1">
                HIPAA 25
              </Typography>
              <Stack direction="row" spacing={2} alignItems="center">
                <Typography variant="body2" color="text.secondary">
                  Last updated {lastUpdated} seconds ago
                </Typography>
                <Stack direction="row" spacing={1}>
                  <IconButton size="small">
                    <RefreshRounded fontSize="small" />
                  </IconButton>
                  <IconButton size="small">
                    <PauseRounded fontSize="small" />
                  </IconButton>
                  <IconButton size="small">
                    <StopRounded fontSize="small" />
                  </IconButton>
                </Stack>
              </Stack>
            </Box>
          </Stack>
        </Container>
      </Box>

      <Container maxWidth="lg" sx={{ py: 3 }}>
        <Paper>
          <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
            <Tabs value={tabValue} onChange={handleTabChange}>
              <Tab label="Configuration" />
              <Tab label="Analytics" />
              <Tab label="Outputs" />
            </Tabs>
          </Box>
          <TabPanel value={tabValue} index={0}>
            <Configuration />
          </TabPanel>
          <TabPanel value={tabValue} index={1}>
            <Analytics />
          </TabPanel>
          <TabPanel value={tabValue} index={2}>
            <Outputs />
          </TabPanel>
        </Paper>
      </Container>
    </Box>
  );
} 