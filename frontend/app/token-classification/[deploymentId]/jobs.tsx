'use client';

import { useEffect, useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

interface Job {
  id: string;
  status: string;
  createdAt: string;
  completedAt?: string;
  type: string;
}

export default function Jobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: Implement job fetching logic
    setLoading(false);
  }, []);

  if (loading) {
    return <div>Loading jobs...</div>;
  }

  return (
    <div className="rounded-lg border bg-card">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Job ID</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Completed</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {jobs.map((job) => (
            <TableRow key={job.id}>
              <TableCell className="font-mono">{job.id}</TableCell>
              <TableCell>{job.type}</TableCell>
              <TableCell>{job.status}</TableCell>
              <TableCell>{new Date(job.createdAt).toLocaleString()}</TableCell>
              <TableCell>
                {job.completedAt
                  ? new Date(job.completedAt).toLocaleString()
                  : '-'}
              </TableCell>
            </TableRow>
          ))}
          {jobs.length === 0 && (
            <TableRow>
              <TableCell colSpan={5} className="text-center text-muted-foreground">
                No jobs found
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
} 