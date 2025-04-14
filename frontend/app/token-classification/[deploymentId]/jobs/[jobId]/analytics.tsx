'use client';

import {
  Box,
  Paper,
  Typography,
  Grid,
  LinearProgress,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';

const tokenStats = [
  { tag: 'VIN', count: 150, percentage: 30 },
  { tag: 'NAME', count: 100, percentage: 20 },
  { tag: 'EMAIL', count: 75, percentage: 15 },
  { tag: 'SSN', count: 50, percentage: 10 },
  { tag: 'ADDRESS', count: 75, percentage: 15 },
  { tag: 'PHONE', count: 35, percentage: 7 },
  { tag: 'DOB', count: 15, percentage: 3 },
];

export default function Analytics() {
  return (
    <Stack spacing={3}>
      {/* Summary Cards */}
      <Grid container spacing={3}>
        {/* Progress Card */}
        <Grid item xs={12} md={4}>
          <Paper elevation={0} variant="outlined">
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Progress
              </Typography>
              <Box sx={{ mt: 2 }}>
                <LinearProgress variant="determinate" value={75} />
                <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                  75% Complete
                </Typography>
              </Box>
            </Box>
          </Paper>
        </Grid>

        {/* Tokens Processed Card */}
        <Grid item xs={12} md={4}>
          <Paper elevation={0} variant="outlined">
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Tokens Processed
              </Typography>
              <Typography variant="h4">500</Typography>
              <Typography variant="body2" color="text.secondary">
                Total tokens identified
              </Typography>
            </Box>
          </Paper>
        </Grid>

        {/* Live Latency Card */}
        <Grid item xs={12} md={4}>
          <Paper elevation={0} variant="outlined">
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Live Latency
              </Typography>
              <Typography variant="h4">250ms</Typography>
              <Typography variant="body2" color="text.secondary">
                Average processing time
              </Typography>
            </Box>
          </Paper>
        </Grid>
      </Grid>

      {/* Identified Tokens Section */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Identified Tokens
          </Typography>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Tag</TableCell>
                  <TableCell align="right">Count</TableCell>
                  <TableCell align="right">Percentage</TableCell>
                  <TableCell>Distribution</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {tokenStats.map((stat) => (
                  <TableRow key={stat.tag}>
                    <TableCell>{stat.tag}</TableCell>
                    <TableCell align="right">{stat.count}</TableCell>
                    <TableCell align="right">{stat.percentage}%</TableCell>
                    <TableCell>
                      <LinearProgress
                        variant="determinate"
                        value={stat.percentage}
                        sx={{ height: 8, borderRadius: 4 }}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Box>
      </Paper>
    </Stack>
  );
} 