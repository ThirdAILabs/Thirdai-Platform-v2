import {
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
} from '@/components/ui/table';
import { Loader2 } from 'lucide-react';
import { ClassifiedTokenDatabaseRecord, ObjectDatabaseRecord, ViewMode } from './types';
import { useMemo } from 'react';

// Define color palettes for tag colors
const PASTELS = ['#E5A49C', '#F6C886', '#FBE7AA', '#99E3B5', '#A6E6E7', '#A5A1E1', '#D8A4E2'];
const DARKERS = ['#D34F3E', '#F09336', '#F7CF5F', '#5CC96E', '#65CFD0', '#597CE2', '#B64DC8'];

// Default color for tags not in the computed colors map
const DEFAULT_COLOR = { text: '#E0E0E0', tag: '#A0A0A0' };

interface HighlightColor {
  text: string;
  tag: string;
}

interface HighlightedTokenProps {
  token: string;
  tag: string;
  tagColors: Record<string, HighlightColor>;
}

function HighlightedToken({ token, tag, tagColors }: HighlightedTokenProps) {
  // If tag is 'O', return the token with spacing but no highlighting
  if (tag === 'O') {
    return (
      <span style={{ marginRight: '4px' }}>
        {token}
      </span>
    );
  }

  const tagColor = tagColors[tag] || DEFAULT_COLOR;

  return (
    <span
      style={{
        backgroundColor: tagColor.text,
        padding: '2px',
        borderRadius: '2px',
        userSelect: 'none',
        display: 'inline-flex',
        alignItems: 'center',
        marginRight: '4px',
      }}
    >
      {token}
      <span
        style={{
          backgroundColor: tagColor.tag,
          color: 'white',
          fontSize: '11px',
          fontWeight: 'bold',
          borderRadius: '2px',
          marginLeft: '4px',
          padding: '1px 3px',
        }}
      >
        {tag}
      </span>
    </span>
  );
}

// Add a new component for highlighting tags
function HighlightedTag({ tag, tagColors }: { tag: string; tagColors: Record<string, HighlightColor> }) {
  if (tag === 'O') {
    return <span>{tag}</span>;
  }

  const tagColor = tagColors[tag] || DEFAULT_COLOR;

  return (
    <span
      style={{
        backgroundColor: tagColor.tag,
        color: 'white',
        fontSize: '11px',
        fontWeight: 'bold',
        borderRadius: '2px',
        padding: '1px 3px',
        display: 'inline-block',
      }}
    >
      {tag}
    </span>
  );
}

interface TableContentProps {
  viewMode: ViewMode;
  objectRecords: ObjectDatabaseRecord[];
  tokenRecords: ClassifiedTokenDatabaseRecord[];
  groupFilters: Record<string, boolean>;
  tagFilters: Record<string, boolean>;
  isLoadingObjectRecords: boolean;
  isLoadingTokenRecords: boolean;
  tags: string[]; // Add tags prop
}

export function TableContent({
  viewMode,
  objectRecords,
  tokenRecords,
  groupFilters,
  tagFilters,
  isLoadingObjectRecords,
  isLoadingTokenRecords,
  tags,
}: TableContentProps) {
  // Compute tag colors based on the provided tags
  const tagColors = useMemo(() => {
    const colors: Record<string, HighlightColor> = {};
    
    // Filter out 'O' tag and assign colors to each tag
    tags
      .filter(tag => tag !== 'O')
      .forEach((tag, index) => {
        colors[tag] = {
          text: PASTELS[index % PASTELS.length],
          tag: DARKERS[index % DARKERS.length],
        };
      });
    
    return colors;
  }, [tags]);

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
                <TableCell
                  style={{
                    maxWidth: '50%',
                    whiteSpace: 'normal',
                    wordBreak: 'break-word',
                    overflowWrap: 'break-word',
                  }}
                >
                  {record.taggedTokens.map((token, tokenIndex) => (
                    <HighlightedToken
                      key={`${index}-${tokenIndex}`}
                      token={token[0]}
                      tag={token[1]}
                      tagColors={tagColors}
                    />
                  ))}
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
              <TableCell>
                <HighlightedTag tag={record.tag} tagColors={tagColors} />
              </TableCell>
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