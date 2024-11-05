import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { TrainReportData, TrainingExample } from '@/lib/backend';

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
  report: TrainReportData;
}

export const ExamplesVisualizer: React.FC<ExamplesVisualizerProps> = ({ report }) => {
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
