'use client';
import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import  StackedChart  from '@/components/ui/StackedChart'; // Import the StackedChart
import { fetchPrometheusData, MetricDataPoint } from '@/lib/backend'; // Fetch the new Prometheus data
import isEqual from 'lodash/isEqual';

type StackedChartDataPoint = {
  x: Date;
  top1: number;
  top2: number;
  top3: number;
  top4Plus: number;
};

export default function AnalyticsPage() {
  const [isClient, setIsClient] = useState(false);
  const [stackedChartData, setStackedChartData] = useState<StackedChartDataPoint[]>([]);

  // Ensure that the component only runs on the client
  useEffect(() => {
    setIsClient(true);
  }, []);

  // Function to fetch and update stacked chart data
  const fetchData = async () => {
    try {
      const data: Record<string, MetricDataPoint[]> = await fetchPrometheusData();
  
      // Ensure that all metrics exist in the response before proceeding
      const top1Selection = data['ndb_top_1_selection'] || [];
      const top2Selection = data['ndb_top_2_selection'] || [];
      const top3Selection = data['ndb_top_3_selection'] || [];
      const top4PlusSelection = data['ndb_top_4_plus_selection'] || [];
  
      // Check if we have data for top 1 selection (assuming this metric is always needed)
      if (top1Selection.length === 0) {
        console.error('No data for top 1 selection');
        return; // Exit the function if there's no data
      }
  
      // Process the chart data, safely handling index access for all metrics
      const chartData = top1Selection.map(([timestamp, top1], index) => {
        const top2 = parseFloat(top2Selection[index]?.[1] || '0'); // Handle case where data might be missing
        const top3 = parseFloat(top3Selection[index]?.[1] || '0');
        const top4Plus = parseFloat(top4PlusSelection[index]?.[1] || '0');
  
        return {
          x: new Date(timestamp * 1000), // Convert to milliseconds
          top1: parseFloat(top1),
          top2,
          top3,
          top4Plus,
        };
      });
  
      // Update the state only if the new data is different
      setStackedChartData((prevData) => {
        if (isEqual(prevData, chartData)) {
          return prevData; // Prevent unnecessary updates if the data is the same
        }
        return chartData;
      });
    } catch (error) {
      console.error('Error fetching stacked chart data:', error);
    }
  };

  // Fetch data periodically every 1 second
  useEffect(() => {
    fetchData(); // Initial fetch
    const intervalId = setInterval(fetchData, 1000); // Fetch every 1 second
    return () => clearInterval(intervalId); // Cleanup on unmount
  }, []);

  if (!isClient) {
    return null;
  }

  // Prepare data for the stacked chart
  const stackedChartDataFormatted = {
    labels: stackedChartData.map((point) => point.x.toLocaleTimeString()),
    datasets: [
      {
        label: 'Top 1',
        data: stackedChartData.map((point) => point.top1),
        backgroundColor: 'rgba(75, 192, 192, 0.8)',
      },
      {
        label: 'Top 2',
        data: stackedChartData.map((point) => point.top2),
        backgroundColor: 'rgba(153, 102, 255, 0.8)',
      },
      {
        label: 'Top 3',
        data: stackedChartData.map((point) => point.top3),
        backgroundColor: 'rgba(255, 159, 64, 0.8)',
      },
      {
        label: 'Top 4+',
        data: stackedChartData.map((point) => point.top4Plus),
        backgroundColor: 'rgba(255, 99, 132, 0.8)',
      },
    ],
  };

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>User Click Behavior</CardTitle>
          <CardDescription>Track the proportion of clicks on top 1, top 2, etc.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <StackedChart data={stackedChartDataFormatted} />
          </div>
        </CardContent>
      </Card>
    </>
  );
}
