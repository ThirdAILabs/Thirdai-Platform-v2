'use client';

import React from 'react';
import { Box, Typography, Tooltip } from '@mui/material';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';

// Define the data structure for metrics
export interface MetricsData {
  entityType: string;
  before: number;
  after: number;
}

interface MetricsChartProps {
  title: string;
  data: MetricsData[];
  color: string;
}

export const MetricsChart: React.FC<MetricsChartProps> = ({ title, data, color }) => {
  return (
    <Box sx={{ p: 2, border: '1px solid #eaeaea', borderRadius: 1, height: '100%' }}>
      <Typography variant="subtitle1" sx={{ fontWeight: 500, mb: 2 }}>
        {title}
      </Typography>

      {data.map((item, index) => {
        const improvement = ((item.after - item.before) / item.before) * 100;
        const showImprovement = improvement > 0;

        return (
          <Box key={index} sx={{ mb: 2.5 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
              <Typography variant="body2" fontWeight={500}>
                {item.entityType}
              </Typography>
              
              {showImprovement && (
                <Tooltip title={`${improvement.toFixed(1)}% improvement`}>
                  <Box sx={{ display: 'flex', alignItems: 'center', color: 'success.main' }}>
                    <ArrowUpwardIcon fontSize="small" sx={{ fontSize: 14, mr: 0.5 }} />
                    <Typography variant="caption">
                      {improvement.toFixed(1)}%
                    </Typography>
                  </Box>
                </Tooltip>
              )}
            </Box>
            
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.5 }}>
              <Typography variant="caption" sx={{ width: 60, color: 'text.secondary' }}>
                Before:
              </Typography>
              <Box sx={{ flex: 1, position: 'relative' }}>
                <Box
                  sx={{
                    height: 8,
                    width: `${item.before * 100}%`,
                    bgcolor: color,
                    opacity: 0.3,
                    borderRadius: 1
                  }}
                />
                <Typography 
                  variant="caption" 
                  sx={{ 
                    position: 'absolute', 
                    right: -40, 
                    top: -2 
                  }}
                >
                  {(item.before * 100).toFixed(1)}%
                </Typography>
              </Box>
            </Box>
            
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <Typography variant="caption" sx={{ width: 60, color: 'text.secondary' }}>
                After:
              </Typography>
              <Box sx={{ flex: 1, position: 'relative' }}>
                <Box
                  sx={{
                    height: 8,
                    width: `${item.after * 100}%`,
                    bgcolor: color,
                    borderRadius: 1
                  }}
                />
                <Typography 
                  variant="caption" 
                  sx={{ 
                    position: 'absolute', 
                    right: -40, 
                    top: -2 
                  }}
                >
                  {(item.after * 100).toFixed(1)}%
                </Typography>
              </Box>
            </Box>
          </Box>
        );
      })}
    </Box>
  );
}; 