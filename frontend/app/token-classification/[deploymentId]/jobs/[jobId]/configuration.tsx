'use client';

import {
  Box,
  Paper,
  Typography,
  Radio,
  RadioGroup,
  FormControlLabel,
  FormControl,
  FormLabel,
  Checkbox,
  Stack,
  Divider,
} from '@mui/material';

const tags = ['VIN', 'NAME', 'EMAIL', 'SSN', 'ADDRESS', 'PHONE', 'DOB'];

export default function Configuration() {
  return (
    <Stack spacing={3}>
      {/* Source Section */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Source
          </Typography>
          <FormControl disabled>
            <RadioGroup defaultValue="s3">
              <FormControlLabel value="s3" control={<Radio />} label="S3 Bucket" />
              <FormControlLabel value="local" control={<Radio />} label="Local Storage" />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                More options coming soon
              </Typography>
            </RadioGroup>
          </FormControl>
        </Box>
      </Paper>

      {/* Tags Section */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Tags
          </Typography>
          <Stack spacing={2}>
            <FormControlLabel
              control={<Checkbox disabled />}
              label="Select All"
            />
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 2 }}>
              {tags.map((tag) => (
                <FormControlLabel
                  key={tag}
                  control={<Checkbox disabled checked />}
                  label={tag}
                />
              ))}
            </Box>
          </Stack>
        </Box>
      </Paper>

      {/* Groups Section */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Groups
          </Typography>
          <Stack spacing={2}>
            <Paper variant="outlined">
              <Box sx={{ p: 2 }}>
                <Typography variant="subtitle1" gutterBottom>
                  Reject
                </Typography>
                <Box
                  component="pre"
                  sx={{
                    p: 2,
                    bgcolor: 'grey.100',
                    borderRadius: 1,
                    fontSize: '0.875rem',
                  }}
                >
                  {`SELECT * FROM tokens
WHERE tag IN ('SSN', 'DOB')
  AND confidence > 0.95`}
                </Box>
              </Box>
            </Paper>
            <Paper
              variant="outlined"
              sx={{
                p: 2,
                borderStyle: 'dashed',
              }}
            >
              <Typography variant="subtitle1" color="text.secondary">
                Define new group
              </Typography>
            </Paper>
          </Stack>
        </Box>
      </Paper>

      {/* Save Groups To Section */}
      <Paper elevation={0} variant="outlined">
        <Box sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>
            Save Groups To
          </Typography>
          <FormControl disabled>
            <RadioGroup defaultValue="local">
              <FormControlLabel value="s3" control={<Radio />} label="S3 Bucket" />
              <FormControlLabel value="local" control={<Radio />} label="Local Storage" />
              <FormControlLabel value="none" control={<Radio />} label="No storage location" />
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                More options coming soon
              </Typography>
            </RadioGroup>
          </FormControl>
        </Box>
      </Paper>
    </Stack>
  );
} 