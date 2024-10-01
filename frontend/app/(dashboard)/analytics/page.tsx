'use client';
import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { UsageDurationChart } from '@/components/ui/charts'; // Import the chart
import Link from 'next/link';
import { Button } from '@mui/material';
import _ from 'lodash';
import { fetchNdbQueryCountStats, MetricDataPoint } from '@/lib/backend';

type ChartDataPoint = {
  x: Date;
  y: number;
};

export default function AnalyticsPage() {
  const [isClient, setIsClient] = useState(false);
  const [ndbQueryCountData, setNdbQueryCountData] = useState<ChartDataPoint[]>([]);

  // Ensure that the component only runs on the client
  useEffect(() => {
    setIsClient(true);
  }, []);

  // Function to fetch and update data
  const fetchData = async () => {
    try {
      const data: MetricDataPoint[] = await fetchNdbQueryCountStats();
      const chartData = data.map(([timestamp, value]) => ({
        x: new Date(timestamp * 1000), // Ensure proper conversion from seconds to milliseconds
        y: parseFloat(value),
      }));
      setNdbQueryCountData((prevData) => {
        // Ensure immutability for proper React state updates
        if (_.isEqual(prevData, chartData)) {
          console.log('prevData', prevData)
          console.log('chartData', chartData)

          return prevData; // Prevent unnecessary updates if the data is the same
        }
        return chartData;
      });
    } catch (error) {
      console.error('Error fetching ndb_query_count data:', error);
    }
  };

  // Fetch data periodically every 1 second
  useEffect(() => {
    fetchData(); // Initial fetch
    const intervalId = setInterval(fetchData, 1000); // Fetch every 1 second
    return () => clearInterval(intervalId); // Cleanup on unmount
  }, []);

  if (!isClient) {
    return null; // Return null on the first render to avoid hydration mismatch
  }

  // Dynamically create the Chart.js data object every render
  const usageDurationChartData = {
    labels: ndbQueryCountData.map((point) => point.x.toLocaleTimeString()), // Convert Date to readable time
    datasets: [
      {
        label: 'NDB Query Count',
        data: ndbQueryCountData.map((point) => point.y),
        borderColor: 'rgb(75, 192, 192)',
        backgroundColor: 'rgba(75, 192, 192, 0.2)',
        fill: false,
      },
    ],
  };

  const thirdaiPlatformBaseUrl = _.trim(process.env.THIRDAI_PLATFORM_BASE_URL || '', '/');
  const grafanaUrl = `${thirdaiPlatformBaseUrl}/grafana`;

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>System Status</CardTitle>
          <CardDescription>Monitor real-time usage and system improvements.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <UsageDurationChart data={usageDurationChartData} />
          </div>

          <div className="mt-4 flex justify-center items-center">
            <Link href={grafanaUrl} passHref legacyBehavior>
              <a target="_blank" rel="noopener noreferrer">
                <Button variant="contained">See more system stats</Button>
              </a>
            </Link>
          </div>
        </CardContent>
      </Card>
    </>
  );
}
