'use client';

import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Paper,
  Divider,
  Chip
} from '@mui/material';

// Types to match the original
interface TrainingExample {
  source: string;
  target: string;
  predictions: string;
  index: number;
}

interface TrainReportData {
  after_train_examples: {
    true_positives: { [key: string]: TrainingExample[] };
    false_positives: { [key: string]: TrainingExample[] };
    false_negatives: { [key: string]: TrainingExample[] };
  };
}

// Mock data for display
const mockReport: TrainReportData = {
  after_train_examples: {
    true_positives: {
      'O': [
        {
          source: 'Contact John at john@example.com for more information about the project.',
          target: 'O O NAME O EMAIL O O O O O O O',
          predictions: 'O O NAME O EMAIL O O O O O O O',
          index: 2
        },
        {
          source: 'Please send your documents to 555 Main Street, New York, NY 10001.',
          target: 'O O O O O ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS',
          predictions: 'O O O O O ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS',
          index: 5
        }
      ],
      'NAME': [
        {
          source: 'Sarah Johnson will be presenting at the conference tomorrow.',
          target: 'NAME NAME O O O O O O O',
          predictions: 'NAME NAME O O O O O O O',
          index: 0
        }
      ],
      'PHONE': [
        {
          source: 'You can reach customer service at (800) 555-1234 from 9am to 5pm.',
          target: 'O O O O O O O PHONE PHONE PHONE O O O O',
          predictions: 'O O O O O O O PHONE PHONE PHONE O O O O',
          index: 7
        }
      ],
      'EMAIL': [
        {
          source: 'For technical support, email support@company.com with your query.',
          target: 'O O O O EMAIL O O O',
          predictions: 'O O O O EMAIL O O O',
          index: 4
        }
      ],
      'ADDRESS': [
        {
          source: 'The new office is located at 123 Business Ave, Suite 400, Chicago, IL.',
          target: 'O O O O O O ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS',
          predictions: 'O O O O O O ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS ADDRESS',
          index: 6
        }
      ]
    },
    false_positives: {
      'O': [
        {
          source: 'Contact Alex from the marketing team for updates.',
          target: 'O NAME O O O O O O',
          predictions: 'O O O O O O O O',
          index: 1
        }
      ],
      'NAME': [
        {
          source: 'The company headquarters is in San Francisco.',
          target: 'O O O O O ADDRESS',
          predictions: 'O O O O O NAME',
          index: 5
        }
      ],
      'EMAIL': [
        {
          source: 'Send a message to the team chat.',
          target: 'O O O O O O O',
          predictions: 'O O O O O EMAIL O',
          index: 5
        }
      ]
    },
    false_negatives: {
      'NAME': [
        {
          source: 'Michael Brown is the new project manager starting next month.',
          target: 'NAME NAME O O O O O O O O',
          predictions: 'O O O O O O O O O O',
          index: 0
        }
      ],
      'PHONE': [
        {
          source: 'Call me at 212-555-6789 when you arrive.',
          target: 'O O O PHONE O O O',
          predictions: 'O O O O O O O',
          index: 3
        }
      ],
      'ADDRESS': [
        {
          source: 'The meeting will be at 350 Fifth Avenue, New York.',
          target: 'O O O O O ADDRESS ADDRESS ADDRESS',
          predictions: 'O O O O O O O O',
          index: 5
        }
      ]
    }
  }
};

// Token highlight component similar to the original
interface TokenHighlightProps {
  text: string;
  index: number;
  highlightIndex: number;
  type: 'tp' | 'fp' | 'fn';
}

const TokenHighlight: React.FC<TokenHighlightProps> = ({ text, index, highlightIndex, type }) => {
  const isHighlighted = index === highlightIndex;

  const getHighlightColor = () => {
    if (!isHighlighted) return {};
    
    switch (type) {
      case 'tp':
        return { 
          backgroundColor: 'rgba(76, 175, 80, 0.1)', 
          border: '2px solid rgba(76, 175, 80, 0.4)' 
        };
      case 'fp':
        return { 
          backgroundColor: 'rgba(244, 67, 54, 0.1)', 
          border: '2px solid rgba(244, 67, 54, 0.4)' 
        };
      case 'fn':
        return { 
          backgroundColor: 'rgba(255, 152, 0, 0.1)', 
          border: '2px solid rgba(255, 152, 0, 0.4)' 
        };
      default:
        return {};
    }
  };

  return (
    <Box
      component="span"
      sx={{
        padding: '2px 4px',
        borderRadius: '4px',
        ...getHighlightColor()
      }}
    >
      {text}
    </Box>
  );
};

