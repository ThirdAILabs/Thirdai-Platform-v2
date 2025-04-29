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
} from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

// Mock data structure to match frontend version
interface LabelMetrics {
  [key: string]: {
    precision: number;
    recall: number;
    fmeasure: number;
  };
}

interface TrainReportData {
  before_train_metrics: LabelMetrics;
  after_train_metrics: LabelMetrics;
  // These would normally include training examples too
}

// Sample mock data
const mockReportData: TrainReportData = {
  before_train_metrics: {
    'O': { precision: 0.90, recall: 0.88, fmeasure: 0.89 },
    'NAME': { precision: 0.83, recall: 0.79, fmeasure: 0.81 },
    'PHONE': { precision: 0.86, recall: 0.82, fmeasure: 0.84 },
    'EMAIL': { precision: 0.90, recall: 0.86, fmeasure: 0.88 },
    'ADDRESS': { precision: 0.81, recall: 0.77, fmeasure: 0.79 }
  },
  after_train_metrics: {
    'O': { precision: 0.95, recall: 0.92, fmeasure: 0.93 },
    'NAME': { precision: 0.88, recall: 0.84, fmeasure: 0.86 },
    'PHONE': { precision: 0.92, recall: 0.88, fmeasure: 0.90 },
    'EMAIL': { precision: 0.96, recall: 0.92, fmeasure: 0.94 },
    'ADDRESS': { precision: 0.89, recall: 0.85, fmeasure: 0.87 }
  }
};

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
          : null,
        'After Training': Number.isFinite(afterMetrics[label]?.precision)
          ? afterMetrics[label].precision * 100
          : null,
      },
      {
        name: 'Recall',
        'Before Training': Number.isFinite(beforeMetrics[label]?.recall)
          ? beforeMetrics[label].recall * 100
          : null,
        'After Training': Number.isFinite(afterMetrics[label]?.recall)
          ? afterMetrics[label].recall * 100
          : null,
      },
      {
        name: 'F1',
        'Before Training': Number.isFinite(beforeMetrics[label]?.fmeasure)
          ? beforeMetrics[label].fmeasure * 100
          : null,
        'After Training': Number.isFinite(afterMetrics[label]?.fmeasure)
          ? afterMetrics[label].fmeasure * 100
          : null,
      },
    ];
  };

  const formatTooltip = (value: string | number | Array<string | number>) => {
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

      {/* Metrics Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Metrics for &quot;{selectedLabel}&quot;</CardTitle>
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
              <YAxis domain={[0, 100]} tickFormatter={(value) => `${value}%`} />
              <Tooltip formatter={formatTooltip} />
              <Legend />
              <Bar dataKey="Before Training" fill="#94a3b8" radius={[4, 4, 0, 0]} />
              <Bar dataKey="After Training" fill="#3b82f6" radius={[4, 4, 0, 0]} />
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
                <span
                  className={`text-sm font-medium ${
                    Number(getMetricDifference(selectedLabel, 'precision')) >= 0
                      ? 'text-green-600'
                      : 'text-red-600'
                  }`}
                >
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
                <span
                  className={`text-sm font-medium ${
                    Number(getMetricDifference(selectedLabel, 'recall')) >= 0
                      ? 'text-green-600'
                      : 'text-red-600'
                  }`}
                >
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
                <span
                  className={`text-sm font-medium ${
                    Number(getMetricDifference(selectedLabel, 'fmeasure')) >= 0
                      ? 'text-green-600'
                      : 'text-red-600'
                  }`}
                >
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

// Performance Summary Component
interface PerformanceSummaryProps {
  beforeMetrics: LabelMetrics;
  afterMetrics: LabelMetrics;
}

const PerformanceSummary: React.FC<PerformanceSummaryProps> = ({ beforeMetrics, afterMetrics }) => {
  // Calculate overall averages
  const calculateAverage = (metrics: LabelMetrics, key: 'precision' | 'recall' | 'fmeasure') => {
    const values = Object.values(metrics).map(m => m[key]);
    return values.reduce((sum, val) => sum + val, 0) / values.length;
  };

  const labels = Array.from(
    new Set([...Object.keys(beforeMetrics), ...Object.keys(afterMetrics)])
  );

  const tableData = labels.map(label => ({
    label,
    beforeF1: beforeMetrics[label]?.fmeasure || 0,
    afterF1: afterMetrics[label]?.fmeasure || 0,
    change: (afterMetrics[label]?.fmeasure || 0) - (beforeMetrics[label]?.fmeasure || 0)
  }));

  // Add overall average row
  const beforeAvgF1 = calculateAverage(beforeMetrics, 'fmeasure');
  const afterAvgF1 = calculateAverage(afterMetrics, 'fmeasure');
  
  tableData.push({
    label: 'Overall Average',
    beforeF1: beforeAvgF1,
    afterF1: afterAvgF1,
    change: afterAvgF1 - beforeAvgF1
  });

  return (
    <div className="mb-8">
      <h3 className="text-lg font-semibold mb-2">Performance Summary</h3>
      <p className="text-sm text-gray-500 mb-4">Overview of F1 score changes across all labels</p>
      
      <div className="overflow-hidden border rounded-md">
        <div className="bg-gray-50 grid grid-cols-4 p-3 border-b text-sm font-medium">
          <div>TAG</div>
          <div>F1 before fine-tuning</div>
          <div>F1 after fine-tuning</div>
          <div>Change</div>
        </div>
        
        <div className="divide-y">
          {tableData.map((row, idx) => (
            <div 
              key={row.label}
              className={`grid grid-cols-4 p-3 text-sm ${
                row.label === 'Overall Average' ? 'bg-gray-50 font-medium' : ''
              }`}
            >
              <div>{row.label}</div>
              <div>{(row.beforeF1 * 100).toFixed(1)}%</div>
              <div>{(row.afterF1 * 100).toFixed(1)}%</div>
              <div className="text-green-600">
                ↑ +{(row.change * 100).toFixed(1)}%
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

interface TrainingResultsProps {
  report?: TrainReportData;
}

const TrainingResults: React.FC<TrainingResultsProps> = ({ report = mockReportData }) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Fine-tuning Metrics Comparison</CardTitle>
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
      </CardContent>
    </Card>
  );
};

export default TrainingResults; 