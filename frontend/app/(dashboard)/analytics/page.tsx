'use client';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import RecentSamples from './samples';
import UpdateButton from './updateButton';
import {
  UsageDurationChart,
  UsageFrequencyChart,
  ReformulatedQueriesChart,
} from './charts';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { chartData } from './chart_data';
import { BAD, GOOD } from './good_bad';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import _ from 'lodash';

export default function AnalyticsPage() {
  const [isClient, setIsClient] = useState(false);

  const params = useSearchParams();

  // Ensure that the component only runs on the client
  useEffect(() => {
    setIsClient(true);
  }, []);

  if (!isClient) {
    return null;  // Return null on the first render to avoid hydration mismatch
  }

  const id = params.get("id") as string;

  const key = id.includes("feedback") ? "empty" : id === GOOD ? "good" : id === BAD ? "bad" : "default"

  const { usageDuration, usageFrequency, reformulatedQueries } = chartData[key];

  const thirdaiPlatformBaseUrl = _.trim(process.env.THIRDAI_PLATFORM_BASE_URL!, '/');
  const grafanaUrl = `${thirdaiPlatformBaseUrl}/grafana`;

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>System Status</CardTitle>
          <CardDescription>
            Monitor real-time usage and system improvements.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <UsageDurationChart data={usageDuration} />
            <UsageFrequencyChart data={usageFrequency} />
            <ReformulatedQueriesChart data={reformulatedQueries} />
          </div>

          <div className="mt-4">
            <Link href={grafanaUrl} passHref>
              <Button>
                Go to Grafana
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
      <RecentSamples id={key}/>
      <UpdateButton />
    </>
  );
}
