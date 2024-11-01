import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { TrainReportData } from '@/lib/backend';
import { ArrowUpIcon, ArrowDownIcon } from 'lucide-react';

interface PerformanceSummaryProps {
  beforeMetrics: TrainReportData['before_train_metrics'];
  afterMetrics: TrainReportData['after_train_metrics'];
}

export const PerformanceSummary: React.FC<PerformanceSummaryProps> = ({ beforeMetrics, afterMetrics }) => {
  const calculateChanges = () => {
    const allLabels = new Set([
      ...Object.keys(beforeMetrics),
      ...Object.keys(afterMetrics)
    ]);

    return Array.from(allLabels).map(label => {
      const beforeF1 = beforeMetrics[label]?.fmeasure || 0;
      const afterF1 = afterMetrics[label]?.fmeasure || 0;
      const change = (afterF1 - beforeF1) * 100;

      return {
        label,
        beforeF1: beforeF1 * 100,
        afterF1: afterF1 * 100,
        change
      };
    }).sort((a, b) => Math.abs(b.change) - Math.abs(a.change)); // Sort by absolute change
  };

  const changes = calculateChanges();
  const overallBefore = changes.reduce((acc, curr) => acc + curr.beforeF1, 0) / changes.length;
  const overallAfter = changes.reduce((acc, curr) => acc + curr.afterF1, 0) / changes.length;
  const overallChange = overallAfter - overallBefore;

  return (
    <div className="space-y-4 mb-12"> {/* Added explicit bottom margin */}
      <div>
        <h3 className="text-lg font-medium">Performance Summary</h3>
        <p className="text-sm text-gray-500">
          Overview of F1 score changes across all labels
        </p>
      </div>
      
      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[200px]">Label</TableHead>
              <TableHead className="text-right">Before F1</TableHead>
              <TableHead className="text-right">After F1</TableHead>
              <TableHead className="text-right">Change</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {changes.map(({ label, beforeF1, afterF1, change }) => (
              <TableRow key={label}>
                <TableCell className="font-medium">{label}</TableCell>
                <TableCell className="text-right">{beforeF1.toFixed(1)}%</TableCell>
                <TableCell className="text-right">{afterF1.toFixed(1)}%</TableCell>
                <TableCell className="text-right">
                  <span className={`flex items-center justify-end gap-1 font-medium
                    ${change > 0 ? 'text-green-600' : change < 0 ? 'text-red-600' : 'text-gray-600'}`}
                  >
                    {change > 0 ? (
                      <ArrowUpIcon className="h-4 w-4" />
                    ) : change < 0 ? (
                      <ArrowDownIcon className="h-4 w-4" />
                    ) : null}
                    {change > 0 ? '+' : ''}{change.toFixed(1)}%
                  </span>
                </TableCell>
              </TableRow>
            ))}
            {/* Overall row */}
            <TableRow className="bg-gray-50 font-medium">
              <TableCell>Overall Average</TableCell>
              <TableCell className="text-right">{overallBefore.toFixed(1)}%</TableCell>
              <TableCell className="text-right">{overallAfter.toFixed(1)}%</TableCell>
              <TableCell className="text-right">
                <span className={`flex items-center justify-end gap-1 font-medium
                  ${overallChange > 0 ? 'text-green-600' : overallChange < 0 ? 'text-red-600' : 'text-gray-600'}`}
                >
                  {overallChange > 0 ? (
                    <ArrowUpIcon className="h-4 w-4" />
                  ) : overallChange < 0 ? (
                    <ArrowDownIcon className="h-4 w-4" />
                  ) : null}
                  {overallChange > 0 ? '+' : ''}{overallChange.toFixed(1)}%
                </span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>
  );
};