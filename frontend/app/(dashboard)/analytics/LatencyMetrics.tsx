import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tag, Clock, Database, Zap } from 'lucide-react';
import { useLabels } from '@/lib/backend';

interface LatencyMetricsProps {
  deploymentUrl: string;
  performanceData?: {
    avg_time_per_sample: number;
    avg_time_per_token: number;
    throughput: number;
    total_time: number;
    total_tokens: number;
    total_samples: number;
  };
}

const LatencyMetrics = ({ deploymentUrl, performanceData }: LatencyMetricsProps) => {
  const {
    recentLabels,
    error: labelError,
    isLoading: isLoadingLabels,
  } = useLabels({ deploymentUrl });

  // Create a set of unique labels and convert it back to an array
  const uniqueLabels = Array.from(new Set(recentLabels));
  const numLabels = uniqueLabels.length;

  // Only show metrics if we have 10 labels and performance data
  if (numLabels !== 10 || !performanceData) {
    return null;
  }

  const formatNumber = (num: number): string => {
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`;
    }
    return num.toString();
  };

  // Convert seconds to milliseconds and round for per-token latency
  const avgLatencyPerTokenMs = (performanceData.avg_time_per_token * 1000).toFixed(2);
  const throughputFormatted = Math.round(performanceData.throughput).toLocaleString();
  const totalSamplesFormatted = formatNumber(performanceData.total_samples);

  return (
    <div className="container mx-auto px-4">
      <Card>
        <CardHeader>
          <CardTitle>Model Performance Metrics</CardTitle>
          <CardDescription>Real-time latency metrics based on production inference</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-4 gap-6">
            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Tag className="h-5 w-5" />
                <span className="text-sm font-medium">Entity Types</span>
              </div>
              <div className="text-3xl font-semibold">{numLabels}</div>
              <div className="text-sm text-gray-500 mt-1">distinct entities recognized</div>
            </div>

            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Clock className="h-5 w-5" />
                <span className="text-sm font-medium">Average Latency</span>
              </div>
              <div className="text-3xl font-semibold">{avgLatencyPerTokenMs}ms</div>
              <div className="text-sm text-gray-500 mt-1">per token processed</div>
            </div>

            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Zap className="h-5 w-5" />
                <span className="text-sm font-medium">Throughput</span>
              </div>
              <div className="text-3xl font-semibold">{throughputFormatted}</div>
              <div className="text-sm text-gray-500 mt-1">tokens per second</div>
            </div>

            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Database className="h-5 w-5" />
                <span className="text-sm font-medium">Sample Size</span>
              </div>
              <div className="text-3xl font-semibold">{totalSamplesFormatted}</div>
              <div className="text-sm text-gray-500 mt-1">inference requests measured</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default LatencyMetrics;
