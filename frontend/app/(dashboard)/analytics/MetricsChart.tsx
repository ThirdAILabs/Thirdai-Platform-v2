import React, { useState } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  TooltipFormatter
} from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { LabelMetrics, TrainReportData, TrainingExample } from '@/lib/backend';
import { ExamplesVisualizer } from './ExamplesVisualizer'
import {PerformanceSummary} from './PerformanceSummary'

interface MetricsChartProps {
  beforeMetrics: LabelMetrics;
  afterMetrics: LabelMetrics;
}

const MetricsChart: React.FC<MetricsChartProps> = ({ beforeMetrics, afterMetrics }) => {
  // Get all unique labels
  const allLabels = Array.from(
    new Set([...Object.keys(beforeMetrics), ...Object.keys(afterMetrics)])
  );
  
  // State for selected label
  const [selectedLabel, setSelectedLabel] = useState(allLabels[0]);

  const prepareChartData = (label: string) => {
    return [
      {
        name: 'Precision',
        'Before Training': Number.isFinite(beforeMetrics[label]?.precision) 
          ? beforeMetrics[label].precision * 100 
          : 0,
        'After Training': Number.isFinite(afterMetrics[label]?.precision)
          ? afterMetrics[label].precision * 100
          : 0,
      },
      {
        name: 'Recall',
        'Before Training': Number.isFinite(beforeMetrics[label]?.recall)
          ? beforeMetrics[label].recall * 100
          : 0,
        'After Training': Number.isFinite(afterMetrics[label]?.recall)
          ? afterMetrics[label].recall * 100
          : 0,
      },
      {
        name: 'F1',
        'Before Training': Number.isFinite(beforeMetrics[label]?.fmeasure)
          ? beforeMetrics[label].fmeasure * 100
          : 0,
        'After Training': Number.isFinite(afterMetrics[label]?.fmeasure)
          ? afterMetrics[label].fmeasure * 100
          : 0,
      },
    ];
  };

  const formatTooltip: TooltipFormatter = (value: string | number | Array<string | number>) => {
    if (typeof value === 'number') {
      return `${value.toFixed(1)}%`;
    }
    return value;
  };

  const getMetricDifference = (label: string, metricKey: 'precision' | 'recall' | 'fmeasure') => {
    const before = beforeMetrics[label]?.[metricKey] || 0;
    const after = afterMetrics[label]?.[metricKey] || 0;
    return ((after - before) * 100).toFixed(1);
  };

  return (
    <div className="space-y-6">
      {/* Label Selection */}
      <div className="flex flex-wrap gap-2">
        {allLabels.map((label) => (
          <button
            key={label}
            onClick={() => setSelectedLabel(label)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors
              ${selectedLabel === label
                ? 'bg-blue-100 text-blue-800'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Metrics Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Metrics for "{selectedLabel}"</CardTitle>
          <CardDescription>Comparing performance before and after training</CardDescription>
        </CardHeader>
        <CardContent>
          {/* Chart */}
          <ResponsiveContainer width="100%" height={300}>
            <BarChart
              data={prepareChartData(selectedLabel)}
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
              barSize={40}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis
                domain={[0, 100]}
                tickFormatter={(value) => `${value}%`}
              />
              <Tooltip formatter={formatTooltip} />
              <Legend />
              <Bar 
                dataKey="Before Training" 
                fill="#94a3b8"
                radius={[4, 4, 0, 0]}
              />
              <Bar 
                dataKey="After Training" 
                fill="#3b82f6"
                radius={[4, 4, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>

          {/* Metrics Summary */}
          <div className="grid grid-cols-3 gap-4 mt-6">
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="text-sm font-medium text-gray-500">Precision</div>
              <div className="mt-2 flex items-baseline justify-between">
                <span className="text-2xl font-semibold">
                  {(afterMetrics[selectedLabel]?.precision * 100).toFixed(1)}%
                </span>
                <span className={`text-sm font-medium ${
                  Number(getMetricDifference(selectedLabel, 'precision')) >= 0 
                    ? 'text-green-600' 
                    : 'text-red-600'
                }`}>
                  {Number(getMetricDifference(selectedLabel, 'precision')) >= 0 ? '+' : ''}
                  {getMetricDifference(selectedLabel, 'precision')}%
                </span>
              </div>
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="text-sm font-medium text-gray-500">Recall</div>
              <div className="mt-2 flex items-baseline justify-between">
                <span className="text-2xl font-semibold">
                  {(afterMetrics[selectedLabel]?.recall * 100).toFixed(1)}%
                </span>
                <span className={`text-sm font-medium ${
                  Number(getMetricDifference(selectedLabel, 'recall')) >= 0 
                    ? 'text-green-600' 
                    : 'text-red-600'
                }`}>
                  {Number(getMetricDifference(selectedLabel, 'recall')) >= 0 ? '+' : ''}
                  {getMetricDifference(selectedLabel, 'recall')}%
                </span>
              </div>
            </div>
            <div className="p-4 bg-gray-50 rounded-lg">
              <div className="text-sm font-medium text-gray-500">F1 Score</div>
              <div className="mt-2 flex items-baseline justify-between">
                <span className="text-2xl font-semibold">
                  {(afterMetrics[selectedLabel]?.fmeasure * 100).toFixed(1)}%
                </span>
                <span className={`text-sm font-medium ${
                  Number(getMetricDifference(selectedLabel, 'fmeasure')) >= 0 
                    ? 'text-green-600' 
                    : 'text-red-600'
                }`}>
                  {Number(getMetricDifference(selectedLabel, 'fmeasure')) >= 0 ? '+' : ''}
                  {getMetricDifference(selectedLabel, 'fmeasure')}%
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

interface ExampleSectionProps {
  title: string;
  examples: TrainingExample[];
  bgColor?: string;
}

const ExampleSection: React.FC<ExampleSectionProps> = ({ title, examples, bgColor = 'bg-gray-50' }) => (
  <div className="space-y-2">
    <h4 className="font-medium">{title}</h4>
    <div className="space-y-1">
      {examples.slice(0, 2).map((example, idx) => (
        <div key={idx} className={`text-sm p-3 ${bgColor} rounded-lg`}>
          <div className="text-gray-700">
            <span className="font-medium">Input:</span> {example.source}
          </div>
          <div className="text-gray-700">
            <span className="font-medium">Prediction:</span> {example.target}
          </div>
        </div>
      ))}
    </div>
  </div>
);

interface TrainingResultsProps {
  report: TrainReportData;
}

export const TrainingResults: React.FC<TrainingResultsProps> = ({ report }) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Training Results Comparison</CardTitle>
        <CardDescription>
          Model performance metrics and example predictions before and after training
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-8">
        <PerformanceSummary 
          beforeMetrics={report.before_train_metrics}
          afterMetrics={report.after_train_metrics}
        />

        <MetricsChart
          beforeMetrics={report.before_train_metrics}
          afterMetrics={report.after_train_metrics}
        />
        <ExamplesVisualizer report={report} />
      </CardContent>
    </Card>
  );
};