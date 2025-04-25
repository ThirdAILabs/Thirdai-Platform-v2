import React from 'react';
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface TokenCount {
  type: string;
  count: number;
}

interface LatencyDataPoint {
  timestamp: string;
  latency: number;
}

interface AnalyticsDashboardProps {
  progress: number;
  tokensProcessed: number;
  latencyData: LatencyDataPoint[];
  tokenTypes: string[];
  tokenCounts: Record<string, number>;
}

const formatNumber = (num: number): string => {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(2)}M`;
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(1)}K`;
  }
  return num.toString();
};

export const AnalyticsDashboard: React.FC<AnalyticsDashboardProps> = ({
  progress,
  tokensProcessed,
  latencyData,
  tokenTypes,
  tokenCounts,
}) => {
  // Convert token counts to chart data format
  const tokenChartData = Object.entries(tokenCounts).map(([type, count]) => ({
    type,
    count,
  }));

  // Calculate min and max latency for the y-axis domain
  const latencies = latencyData.map(d => d.latency);
  const minLatency = Math.min(...latencies);
  const maxLatency = Math.max(...latencies);
  const latencyPadding = (maxLatency - minLatency) * 0.1; // Add 10% padding

  return (
    <div className="space-y-6 w-full">
      {/* Top Widgets */}
      <div className="grid grid-cols-3 gap-4">
        {/* Progress Widget */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center justify-center">
              <div className="relative h-32 w-32">
                <svg className="h-full w-full" viewBox="0 0 100 100">
                  {/* Background circle */}
                  <circle
                    className="stroke-muted"
                    cx="50"
                    cy="50"
                    r="40"
                    fill="none"
                    strokeWidth="10"
                  />
                  {/* Progress circle */}
                  <circle
                    className="stroke-primary"
                    cx="50"
                    cy="50"
                    r="40"
                    fill="none"
                    strokeWidth="10"
                    strokeDasharray={`${progress * 2.51327} 251.327`}
                    transform="rotate(-90 50 50)"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-2xl font-bold">{progress}%</span>
                </div>
              </div>
              <h3 className="mt-4 text-sm font-medium text-muted-foreground">Progress</h3>
            </div>
          </CardContent>
        </Card>

        {/* Tokens Processed Widget */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center justify-center h-full">
              <span className="text-4xl font-semibold">{formatNumber(tokensProcessed)}</span>
              <h3 className="mt-2 text-sm text-muted-foreground">Tokens Processed</h3>
            </div>
          </CardContent>
        </Card>

        {/* Live Latency Widget */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center justify-center h-full">
              <div className="w-full h-24">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={latencyData.slice(-20)} margin={{ top: 5, right: 10, bottom: 16, left: 25 }}>
                    <CartesianGrid 
                      strokeDasharray="3 3" 
                      horizontal={true}
                      vertical={false}
                      stroke="rgba(0,0,0,0.1)"
                    />
                    <XAxis 
                      dataKey="timestamp"
                      tickFormatter={(value) => {
                        const date = new Date(value);
                        return date.toLocaleTimeString([], { 
                          hour: '2-digit', 
                          minute: '2-digit',
                          second: '2-digit',
                          hour12: false 
                        });
                      }}
                      tick={{ fontSize: 10 }}
                      axisLine={false}
                      tickLine={false}
                      interval="preserveStartEnd"
                      minTickGap={30}
                    />
                    <YAxis 
                      domain={[
                        Math.max(0, minLatency - latencyPadding), 
                        maxLatency + latencyPadding
                      ]} 
                      tickFormatter={(value) => `${value.toFixed(1)}`}
                      tick={{ fontSize: 10 }}
                      axisLine={false}
                      tickLine={false}
                      style={{ fontSize: '10px' }}
                      width={20}
                    />
                    <Tooltip
                      formatter={(value: number) => `${value.toFixed(1)}ms`}
                      labelFormatter={(timestamp) => {
                        const date = new Date(timestamp as string);
                        return date.toLocaleTimeString([], { 
                          hour: '2-digit', 
                          minute: '2-digit',
                          second: '2-digit',
                          hour12: false 
                        });
                      }}
                      contentStyle={{
                        backgroundColor: 'rgba(0, 0, 0, 0.8)',
                        border: 'none',
                        borderRadius: '4px',
                        padding: '4px 8px',
                      }}
                      itemStyle={{ color: '#fff' }}
                    />
                    <Line
                      type="linear"
                      dataKey="latency"
                      stroke="#3b82f6"
                      strokeWidth={2}
                      dot={false}
                      isAnimationActive={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
              <h3 className="mt-2 text-sm text-muted-foreground">Live Latency</h3>
              <span className="text-sm text-muted-foreground">
                {latencyData[latencyData.length - 1]?.latency.toFixed(1)}ms/token
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Token Distribution Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Identified Tokens</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={tokenChartData}
                layout="vertical"
                margin={{ top: 20, right: 30, left: 70, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis type="number" />
                <YAxis dataKey="type" type="category" />
                <Tooltip
                  formatter={(value: number) => formatNumber(value)}
                  labelFormatter={(label) => `Type: ${label}`}
                />
                <Bar dataKey="count" fill="#3b82f6" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}; 