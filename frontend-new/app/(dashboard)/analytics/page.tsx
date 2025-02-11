'use client';
import RecentFeedbacks from './recentFeedbacks';
import UpdateButton from './updateButton';
import ModelUpdate from './ModelUpdate';
import UpdateButtonNDB from './updateButtonNDB';
import UsageStats from './usageStats';
import UsageStatsUDT from './usageStatsUDT';
import { useEffect, useState, Suspense, use } from 'react';
import { useSearchParams } from 'next/navigation';
import {
  getWorkflowDetails,
  deploymentBaseUrl,
  useTokenClassificationEndpoints,
} from '@/lib/backend';
import _ from 'lodash';
import { Workflow, fetchWorkflows } from '@/lib/backend';

function AnalyticsContent() {
  const [isClient, setIsClient] = useState(false);
  const searchParams = useSearchParams();
  const workflowId = searchParams.get('id');
  const [deploymentUrl, setDeploymentUrl] = useState<string | undefined>();
  const [modelName, setModelName] = useState<string>('');
  const [username, setUsername] = useState<string>('');
  const [workflowtype, setWorkflowType] = useState<string>('');
  const [deployStatus, setDeployStatus] = useState<string>('not_started');
  const [ndbModelId, setNdbModelId] = useState<string | undefined>();
  const [tags, setTags] = useState<string[]>([]);
  useEffect(() => {
    setIsClient(true);

    const init = async () => {
      if (workflowId) {
        try {
          const workflowDetails = await getWorkflowDetails(workflowId);

          console.log('workflowDetails', workflowDetails);
          setWorkflowType(workflowDetails.type);

          // Set deploy status
          setDeployStatus(workflowDetails.deploy_status);

          if (workflowDetails.type === 'ndb') {
            setNdbModelId(workflowDetails.model_id);
            setModelName(workflowDetails.model_name);
          } else {
            setDeploymentUrl(`${deploymentBaseUrl}/${workflowDetails.model_id}`);
            setModelName(workflowDetails.model_name);
            setUsername(workflowDetails.username);
          }
        } catch (err) {
          console.error('Error fetching workflow details:', err);
        }
      }
    };

    init();
  }, [workflowId]);

  const [workflows, setWorkflows] = useState<Workflow[]>([]);

  useEffect(() => {
    async function getWorkflows() {
      try {
        const fetchedWorkflows = await fetchWorkflows();
        setWorkflows(fetchedWorkflows);
      } catch (err) {
        if (err instanceof Error) {
          console.log(err.message);
        } else {
          console.log('An unknown error occurred');
        }
      }
    }

    getWorkflows();
  }, []);

  const workflowNames = workflows.map((workflow) => workflow.model_name);

  if (!isClient) {
    return null; // Return null on the first render to avoid hydration mismatch
  }

  if (workflowtype == 'nlp-token' || workflowtype == 'nlp-text')
    return (
      <div className="container mx-auto px-4 py-8">
        {modelName && deploymentUrl && workflowId && (
          <>
            <UsageStatsUDT />
            <div className="mt-6">
              <ModelUpdate
                username={username}
                modelName={modelName}
                deploymentUrl={deploymentUrl}
                workflowNames={workflowNames}
                deployStatus={deployStatus}
                modelId={workflowId}
              />
            </div>
          </>
        )}
      </div>
    );
  else if (workflowtype == 'ndb' || workflowtype == 'enterprise-search') {
    console.log('update button, ', ndbModelId);
    return (
      <>
        <UsageStats />
        <RecentFeedbacks modelId={ndbModelId} />
        {modelName && <UpdateButtonNDB modelName={modelName} />}
      </>
    );
  }
}

export default function AnalyticsPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <AnalyticsContent />
    </Suspense>
  );
}
