import React, { useState } from 'react';
import { Box, Card, CardContent, Typography, Grid, Paper, Button } from '@mui/material';

// Keep only the table data
const entityLabels = ['O', 'NAME', 'PHONE', 'EMAIL', 'ADDRESS'];
const tableData = [
  { label: 'O', beforeF1: 0.89, afterF1: 0.93, change: 0.04 },
  { label: 'NAME', beforeF1: 0.81, afterF1: 0.86, change: 0.05 },
  { label: 'PHONE', beforeF1: 0.84, afterF1: 0.90, change: 0.06 },
  { label: 'EMAIL', beforeF1: 0.88, afterF1: 0.94, change: 0.06 },
  { label: 'ADDRESS', beforeF1: 0.79, afterF1: 0.87, change: 0.08 },
  { label: 'Overall Average', beforeF1: 0.842, afterF1: 0.90, change: 0.058 }
];

// Metrics for specific entity (for the detailed view)
const entitySpecificMetrics = {
  'O': { 
    precision: 0.95, 
    recall: 0.92, 
    f1Score: 0.93,
    beforeValues: { precision: 0.90, recall: 0.88, f1Score: 0.89 }
  },
  'NAME': { 
    precision: 0.88, 
    recall: 0.84, 
    f1Score: 0.86,
    beforeValues: { precision: 0.83, recall: 0.79, f1Score: 0.81 }
  },
  'PHONE': { 
    precision: 0.92, 
    recall: 0.88, 
    f1Score: 0.90,
    beforeValues: { precision: 0.86, recall: 0.82, f1Score: 0.84 }
  },
  'EMAIL': { 
    precision: 0.96, 
    recall: 0.92, 
    f1Score: 0.94,
    beforeValues: { precision: 0.90, recall: 0.86, f1Score: 0.88 }
  },
  'ADDRESS': { 
    precision: 0.89, 
    recall: 0.85, 
    f1Score: 0.87,
    beforeValues: { precision: 0.81, recall: 0.77, f1Score: 0.79 }
  }
};

