import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tag, Clock, Database, Zap } from 'lucide-react';
import { useLabels } from '@/lib/backend';

interface PerformanceData {
  avg_time_per_sample: number;
  avg_time_per_token: number;
  throughput: number;
  total_time: number;
  total_tokens: number;
  total_samples: number;
}

interface LatencyMetricsProps {
  deploymentUrl: string;
  performanceData?: PerformanceData;
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

  // If still loading or no performance data, return null
  if (isLoadingLabels || !performanceData) {
    return null;
  }

  // Performance data for different model sizes
  const modelData = {
    5: {
      avg_time_per_sample: 0.00140074630305171,
      avg_time_per_token: 6.0522040020554004e-05,
      throughput: 16522.90636040008,
      total_time: 14.007463030517101,
      total_tokens: 231444,
      total_samples: 10000
    },
    10: {
      avg_time_per_sample: 0.0009167316736653447,
      avg_time_per_token: 3.946547476463791e-05,
      throughput: 25338.60306923321,
      total_time: 9.167316736653447,
      total_tokens: 232287,
      total_samples: 10000
    },
    16: {
      avg_time_per_sample: 0.0009164524495601654,
      avg_time_per_token: 3.9399177563880944e-05,
      throughput: 25381.240468246386,
      total_time: 9.164524495601654,
      total_tokens: 232607,
      total_samples: 10000
    }
  };

  // Get the correct performance data based on number of labels
  const relevantPerformanceData = modelData[numLabels as keyof typeof modelData];
  
  // If we don't have performance data for this number of labels, return null
  if (!relevantPerformanceData) {
    return null;
  }

  const formatNumber = (num: number): string => {
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`;
    }
    return num.toString();
  };

  // Convert seconds to milliseconds and round for per-token latency
  const avgLatencyPerTokenMs = (relevantPerformanceData.avg_time_per_token * 1000).toFixed(2);
  const throughputFormatted = Math.round(relevantPerformanceData.throughput).toLocaleString();
  const totalSamplesFormatted = formatNumber(relevantPerformanceData.total_samples);

  return (
    <div className="container mx-auto px-4">
      <Card>
        <CardHeader>
          <CardTitle>Model Performance Metrics</CardTitle>
          <CardDescription>
            Performance metrics for {numLabels}-tag NER model
          </CardDescription>
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
