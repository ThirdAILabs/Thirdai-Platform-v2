'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Pagination,
  FormControl,
  Select,
  MenuItem,
  InputLabel,
  CircularProgress
} from '@mui/material';

// Mock data for recent samples
const mockSamples = [
  {
    id: 's1',
    text: 'Please contact John Doe at john.doe@example.com or 555-123-4567.',
    tokens: [
      { text: 'Please', label: 'O' },
      { text: 'contact', label: 'O' },
      { text: 'John', label: 'PERSON' },
      { text: 'Doe', label: 'PERSON' },
      { text: 'at', label: 'O' },
      { text: 'john.doe@example.com', label: 'EMAIL' },
      { text: 'or', label: 'O' },
      { text: '555-123-4567', label: 'PHONE' }
    ],
    timestamp: '2023-11-15T14:32:00Z'
  },
  {
    id: 's2',
    text: 'Our office is located at 123 Business Ave, Suite 200, New York, NY 10001.',
    tokens: [
      { text: 'Our', label: 'O' },
      { text: 'office', label: 'O' },
      { text: 'is', label: 'O' },
      { text: 'located', label: 'O' },
      { text: 'at', label: 'O' },
      { text: '123', label: 'ADDRESS' },
      { text: 'Business', label: 'ADDRESS' },
      { text: 'Ave,', label: 'ADDRESS' },
      { text: 'Suite', label: 'ADDRESS' },
      { text: '200,', label: 'ADDRESS' },
      { text: 'New', label: 'ADDRESS' },
      { text: 'York,', label: 'ADDRESS' },
      { text: 'NY', label: 'ADDRESS' },
      { text: '10001', label: 'ADDRESS' }
    ],
    timestamp: '2023-11-14T09:18:00Z'
  },
  {
    id: 's3',
    text: 'Sarah Johnson will present the quarterly report tomorrow.',
    tokens: [
      { text: 'Sarah', label: 'PERSON' },
      { text: 'Johnson', label: 'PERSON' },
      { text: 'will', label: 'O' },
      { text: 'present', label: 'O' },
      { text: 'the', label: 'O' },
      { text: 'quarterly', label: 'O' },
      { text: 'report', label: 'O' },
      { text: 'tomorrow', label: 'DATE' }
    ],
    timestamp: '2023-11-13T16:45:00Z'
  },
  {
    id: 's4',
    text: 'The customer paid $125.99 for the product on January 15th.',
    tokens: [
      { text: 'The', label: 'O' },
      { text: 'customer', label: 'O' },
      { text: 'paid', label: 'O' },
      { text: '$125.99', label: 'MONEY' },
      { text: 'for', label: 'O' },
      { text: 'the', label: 'O' },
      { text: 'product', label: 'O' },
      { text: 'on', label: 'O' },
      { text: 'January', label: 'DATE' },
      { text: '15th', label: 'DATE' }
    ],
    timestamp: '2023-11-12T11:20:00Z'
  },
  {
    id: 's5',
    text: 'For assistance, call our help desk at (800) 555-HELP.',
    tokens: [
      { text: 'For', label: 'O' },
      { text: 'assistance,', label: 'O' },
      { text: 'call', label: 'O' },
      { text: 'our', label: 'O' },
      { text: 'help', label: 'O' },
      { text: 'desk', label: 'O' },
      { text: 'at', label: 'O' },
      { text: '(800)', label: 'PHONE' },
      { text: '555-HELP', label: 'PHONE' }
    ],
    timestamp: '2023-11-11T13:52:00Z'
  }
];

interface SampleToken {
  text: string;
  label: string;
}

interface Sample {
  id: string;
  text: string;
  tokens: SampleToken[];
  timestamp: string;
}

interface RecentSamplesProps {
  deploymentUrl?: string;
}

