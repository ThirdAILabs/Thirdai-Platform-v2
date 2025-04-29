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
import _ from 'lodash';

interface TokenCount {
  type: string;
  count: number;
}

interface LatencyDataPoint {
  timestamp: string;
  latency: number;
}

interface ClusterSpecs {
  cpus: number;
  vendorId: string;
  modelName: string;
  cpuMhz: number;
}

interface AnalyticsDashboardProps {
  progress: number;
  tokensProcessed: number;
  latencyData: LatencyDataPoint[];
  tokenTypes: string[];
  tokenCounts: Record<string, number>;
  clusterSpecs: ClusterSpecs;
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

export function AnalyticsDashboard({ 
  progress, 
  tokensProcessed,
  latencyData,
  tokenTypes,
  tokenCounts,
  clusterSpecs
}: AnalyticsDashboardProps) {
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
      <div className="grid grid-cols-4 gap-4">
        {/* Progress Widget */}
        <Card className="flex flex-col justify-between">
          <CardContent className="flex flex-col items-center justify-center flex-1 pt-6">
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
            <h3 className="mt-auto text-sm text-muted-foreground">Progress</h3>
          </CardContent>
        </Card>

        {/* Tokens Processed Widget */}
        <Card className="flex flex-col justify-between">
          <CardContent className="flex flex-col items-center pt-6 h-full">
            <div className="flex-1 flex items-center">
              <span className="text-4xl font-semibold">{formatNumber(tokensProcessed)}</span>
            </div>
            <h3 className="text-sm text-muted-foreground">Tokens Processed</h3>
          </CardContent>
        </Card>

        {/* Live Latency Widget */}
        <Card className="flex flex-col justify-between">
          <CardContent className="flex flex-col pt-6 h-full">
            <div className="flex-1">
              <div className="w-full h-[120px]">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={latencyData.slice(-20)}>
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
            </div>
            <div className="flex flex-col items-center">
              <span className="text-sm text-muted-foreground">
                {_.mean(latencyData.map(d => d.latency)).toFixed(3)}ms/token
              </span>
              <h3 className="text-sm text-muted-foreground">Live Latency</h3>
            </div>
          </CardContent>
        </Card>

        {/* Cluster Specs Widget */}
        <Card className="flex flex-col justify-between">
          <CardContent className="flex flex-col pt-6 h-full">
            <div className="flex-1 space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">CPU(s):</span>
                <span className="font-medium">{clusterSpecs.cpus}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Vendor ID:</span>
                <span className="font-medium">{clusterSpecs.vendorId}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Model name:</span>
                <span className="font-medium text-xs">{clusterSpecs.modelName}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">CPU MHz:</span>
                <span className="font-medium">{clusterSpecs.cpuMhz.toFixed(3)}</span>
              </div>
            </div>
            <h3 className="text-sm text-muted-foreground mt-4 text-center">Cluster Specs</h3>
          </CardContent>
        </Card>
      </div>

      {/* Token Distribution Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Identified Tokens</CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ height: `${Math.max(300, tokenChartData.length * 50)}px` }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={tokenChartData}
                layout="vertical"
                margin={{ top: 20, right: 30, left: 50, bottom: 30 }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  type="number" 
                  label={{ 
                    value: "Number of tokens", 
                    position: "bottom",
                    offset: 15
                  }}
                  tickFormatter={formatNumber}
                />
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
} 