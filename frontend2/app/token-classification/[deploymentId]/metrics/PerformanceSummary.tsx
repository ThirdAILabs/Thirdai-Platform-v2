'use client';

import React from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  Grid,
  LinearProgress,
  Divider
} from '@mui/material';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import RemoveIcon from '@mui/icons-material/Remove';

// Types for metrics
interface Metrics {
  precision: number;
  recall: number;
  f1: number;
  [key: string]: number;
}

interface PerformanceSummaryProps {
  beforeMetrics: Metrics;
  afterMetrics: Metrics;
}

export const PerformanceSummary: React.FC<PerformanceSummaryProps> = ({
  beforeMetrics,
  afterMetrics
}) => {
  // Calculate changes
  const calculateChange = (before: number, after: number) => {
    return {
      value: after - before,
      percentage: before === 0 ? 100 : Math.round((after - before) / before * 100)
    };
  };

  const precisionChange = calculateChange(beforeMetrics.precision, afterMetrics.precision);
  const recallChange = calculateChange(beforeMetrics.recall, afterMetrics.recall);
  const f1Change = calculateChange(beforeMetrics.f1, afterMetrics.f1);

  // Helper to format percentages
  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(1)}%`;
  };

  // Helper to render change indicator
  const renderChangeIndicator = (change: { value: number; percentage: number }) => {
    if (change.value > 0) {
      return (
        <Box sx={{ display: 'flex', alignItems: 'center', color: 'success.main' }}>
          <ArrowUpwardIcon fontSize="small" sx={{ mr: 0.5 }} />
          <Typography variant="body2" sx={{ fontWeight: 500 }}>
            {change.value.toFixed(3)} ({change.percentage > 0 ? '+' : ''}{change.percentage}%)
          </Typography>
        </Box>
      );
    } else if (change.value < 0) {
      return (
        <Box sx={{ display: 'flex', alignItems: 'center', color: 'error.main' }}>
          <ArrowDownwardIcon fontSize="small" sx={{ mr: 0.5 }} />
          <Typography variant="body2" sx={{ fontWeight: 500 }}>
            {change.value.toFixed(3)} ({change.percentage}%)
          </Typography>
        </Box>
      );
    } else {
      return (
        <Box sx={{ display: 'flex', alignItems: 'center', color: 'text.secondary' }}>
          <RemoveIcon fontSize="small" sx={{ mr: 0.5 }} />
          <Typography variant="body2" sx={{ fontWeight: 500 }}>
            No change
          </Typography>
        </Box>
      );
    }
  };

  // Helper to render metric comparison
  const renderMetricComparison = (
    label: string,
    before: number,
    after: number,
    change: { value: number; percentage: number },
    color: string
  ) => {
    return (
      <Grid item xs={12} md={4}>
        <Box sx={{ p: 2, height: '100%' }}>
          <Typography variant="subtitle2" sx={{ mb: 1, fontWeight: 500 }}>
            {label}
          </Typography>
          
          <Box sx={{ mb: 3 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
              <Typography variant="body2" color="text.secondary">
                Before
              </Typography>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {formatPercentage(before)}
              </Typography>
            </Box>
            <LinearProgress 
              variant="determinate" 
              value={before * 100} 
              sx={{ 
                height: 8, 
                borderRadius: 4,
                backgroundColor: 'rgba(0, 0, 0, 0.08)',
                '& .MuiLinearProgress-bar': {
                  backgroundColor: color,
                  opacity: 0.6
                }
              }} 
            />
          </Box>
          
          <Box sx={{ mb: 1 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
              <Typography variant="body2" color="text.secondary">
                After
              </Typography>
              <Typography variant="body2" sx={{ fontWeight: 500 }}>
                {formatPercentage(after)}
              </Typography>
            </Box>
            <LinearProgress 
              variant="determinate" 
              value={after * 100} 
              sx={{ 
                height: 8, 
                borderRadius: 4,
                backgroundColor: 'rgba(0, 0, 0, 0.08)',
                '& .MuiLinearProgress-bar': {
                  backgroundColor: color
                }
              }} 
            />
          </Box>
          
          <Box sx={{ pt: 1 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
              Improvement
            </Typography>
            {renderChangeIndicator(change)}
          </Box>
        </Box>
      </Grid>
    );
  };

  return (
    <Box sx={{ mb: 4 }}>
      <Box sx={{ mb: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 500, fontSize: '1.125rem' }}>
          Performance Summary
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Comparison of model performance metrics before and after fine-tuning
        </Typography>
      </Box>
      
      <Paper sx={{ border: '1px solid rgba(0, 0, 0, 0.12)', boxShadow: 'none' }}>
        <Grid container>
          {renderMetricComparison(
            'Precision',
            beforeMetrics.precision,
            afterMetrics.precision,
            precisionChange,
            '#4caf50'
          )}
          
          <Grid item xs={12} md={4}>
            <Divider orientation="vertical" sx={{ height: '100%' }} />
          </Grid>
          
          {renderMetricComparison(
            'Recall',
            beforeMetrics.recall,
            afterMetrics.recall,
            recallChange,
            '#2196f3'
          )}
          
          <Grid item xs={12} md={4}>
            <Divider orientation="vertical" sx={{ height: '100%' }} />
          </Grid>
          
          {renderMetricComparison(
            'F1 Score',
            beforeMetrics.f1,
            afterMetrics.f1,
            f1Change,
            '#ff9800'
          )}
        </Grid>
        
        <Divider />
        
        <Box sx={{ p: 2, backgroundColor: 'rgba(0, 0, 0, 0.02)' }}>
          <Typography variant="body2" color="text.secondary">
            Overall Performance Change:
          </Typography>
          <Typography variant="body1" sx={{ fontWeight: 500, mt: 0.5 }}>
            The model's F1 score {f1Change.value > 0 ? 'improved by' : 'decreased by'} {Math.abs(f1Change.percentage)}% 
            after fine-tuning, with {precisionChange.value >= 0 ? 'better' : 'reduced'} precision and
            {recallChange.value >= 0 ? ' better' : ' reduced'} recall.
          </Typography>
        </Box>
      </Paper>
    </Box>
  );
}; 