const RecentSamples: React.FC<RecentSamplesProps> = ({ deploymentUrl }) => {
  const [samples, setSamples] = useState<Sample[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [rowsPerPage, setRowsPerPage] = useState(3);
  const [uniqueLabels, setUniqueLabels] = useState<string[]>([]);
  const [selectedLabel, setSelectedLabel] = useState<string>('all');

  // Fetch samples (using mock data for now)
  useEffect(() => {
    const fetchSamples = async () => {
      try {
        setLoading(true);
        
        // In a real app, this would be an API call using the deploymentUrl
        // For now, simulate a delay and use mock data
        await new Promise(resolve => setTimeout(resolve, 800));
        
        setSamples(mockSamples);
        
        // Extract unique labels
        const allLabels = new Set<string>();
        mockSamples.forEach(sample => {
          sample.tokens.forEach(token => {
            if (token.label !== 'O') { // Exclude "Outside" label
              allLabels.add(token.label);
            }
          });
        });
        
        setUniqueLabels(Array.from(allLabels));
      } catch (error) {
        console.error('Error fetching samples:', error);
      } finally {
        setLoading(false);
      }
    };
    
    fetchSamples();
  }, [deploymentUrl]);

  const handleChangePage = (event: React.ChangeEvent<unknown>, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<{ value: unknown }>) => {
    setRowsPerPage(parseInt(event.target.value as string, 10));
    setPage(1);
  };

  const handleLabelChange = (event: React.ChangeEvent<{ value: unknown }>) => {
    setSelectedLabel(event.target.value as string);
    setPage(1); // Reset to first page on filter change
  };

  // Filter samples based on selected label
  const filteredSamples = selectedLabel === 'all'
    ? samples
    : samples.filter(sample => 
        sample.tokens.some(token => token.label === selectedLabel)
      );

  // Calculate pagination
  const startIndex = (page - 1) * rowsPerPage;
  const endIndex = startIndex + rowsPerPage;
  const paginatedSamples = filteredSamples.slice(startIndex, endIndex);
  const totalPages = Math.ceil(filteredSamples.length / rowsPerPage);

  // Format date function
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  // Highlight tokens based on label
  const highlightToken = (token: SampleToken, index: number) => {
    const getColor = () => {
      switch(token.label) {
        case 'PERSON':
          return { bg: '#e3f2fd', text: '#1565c0', border: '#90caf9' };
        case 'EMAIL':
          return { bg: '#f3e5f5', text: '#7b1fa2', border: '#ce93d8' };
        case 'PHONE':
          return { bg: '#e8f5e9', text: '#2e7d32', border: '#a5d6a7' };
        case 'ADDRESS':
          return { bg: '#fff3e0', text: '#e65100', border: '#ffcc80' };
        case 'DATE':
          return { bg: '#e0f7fa', text: '#00838f', border: '#80deea' };
        case 'MONEY':
          return { bg: '#f1f8e9', text: '#33691e', border: '#c5e1a5' };
        default:
          return token.label === 'O' 
            ? { bg: 'transparent', text: 'text.primary', border: 'transparent' } 
            : { bg: '#f5f5f5', text: '#757575', border: '#e0e0e0' };
      }
    };
    
    const { bg, text, border } = getColor();
    
    return (
      <Chip
        key={`${token.text}-${index}`}
        label={`${token.text} (${token.label})`}
        size="small"
        sx={{ 
          m: 0.5,
          backgroundColor: bg,
          color: text,
          border: token.label === 'O' ? 'none' : `1px solid ${border}`,
          '& .MuiChip-label': {
            px: 1,
            py: 0.5,
          }
        }}
      />
    );
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="subtitle1" sx={{ fontWeight: 500, mb: 2 }}>
        Recent Annotated Samples
      </Typography>
      
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <FormControl variant="outlined" size="small" sx={{ minWidth: 200 }}>
          <InputLabel id="label-filter-label">Filter by Label</InputLabel>
          <Select
            labelId="label-filter-label"
            value={selectedLabel}
            onChange={handleLabelChange as any}
            label="Filter by Label"
          >
            <MenuItem value="all">All Labels</MenuItem>
            {uniqueLabels.map((label) => (
              <MenuItem key={label} value={label}>{label}</MenuItem>
            ))}
          </Select>
        </FormControl>
        
        <FormControl variant="outlined" size="small" sx={{ minWidth: 120 }}>
          <InputLabel id="rows-per-page-label">Show</InputLabel>
          <Select
            labelId="rows-per-page-label"
            value={rowsPerPage}
            onChange={handleChangeRowsPerPage as any}
            label="Show"
          >
            <MenuItem value={3}>3 rows</MenuItem>
            <MenuItem value={5}>5 rows</MenuItem>
            <MenuItem value={10}>10 rows</MenuItem>
          </Select>
        </FormControl>
      </Box>
      
      {paginatedSamples.length > 0 ? (
        <>
          <TableContainer component={Paper} variant="outlined" sx={{ mb: 2 }}>
            <Table>
              <TableHead sx={{ backgroundColor: 'rgba(0, 0, 0, 0.03)' }}>
                <TableRow>
                  <TableCell sx={{ fontWeight: 500 }}>Sample Text & Tokens</TableCell>
                  <TableCell sx={{ fontWeight: 500, width: '180px' }}>Date</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {paginatedSamples.map((sample) => (
                  <TableRow key={sample.id}>
                    <TableCell>
                      <Box sx={{ mb: 1 }}>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>
                          {sample.text}
                        </Typography>
                      </Box>
                      <Box sx={{ display: 'flex', flexWrap: 'wrap' }}>
                        {sample.tokens.map((token, index) => 
                          highlightToken(token, index)
                        )}
                      </Box>
                    </TableCell>
                    <TableCell>
                      {formatDate(sample.timestamp)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Showing {paginatedSamples.length} of {filteredSamples.length} samples
            </Typography>
            <Pagination 
              count={totalPages} 
              page={page} 
              onChange={handleChangePage}
              color="primary"
              size="small"
            />
          </Box>
        </>
      ) : (
        <Paper 
          variant="outlined" 
          sx={{ p: 3, textAlign: 'center', color: 'text.secondary' }}
        >
          No samples found with the selected filter.
        </Paper>
      )}
    </Box>
  );
};

export default RecentSamples; 