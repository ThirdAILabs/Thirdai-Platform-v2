'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

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

interface TokenHighlightProps {
  text: string;
  index: number;
  highlightIndex: number;
  type: 'tp' | 'fp' | 'fn';
}

const TokenHighlight: React.FC<TokenHighlightProps> = ({ text, index, highlightIndex, type }) => {
  const isHighlighted = index === highlightIndex;

  const getHighlightColor = () => {
    if (!isHighlighted) return 'bg-transparent';
    switch (type) {
      case 'tp':
        return 'bg-green-100 border-green-400';
      case 'fp':
        return 'bg-red-100 border-red-400';
      case 'fn':
        return 'bg-yellow-100 border-yellow-400';
      default:
        return 'bg-transparent';
    }
  };

  return (
    <span
      className={`px-1 py-0.5 rounded ${getHighlightColor()} ${isHighlighted ? 'border-2' : ''}`}
    >
      {text}
    </span>
  );
};

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
        return 'text-green-700 bg-green-50';
      case 'fp':
        return 'text-red-700 bg-red-50';
      case 'fn':
        return 'text-yellow-700 bg-yellow-50';
    }
  };

  return (
    <div className="border rounded-lg p-4 space-y-3">
      <div className={`inline-block px-2 py-1 rounded-full text-xs font-medium ${getTypeColor()}`}>
        {getTypeLabel()}
      </div>

      <div className="space-y-2">
        <div className="space-y-1">
          <div className="text-sm font-medium text-gray-500">Input</div>
          <div className="p-2 bg-gray-50 rounded">
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
          </div>
        </div>

        <div className="space-y-1">
          <div className="text-sm font-medium text-gray-500">Ground Truth</div>
          <div className="p-2 bg-gray-50 rounded">
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
          </div>
        </div>

        <div className="space-y-1">
          <div className="text-sm font-medium text-gray-500">Prediction</div>
          <div className="p-2 bg-gray-50 rounded">
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
          </div>
        </div>
      </div>
    </div>
  );
};

interface ExamplesVisualizerProps {
  report?: TrainReportData;
}

const ExamplesVisualizer: React.FC<ExamplesVisualizerProps> = ({ report = mockReport }) => {
  const allLabels = Object.keys(report.after_train_examples.true_positives);
  const [selectedLabel, setSelectedLabel] = useState(allLabels[0]);
  const [selectedType, setSelectedType] = useState<'tp' | 'fp' | 'fn'>('tp');

  const predictionTypes = [
    { id: 'tp', label: 'True Positives', color: 'bg-green-100 hover:bg-green-200' },
    { id: 'fp', label: 'False Positives', color: 'bg-red-100 hover:bg-red-200' },
    { id: 'fn', label: 'False Negatives', color: 'bg-yellow-100 hover:bg-yellow-200' },
  ] as const;

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

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sample Predictions</CardTitle>
        <CardDescription>Analyze model predictions with token-level details</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Label Selection */}
        <div className="space-y-2">
          <div className="text-sm font-medium text-gray-500">Select Label</div>
          <div className="flex flex-wrap gap-2">
            {allLabels.map((label) => (
              <button
                key={label}
                onClick={() => setSelectedLabel(label)}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors
                  ${
                    selectedLabel === label
                      ? 'bg-blue-100 text-blue-800'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  }`}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        {/* Prediction Type Selection */}
        <div className="space-y-2">
          <div className="text-sm font-medium text-gray-500">Select Prediction Type</div>
          <div className="flex flex-wrap gap-2">
            {predictionTypes.map(({ id, label, color }) => (
              <button
                key={id}
                onClick={() => setSelectedType(id)}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors
                  ${selectedType === id ? color : 'bg-gray-100 hover:bg-gray-200'}`}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        {/* Examples */}
        <div className="space-y-4">
          {getExamples().map((example, idx) => (
            <ExamplePair key={idx} example={example} type={selectedType} />
          ))}
          {getExamples().length === 0 && (
            <div className="text-center py-8 text-gray-500">
              No examples found for this combination
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};

export default ExamplesVisualizer; 