const TrainingResults = () => {
  const [selectedEntity, setSelectedEntity] = useState('O');

  // Handle entity selection for detailed view
  const handleEntityChange = (entity: string) => {
    setSelectedEntity(entity);
  };

  return (
    <Card sx={{ mb: 4, backgroundColor: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}>
      <CardContent>
        <Typography variant="h6" sx={{ fontWeight: 500, mb: 1 }}>
          Fine-tuning Metrics Comparison
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Performance metrics comparison before and after fine-tuning the model with your data
        </Typography>
        
        <Typography variant="h6" fontWeight={500} mt={2} mb={1}>
          Performance Summary
        </Typography>
        <Typography variant="body2" color="text.secondary" mb={2}>
          Overview of F1 score changes across all labels
        </Typography>
        
        <Paper sx={{ width: '100%', boxShadow: 'none', border: '1px solid rgba(0, 0, 0, 0.12)', borderRadius: '4px', mb: 4 }}>
          <Box sx={{ 
            display: 'grid',
            gridTemplateColumns: '1fr 1fr 1fr 1fr',
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
            bgcolor: 'rgba(0, 0, 0, 0.02)',
            p: 2
          }}>
            <Typography variant="subtitle2" fontWeight={500}>TAG</Typography>
            <Typography variant="subtitle2" fontWeight={500}>F1 before fine-tuning</Typography>
            <Typography variant="subtitle2" fontWeight={500}>F1 after fine-tuning</Typography>
            <Typography variant="subtitle2" fontWeight={500}>Change</Typography>
          </Box>
          
          {tableData.map((row, index) => (
            <Box 
              key={row.label}
              sx={{ 
                display: 'grid',
                gridTemplateColumns: '1fr 1fr 1fr 1fr',
                p: 2,
                borderBottom: index < tableData.length - 1 ? '1px solid rgba(0, 0, 0, 0.06)' : 'none',
                bgcolor: row.label === 'Overall Average' ? 'rgba(0, 0, 0, 0.02)' : 'transparent',
                fontWeight: row.label === 'Overall Average' ? 500 : 400,
              }}
            >
              <Typography variant="body2" fontWeight={row.label === 'Overall Average' ? 500 : 400}>
                {row.label}
              </Typography>
              <Typography variant="body2">{(row.beforeF1 * 100).toFixed(1)}%</Typography>
              <Typography variant="body2">{(row.afterF1 * 100).toFixed(1)}%</Typography>
              <Box sx={{ display: 'flex', alignItems: 'center' }}>
                <Box 
                  component="span" 
                  sx={{ 
                    color: '#4caf50',
                    display: 'flex',
                    alignItems: 'center'
                  }}
                >
                  <Box component="span" sx={{ mr: 0.5 }}>↑</Box>
                  +{(row.change * 100).toFixed(1)}%
                </Box>
              </Box>
            </Box>
          ))}
        </Paper>
        
        {/* Filter tabs */}
        <Box sx={{ mb: 2 }}>
          <Box sx={{ display: 'flex', gap: 1, mb: 3 }}>
            {entityLabels.map(label => (
              <Button
                key={label}
                variant={selectedEntity === label ? 'contained' : 'outlined'}
                color="primary"
                size="small"
                sx={{ 
                  borderRadius: '16px', 
                  px: 2, 
                  py: 0.5,
                  backgroundColor: selectedEntity === label ? '#E3F2FD' : '#F5F7FA',
                  color: selectedEntity === label ? '#1976d2' : 'rgba(0, 0, 0, 0.6)',
                  border: selectedEntity === label ? '1px solid #90CAF9' : '1px solid rgba(0, 0, 0, 0.12)',
                  '&:hover': {
                    backgroundColor: selectedEntity === label ? '#BBDEFB' : '#ECEFF1'
                  }
                }}
                onClick={() => handleEntityChange(label)}
              >
                {label}
              </Button>
            ))}
          </Box>
        </Box>
        
        {/* Entity-specific metrics */}
        <Paper sx={{ p: 3, border: '1px solid rgba(0, 0, 0, 0.12)', boxShadow: 'none', borderRadius: '4px' }}>
          <Typography variant="h6" fontWeight={500} mb={1}>
            Metrics for "{selectedEntity}"
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={3}>
            Comparing performance before and after training
          </Typography>
          
          <Box sx={{ 
            height: 300, 
            px: 3, 
            display: 'flex', 
            justifyContent: 'space-around', 
            alignItems: 'flex-end',
            borderBottom: '1px solid rgba(0, 0, 0, 0.12)',
            position: 'relative',
            mb: 3
          }}>
            {/* Y-axis labels */}
            <Box sx={{ position: 'absolute', left: 0, top: 0, bottom: 0, display: 'flex', flexDirection: 'column', justifyContent: 'space-between', py: 3 }}>
              <Typography variant="caption" color="text.secondary">100%</Typography>
              <Typography variant="caption" color="text.secondary">75%</Typography>
              <Typography variant="caption" color="text.secondary">50%</Typography>
              <Typography variant="caption" color="text.secondary">25%</Typography>
              <Typography variant="caption" color="text.secondary">0%</Typography>
            </Box>
            
            {/* Grid lines */}
            <Box sx={{ position: 'absolute', left: 30, right: 0, top: 0, bottom: 0, display: 'flex', flexDirection: 'column', justifyContent: 'space-between', py: 3 }}>
              <Box sx={{ borderTop: '1px dashed rgba(0, 0, 0, 0.1)', width: '100%' }} />
              <Box sx={{ borderTop: '1px dashed rgba(0, 0, 0, 0.1)', width: '100%' }} />
              <Box sx={{ borderTop: '1px dashed rgba(0, 0, 0, 0.1)', width: '100%' }} />
              <Box sx={{ borderTop: '1px dashed rgba(0, 0, 0, 0.1)', width: '100%' }} />
              <Box sx={{ borderTop: '1px dashed rgba(0, 0, 0, 0.1)', width: '100%' }} />
            </Box>
            
            {/* Bar charts */}
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', zIndex: 1, width: '80px' }}>
              <Typography variant="body2" mb={2}>Precision</Typography>
              <Box sx={{ display: 'flex', width: '100%', height: '200px', alignItems: 'flex-end' }}>
                <Box sx={{ 
                  width: '45%', 
                  height: `${entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].beforeValues.precision * 200}px`, 
                  bgcolor: '#94bbf7',
                  mr: 1
                }} />
                <Box sx={{ 
                  width: '45%', 
                  height: `${entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].precision * 200}px`, 
                  bgcolor: '#4A90E2'
                }} />
              </Box>
            </Box>
            
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', zIndex: 1, width: '80px' }}>
              <Typography variant="body2" mb={2}>Recall</Typography>
              <Box sx={{ display: 'flex', width: '100%', height: '200px', alignItems: 'flex-end' }}>
                <Box sx={{ 
                  width: '45%', 
                  height: `${entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].beforeValues.recall * 200}px`, 
                  bgcolor: '#94d3a7',
                  mr: 1
                }} />
                <Box sx={{ 
                  width: '45%', 
                  height: `${entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].recall * 200}px`, 
                  bgcolor: '#50C878'
                }} />
              </Box>
            </Box>
            
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', zIndex: 1, width: '80px' }}>
              <Typography variant="body2" mb={2}>F1</Typography>
              <Box sx={{ display: 'flex', width: '100%', height: '200px', alignItems: 'flex-end' }}>
                <Box sx={{ 
                  width: '45%', 
                  height: `${entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].beforeValues.f1Score * 200}px`, 
                  bgcolor: '#ffcc90',
                  mr: 1
                }} />
                <Box sx={{ 
                  width: '45%', 
                  height: `${entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].f1Score * 200}px`, 
                  bgcolor: '#FF8C00'
                }} />
              </Box>
            </Box>
          </Box>
          
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', mb: 3 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mr: 3 }}>
              <Box sx={{ width: 16, height: 16, bgcolor: '#94bbf7', mr: 1, borderRadius: 1 }} />
              <Typography variant="caption">Before Training</Typography>
            </Box>
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <Box sx={{ width: 16, height: 16, bgcolor: '#4A90E2', mr: 1, borderRadius: 1 }} />
              <Typography variant="caption">After Training</Typography>
            </Box>
          </Box>
          
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 2 }}>
            <Box sx={{ 
              p: 2, 
              bgcolor: 'rgba(0, 0, 0, 0.02)', 
              borderRadius: '4px', 
              flex: 1, 
              mr: 2, 
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center' 
            }}>
              <Typography variant="body2" color="text.secondary" mb={1}>Precision</Typography>
              <Typography variant="h5" fontWeight={500}>{(entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].precision * 100).toFixed(1)}%</Typography>
              <Box sx={{ 
                color: '#4caf50', 
                display: 'flex', 
                alignItems: 'center',
                mt: 1
              }}>
                <Box component="span" sx={{ mr: 0.5 }}>+</Box>
                {((entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].precision - 
                  entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].beforeValues.precision) * 100).toFixed(1)}%
              </Box>
            </Box>
            
            <Box sx={{ 
              p: 2, 
              bgcolor: 'rgba(0, 0, 0, 0.02)', 
              borderRadius: '4px', 
              flex: 1, 
              mr: 2,
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center'
            }}>
              <Typography variant="body2" color="text.secondary" mb={1}>Recall</Typography>
              <Typography variant="h5" fontWeight={500}>{(entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].recall * 100).toFixed(1)}%</Typography>
              <Box sx={{ 
                color: '#4caf50', 
                display: 'flex', 
                alignItems: 'center',
                mt: 1
              }}>
                <Box component="span" sx={{ mr: 0.5 }}>+</Box>
                {((entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].recall - 
                  entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].beforeValues.recall) * 100).toFixed(1)}%
              </Box>
            </Box>
            
            <Box sx={{ 
              p: 2, 
              bgcolor: 'rgba(0, 0, 0, 0.02)', 
              borderRadius: '4px', 
              flex: 1,
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center'
            }}>
              <Typography variant="body2" color="text.secondary" mb={1}>F1 Score</Typography>
              <Typography variant="h5" fontWeight={500}>{(entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].f1Score * 100).toFixed(1)}%</Typography>
              <Box sx={{ 
                color: '#4caf50', 
                display: 'flex', 
                alignItems: 'center',
                mt: 1
              }}>
                <Box component="span" sx={{ mr: 0.5 }}>+</Box>
                {((entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].f1Score - 
                  entitySpecificMetrics[selectedEntity as keyof typeof entitySpecificMetrics].beforeValues.f1Score) * 100).toFixed(1)}%
              </Box>
            </Box>
          </Box>
        </Paper>
      </CardContent>
    </Card>
  );
};

export default TrainingResults; 