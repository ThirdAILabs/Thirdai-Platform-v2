'use client';

import { useEffect, useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink } from '@/components/ui/breadcrumb';
import * as _ from 'lodash';
import { useTokenClassificationEndpoints, getTrainReport } from '@/lib/backend';
import Interact from './interact';
import Dashboard from './dashboard';
import Jobs from './jobs';
import { useSearchParams } from 'next/navigation';
import { TrainingResults } from '@/app/(dashboard)/analytics/MetricsChart';
import { Card, CardContent } from '@/components/ui/card';
import { Alert } from '@mui/material';
import type { TrainReportData, LabelMetrics, ExampleCategories } from '@/lib/backend';

const emptyMetrics: LabelMetrics = {
  'O': {
    precision: 0,
    recall: 0,
    fmeasure: 0
  }
};

const emptyExamples: ExampleCategories = {
  true_positives: {},
  false_positives: {},
  false_negatives: {}
};

const emptyReport: TrainReportData = {
  before_train_metrics: emptyMetrics,
  after_train_metrics: emptyMetrics,
  after_train_examples: emptyExamples
};

export default function Page() {
  const { workflowName } = useTokenClassificationEndpoints();
  const searchParams = useSearchParams();
  const tab = searchParams.get('tab') || 'testing';
  const [trainReport, setTrainReport] = useState<TrainReportData>(emptyReport);
  const [isLoadingReport, setIsLoadingReport] = useState(true);
  const [reportError, setReportError] = useState('');

  // Fetch training report for monitoring tab
  useEffect(() => {
    const fetchReport = async () => {
      try {
        setIsLoadingReport(true);
        setReportError('');
        const response = await getTrainReport(workflowName);
        setTrainReport(response.data);
      } catch (error) {
        setReportError(error instanceof Error ? error.message : 'Failed to fetch training report');
        // Even on error, we want to show the TrainingResults component with empty data
        setTrainReport(emptyReport);
      } finally {
        setIsLoadingReport(false);
      }
    };

    fetchReport();
  }, [workflowName]);

  return (
    <div className="bg-muted min-h-screen">
      <header className="w-full p-4 bg-muted border-b">
        <div className="max-w-7xl mx-auto space-y-4">
          <Breadcrumb>
            <BreadcrumbItem>
              <BreadcrumbLink href="#" aria-current="page">Token Classification</BreadcrumbLink>
            </BreadcrumbItem>
          </Breadcrumb>
          
          <div className="font-bold text-xl truncate" title={workflowName}>
            {workflowName}
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto p-4">
        <Tabs defaultValue={tab} className="w-full">
          <TabsList className="mb-6">
            <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
            <TabsTrigger value="testing">Testing</TabsTrigger>
            <TabsTrigger value="jobs">Jobs</TabsTrigger>
          </TabsList>
          
          <TabsContent value="monitoring" className="mt-0">
            {isLoadingReport ? (
              <Card>
                <CardContent>
                  <div className="text-center py-8">Loading training report...</div>
                </CardContent>
              </Card>
            ) : (
              <TrainingResults report={trainReport} />
            )}
          </TabsContent>
          
          <TabsContent value="testing" className="mt-0">
            <Interact />
          </TabsContent>
          
          <TabsContent value="jobs" className="mt-0">
            <Jobs />
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}
