'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import RefreshIcon from '@mui/icons-material/Refresh';
import { IconButton, Alert, AlertTitle } from '@mui/material';
import { Play } from 'lucide-react';

const mockLabels = ['NAME', 'PHONE', 'EMAIL', 'ADDRESS', 'DATE', 'ORGANIZATION', 'O'];

// Mock data for display
const mockSamples = [
  {
    id: 1,
    text: "Please contact John Doe at john.doe@example.com or 555-123-4567.",
    tokens: ["Please", "contact", "John", "Doe", "at", "john.doe@example.com", "or", "555-123-4567"],
    tags: ["O", "O", "NAME", "NAME", "O", "EMAIL", "O", "PHONE"],
    timestamp: "11/15/2023, 8:32:00 AM"
  },
  {
    id: 2,
    text: "Our office is located at 123 Main St",
    tokens: ["Our", "office", "is", "located", "at", "123", "Main", "St"],
    tags: ["O", "O", "O", "O", "O", "ADDRESS", "ADDRESS", "ADDRESS"],
    timestamp: "11/14/2023, 3:18:00 AM"
  },
  {
    id: 3,
    text: "Robert Johnson will visit on January 15th",
    tokens: ["Robert", "Johnson", "will", "visit", "on", "January", "15th"],
    tags: ["NAME", "NAME", "O", "O", "O", "DATE", "DATE"],
    timestamp: "11/13/2023, 10:45:00 AM"
  },
  {
    id: 4,
    text: "The meeting is at Google headquarters",
    tokens: ["The", "meeting", "is", "at", "Google", "headquarters"],
    tags: ["O", "O", "O", "O", "ORGANIZATION", "O"],
    timestamp: "11/12/2023, 9:30:00 AM"
  },
  {
    id: 5,
    text: "Please email Jane Doe at jane.doe@example.com",
    tokens: ["Please", "email", "Jane", "Doe", "at", "jane.doe@example.com"],
    tags: ["O", "O", "NAME", "NAME", "O", "EMAIL"],
    timestamp: "11/11/2023, 2:15:00 PM"
  }
];

const Separator: React.FC = () => <hr className="my-3 border-t border-gray-200" />;

interface HighlightColor {
  text: string;
  tag: string;
}

interface HighlightedSampleProps {
  tokens: string[];
  tags: string[];
  tagColors: Record<string, HighlightColor>;
}

function HighlightedSample({ tokens, tags, tagColors }: HighlightedSampleProps) {
  return (
    <div className="mb-4 leading-relaxed">
      {tokens.map((token, index) => (
        <span
          key={index}
          className={`inline-block px-1 py-0.5 rounded transition-colors duration-200 ease-in-out mr-1`}
          style={{
            backgroundColor: tagColors[tags[index]]?.text || 'transparent',
            userSelect: 'none',
          }}
        >
          {token}
          {tagColors[tags[index]] && (index === tokens.length - 1 || tags[index] !== tags[index + 1]) && (
            <span
              className="text-xs font-bold text-white rounded px-1 py-0.5 ml-1 align-text-top"
              style={{ backgroundColor: tagColors[tags[index]].tag }}
            >
              {tags[index]}
            </span>
          )}
        </span>
      ))}
    </div>
  );
}

interface RecentSamplesProps {
  deploymentUrl?: string;
}

const RecentSamples: React.FC<RecentSamplesProps> = ({ deploymentUrl }) => {
  const [samples] = useState(mockSamples);
  const [isLoadingLabels, setIsLoadingLabels] = useState(false);
  const [isLoadingSamples, setIsLoadingSamples] = useState(false);
  const [tagColors, setTagColors] = useState<Record<string, HighlightColor>>({});
  const colorAssignmentsRef = useRef<Record<string, HighlightColor>>({});

  const predefinedColors: HighlightColor[] = [
    { text: '#FEE2E2', tag: '#EF4444' }, // Red
    { text: '#FEF3C7', tag: '#F59E0B' }, // Amber
    { text: '#D1FAE5', tag: '#10B981' }, // Emerald
    { text: '#DBEAFE', tag: '#3B82F6' }, // Blue
    { text: '#E0E7FF', tag: '#6366F1' }, // Indigo
    { text: '#EDE9FE', tag: '#8B5CF6' }, // Violet
    { text: '#FCE7F3', tag: '#EC4899' }, // Pink
  ];

  // Extract all unique tags from samples (excluding 'O')
  const uniqueLabels = Array.from(
    new Set(samples.flatMap(sample => sample.tags))
  ).filter(tag => tag !== 'O');

  useEffect(() => {
    const allTags = samples.flatMap((sample) => sample.tags);
    const uniqueTags = Array.from(new Set(allTags)).filter((tag) => tag !== 'O');
    const newColors: Record<string, HighlightColor> = {};

    uniqueTags.forEach((tag, index) => {
      if (colorAssignmentsRef.current[tag]) {
        newColors[tag] = colorAssignmentsRef.current[tag];
      } else if (index < predefinedColors.length) {
        newColors[tag] = predefinedColors[index];
      } else {
        const hue = (index * 137.508) % 360;
        newColors[tag] = {
          text: `hsl(${hue}, 70%, 90%)`,
          tag: `hsl(${hue}, 70%, 40%)`,
        };
      }
      colorAssignmentsRef.current[tag] = newColors[tag];
    });

    setTagColors(newColors);
  }, [samples]);

  const handleRefresh = async () => {
    setIsLoadingLabels(true);
    setIsLoadingSamples(true);
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 800));
    
    setIsLoadingLabels(false);
    setIsLoadingSamples(false);
  };

  return (
    <div className="flex flex-col w-full box-border">
      <div className="flex justify-end mb-4">
        <IconButton
          onClick={handleRefresh}
          disabled={isLoadingLabels || isLoadingSamples}
          color="primary"
          size="large"
          className={isLoadingLabels || isLoadingSamples ? 'animate-spin' : ''}
        >
          <RefreshIcon />
        </IconButton>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="h-[calc(100vh-16rem)] overflow-hidden">
          <CardHeader className="pb-2">
            <CardTitle className="text-xl font-semibold">Recent TAGs</CardTitle>
            <CardDescription>List of active TAGs</CardDescription>
          </CardHeader>
          <CardContent className="overflow-y-auto h-[calc(100%-5rem)]">
            {uniqueLabels.map((label, idx) => (
              <React.Fragment key={idx}>
                {idx > 0 && <Separator />}
                <div className="mb-2 p-2 bg-gray-100 rounded-md">
                  <span className="font-medium">{label}</span>
                </div>
              </React.Fragment>
            ))}
          </CardContent>
        </Card>

        <Card className="h-[calc(100vh-16rem)] overflow-hidden">
          <CardHeader className="pb-2">
            <CardTitle className="text-xl font-semibold">Recent Samples from Users</CardTitle>
            <CardDescription>The latest inserted samples</CardDescription>
          </CardHeader>
          <CardContent className="overflow-y-auto h-[calc(100%-5rem)]">
            {samples.map((sample, idx) => (
              <React.Fragment key={idx}>
                {idx > 0 && <Separator />}
                <HighlightedSample
                  tokens={sample.tokens}
                  tags={sample.tags}
                  tagColors={tagColors}
                />
              </React.Fragment>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default RecentSamples; 