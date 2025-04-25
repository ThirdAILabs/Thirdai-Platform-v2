'use client';

import React, { useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Box, Typography } from '@mui/material';
import { StorageOptionButton, GroupButton, AddGroupButton } from './ConfigurationCardButtons';

// Predefined tags based on the wireframe
const PREDEFINED_TAGS = [
  'VIN', 'NAME', 'ORG', 'ADDRESS', 'EMAIL', 'SSN', 'PHONE',
  'POLICY_ID', 'MED_REC_NO', 'LICENSE', 'EMPLOYER', 'ID', 'USERNAME',
  'URL', 'IP_ADDR', 'ZIP_CODE', 'ACCOUNT', 'INS_PROV', 'PROCEDURE',
  'DATE', 'NATIONALITY', 'SERIAL_NO', 'CRED_CARD_NUM', 'CVV'
];

const BUTTON_STYLES = {
  default: "border border-blue-500 bg-blue-500 text-white hover:bg-blue-600",
  outline: "border border-gray-200 hover:border-blue-500 hover:text-blue-500"
};

interface StorageConfig {
  name: string;
}

interface Group {
  name: string;
  definition: string;
}

interface ConfigurationCardProps {
  sourceS3Config?: StorageConfig;
  sourceLocalConfig?: StorageConfig;
  saveS3Config?: StorageConfig;
  saveLocalConfig?: StorageConfig;
  selectedSource?: 's3' | 'local' | null;
  selectedSaveLocation?: 's3' | 'local' | 'none' | null;
  initialGroups?: Group[];
  jobStarted?: boolean;
  onSourceSelect?: (source: 's3' | 'local') => void;
  onSaveLocationSelect?: (location: 's3' | 'local' | 'none') => void;
  onGroupsChange?: (groups: Group[]) => void;
  onTagsChange?: (tags: string[]) => void;
  onConfigureSource?: (type: 's3' | 'local', name: string) => void;
  onConfigureSaveLocation?: (type: 's3' | 'local', name: string) => void;
}

interface SectionTitleProps {
  children: React.ReactNode;
}

const SectionTitle: React.FC<SectionTitleProps> = ({ children }) => (
  <Typography 
    variant="h3" 
    className="text-gray-900"
    sx={{
      fontSize: '1.5rem',
      fontWeight: 600,
      lineHeight: 1.2,
      letterSpacing: '-0.025em'
    }}
  >
    {children}
  </Typography>
);

