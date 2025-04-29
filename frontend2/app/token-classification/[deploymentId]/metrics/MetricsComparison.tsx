import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Tabs, Tab, Divider, useTheme } from '@mui/material';
import { MetricsChart } from './MetricsChart';

// Mock data for metrics
const mockMetricsData = {
  precision: {
    before: {
      Person: 0.82,
      Organization: 0.75,
      Location: 0.78,
      Date: 0.85,
      Product: 0.70
    },
    after: {
      Person: 0.88,
      Organization: 0.80,
      Location: 0.83,
      Date: 0.87,
      Product: 0.76
    }
  },
  recall: {
    before: {
      Person: 0.79,
      Organization: 0.72,
      Location: 0.75,
      Date: 0.81,
      Product: 0.68
    },
    after: {
      Person: 0.85,
      Organization: 0.78,
      Location: 0.81,
      Date: 0.84,
      Product: 0.73
    }
  },
  f1: {
    before: {
      Person: 0.80,
      Organization: 0.73,
      Location: 0.76,
      Date: 0.83,
      Product: 0.69
    },
    after: {
      Person: 0.86,
      Organization: 0.79,
      Location: 0.82,
      Date: 0.85,
      Product: 0.74
    }
  }
};

const MetricsComparison: React.FC = () => {
  const theme = useTheme();
  const [currentMetric, setCurrentMetric] = useState<'precision' | 'recall' | 'f1'>('f1');

  const handleMetricChange = (_: React.SyntheticEvent, newValue: 'precision' | 'recall' | 'f1') => {
    setCurrentMetric(newValue);
  };

  return (
    <Card 
      sx={{ 
        mb: 3, 
        boxShadow: '0px 2px 4px rgba(0, 0, 0, 0.05)',
        backgroundColor: '#FFFFFF',
        borderRadius: 1,
      }}
    >
      <CardContent>
        <Typography variant="h6" fontWeight={600} mb={2}>
          Training Metric Comparison
        </Typography>
        <Typography variant="body2" color="text.secondary" mb={3}>
          Compare model performance metrics before and after fine-tuning
        </Typography>
        
        <Tabs 
          value={currentMetric} 
          onChange={handleMetricChange}
          sx={{
            minHeight: '40px',
            mb: 2,
            '& .MuiTabs-indicator': {
              backgroundColor: theme.palette.primary.main,
            },
            '& .MuiTab-root': {
              textTransform: 'none',
              minHeight: '40px',
              fontWeight: 500,
              fontSize: '0.875rem',
              '&.Mui-selected': {
                color: theme.palette.primary.main,
                fontWeight: 600,
              }
            }
          }}
        >
          <Tab value="precision" label="Precision" />
          <Tab value="recall" label="Recall" />
          <Tab value="f1" label="F1 Score" />
        </Tabs>
        
        <Divider sx={{ mb: 3 }} />
        
        <Box sx={{ height: 350 }}>
          <MetricsChart metrics={mockMetricsData[currentMetric]} metricType={currentMetric} />
        </Box>
      </CardContent>
    </Card>
  );
};

export default MetricsComparison; 