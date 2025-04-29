import { useEffect, useRef, useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Table } from '@/components/ui/table';
import { DatabaseTableProps, ViewMode, ClassifiedTokenDatabaseRecord, ObjectDatabaseRecord } from './types';
import { FilterSection } from './FilterSection';
import { HeaderContent } from './HeaderContent';
import { TableContent } from './TableContent';

export function DatabaseTable({
  loadMoreObjectRecords,
  loadMoreClassifiedTokenRecords,
  groups,
  tags,
}: DatabaseTableProps) {
  // Loading states and refs
  const [isLoadingTokenRecords, setIsLoadingTokenRecords] = useState(false);
  const [isLoadingObjectRecords, setIsLoadingObjectRecords] = useState(false);
  const loadedInitialTokenRecords = useRef(false);
  const loadedInitialObjectRecords = useRef(false);
  const tableScrollRef = useRef<HTMLDivElement>(null);
  const [showTableShadow, setShowTableShadow] = useState(false);

  // Data states
  const [tokenRecords, setTokenRecords] = useState<ClassifiedTokenDatabaseRecord[]>([]);
  const [objectRecords, setObjectRecords] = useState<ObjectDatabaseRecord[]>([]);
  const [viewMode, setViewMode] = useState<ViewMode>('object');
  const [query, setQuery] = useState('');

  // Filter states
  const [groupFilters, setGroupFilters] = useState<Record<string, boolean>>(
    Object.fromEntries(groups.map((group) => [group, true]))
  );
  const [tagFilters, setTagFilters] = useState<Record<string, boolean>>(
    Object.fromEntries(tags.map((tag) => [tag, true]))
  );

  // Load records functions
  const loadTokenRecords = () => {
    setIsLoadingTokenRecords(true);
    loadMoreClassifiedTokenRecords().then((records) => {
      setTokenRecords((prev) => [...prev, ...records]);
      setIsLoadingTokenRecords(false);
    });
  };

  const loadObjectRecords = () => {
    setIsLoadingObjectRecords(true);
    loadMoreObjectRecords().then((records) => {
      setObjectRecords((prev) => [...prev, ...records]);
      setIsLoadingObjectRecords(false);
    });
  };

  // Initial load
  useEffect(() => {
    if (!loadedInitialTokenRecords.current) {
      loadTokenRecords();
      loadedInitialTokenRecords.current = true;
    }
    if (!loadedInitialObjectRecords.current) {
      loadObjectRecords();
      loadedInitialObjectRecords.current = true;
    }
  }, []);

  // Scroll handler for infinite loading
  const handleTableScroll = () => {
    if (tableScrollRef.current) {
      setShowTableShadow(tableScrollRef.current.scrollTop > 0);

      // Check if we're near the bottom
      const { scrollTop, scrollHeight, clientHeight } = tableScrollRef.current;
      const bottomThreshold = 100; // pixels from bottom to trigger load
      
      if (scrollHeight - (scrollTop + clientHeight) < bottomThreshold) {
        // Load more records based on view mode
        if (viewMode === 'object' && !isLoadingObjectRecords) {
          loadObjectRecords();
        } else if (viewMode === 'classified-token' && !isLoadingTokenRecords) {
          loadTokenRecords();
        }
      }
    }
  };

  // Filter handlers
  const handleGroupFilterChange = (filterKey: string) => {
    setGroupFilters((prev) => ({
      ...prev,
      [filterKey]: !prev[filterKey],
    }));
  };

  const handleTagFilterChange = (filterKey: string) => {
    setTagFilters((prev) => ({
      ...prev,
      [filterKey]: !prev[filterKey],
    }));
  };

  const handleSelectAllGroups = () => {
    setGroupFilters(Object.fromEntries(groups.map((group) => [group, true])));
  };

  const handleDeselectAllGroups = () => {
    setGroupFilters(Object.fromEntries(groups.map((group) => [group, false])));
  };

  const handleSelectAllTags = () => {
    setTagFilters(Object.fromEntries(tags.map((tag) => [tag, true])));
  };

  const handleDeselectAllTags = () => {
    setTagFilters(Object.fromEntries(tags.map((tag) => [tag, false])));
  };

  // Other handlers
  const handleQueryChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(event.target.value);
  };

  const handleSave = () => {
    console.log('Saving...');
  };

  return (
    <Card className="h-[70vh]">
      <CardContent className="p-0 h-full">
        <div className="flex h-full">
          <FilterSection
            groups={groups}
            tags={tags}
            groupFilters={groupFilters}
            tagFilters={tagFilters}
            onGroupFilterChange={handleGroupFilterChange}
            onTagFilterChange={handleTagFilterChange}
            onSelectAllGroups={handleSelectAllGroups}
            onDeselectAllGroups={handleDeselectAllGroups}
            onSelectAllTags={handleSelectAllTags}
            onDeselectAllTags={handleDeselectAllTags}
          />

          <div className="flex-1 flex flex-col h-full">
            <HeaderContent
              viewMode={viewMode}
              query={query}
              onQueryChange={handleQueryChange}
              onViewModeChange={setViewMode}
              onSave={handleSave}
            />

            <div
              ref={tableScrollRef}
              className="flex-1 overflow-auto"
              onScroll={handleTableScroll}
              style={{
                boxShadow: showTableShadow
                  ? 'inset 0 4px 6px -4px rgba(0, 0, 0, 0.1)'
                  : 'none',
              }}
            >
              <div className="px-6">
                <TableContent
                  viewMode={viewMode}
                  objectRecords={objectRecords}
                  tokenRecords={tokenRecords}
                  groupFilters={groupFilters}
                  tagFilters={tagFilters}
                  isLoadingObjectRecords={isLoadingObjectRecords}
                  isLoadingTokenRecords={isLoadingTokenRecords}
                  tags={tags}
                />
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
} 