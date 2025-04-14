'use client';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Card } from '@/components/ui/card';
import Link from 'next/link';
import { useParams } from 'next/navigation';

interface Job {
  id: string;
  name: string;
  source: string;
  initiated: string;
  progress: {
    status: 'completed' | 'in_progress';
    text: string;
    percentage?: number;
  };
}

const jobs: Job[] = [
  {
    id: '1',
    name: 'Daniel Docs 1',
    source: 'S3 Bucket - Daniel Docs 1',
    initiated: '2 months ago',
    progress: {
      status: 'completed',
      text: 'Completed 2 months ago',
    },
  },
  {
    id: '2',
    name: 'Daniel Docs 2',
    source: 'S3 Bucket - Daniel Docs 2',
    initiated: '3 days ago',
    progress: {
      status: 'in_progress',
      text: '52M / 92M',
      percentage: 56.5,
    },
  },
];

export default function Jobs() {
  const params = useParams();
  
  return (
    <Card>
      <div className="p-6">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-lg font-semibold">Jobs</h2>
          <Link href={`/token-classification/${params.deploymentId}/jobs/new`}>
            <Button>New</Button>
          </Link>
        </div>

        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="font-semibold">Name</TableHead>
                <TableHead className="font-semibold">Source</TableHead>
                <TableHead className="font-semibold">Initiated</TableHead>
                <TableHead className="font-semibold">Progress</TableHead>
                <TableHead className="font-semibold">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {jobs.map((job, index) => (
                <TableRow
                  key={job.id}
                  className={index % 2 === 0 ? 'bg-white' : 'bg-muted/50'}
                >
                  <TableCell>{job.name}</TableCell>
                  <TableCell>{job.source}</TableCell>
                  <TableCell>{job.initiated}</TableCell>
                  <TableCell>
                    {job.progress.status === 'completed' ? (
                      job.progress.text
                    ) : (
                      <div className="flex items-center gap-3">
                        <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-primary rounded-full transition-all"
                            style={{ width: `${job.progress.percentage}%` }}
                          />
                        </div>
                        <span className="text-sm text-muted-foreground whitespace-nowrap">
                          {job.progress.text}
                        </span>
                      </div>
                    )}
                  </TableCell>
                  <TableCell>
                    <Link
                      href={`/token-classification/${params.deploymentId}/jobs/${job.id}`}
                      className="text-primary hover:underline"
                    >
                      View
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </Card>
  );
} 