import React from 'react';
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

    return Array.from(allLabels).map(label => {
      // Handle potential NaN or undefined values
      const beforeP = beforeMetrics[label]?.precision || 0;
      const afterP = afterMetrics[label]?.precision || 0;
      const beforeR = beforeMetrics[label]?.recall || 0;
      const afterR = afterMetrics[label]?.recall || 0;
      const beforeF1 = beforeMetrics[label]?.fmeasure || 0;
      const afterF1 = afterMetrics[label]?.fmeasure || 0;

      return {
        label,
        'Before Precision': Number.isFinite(beforeP) ? beforeP * 100 : 0,
        'After Precision': Number.isFinite(afterP) ? afterP * 100 : 0,
        'Before Recall': Number.isFinite(beforeR) ? beforeR * 100 : 0,
        'After Recall': Number.isFinite(afterR) ? afterR * 100 : 0,
        'Before F1': Number.isFinite(beforeF1) ? beforeF1 * 100 : 0,
        'After F1': Number.isFinite(afterF1) ? afterF1 * 100 : 0,
      };
    });
  };

  const formatTooltip: TooltipFormatter = (value: string | number | Array<string | number>, name: string) => {
    if (typeof value === 'number') {
      return [`${value.toFixed(1)}%`, name];
    }
    return [value, name];
  };

  return (
    <ResponsiveContainer width="100%" height={400}>
      <BarChart
        data={prepareChartData()}
        margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="label" />
        <YAxis
          label={{ value: 'Score (%)', angle: -90, position: 'insideLeft' }}
          domain={[0, 100]}
        />
        <Tooltip formatter={formatTooltip} />
        <Legend />
        <Bar dataKey="Before Precision" fill="#8884d8" />
        <Bar dataKey="After Precision" fill="#82ca9d" />
        <Bar dataKey="Before Recall" fill="#ffc658" />
        <Bar dataKey="After Recall" fill="#ff7300" />
        <Bar dataKey="Before F1" fill="#a4de6c" />
        <Bar dataKey="After F1" fill="#de6ca4" />
      </BarChart>
    </ResponsiveContainer>
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
      <CardContent>
        <MetricsChart
          beforeMetrics={report.before_train_metrics}
          afterMetrics={report.after_train_metrics}
        />
        
        <div className="mt-8 space-y-6">
          <h3 className="text-xl font-semibold">Example Predictions</h3>
          
          {Object.entries(report.after_train_examples.true_positives).map(([label, examples]) => (
            <Card key={label} className="p-4">
              <h3 className="text-lg font-semibold mb-4">{label}</h3>
              <div className="space-y-4">
                <ExampleSection
                  title="True Positives"
                  examples={examples}
                  bgColor="bg-green-50"
                />
                <ExampleSection
                  title="False Positives"
                  examples={report.after_train_examples.false_positives[label] || []}
                  bgColor="bg-red-50"
                />
                <ExampleSection
                  title="False Negatives"
                  examples={report.after_train_examples.false_negatives[label] || []}
                  bgColor="bg-yellow-50"
                />
              </div>
            </Card>
          ))}
        </div>
      </CardContent>
    </Card>
  );
};