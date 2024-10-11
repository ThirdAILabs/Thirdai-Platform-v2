'use client';
import React, { useState, useEffect, useRef, MouseEvent as ReactMouseEvent } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useLabels, useRecentSamples } from '@/lib/backend';
import { associations, reformulations, upvotes } from './mock_samples';
import useRollingSamples from './rolling';
import axios from 'axios';

interface TextPairsProps {
  timestamp: string;
  label1: string;
  label2: string;
  text1: string;
  text2: string;
}

function TextPairs({ timestamp, label1, label2, text1, text2 }: TextPairsProps) {
  return (
    <div
      className="text-md"
      style={{ display: 'flex', flexDirection: 'column', marginBottom: '10px' }}
    >
      <CardDescription>{timestamp}</CardDescription>
      <div className="text-md" style={{ display: 'flex', flexDirection: 'row' }}>
        <span style={{ fontWeight: 'bold', marginRight: '5px' }}>{label1}:</span>
        <span style={{}}>{text1}</span>
      </div>
      <div className="text-md" style={{ display: 'flex', flexDirection: 'row' }}>
        <span style={{ fontWeight: 'bold', marginRight: '5px' }}>{label2}:</span>
        <span>{text2}</span>
      </div>
    </div>
  );
}

interface ReformulationProps {
  timestamp: string;
  original: string;
  reformulations: string[];
}

function Reformulation({ timestamp, original, reformulations }: ReformulationProps) {
  return (
    <div
      className="text-md"
      style={{ display: 'flex', flexDirection: 'column', marginBottom: '10px' }}
    >
      <CardDescription>{timestamp}</CardDescription>
      <span className="text-md" style={{ fontWeight: 'bold' }}>
        {original}
      </span>
      {reformulations.map((r, i) => (
        <span key={i} className="text-md" style={{ marginLeft: '10px' }}>
          {r}
        </span>
      ))}
    </div>
  );
}

interface Token {
  text: string;
  tag: string;
}

interface HighlightColor {
  text: string;
  tag: string;
}

interface HighlightProps {
  currentToken: Token;
  nextToken?: Token | null;
  tagColors: Record<string, HighlightColor>;
  onMouseOver: (e: ReactMouseEvent<HTMLSpanElement>) => void;
  onMouseDown: (e: ReactMouseEvent<HTMLSpanElement>) => void;
  selecting: boolean;
  selected: boolean;
}

const SELECTING_COLOR = '#EFEFEF';
const SELECTED_COLOR = '#DFDFDF';

const Highlight: React.FC<HighlightProps> = ({
  currentToken,
  nextToken,
  tagColors,
  onMouseOver,
  onMouseDown,
  selecting,
  selected,
}) => {
  const [hover, setHover] = useState<boolean>(false);

  return (
    <>
      <span
        style={{
          backgroundColor:
            hover || selecting
              ? SELECTING_COLOR
              : selected
              ? SELECTED_COLOR
              : tagColors[currentToken.tag]?.text || 'transparent',
          padding: '2px',
          borderRadius: '2px',
          cursor: hover ? 'pointer' : 'default',
          userSelect: 'none',
        }}
        onMouseOver={(e) => {
          setHover(true);
          onMouseOver(e);
        }}
        onMouseLeave={() => {
          setHover(false);
        }}
        onMouseDown={onMouseDown}
      >
        {currentToken.text}
        {tagColors[currentToken.tag] && nextToken?.tag !== currentToken.tag && (
          <span
            style={{
              backgroundColor: tagColors[currentToken.tag].tag,
              color: 'white',
              fontSize: '11px',
              fontWeight: 'bold',
              borderRadius: '2px',
              marginLeft: '4px',
              padding: '5px 3px 1px 3px',
              marginBottom: '1px',
            }}
          >
            {currentToken.tag}
          </span>
        )}
      </span>
      <span
        style={{ cursor: hover ? 'pointer' : 'default', userSelect: 'none' }}
        onMouseOver={(e) => {
          setHover(true);
          onMouseOver(e);
        }}
        onMouseLeave={() => {
          setHover(false);
        }}
        onMouseDown={onMouseDown}
      >
        {' '}
      </span>
    </>
  );
};

interface HighlightedSampleProps {
  tokens: string[];
  tags: string[];
  tagColors: Record<string, HighlightColor>;
}

