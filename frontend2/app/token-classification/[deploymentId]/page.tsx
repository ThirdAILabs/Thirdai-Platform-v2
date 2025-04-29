'use client';

import { useEffect, useState } from 'react';
import { Tabs, Tab, Box } from '@mui/material';
import { Breadcrumbs, Link as MuiLink, Typography } from '@mui/material';
import * as _ from 'lodash';
import { useParams, useSearchParams } from 'next/navigation';
import { Card, CardContent } from '@mui/material';
import { Alert } from '@mui/material';

// Import our implemented components
import Interact from './interact';
import Dashboard from './dashboard';
import Jobs from './jobs';

interface ModelUpdateProps {
  username: string;
  modelName: string;
  deploymentUrl: string;
  workflowNames: string[];
  deployStatus: string;
  modelId: string;
}

// Types included for completeness, but using Dashboard instead of ModelUpdate
type LabelMetrics = {
  [key: string]: {
    precision: number;
    recall: number;
    fmeasure: number;
  };
};

type ExampleCategories = {
  true_positives: Record<string, any>;
  false_positives: Record<string, any>;
  false_negatives: Record<string, any>;
};

type TrainReportData = {
  before_train_metrics: LabelMetrics;
  after_train_metrics: LabelMetrics;
  after_train_examples: ExampleCategories;
};

const emptyMetrics: LabelMetrics = {
  'O': {
    precision: 0,
    recall: 0,
    fmeasure: 0
  }
};

const emptyExamples: ExampleCategories = {
  true_positives: {},
  false_positives: {},
  false_negatives: {}
};

const emptyReport: TrainReportData = {
  before_train_metrics: emptyMetrics,
  after_train_metrics: emptyMetrics,
  after_train_examples: emptyExamples
};

// Mock API function until we implement the real one
const getTrainReport = async (workflowName: string) => {
  return { data: emptyReport };
};

export default function Page() {
  const params = useParams();
  const workflowName = params.deploymentId as string || 'PII';
  const searchParams = useSearchParams();
  const defaultTab = searchParams.get('tab') || 'testing';
  const [tabValue, setTabValue] = useState(defaultTab);
  const [trainReport, setTrainReport] = useState<TrainReportData>(emptyReport);
  const [isLoadingReport, setIsLoadingReport] = useState(false);
  const [reportError, setReportError] = useState('');

  const handleTabChange = (_event: React.SyntheticEvent, newValue: string) => {
    setTabValue(newValue);
  };

  // Fetch training report for monitoring tab
  useEffect(() => {
    const fetchReport = async () => {
      try {
        setIsLoadingReport(true);
        setReportError('');
        const response = await getTrainReport(workflowName);
        setTrainReport(response.data);
      } catch (error) {
        setReportError(error instanceof Error ? error.message : 'Failed to fetch training report');
        // Even on error, we want to show the TrainingResults component with empty data
        setTrainReport(emptyReport);
      } finally {
        setIsLoadingReport(false);
      }
    };

    fetchReport();
  }, [workflowName]);

  // Mock data for ModelUpdate component
  const mockModelData: ModelUpdateProps = {
    username: 'testuser',
    modelName: 'token-classifier',
    deploymentUrl: 'https://mock-deployment-url.com',
    workflowNames: ['token-classifier', 'token-classifier-v2'],
    deployStatus: 'complete',
    modelId: 'mock-model-id-123'
  };

  return (
    <div style={{ backgroundColor: '#F5F7FA', minHeight: '100vh' }}>
      <header style={{ width: '100%', padding: '16px', borderBottom: '1px solid #e0e0e0' }}>
        <div style={{ maxWidth: '1200px', margin: '0 auto', marginBottom: '16px' }}>
          <Typography 
            variant="body1" 
            color="textSecondary" 
            gutterBottom
            style={{ fontWeight: 500 }}
          >
            Token Classification
          </Typography>
          
          <Typography 
            variant="h5" 
            style={{ 
              fontWeight: 'bold', 
              overflow: 'hidden', 
              textOverflow: 'ellipsis', 
              whiteSpace: 'nowrap' 
            }} 
            title={workflowName}
          >
            {workflowName}
          </Typography>
        </div>
      </header>

      <main style={{ maxWidth: '1200px', margin: '0 auto', padding: '16px' }}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
          <Tabs 
            value={tabValue} 
            onChange={handleTabChange}
            aria-label="token classification tabs"
            sx={{
              '& .MuiTab-root': {
                textTransform: 'none',
                fontWeight: 500,
                fontSize: '1rem',
                color: '#5F6368',
                minWidth: 100,
                padding: '12px 16px',
                '&.Mui-selected': {
                  color: '#1a73e8',
                  fontWeight: 500
                }
              },
              '& .MuiTabs-indicator': {
                backgroundColor: '#1a73e8'
              }
            }}
          >
            <Tab label="Monitoring" value="monitoring" />
            <Tab label="Testing" value="testing" />
            <Tab label="Jobs" value="jobs" />
          </Tabs>
        </Box>
        
        {tabValue === 'monitoring' && (
          <Box>
            <Dashboard />
          </Box>
        )}
        
        {tabValue === 'testing' && (
          <Box>
            <Interact />
          </Box>
        )}
        
        {tabValue === 'jobs' && (
          <Box>
            <Jobs />
          </Box>
        )}
      </main>
    </div>
  );
} 