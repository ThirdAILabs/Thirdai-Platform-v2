'use client';

import { useEffect, useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink } from '@/components/ui/breadcrumb';
import * as _ from 'lodash';
import { useTokenClassificationEndpoints } from '@/lib/backend';
import Interact from './interact';
import Dashboard from './dashboard';
import Jobs from './jobs';

export default function Page() {
  const { workflowName } = useTokenClassificationEndpoints();

  return (
    <div className="bg-muted min-h-screen">
      <header className="w-full p-4 bg-muted border-b">
        <div className="max-w-7xl mx-auto space-y-4">
          <Breadcrumb>
            <BreadcrumbItem>
              <BreadcrumbLink href="/">Home</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbItem>
              <BreadcrumbLink href="#" aria-current="page">{workflowName}</BreadcrumbLink>
            </BreadcrumbItem>
          </Breadcrumb>
          
          <div className="font-bold text-xl truncate" title={workflowName}>
            {workflowName}
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto p-4">
        <Tabs defaultValue="testing" className="w-full">
          <TabsList className="mb-6">
            <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
            <TabsTrigger value="testing">Testing</TabsTrigger>
            <TabsTrigger value="jobs">Jobs</TabsTrigger>
          </TabsList>
          
          <TabsContent value="monitoring" className="mt-0">
            <Dashboard />
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
