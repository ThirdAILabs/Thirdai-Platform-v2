'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import RecentSamples from './samples';
import UpdateButton from './updateButton';
import RetrainButton from './retrainButton';
import { UsageDurationChart, UsageFrequencyChart, ReformulatedQueriesChart } from './charts';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Button } from '@mui/material';
import { useSearchParams } from 'next/navigation';
import { getWorkflowDetails, deploymentBaseUrl } from '@/lib/backend';
import _ from 'lodash';

export default function AnalyticsPage() {
  const [isClient, setIsClient] = useState(false);

  const searchParams = useSearchParams();
  const workflowid = searchParams.get('id');
  const [deploymentUrl, setDeploymentUrl] = useState<string | undefined>();
  const [modelName, setModelName] = useState<string>('');
  const [username, setUsername] = useState<string>('');

  // Ensure that the component only runs on the client
  useEffect(() => {
    setIsClient(true);

    const init = async () => {
      if (workflowid) {
        try {
          const workflowDetails = await getWorkflowDetails(workflowid);

          console.log('workflowDetails',workflowDetails)
          for (const model of workflowDetails.data.models) {
            if (model.component === 'nlp') {
              console.log(`here is: ${deploymentBaseUrl}/${model.model_id}`)
              setDeploymentUrl(`${deploymentBaseUrl}/${model.model_id}`);
              setModelName(model.model_name);
              setUsername(model.username);
              break;
            }
          }
        } catch (err) {
          console.error('Error fetching workflow details:', err);
        }
      }
    };

    init();
  }, [workflowid]);

  if (!isClient) {
    return null; // Return null on the first render to avoid hydration mismatch
  }

  const usageDurationData = {
    labels: ['7/17-7/20', '7/18-7/21', '7/22-7/25', '7/26-7/29', '7/30-8/2', '8/3-8/6', '8/7-8/11'],
    datasets: [
      {
        label: 'Hours per Day',
        data: [2.5, 3.2, 2.8, 3.5, 4.1, 3.9, 4.5],
        borderColor: 'rgb(75, 192, 192)',
        backgroundColor: 'rgba(75, 192, 192, 0.2)',
      },
      {
        label: 'Hours per Month',
        data: [75, 90, 86, 105, 128, 121, 135],
        borderColor: 'rgb(54, 162, 235)',
        backgroundColor: 'rgba(54, 162, 235, 0.2)',
      },
    ],
  };

  const usageFrequencyData = {
    labels: ['7/17-7/20', '7/18-7/21', '7/22-7/25', '7/26-7/29', '7/30-8/2', '8/3-8/6', '8/7-8/11'],
    datasets: [
      {
        label: 'Queries per Day',
        data: [25, 30, 28, 35, 40, 39, 45],
        borderColor: 'rgb(255, 99, 132)',
        backgroundColor: 'rgba(255, 99, 132, 0.2)',
      },
      {
        label: 'Clicks per Day',
        data: [15, 18, 20, 22, 24, 26, 28],
        borderColor: 'rgb(153, 102, 255)',
        backgroundColor: 'rgba(153, 102, 255, 0.2)',
      },
      {
        label: 'Upvotes per Day',
        data: [16, 13, 16, 13, 11, 10, 8],
        borderColor: 'rgb(255, 159, 64)',
        backgroundColor: 'rgba(255, 159, 64, 0.2)',
      },
      {
        label: 'Associates per Day',
        data: [5, 6, 5, 6, 7, 6, 7],
        borderColor: 'rgb(75, 192, 192)',
        backgroundColor: 'rgba(75, 192, 192, 0.2)',
      },
    ],
  };

  const reformulatedQueriesData = {
    labels: ['7/17-7/20', '7/18-7/21', '7/22-7/25', '7/26-7/29', '7/30-8/2', '8/3-8/6', '8/7-8/11'],
    datasets: [
      {
        label: 'Reformulated Queries',
        data: [18, 15, 17, 14, 13, 12, 10],
        borderColor: 'rgb(255, 205, 86)',
        backgroundColor: 'rgba(255, 205, 86, 0.2)',
      },
    ],
  };

  const thirdaiPlatformBaseUrl = _.trim(process.env.THIRDAI_PLATFORM_BASE_URL!, '/');
  const grafanaUrl = `${thirdaiPlatformBaseUrl}/grafana`;

  return (
    <>
      {deploymentUrl && <RecentSamples deploymentUrl={deploymentUrl} />}

      {deploymentUrl && modelName && username && (
        <UpdateButton 
          modelName={modelName} 
        />
      )}
    </>
  );
}