const HighlightedSample: React.FC<HighlightedSampleProps> = ({ tokens, tags, tagColors }) => {
  const handleMouseOver = (e: ReactMouseEvent<HTMLSpanElement>) => {
    // Handle mouse over event if needed
  };

  const handleMouseDown = (e: ReactMouseEvent<HTMLSpanElement>) => {
    // Handle mouse down event if needed
  };

  return (
    <div style={{ lineHeight: 2, marginBottom: '10px' }}>
      {tokens.map((token, index) => (
        <Highlight
          key={index}
          currentToken={{ text: token, tag: tags[index] }}
          nextToken={index < tokens.length - 1 ? { text: tokens[index + 1], tag: tags[index + 1] } : null}
          tagColors={tagColors}
          onMouseOver={handleMouseOver}
          onMouseDown={handleMouseDown}
          selecting={false}
          selected={false}
        />
      ))}
    </div>
  );
};

interface RecentSamplesProps {
  deploymentUrl: string;
}

export default function RecentSamples({ deploymentUrl }: RecentSamplesProps) {
  const { recentLabels, error } = useLabels({ deploymentUrl });
  const { recentSamples, error: sampleError } = useRecentSamples({ deploymentUrl });

  const recentUpvotes = useRollingSamples(
    /* samples= */ upvotes,
    /* numSamples= */ 5,
    /* maxNewSamples= */ 2,
    /* probabilityNewSamples= */ 0.2,
    /* intervalSeconds= */ 2
  );

  const recentAssociations = useRollingSamples(
    /* samples= */ associations,
    /* numSamples= */ 5,
    /* maxNewSamples= */ 2,
    /* probabilityNewSamples= */ 0.1,
    /* intervalSeconds= */ 3
  );

  const recentReformulations = useRollingSamples(
    /* samples= */ reformulations,
    /* numSamples= */ 3,
    /* maxNewSamples= */ 1,
    /* probabilityNewSamples= */ 0.4,
    /* intervalSeconds= */ 2
  );

  const predefinedColors = [
    { text: '#E5A49C', tag: '#D34F3E' },
    { text: '#F6C886', tag: '#F09336' },
    { text: '#FBE7AA', tag: '#F7CF5F' },
    { text: '#99E3B5', tag: '#5CC96E' },
    { text: '#A6E6E7', tag: '#65CFD0' },
    { text: '#A5A1E1', tag: '#597CE2' },
    { text: '#D8A4E2', tag: '#B64DC8' },
  ];

  const generateColor = (index: number): HighlightColor => {
    const hue = (index * 137.508) % 360; // Use golden angle approximation
    return {
      text: `hsl(${hue}, 70%, 85%)`,
      tag: `hsl(${hue}, 70%, 35%)`,
    };
  };

  const [tagColors, setTagColors] = useState<Record<string, HighlightColor>>({});
  const colorAssignmentsRef = useRef<Record<string, HighlightColor>>({});

  useEffect(() => {
    const updateTagColors = () => {
      const allTags = recentSamples.flatMap(sample => sample.tags);
      const uniqueTags = Array.from(new Set(allTags)).filter((tag) => tag !== 'O');
      const newColors: Record<string, HighlightColor> = {};

      uniqueTags.forEach((tag, index) => {
        if (colorAssignmentsRef.current[tag]) {
          newColors[tag] = colorAssignmentsRef.current[tag];
        } else if (index < predefinedColors.length) {
          newColors[tag] = predefinedColors[index];
        } else {
          newColors[tag] = generateColor(Object.keys(colorAssignmentsRef.current).length + index);
        }
        colorAssignmentsRef.current[tag] = newColors[tag];
      });

      setTagColors(newColors);
    };

    updateTagColors();
  }, [recentSamples]);

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'row',
        flexWrap: 'wrap',
        justifyContent: 'space-between',
        width: '100%',
      }}
    >
      <Card style={{ width: '32.5%', height: '45rem', marginBottom: '1rem' }}>
        <CardHeader>
          <CardTitle>Recent Labels</CardTitle>
          <CardDescription>The latest added labels</CardDescription>
        </CardHeader>
        <CardContent>
          {error && <div>Error fetching labels: {error.message}</div>}
          {recentLabels.map((label, idx) => (
            <div key={idx} className="text-md" style={{ marginBottom: '10px' }}>
              <span style={{ fontWeight: 'bold' }}>{label}</span>
            </div>
          ))}
        </CardContent>
      </Card>
      <Card style={{ width: '32.5%', height: '45rem', marginBottom: '1rem' }}>
        <CardHeader>
          <CardTitle>Recent Samples</CardTitle>
          <CardDescription>The latest inserted samples</CardDescription>
        </CardHeader>
        <CardContent>
          {sampleError && <div>Error fetching samples: {sampleError.message}</div>}
          {recentSamples.map((sample, idx) => (
            <HighlightedSample key={idx} tokens={sample.tokens} tags={sample.tags} tagColors={tagColors} />
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
