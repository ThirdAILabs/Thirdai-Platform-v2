'use client';

import { useEffect, useState, useRef } from 'react';
import { TableHead, TableRow, TableHeader, TableBody, Table, TableCell } from '@/components/ui/table';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import SaveButton from '@/app/semantic-search/[deploymentId]/components/buttons/SaveButton';
import { ChevronDown, ChevronUp, ListFilter, Loader2 } from 'lucide-react';

interface ObjectDatabaseRecord {
  taggedTokens: [string, string][];
  sourceObject: string;
  groups: string[];
}

interface ClassifiedTokenDatabaseRecord {
  token: string;
  tag: string;
  sourceObject: string;
  groups: string[];
}

interface DatabaseTableProps {
  loadMoreObjectRecords: () => Promise<ObjectDatabaseRecord[]>;
  loadMoreClassifiedTokenRecords: () => Promise<ClassifiedTokenDatabaseRecord[]>;
  groups: string[];
  tags: string[];
}

export function DatabaseTable({ loadMoreObjectRecords, loadMoreClassifiedTokenRecords, groups, tags }: DatabaseTableProps) {
    const [isLoadingTokenRecords, setIsLoadingTokenRecords] = useState(false);
    const [isLoadingObjectRecords, setIsLoadingObjectRecords] = useState(false);

    const loadTokenRecords = () => {
        setIsLoadingTokenRecords(true);
        loadMoreClassifiedTokenRecords().then((records) => {
            setTokenRecords(prev => [...prev, ...records]);
            setIsLoadingTokenRecords(false);
        });
    }

    const loadObjectRecords = () => {
        setIsLoadingObjectRecords(true);
        loadMoreObjectRecords().then((records) => {
            setObjectRecords(prev => [...prev, ...records]);
            setIsLoadingObjectRecords(false);
        });
    }

    useEffect(() => {
        loadTokenRecords();
        loadObjectRecords();
    }, []);

  const [tokenRecords, setTokenRecords] = useState<ClassifiedTokenDatabaseRecord[]>([]);
  const [objectRecords, setObjectRecords] = useState<ObjectDatabaseRecord[]>([]);

  // Separate states for groups and tags
  const [groupFilters, setGroupFilters] = useState<Record<string, boolean>>(Object.fromEntries(groups.map(group => [group, true])) );

  const [tagFilters, setTagFilters] = useState<Record<string, boolean>>(Object.fromEntries(tags.map(tag => [tag, true])) );

  const [viewMode, setViewMode] = useState<'object' | 'classified-token'>('object');
  const [query, setQuery] = useState('');
  const [isGroupsExpanded, setIsGroupsExpanded] = useState(true);
  const [isTagsExpanded, setIsTagsExpanded] = useState(true);
  const [showShadow, setShowShadow] = useState(false);
  const [showTableShadow, setShowTableShadow] = useState(false);
  const filterScrollRef = useRef<HTMLDivElement>(null);
  const tableScrollRef = useRef<HTMLDivElement>(null);

  // Filter handling
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

  // View mode handling
  const handleViewModeChange = (value: 'object' | 'classified-token') => {
    setViewMode(value);
  };

  // Query handling
  const handleQueryChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(event.target.value);
  };

  const handleSave = () => {
    console.log('Saving...');
  };

  const toggleGroups = () => setIsGroupsExpanded(!isGroupsExpanded);
  const toggleTags = () => setIsTagsExpanded(!isTagsExpanded);

  const handleSelectAllGroups = () => {
    setGroupFilters(Object.fromEntries(groups.map(group => [group, true])));
  };

  const handleDeselectAllGroups = () => {
    setGroupFilters(Object.fromEntries(groups.map(group => [group, false])));
  };

  const handleSelectAllTags = () => {
    setTagFilters(Object.fromEntries(tags.map(tag => [tag, true])));
  };

  const handleDeselectAllTags = () => {
    setTagFilters(Object.fromEntries(tags.map(tag => [tag, false])));
  };

  // Handle scroll for shadows
  const handleFilterScroll = () => {
    if (filterScrollRef.current) {
      setShowShadow(filterScrollRef.current.scrollTop > 0);
    }
  };

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

  const renderTableHeader = () => {
    if (viewMode === 'object') {
        return (
          <TableHeader>
            <TableRow>
              <TableHead>Tagged Tokens</TableHead>
              <TableHead>Source Object</TableHead>
              <TableHead>Groups</TableHead>
            </TableRow>
        </TableHeader>
        );
      }
  
      return (
        <TableHeader>
            <TableRow>
              <TableHead>Token</TableHead>
              <TableHead>Tag</TableHead>
              <TableHead>Source Object</TableHead>
              <TableHead>Groups</TableHead>
            </TableRow>
        </TableHeader>
      );
  }

  const renderTableContent = () => {
    if (viewMode === 'object') {
      return (
        <TableBody>
            {objectRecords.map((record, index) => (
                <TableRow key={index}>
                <TableCell>{record.taggedTokens.map((token, index) => `${token[0]} (${token[1]})`).join(' ')}</TableCell>
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
      );
    }
  
    return (
        <TableBody>
            {tokenRecords.map((record, index) => (
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
    );
  };

  return (
    <Card className="h-[40vh]">
      <CardContent className="p-0 h-full">
        <div className="flex h-full">
          {/* Left Column - Filter Section */}
          <div className="w-64 flex flex-col border-r relative">
            {/* Fixed Filter Header */}
            <div className="sticky top-0 p-6 pb-2 pt-4 z-10">
              <div className="flex items-center gap-2">
                <ListFilter className="h-5 w-5" />
                <span className="flex font-medium h-[40px] items-center">Filter</span>
              </div>
            </div>

            {/* Scrollable Filter Content with Shadow */}
            <div 
              ref={filterScrollRef}
              className="flex-1 overflow-y-auto"
              onScroll={handleFilterScroll}
              style={{
                boxShadow: showShadow ? 'inset 0 4px 6px -4px rgba(0, 0, 0, 0.1)' : 'none'
              }}
            >
              <div className="p-6 pt-4 space-y-6">
                {/* Groups Section */}
                <div>
                  <div 
                    className="flex items-center justify-between text-sm text-gray-600 mb-2 cursor-pointer hover:text-gray-800"
                    onClick={toggleGroups}
                  >
                    <span>Groups</span>
                    {isGroupsExpanded ? (
                      <ChevronUp className="h-4 w-4" />
                    ) : (
                      <ChevronDown className="h-4 w-4" />
                    )}
                  </div>
                  {isGroupsExpanded && (
                    <>
                      <div className="flex gap-2 mb-2">
                        <button
                          onClick={handleSelectAllGroups}
                          className="text-xs text-blue-600 hover:text-blue-800"
                        >
                          Select All
                        </button>
                        <span className="text-gray-300">|</span>
                        <button
                          onClick={handleDeselectAllGroups}
                          className="text-xs text-blue-600 hover:text-blue-800"
                        >
                          Deselect All
                        </button>
                      </div>
                      <div className="space-y-2">
                        {groups.map((filter) => (
                          <label key={filter} className="flex items-center">
                            <input
                              type="checkbox"
                              checked={groupFilters[filter]}
                              onChange={() => handleGroupFilterChange(filter)}
                              className="mr-2"
                            />
                            {filter.charAt(0).toUpperCase() + filter.slice(1)}
                          </label>
                        ))}
                      </div>
                    </>
                  )}
                </div>
                
                {/* Tags Section */}
                <div>
                  <div 
                    className="flex items-center justify-between text-sm text-gray-600 mb-2 cursor-pointer hover:text-gray-800"
                    onClick={toggleTags}
                  >
                    <span>Tags</span>
                    {isTagsExpanded ? (
                      <ChevronUp className="h-4 w-4" />
                    ) : (
                      <ChevronDown className="h-4 w-4" />
                    )}
                  </div>
                  {isTagsExpanded && (
                    <>
                      <div className="flex gap-2 mb-2">
                        <button
                          onClick={handleSelectAllTags}
                          className="text-xs text-blue-600 hover:text-blue-800"
                        >
                          Select All
                        </button>
                        <span className="text-gray-300">|</span>
                        <button
                          onClick={handleDeselectAllTags}
                          className="text-xs text-blue-600 hover:text-blue-800"
                        >
                          Deselect All
                        </button>
                      </div>
                      <div className="space-y-2">
                        {tags.map((filter) => (
                          <label key={filter} className="flex items-center">
                            <input
                              type="checkbox"
                              checked={tagFilters[filter]}
                              onChange={() => handleTagFilterChange(filter)}
                              className="mr-2"
                            />
                            {filter.charAt(0).toUpperCase() + filter.slice(1)}
                          </label>
                        ))}
                      </div>
                    </>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Right Column - Main Content */}
          <div className="flex-1 flex flex-col h-full">
            {/* Fixed Header Content */}
            <div className="p-6 pb-2 pt-4 bg-white">
              <div className="flex items-center space-x-4">
                <div className="font-medium">View By</div>
                <Tabs value={viewMode} onValueChange={handleViewModeChange as any}>
                  <TabsList>
                    <TabsTrigger value="object">Object</TabsTrigger>
                    <TabsTrigger value="classified-token">Classified Token</TabsTrigger>
                  </TabsList>
                </Tabs>
                <div className="font-medium pl-2">Query</div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <div className="flex-1">
                      <Input
                        type="text"
                        placeholder="Enter query..."
                        value={query}
                        onChange={handleQueryChange}
                      />
                    </div>
                    <SaveButton 
                      onClick={handleSave}
                      style={{
                        width: '40px',
                        height: '40px',
                        minWidth: '40px',
                        padding: '8px'
                      }}
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Table Container */}
            <div 
              ref={tableScrollRef}
              className="flex-1 overflow-auto"
              onScroll={handleTableScroll}
              style={{
                boxShadow: showTableShadow ? 'inset 0 4px 6px -4px rgba(0, 0, 0, 0.1)' : 'none'
              }}
            >
              <div className="px-6">
                <Table>
                  {renderTableHeader()}
                  {renderTableContent()}
                </Table>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}