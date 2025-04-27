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

// TODO: MOCK DATA
const mockReports: Report[] = [
  {
    name: "Medical Records Review",
    report_id: "tcr_123456789",
    status: "completed", 
    submitted_at: "2024-01-15T10:30:00Z",
    updated_at: "2024-01-15T10:35:00Z",
    documents: [
      {
        path: "medical_records.txt",
        location: "s3://bucket/medical_records.txt",
        source_id: "doc_123",
        options: {},
        metadata: {}
      }
    ],
    msg: null,
    content: {
      report_id: "tcr_123456789",
      results: [
        {
          "medical_records.txt": [
            { text: "123-45-6789", tag: "SSN" },
            { text: "01/15/1980", tag: "DOB" },
            { text: "555-123-4567", tag: "PHONE" },
            { text: "123 Main St, Apt 4B", tag: "ADDRESS" },
            { text: "robert.chen1982@gmail.com", tag: "EMAIL" },
            { text: "Diabetes Type 2", tag: "MEDICAL_CONDITION" },
            { text: "Metformin 500mg", tag: "MEDICATION" },
            { text: "Dr. Sarah Wilson", tag: "PROVIDER_NAME" },
            { text: "Memorial Hospital", tag: "FACILITY_NAME" },
            { text: "INS-987654321", tag: "INSURANCE_ID" }
          ]
        }
      ]
    }
  },
  {
    name: "Insurance Claims Processing",
    report_id: "tcr_987654321", 
    status: "completed",
    submitted_at: "2024-01-15T11:30:00Z",
    updated_at: "2024-01-15T11:35:00Z",
    documents: [
      {
        path: "insurance_claims.txt",
        location: "s3://bucket/insurance_claims.txt",
        source_id: "doc_456",
        options: {},
        metadata: {}
      }
    ],
    msg: null,
    content: {
      report_id: "tcr_987654321",
      results: [
        {
          "insurance_claims.txt": [
            { text: "CLM-123456", tag: "CLAIM_ID" },
            { text: "POL-789012", tag: "POLICY_NUMBER" },
            { text: "John Smith", tag: "NAME" },
            { text: "Hypertension", tag: "MEDICAL_CONDITION" },
            { text: "Lisinopril", tag: "MEDICATION" },
            { text: "Dr. Michael Brown", tag: "PROVIDER_NAME" },
            { text: "City Medical Center", tag: "FACILITY_NAME" }
          ]
        }
      ]
    }
  },
  {
    name: "Customer Support Chat Logs",
    report_id: "tcr_456789123",
    status: "completed",
    submitted_at: "2024-01-15T12:30:00Z",
    updated_at: "2024-01-15T12:35:00Z",
    documents: [
      {
        path: "chat_logs.txt",
        location: "s3://bucket/chat_logs.txt",
        source_id: "doc_789",
        options: {},
        metadata: {}
      }
    ],
    msg: null,
    content: {
      report_id: "tcr_456789123",
      results: [
        {
          "chat_logs.txt": [
            { text: "4832-5691-2748-1035", tag: "CREDIT_CARD" },
            { text: "09/27", tag: "EXPIRATION_DATE" },
            { text: "382", tag: "CVV" },
            { text: "jane.doe@email.com", tag: "EMAIL" },
            { text: "987-654-3210", tag: "PHONE" },
            { text: "532-48-1095", tag: "SSN" }
          ]
        }
      ]
    }
  },
];

export default function Jobs() {
  const params = useParams();
  const [reports, setReports] = useState<Report[]>(mockReports);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // useEffect(() => {
  //   const fetchReports = async () => {
  //     try {
  //       const deploymentId = params.deploymentId as string;
  //       const reportsData = await listReports(deploymentId);
  //       setReports(reportsData);
  //     } catch (err) {
  //       setError('Failed to fetch reports');
  //       console.error('Error fetching reports:', err);
  //     } finally {
  //       setLoading(false);
  //     }
  //   };

  //   fetchReports();
  // }, [params.deploymentId]);

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
                <TableHead className="font-semibold">Name</TableHead>
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
                    <TableCell>{report.name}</TableCell>
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