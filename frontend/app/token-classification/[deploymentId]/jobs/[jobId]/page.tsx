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
  Card,
  CardContent,
  Button,
} from '@mui/material';
import { RefreshRounded, PauseRounded, StopRounded, ArrowBack } from '@mui/icons-material';
import { getReportStatus, Report } from '@/lib/backend';
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
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

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

function ReportContentDisplay({ report }: { report: Report }) {
  const [tagColors, setTagColors] = useState<Record<string, HighlightColor>>({});

  useEffect(() => {
    // Generate colors for tags
    const pastels = ['#E5A49C', '#F6C886', '#FBE7AA', '#99E3B5', '#A6E6E7', '#A5A1E1', '#D8A4E2'];
    const darkers = ['#D34F3E', '#F09336', '#F7CF5F', '#5CC96E', '#65CFD0', '#597CE2', '#B64DC8'];

    if (report.content) {
      const colors: Record<string, HighlightColor> = {};
      report.content.results.forEach(result => {
        const [_, results] = Object.entries(result)[0];
        results.forEach(item => {
          if (item.predicted_tags) {
            item.predicted_tags.forEach((tag: string, index: number) => {
              if (tag !== 'O' && !colors[tag]) {
                const i = Object.keys(colors).length;
                colors[tag] = {
                  text: pastels[i % pastels.length],
                  tag: darkers[i % darkers.length],
                };
              }
            });
          }
        });
      });
      setTagColors(colors);
    }
  }, [report.content]);

  if (!report.content) {
    return (
      <Card>
        <CardContent>
          <Typography variant="body1">No content available yet.</Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Stack spacing={2}>
          {report.content.results.map((result, index) => {
            const [docPath, results] = Object.entries(result)[0];
            return (
              <Paper key={index} elevation={1} sx={{ p: 2 }}>
                <Typography variant="h6" gutterBottom>
                  Document: {docPath}
                </Typography>
                <Box sx={{ pl: 2 }}>
                  {results.map((item, itemIndex) => {
                    if (item.tokens && item.predicted_tags) {
                      const tokens: Token[] = item.tokens.map((token: string, i: number) => ({
                        text: token,
                        tag: item.predicted_tags[i],
                      }));

                      return (
                        <Box key={itemIndex} sx={{ mb: 2 }}>
                          <Typography variant="body1" component="div">
                            {tokens.map((token, tokenIndex) => (
                              <Highlight
                                key={tokenIndex}
                                currentToken={token}
                                tagColors={tagColors}
                              />
                            ))}
                          </Typography>
                        </Box>
                      );
                    }
                    return null;
                  })}
                </Box>
              </Paper>
            );
          })}
        </Stack>
      </CardContent>
    </Card>
  );
}

export default function JobDetail() {
  const params = useParams();
  const [lastUpdated, setLastUpdated] = useState(0);
  const [tabValue, setTabValue] = useState(0);
  const [report, setReport] = useState<Report | null>(null);
  const [loading, setLoading] = useState(true);

  // Fetch report status when component mounts
  useEffect(() => {
    const fetchReportStatus = async () => {
      try {
        const deploymentId = params.deploymentId as string;
        const jobId = params.jobId as string;
        const reportData = await getReportStatus(deploymentId, jobId);
        setReport(reportData);
      } catch (error) {
        console.error('Error fetching report status:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchReportStatus();
  }, [params.deploymentId, params.jobId, lastUpdated]);

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

  if (loading) {
    return (
      <Container>
        <Box sx={{ p: 3 }}>
          <Typography>Loading...</Typography>
        </Box>
      </Container>
    );
  }

  if (!report) {
    return (
      <Container>
        <Box sx={{ p: 3 }}>
          <Typography color="error">Failed to load report</Typography>
        </Box>
      </Container>
    );
  }

  return (
    <Container>
      <Box sx={{ p: 3 }}>
        <Breadcrumbs aria-label="breadcrumb">
          <MuiLink component={Link} href={`/token-classification/${params.deploymentId}/jobs`}>
            Jobs
          </MuiLink>
          <Typography color="text.primary">Report {report.report_id}</Typography>
        </Breadcrumbs>

        <Box sx={{ mt: 3 }}>
          <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
            <Stack direction="row" spacing={2} alignItems="center">
              <Typography variant="h5">Report Details</Typography>
              <IconButton onClick={() => setLastUpdated(prev => prev + 1)}>
                <RefreshRounded />
              </IconButton>
            </Stack>
            <Button
              variant="outlined"
              startIcon={<ArrowBack />}
              component={Link}
              href={`/token-classification/${params.deploymentId}?tab=jobs`}
            >
              Back to Jobs
            </Button>
          </Stack>

          <Box sx={{ mt: 2 }}>
            <Tabs value={tabValue} onChange={handleTabChange}>
              <Tab label="Content" />
              <Tab label="Configuration" />
              <Tab label="Analytics" />
              <Tab label="Outputs" />
            </Tabs>

            <TabPanel value={tabValue} index={0}>
              <ReportContentDisplay report={report} />
            </TabPanel>
            <TabPanel value={tabValue} index={1}>
              <Configuration />
            </TabPanel>
            <TabPanel value={tabValue} index={2}>
              <Analytics />
            </TabPanel>
            <TabPanel value={tabValue} index={3}>
              <Outputs />
            </TabPanel>
          </Box>
        </Box>
      </Box>
    </Container>
  );
} 