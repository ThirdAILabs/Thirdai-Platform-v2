'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { getTagCounts, TagCount } from '@/lib/backend';

export default function Analytics() {
  const params = useParams();
  const [tagCounts, setTagCounts] = useState<TagCount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTagCounts = async () => {
      try {
        const deploymentId = params.deploymentId as string;
        const jobId = params.jobId as string;
        const counts = await getTagCounts(deploymentId, jobId);
        setTagCounts(counts);
      } catch (err) {
        setError('Failed to fetch tag counts');
        console.error('Error fetching tag counts:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchTagCounts();
  }, [params.deploymentId, params.jobId]);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Analytics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-4">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Analytics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-4 text-red-500">{error}</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Tag Distribution</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Tag</TableHead>
                <TableHead>Count</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tagCounts.length > 0 ? (
                tagCounts.map((tagCount) => (
                  <TableRow key={tagCount.tag}>
                    <TableCell>{tagCount.tag}</TableCell>
                    <TableCell>{tagCount.count}</TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={2} className="text-center py-4 text-muted-foreground">
                    No tag data available
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
} 