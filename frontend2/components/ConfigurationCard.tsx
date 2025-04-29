'use client';

import React from 'react';
import { Typography, Box, Chip, Paper } from '@mui/material';

interface ConfigGroup {
  name: string;
  definition: string;
}

interface ConfigSource {
  name: string;
}

interface ConfigurationCardProps {
  sourceS3Config: ConfigSource;
  sourceLocalConfig: ConfigSource;
  saveS3Config: ConfigSource;
  saveLocalConfig: ConfigSource;
  selectedSource: string | null;
  selectedSaveLocation: string;
  initialGroups: ConfigGroup[];
  jobStarted: boolean;
}

const ConfigurationCard: React.FC<ConfigurationCardProps> = ({
  sourceS3Config,
  sourceLocalConfig,
  saveS3Config,
  saveLocalConfig,
  selectedSource,
  selectedSaveLocation,
  initialGroups,
  jobStarted
}) => {
  const tags = ['NAME', 'VIN', 'ORG', 'ID', 'SSN', 'ADDRESS', 'EMAIL'];
  
  const renderSource = () => {
    if (selectedSource === 's3' || (!selectedSource && sourceS3Config.name)) {
      return sourceS3Config.name;
    } else if (selectedSource === 'local' || (!selectedSource && sourceLocalConfig.name)) {
      return sourceLocalConfig.name;
    }
    return 'No source selected';
  };

  const renderSaveLocation = () => {
    if (selectedSaveLocation === 's3') {
      return saveS3Config.name;
    } else if (selectedSaveLocation === 'local') {
      return saveLocalConfig.name;
    }
    return 'No save location selected';
  };

  return (
    <Box sx={{ bgcolor: 'background.paper', p: 3, borderRadius: 1 }}>
      <Typography variant="h6" gutterBottom>Configuration</Typography>
      
      <Box sx={{ mt: 2 }}>
        <Typography variant="subtitle1">Source</Typography>
        <Paper variant="outlined" sx={{ p: 2, mt: 1, bgcolor: '#f5f5f5' }}>
          <Typography>{renderSource()}</Typography>
        </Paper>
      </Box>
      
      <Box sx={{ mt: 3 }}>
        <Typography variant="subtitle1">Save Location</Typography>
        <Paper variant="outlined" sx={{ p: 2, mt: 1, bgcolor: '#f5f5f5' }}>
          <Typography>{renderSaveLocation()}</Typography>
        </Paper>
      </Box>
      
      <Box sx={{ mt: 3 }}>
        <Typography variant="subtitle1">Tags</Typography>
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1 }}>
          {tags.map(tag => (
            <Chip 
              key={tag} 
              label={tag} 
              sx={{ 
                bgcolor: '#e3f2fd', 
                color: '#1976d2',
                borderRadius: 1
              }} 
            />
          ))}
        </Box>
      </Box>
      
      <Box sx={{ mt: 3 }}>
        <Typography variant="subtitle1">Groups</Typography>
        <Box sx={{ mt: 1 }}>
          {initialGroups.map((group, index) => (
            <Paper 
              key={index} 
              variant="outlined" 
              sx={{ 
                mb: index < initialGroups.length - 1 ? 2 : 0,
                borderRadius: 1
              }}
            >
              <Box sx={{ p: 2, borderBottom: '1px solid #e0e0e0' }}>
                <Typography variant="subtitle2">{group.name}</Typography>
              </Box>
              <Box sx={{ p: 2 }}>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  {group.definition}
                </Typography>
              </Box>
            </Paper>
          ))}
        </Box>
      </Box>
    </Box>
  );
};

export default ConfigurationCard; 