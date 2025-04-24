import {
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
} from '@/components/ui/table';
import { Loader2 } from 'lucide-react';
import { ClassifiedTokenDatabaseRecord, ObjectDatabaseRecord, ViewMode } from './types';

interface TableContentProps {
  viewMode: ViewMode;
  objectRecords: ObjectDatabaseRecord[];
  tokenRecords: ClassifiedTokenDatabaseRecord[];
  groupFilters: Record<string, boolean>;
  tagFilters: Record<string, boolean>;
  isLoadingObjectRecords: boolean;
  isLoadingTokenRecords: boolean;
}

export function TableContent({
  viewMode,
  objectRecords,
  tokenRecords,
  groupFilters,
  tagFilters,
  isLoadingObjectRecords,
  isLoadingTokenRecords,
}: TableContentProps) {
  if (viewMode === 'object') {
    return (
      <>
        <TableHeader>
          <TableRow>
            <TableHead>Tagged Tokens</TableHead>
            <TableHead>Source Object</TableHead>
            <TableHead>Groups</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {objectRecords
            .filter((record) => {
              return (
                record.groups.some((group) => groupFilters[group]) &&
                record.taggedTokens
                  .map((v) => v[1])
                  .some((tag) => tagFilters[tag])
              );
            })
            .map((record, index) => (
              <TableRow key={index}>
                <TableCell>
                  {record.taggedTokens
                    .map((token, index) => `${token[0]} (${token[1]})`)
                    .join(' ')}
                </TableCell>
                <TableCell>{record.sourceObject}</TableCell>
                <TableCell>{record.groups.join(', ')}</TableCell>
              </TableRow>
            ))}
          {isLoadingObjectRecords && (
            <TableRow>
              <TableCell colSpan={3} className="text-center py-4 text-gray-500">
                <div className="flex items-center justify-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading more records...
                </div>
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </>
    );
  }

  return (
    <>
      <TableHeader>
        <TableRow>
          <TableHead>Token</TableHead>
          <TableHead>Tag</TableHead>
          <TableHead>Source Object</TableHead>
          <TableHead>Groups</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {tokenRecords
          .filter((record) => {
            return (
              record.groups.some((group) => groupFilters[group]) &&
              tagFilters[record.tag]
            );
          })
          .map((record, index) => (
            <TableRow key={index}>
              <TableCell>{record.token}</TableCell>
              <TableCell>{record.tag}</TableCell>
              <TableCell>{record.sourceObject}</TableCell>
              <TableCell>{record.groups.join(', ')}</TableCell>
            </TableRow>
          ))}
        {isLoadingTokenRecords && (
          <TableRow>
            <TableCell colSpan={4} className="text-center py-4 text-gray-500">
              <div className="flex items-center justify-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading more records...
              </div>
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </>
  );
} 