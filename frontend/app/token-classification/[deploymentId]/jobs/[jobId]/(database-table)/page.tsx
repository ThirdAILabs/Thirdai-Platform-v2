import { useState, useEffect } from 'react';
import { DatabaseTable } from './DatabaseTable';
import { ObjectDatabaseRecord, ClassifiedTokenDatabaseRecord } from './types';

export default function DatabaseTablePage({
  params,
}: {
  params: { deploymentId: string; jobId: string };
}) {
  const [tags, setTags] = useState<string[]>([]);
  const [groups, setGroups] = useState<string[]>([]);
  const [objectRecords, setObjectRecords] = useState<ObjectDatabaseRecord[]>([]);
  const [classifiedTokenRecords, setClassifiedTokenRecords] = useState<ClassifiedTokenDatabaseRecord[]>([]);

  const loadMoreObjectRecords = async (): Promise<ObjectDatabaseRecord[]> => {
    // Implementation will be added later
    return [];
  };

  const loadMoreClassifiedTokenRecords = async (): Promise<ClassifiedTokenDatabaseRecord[]> => {
    // Implementation will be added later
    return [];
  };

  useEffect(() => {
    const fetchTags = async () => {
      try {
        const response = await fetch(`/api/token-classification/${params.deploymentId}/jobs/${params.jobId}/tags`);
        if (response.ok) {
          const data = await response.json();
          setTags(data.tags);
        }
      } catch (error) {
        console.error('Error fetching tags:', error);
      }
    };

    fetchTags();
  }, [params.deploymentId, params.jobId]);

  return (
    <div className="flex flex-col h-full">
      <DatabaseTable
        loadMoreObjectRecords={loadMoreObjectRecords}
        loadMoreClassifiedTokenRecords={loadMoreClassifiedTokenRecords}
        groups={groups}
        tags={tags}
      />
    </div>
  );
} 