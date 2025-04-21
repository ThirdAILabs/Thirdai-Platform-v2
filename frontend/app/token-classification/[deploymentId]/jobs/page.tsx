'use client';

import { useParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import Link from 'next/link';

export default function JobsPage() {
  const params = useParams();

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold">Jobs</h1>
        <Button asChild>
          <Link href={`/token-classification/${params.deploymentId}/jobs/new`}>
            <Plus className="mr-2 h-4 w-4" />
            New Job
          </Link>
        </Button>
      </div>
    </div>
  );
} 