export const ConfigurationCard: React.FC<ConfigurationCardProps> = ({ 
  sourceS3Config = { name: '' },
  sourceLocalConfig = { name: '' },
  saveS3Config = { name: '' },
  saveLocalConfig = { name: '' },
  selectedSource: controlledSource = null,
  selectedSaveLocation: controlledSaveLocation = null,
  initialGroups = [],
  jobStarted = false,
  onSourceSelect,
  onSaveLocationSelect,
  onGroupsChange,
  onTagsChange,
  onConfigureSource,
  onConfigureSaveLocation,
}) => {
  // Internal state for selections when not controlled
  const [internalSelectedSource, setInternalSelectedSource] = useState<'s3' | 'local' | null>(controlledSource);
  const [internalSelectedSaveLocation, setInternalSelectedSaveLocation] = useState<'s3' | 'local' | 'none' | null>(controlledSaveLocation);
  
  // Use controlled state if provided, otherwise use internal state
  const selectedSource = controlledSource ?? internalSelectedSource;
  const selectedSaveLocation = controlledSaveLocation ?? internalSelectedSaveLocation;

  // State for selected tags
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  
  // State for groups
  const [groups, setGroups] = useState<Group[]>(initialGroups);

  // Handle source selection and configuration
  const handleSourceSelect = (source: 's3' | 'local') => {
    if (jobStarted) return;
    const name = prompt(`Configure ${source === 's3' ? 'S3 Bucket' : 'Local Storage'} name:`);
    if (name) {
      onConfigureSource?.(source, name);
    }
    
    setInternalSelectedSource(source);
    onSourceSelect?.(source);
  };

  // Handle tag selection
  const handleTagSelect = (tag: string) => {
    if (jobStarted) return;
    
    const newTags = selectedTags.includes(tag)
      ? selectedTags.filter(t => t !== tag)
      : [...selectedTags, tag];
    
    setSelectedTags(newTags);
    onTagsChange?.(newTags);
  };

  // Handle select all tags
  const handleSelectAllTags = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (jobStarted) return;
    
    const newTags = e.target.checked ? PREDEFINED_TAGS : [];
    setSelectedTags(newTags);
    onTagsChange?.(newTags);
  };

  // Handle adding a new group
  const handleAddGroup = () => {
    if (jobStarted) return;
    
    const newGroup: Group = {
      name: `Group ${groups.length + 1}`,
      definition: ''
    };
    const updatedGroups = [...groups, newGroup];
    setGroups(updatedGroups);
    onGroupsChange?.(updatedGroups);
  };

  // Handle removing a group
  const handleRemoveGroup = (indexToRemove: number) => {
    if (jobStarted) return;
    
    const updatedGroups = groups.filter((_, index) => index !== indexToRemove);
    setGroups(updatedGroups);
    onGroupsChange?.(updatedGroups);
  };

  // Handle updating group
  const handleUpdateGroup = (index: number, name: string, definition: string) => {
    if (jobStarted) return;
    
    const updatedGroups = [...groups];
    updatedGroups[index] = { name, definition };
    setGroups(updatedGroups);
    onGroupsChange?.(updatedGroups);
  };

  // Handle save location selection and configuration
  const handleSaveLocationSelect = (location: 's3' | 'local' | 'none') => {
    if (jobStarted) return;

    if (location !== 'none') {
      const name = prompt(`Configure ${location === 's3' ? 'S3 Bucket' : 'Local Storage'} name:`);
      if (name) {
        onConfigureSaveLocation?.(location, name);
      }
    }
    
    setInternalSelectedSaveLocation(location);
    onSaveLocationSelect?.(location);
  };

  return (
    <Card className="w-full relative">
      {/* Read-only notice */}
      {jobStarted && (
        <div className="absolute top-4 right-6 text-sm text-muted-foreground">
          The configuration of an initiated job is read only
        </div>
      )}

      <CardContent className="p-6">
        {/* Source Section */}
        <Box className="mb-8">
          <SectionTitle>Source</SectionTitle>
          <div className="flex gap-4 mt-4">
            <StorageOptionButton
              title="S3 Bucket"
              description={sourceS3Config.name || "Configure now"}
              isSelected={selectedSource === 's3'}
              onClick={() => handleSourceSelect('s3')}
              showEditIcon={!jobStarted}
              disabled={jobStarted}
            />
            <StorageOptionButton
              title="Local Storage"
              description={sourceLocalConfig.name || "Configure now"}
              isSelected={selectedSource === 'local'}
              onClick={() => handleSourceSelect('local')}
              showEditIcon={!jobStarted}
              disabled={jobStarted}
            />
            <StorageOptionButton
              title="More options"
              description="coming soon"
              disabled
            />
          </div>
        </Box>

        {/* Tags Section */}
        <Box className="mb-8">
          <div className="flex items-center justify-between">
            <SectionTitle>Tags</SectionTitle>
            <label className="flex items-center">
              <input 
                type="checkbox" 
                className="mr-2"
                checked={selectedTags.length === PREDEFINED_TAGS.length}
                onChange={handleSelectAllTags}
                disabled={jobStarted}
              />
              <span>Select All</span>
            </label>
          </div>
          <div className="flex flex-wrap gap-2 mt-4">
            {PREDEFINED_TAGS.map((tag) => (
              <Button
                key={tag}
                variant={selectedTags.includes(tag) ? "default" : "outline"}
                onClick={() => handleTagSelect(tag)}
                className={`text-sm ${selectedTags.includes(tag) ? BUTTON_STYLES.default : BUTTON_STYLES.outline}`}
                disabled={jobStarted}
              >
                {tag}
              </Button>
            ))}
          </div>
        </Box>

        {/* Groups Section */}
        <Box className="mb-8">
          <SectionTitle>Groups</SectionTitle>
          <div className="flex flex-wrap gap-4 mt-4">
            {groups.map((group, index) => (
              <GroupButton
                key={index}
                group={group}
                onEdit={(name, definition) => handleUpdateGroup(index, name, definition)}
                onDelete={() => handleRemoveGroup(index)}
              />
            ))}
            {!jobStarted && <AddGroupButton onClick={handleAddGroup} />}
          </div>
        </Box>

        {/* Save Groups To Section */}
        <Box>
          <SectionTitle>Save Groups To</SectionTitle>
          <div className="flex flex-wrap gap-4 mt-4">
            <StorageOptionButton
              title="S3 Bucket"
              description={saveS3Config.name || "Configure now"}
              isSelected={selectedSaveLocation === 's3'}
              onClick={() => handleSaveLocationSelect('s3')}
              showEditIcon={!jobStarted}
              disabled={jobStarted}
            />
            <StorageOptionButton
              title="Local Storage"
              description={saveLocalConfig.name || "Configure now"}
              isSelected={selectedSaveLocation === 'local'}
              onClick={() => handleSaveLocationSelect('local')}
              showEditIcon={!jobStarted}
              disabled={jobStarted}
            />
            <StorageOptionButton
              title="No storage location"
              description="You can still save groups"
              isSelected={selectedSaveLocation === 'none'}
              onClick={() => handleSaveLocationSelect('none')}
              disabled={jobStarted}
            />
            <StorageOptionButton
              title="More options"
              description="coming soon"
              disabled
            />
          </div>
        </Box>
      </CardContent>
    </Card>
  );
};

export default ConfigurationCard; 