import React from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { LabelMetrics, TrainReportResponse } from '@/lib/backend';

interface MetricsChartProps {
  beforeMetrics: LabelMetrics;
  afterMetrics: LabelMetrics;
}

const MetricsChart: React.FC<MetricsChartProps> = ({ beforeMetrics, afterMetrics }) => {
  const prepareChartData = () => {
    const allLabels = new Set([
      ...Object.keys(beforeMetrics),
      ...Object.keys(afterMetrics)
    ]);

    return Array.from(allLabels).map(label => ({
      label,
      'Before P': beforeMetrics[label]?.precision * 100 || 0,
      'After P': afterMetrics[label]?.precision * 100 || 0,
      'Before R': beforeMetrics[label]?.recall * 100 || 0,
      'After R': afterMetrics[label]?.recall * 100 || 0,
      'Before F1': beforeMetrics[label]?.fmeasure * 100 || 0,
      'After F1': afterMetrics[label]?.fmeasure * 100 || 0,
    }));
  };

  return (
    <ResponsiveContainer width="100%" height={400}>
      <BarChart
        data={prepareChartData()}
        margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="label" />
        <YAxis label={{ value: 'Score (%)', angle: -90, position: 'insideLeft' }} />
        <Tooltip />
        <Legend />
        <Bar dataKey="Before P" fill="#8884d8" name="Precision (Before)" />
        <Bar dataKey="After P" fill="#82ca9d" name="Precision (After)" />
        <Bar dataKey="Before R" fill="#ffc658" name="Recall (Before)" />
        <Bar dataKey="After R" fill="#ff7300" name="Recall (After)" />
        <Bar dataKey="Before F1" fill="#a4de6c" name="F1 (Before)" />
        <Bar dataKey="After F1" fill="#de6ca4" name="F1 (After)" />
      </BarChart>
    </ResponsiveContainer>
  );
};

interface TrainingResultsProps {
  report: TrainReportResponse;
}

export const TrainingResults: React.FC<TrainingResultsProps> = ({ report }) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Training Results Comparison</CardTitle>
        <CardDescription>
          Compare model performance metrics before and after training
        </CardDescription>
      </CardHeader>
      <CardContent>
        <MetricsChart
          beforeMetrics={report.before_train_metrics}
          afterMetrics={report.after_train_metrics}
        />
        
        {/* Examples Section */}
        <div className="mt-6 space-y-4">
          <h3 className="text-lg font-semibold">Example Predictions</h3>
          
          {Object.entries(report.after_train_examples.true_positives).map(([label, examples]) => (
            <div key={label} className="space-y-2">
              <h4 className="font-medium">{label} - True Positives</h4>
              <div className="space-y-1">
                {examples.slice(0, 3).map((example, idx) => (
                  <div key={idx} className="text-sm p-2 bg-gray-50 rounded">
                    <div className="text-gray-700">Input: {example.source}</div>
                    <div className="text-gray-700">Prediction: {example.target}</div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
};