// Example pair component similar to the original
interface ExamplePairProps {
  example: TrainingExample;
  type: 'tp' | 'fp' | 'fn';
}

const ExamplePair: React.FC<ExamplePairProps> = ({ example, type }) => {
  const sourceTokens = example.source.split(' ');
  const targetTokens = example.target.split(' ');
  const predictionTokens = example.predictions.split(' ');

  const getTypeLabel = () => {
    switch (type) {
      case 'tp':
        return 'True Positive';
      case 'fp':
        return 'False Positive';
      case 'fn':
        return 'False Negative';
    }
  };

  const getTypeColor = () => {
    switch (type) {
      case 'tp':
        return { 
          color: '#2e7d32', 
          backgroundColor: '#e8f5e9' 
        };
      case 'fp':
        return { 
          color: '#c62828', 
          backgroundColor: '#ffebee' 
        };
      case 'fn':
        return { 
          color: '#ef6c00', 
          backgroundColor: '#fff3e0' 
        };
    }
  };

  return (
    <Paper 
      variant="outlined" 
      sx={{ 
        p: 3, 
        mb: 2,
        borderRadius: '8px'
      }}
    >
      <Chip
        label={getTypeLabel()}
        size="small"
        sx={{
          mb: 2,
          fontWeight: 500,
          ...getTypeColor(),
          borderRadius: '16px'
        }}
      />

      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        <Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontWeight: 500 }}>
            Input
          </Typography>
          <Paper 
            variant="outlined" 
            sx={{ 
              p: 1.5, 
              backgroundColor: 'rgba(0, 0, 0, 0.02)',
              borderColor: 'rgba(0, 0, 0, 0.08)'
            }}
          >
            {sourceTokens.map((token, idx) => (
              <React.Fragment key={idx}>
                <TokenHighlight
                  text={token}
                  index={idx}
                  highlightIndex={example.index}
                  type={type}
                />
                {idx < sourceTokens.length - 1 && ' '}
              </React.Fragment>
            ))}
          </Paper>
        </Box>

        <Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontWeight: 500 }}>
            Ground Truth
          </Typography>
          <Paper 
            variant="outlined" 
            sx={{ 
              p: 1.5, 
              backgroundColor: 'rgba(0, 0, 0, 0.02)',
              borderColor: 'rgba(0, 0, 0, 0.08)'
            }}
          >
            {targetTokens.map((token, idx) => (
              <React.Fragment key={idx}>
                <TokenHighlight
                  text={token}
                  index={idx}
                  highlightIndex={example.index}
                  type={type}
                />
                {idx < targetTokens.length - 1 && ' '}
              </React.Fragment>
            ))}
          </Paper>
        </Box>

        <Box>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontWeight: 500 }}>
            Prediction
          </Typography>
          <Paper 
            variant="outlined" 
            sx={{ 
              p: 1.5, 
              backgroundColor: 'rgba(0, 0, 0, 0.02)',
              borderColor: 'rgba(0, 0, 0, 0.08)'
            }}
          >
            {predictionTokens.map((token, idx) => (
              <React.Fragment key={idx}>
                <TokenHighlight
                  text={token}
                  index={idx}
                  highlightIndex={example.index}
                  type={type}
                />
                {idx < predictionTokens.length - 1 && ' '}
              </React.Fragment>
            ))}
          </Paper>
        </Box>
      </Box>
    </Paper>
  );
};

