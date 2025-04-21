'use client';

import { useEffect, useState } from 'react';
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
import { listReports, Report } from '@/lib/backend';

export default function Jobs() {
  const params = useParams();
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchReports = async () => {
      try {
        const deploymentId = params.deploymentId as string;
        const reportsData = await listReports(deploymentId);
        setReports(reportsData);
      } catch (err) {
        setError('Failed to fetch reports');
        console.error('Error fetching reports:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchReports();
  }, [params.deploymentId]);

  if (loading) {
    return (
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-lg font-semibold">Jobs</h2>
            <Link href={`/token-classification/${params.deploymentId}/jobs/new`}>
              <Button>New</Button>
            </Link>
          </div>
          <div className="text-center py-4">Loading...</div>
        </div>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-lg font-semibold">Jobs</h2>
            <Link href={`/token-classification/${params.deploymentId}/jobs/new`}>
              <Button>New</Button>
            </Link>
          </div>
          <div className="text-center py-4 text-red-500">{error}</div>
        </div>
      </Card>
    );
  }

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
                <TableHead className="font-semibold">Report ID</TableHead>
                <TableHead className="font-semibold">Status</TableHead>
                <TableHead className="font-semibold">Submitted At</TableHead>
                <TableHead className="font-semibold">Updated At</TableHead>
                <TableHead className="font-semibold">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {reports && reports.length > 0 ? (
                reports.map((report, index) => (
                  <TableRow
                    key={report.report_id}
                    className={index % 2 === 0 ? 'bg-white' : 'bg-muted/50'}
                  >
                    <TableCell>{report.report_id}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className={`h-full rounded-full transition-all ${
                              report.status === 'completed' ? 'bg-green-500' : 'bg-primary'
                            }`}
                            style={{ width: report.status === 'completed' ? '100%' : '50%' }}
                          />
                        </div>
                        <span className="text-sm text-muted-foreground whitespace-nowrap">
                          {report.status}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>{new Date(report.submitted_at).toLocaleString()}</TableCell>
                    <TableCell>{new Date(report.updated_at).toLocaleString()}</TableCell>
                    <TableCell>
                      <Link
                        href={`/token-classification/${params.deploymentId}/jobs/${report.report_id}`}
                        className="text-primary hover:underline"
                      >
                        View
                      </Link>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-4 text-muted-foreground">
                    No reports found. Create a new report to get started.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </Card>
  );
} 