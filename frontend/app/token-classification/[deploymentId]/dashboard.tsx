'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface UsageMetrics {
  totalRequests: number;
  averageLatency: number;
  successRate: number;
  lastDayRequests: number;
}

export default function Dashboard() {
  const [metrics, setMetrics] = useState<UsageMetrics>({
    totalRequests: 0,
    averageLatency: 0,
    successRate: 0,
    lastDayRequests: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: Implement metrics fetching logic
    setLoading(false);
  }, []);

  if (loading) {
    return <div>Loading metrics...</div>;
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{metrics.totalRequests}</div>
          <p className="text-xs text-muted-foreground">
            All-time inference requests
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Average Latency</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {metrics.averageLatency.toFixed(2)}ms
          </div>
          <p className="text-xs text-muted-foreground">
            Response time per request
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {(metrics.successRate * 100).toFixed(1)}%
          </div>
          <p className="text-xs text-muted-foreground">
            Successful inference requests
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Last 24 Hours</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{metrics.lastDayRequests}</div>
          <p className="text-xs text-muted-foreground">
            Requests in past 24 hours
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