// Main ExamplesVisualizer component
export const ExamplesVisualizer: React.FC = () => {
  const report = mockReport; // Use mock data for display
  const allLabels = Object.keys(report.after_train_examples.true_positives);
  const [selectedLabel, setSelectedLabel] = useState(allLabels[0]);
  const [selectedType, setSelectedType] = useState<'tp' | 'fp' | 'fn'>('tp');

  const predictionTypes = [
    { id: 'tp', label: 'True Positives', color: { light: '#e8f5e9', hover: '#c8e6c9' }},
    { id: 'fp', label: 'False Positives', color: { light: '#ffebee', hover: '#ffcdd2' }},
    { id: 'fn', label: 'False Negatives', color: { light: '#fff3e0', hover: '#ffe0b2' }},
  ] as const;

  const handleLabelChange = (label: string) => {
    setSelectedLabel(label);
  };

  const handleTypeChange = (type: 'tp' | 'fp' | 'fn') => {
    setSelectedType(type);
  };

  const getExamples = () => {
    switch (selectedType) {
      case 'tp':
        return report.after_train_examples.true_positives[selectedLabel] || [];
      case 'fp':
        return report.after_train_examples.false_positives[selectedLabel] || [];
      case 'fn':
        return report.after_train_examples.false_negatives[selectedLabel] || [];
    }
  };

  const examples = getExamples();

  return (
    <Card sx={{ mb: 4, backgroundColor: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.12)' }}>
      <CardContent>
        <Typography variant="h6" sx={{ fontWeight: 500, mb: 1 }}>
          Sample Predictions
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Analyze model predictions with token-level details
        </Typography>

        {/* Label Selection */}
        <Box sx={{ mb: 3 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1, fontWeight: 500 }}>
            Select Label
          </Typography>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
            {allLabels.map((label) => (
              <Button
                key={label}
                variant={selectedLabel === label ? 'contained' : 'outlined'}
                size="small"
                onClick={() => handleLabelChange(label)}
                sx={{
                  textTransform: 'none',
                  backgroundColor: selectedLabel === label ? '#bbdefb' : '#f5f5f5',
                  color: selectedLabel === label ? '#0d47a1' : '#616161',
                  border: selectedLabel === label ? '1px solid #90caf9' : '1px solid #e0e0e0',
                  '&:hover': {
                    backgroundColor: selectedLabel === label ? '#90caf9' : '#e0e0e0',
                  },
                  boxShadow: 'none',
                }}
              >
                {label}
              </Button>
            ))}
          </Box>
        </Box>

        {/* Prediction Type Selection */}
        <Box sx={{ mb: 3 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1, fontWeight: 500 }}>
            Select Prediction Type
          </Typography>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
            {predictionTypes.map(({ id, label, color }) => (
              <Button
                key={id}
                variant={selectedType === id ? 'contained' : 'outlined'}
                size="small"
                onClick={() => handleTypeChange(id)}
                sx={{
                  textTransform: 'none',
                  backgroundColor: selectedType === id ? color.light : '#f5f5f5',
                  color: selectedType === id ? 
                    (id === 'tp' ? '#2e7d32' : id === 'fp' ? '#c62828' : '#ef6c00') : 
                    '#616161',
                  border: selectedType === id ? 
                    `1px solid ${id === 'tp' ? '#a5d6a7' : id === 'fp' ? '#ef9a9a' : '#ffcc80'}` : 
                    '1px solid #e0e0e0',
                  '&:hover': {
                    backgroundColor: selectedType === id ? color.hover : '#e0e0e0',
                  },
                  boxShadow: 'none',
                }}
              >
                {label}
              </Button>
            ))}
          </Box>
        </Box>

        {/* Examples display */}
        <Box sx={{ mt: 3 }}>
          {examples.length > 0 ? (
            examples.map((example, idx) => (
              <ExamplePair key={idx} example={example} type={selectedType} />
            ))
          ) : (
            <Box 
              sx={{ 
                textAlign: 'center', 
                py: 4, 
                color: 'text.secondary',
                border: '1px dashed #ccc',
                borderRadius: '8px'
              }}
            >
              No examples found for this combination
            </Box>
          )}
        </Box>
      </CardContent>
    </Card>
  );
};

export default ExamplesVisualizer; 