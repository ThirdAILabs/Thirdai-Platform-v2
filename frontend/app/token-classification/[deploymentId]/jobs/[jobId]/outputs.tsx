'use client';

import {
  Box,
  Paper,
  Typography,
  Stack,
  FormControlLabel,
  Checkbox,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Tooltip,
} from '@mui/material';
import { Download, RefreshCw, Trash2 } from 'lucide-react';

const outputs = [
  {
    id: '1',
    name: 'output_1.json',
    size: '2.5 MB',
    date: '2024-03-15 14:30:00',
    status: 'Completed',
  },
  {
    id: '2',
    name: 'output_2.json',
    size: '1.8 MB',
    date: '2024-03-15 14:35:00',
    status: 'Completed',
  },
  {
    id: '3',
    name: 'output_3.json',
    size: '3.2 MB',
    date: '2024-03-15 14:40:00',
    status: 'Completed',
  },
];

export default function Outputs() {
  return (
    <Stack spacing={3}>
      {/* Output Options */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Output Options
          </Typography>
          <Stack spacing={2}>
            <FormControlLabel
              control={<Checkbox defaultChecked />}
              label="Include raw text"
            />
            <FormControlLabel
              control={<Checkbox defaultChecked />}
              label="Include confidence scores"
            />
            <FormControlLabel
              control={<Checkbox defaultChecked />}
              label="Include metadata"
            />
          </Stack>
        </Box>
      </Paper>

      {/* Output Files */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Output Files
          </Typography>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Size</TableCell>
                  <TableCell>Date</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {outputs.map((output) => (
                  <TableRow key={output.id}>
                    <TableCell>{output.name}</TableCell>
                    <TableCell>{output.size}</TableCell>
                    <TableCell>{output.date}</TableCell>
                    <TableCell>{output.status}</TableCell>
                    <TableCell align="right">
                      <Stack direction="row" spacing={1} justifyContent="flex-end">
                        <Tooltip title="Download">
                          <IconButton size="small">
                            <Download size={16} />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Refresh">
                          <IconButton size="small">
                            <RefreshCw size={16} />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete">
                          <IconButton size="small">
                            <Trash2 size={16} />
                          </IconButton>
                        </Tooltip>
                      </Stack>
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