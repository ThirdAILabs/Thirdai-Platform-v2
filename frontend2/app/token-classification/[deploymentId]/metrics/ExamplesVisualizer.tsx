'use client';

import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Tabs,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  Pagination,
  FormControl,
  InputLabel,
  Select,
  MenuItem
} from '@mui/material';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';

// Types for examples
interface TokenExample {
  id: string;
  text: string;
  trueLabel: string;
  predictedLabel: string;
  context: string;
  isCorrect: boolean;
}

interface ExamplesVisualizerProps {
  examples: TokenExample[];
}

export const ExamplesVisualizer: React.FC<ExamplesVisualizerProps> = ({ 
  examples = [] // Default to empty array if not provided
}) => {
  const [tabValue, setTabValue] = useState(0);
  const [page, setPage] = useState(1);
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const [selectedLabel, setSelectedLabel] = useState<string>('all');
  
  // Get unique labels for filter
  const uniqueLabels = ['all', ...new Set([
    ...examples.map(ex => ex.trueLabel),
    ...examples.map(ex => ex.predictedLabel)
  ])];
  
  // Filter examples based on tab and selected label
  const filteredExamples = examples.filter(example => {
    // Filter by correct/incorrect based on tab
    const matchesTab = tabValue === 0 ? true : 
                       tabValue === 1 ? example.isCorrect : 
                       tabValue === 2 ? !example.isCorrect : false;
    
    // Filter by selected label if not 'all'
    const matchesLabel = selectedLabel === 'all' ? true :
                         example.trueLabel === selectedLabel || 
                         example.predictedLabel === selectedLabel;
    
    return matchesTab && matchesLabel;
  });
  
  // Calculate pagination
  const startIndex = (page - 1) * rowsPerPage;
  const endIndex = startIndex + rowsPerPage;
  const paginatedExamples = filteredExamples.slice(startIndex, endIndex);
  const totalPages = Math.ceil(filteredExamples.length / rowsPerPage);
  
  // Handle tab change
  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
    setPage(1); // Reset to first page when changing tabs
  };
  
  // Handle page change
  const handlePageChange = (event: React.ChangeEvent<unknown>, value: number) => {
    setPage(value);
  };
  
  // Handle rows per page change
  const handleRowsPerPageChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setRowsPerPage(event.target.value as number);
    setPage(1); // Reset to first page
  };
  
  // Handle label filter change
  const handleLabelChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setSelectedLabel(event.target.value as string);
    setPage(1); // Reset to first page
  };
  
  // Helper function to highlight token in context
  const highlightToken = (context: string, token: string) => {
    if (!context || !token) return context;
    
    try {
      const parts = context.split(new RegExp(`(${token})`, 'gi'));
      return (
        <>
          {parts.map((part, i) => 
            part.toLowerCase() === token.toLowerCase() ? (
              <Box 
                component="span" 
                key={i} 
                sx={{ 
                  backgroundColor: 'primary.light',
                  padding: '0 2px',
                  borderRadius: '2px'
                }}
              >
                {part}
              </Box>
            ) : part
          )}
        </>
      );
    } catch (e) {
      return context;
    }
  };
  
  return (
    <Box sx={{ mb: 6 }}>
      <Box sx={{ mb: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 500, fontSize: '1.125rem' }}>
          Example Tokens
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Explore tokens that were correctly or incorrectly classified during fine-tuning
        </Typography>
      </Box>
      
      <Paper sx={{ 
        border: '1px solid rgba(0, 0, 0, 0.12)', 
        boxShadow: 'none' 
      }}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs 
            value={tabValue} 
            onChange={handleTabChange}
            sx={{
              '& .MuiTabs-indicator': {
                backgroundColor: 'primary.main',
              },
              '& .MuiTab-root': {
                textTransform: 'none',
                fontSize: '0.875rem',
                fontWeight: 500,
                color: 'text.primary',
                '&.Mui-selected': {
                  color: 'primary.main',
                  fontWeight: 600,
                }
              }
            }}
          >
            <Tab label="All Examples" />
            <Tab 
              label="Correct Predictions" 
              icon={<CheckCircleOutlineIcon fontSize="small" />} 
              iconPosition="start"
            />
            <Tab 
              label="Incorrect Predictions" 
              icon={<ErrorOutlineIcon fontSize="small" />} 
              iconPosition="start"
            />
          </Tabs>
        </Box>
        
        <Box sx={{ p: 3 }}>
          <Box sx={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            mb: 2
          }}>
            <FormControl variant="outlined" size="small" sx={{ minWidth: 200 }}>
              <InputLabel id="label-filter-label">Filter by Label</InputLabel>
              <Select
                labelId="label-filter-label"
                value={selectedLabel}
                onChange={handleLabelChange as any}
                label="Filter by Label"
              >
                {uniqueLabels.map((label) => (
                  <MenuItem key={label} value={label}>
                    {label === 'all' ? 'All Labels' : label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            
            <FormControl variant="outlined" size="small" sx={{ minWidth: 120 }}>
              <InputLabel id="rows-per-page-label">Rows</InputLabel>
              <Select
                labelId="rows-per-page-label"
                value={rowsPerPage}
                onChange={handleRowsPerPageChange as any}
                label="Rows"
              >
                {[5, 10, 25].map((count) => (
                  <MenuItem key={count} value={count}>
                    {count}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>
          
          <TableContainer>
            <Table aria-label="token examples table">
              <TableHead>
                <TableRow>
                  <TableCell>Token</TableCell>
                  <TableCell>Context</TableCell>
                  <TableCell>True Label</TableCell>
                  <TableCell>Predicted Label</TableCell>
                  <TableCell>Status</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {paginatedExamples.length > 0 ? (
                  paginatedExamples.map((example) => (
                    <TableRow key={example.id}>
                      <TableCell 
                        sx={{ 
                          fontWeight: 500,
                          maxWidth: '150px',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap'
                        }}
                      >
                        {example.text}
                      </TableCell>
                      <TableCell sx={{ maxWidth: '300px' }}>
                        {highlightToken(example.context, example.text)}
                      </TableCell>
                      <TableCell>
                        <Chip 
                          label={example.trueLabel} 
                          size="small"
                          sx={{ 
                            backgroundColor: example.isCorrect ? 'success.light' : 'error.light',
                            color: example.isCorrect ? 'success.dark' : 'error.dark',
                          }}
                        />
                      </TableCell>
                      <TableCell>
                        <Chip 
                          label={example.predictedLabel} 
                          size="small"
                          sx={{ 
                            backgroundColor: example.isCorrect ? 'success.light' : 'warning.light',
                            color: example.isCorrect ? 'success.dark' : 'warning.dark',
                          }}
                        />
                      </TableCell>
                      <TableCell>
                        {example.isCorrect ? (
                          <Chip
                            icon={<CheckCircleOutlineIcon />}
                            label="Correct"
                            size="small"
                            color="success"
                          />
                        ) : (
                          <Chip
                            icon={<ErrorOutlineIcon />}
                            label="Incorrect"
                            size="small"
                            color="error"
                          />
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={5} align="center" sx={{ py: 3 }}>
                      <Typography variant="body1" color="text.secondary">
                        No examples found
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
          
          {filteredExamples.length > 0 && (
            <Box sx={{ 
              display: 'flex', 
              justifyContent: 'flex-end',
              alignItems: 'center', 
              mt: 2 
            }}>
              <Typography variant="body2" color="text.secondary" sx={{ mr: 2 }}>
                Showing {paginatedExamples.length} of {filteredExamples.length} examples
              </Typography>
              <Pagination 
                count={totalPages} 
                page={page} 
                onChange={handlePageChange}
                color="primary"
                size="small"
              />
            </Box>
          )}
        </Box>
      </Paper>
    </Box>
  );
}; 