'use client';

import React from 'react';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { CheckCircle, ArrowLeft, RefreshCw, Pause, Square, Plus, Edit } from 'lucide-react';
import { AnalyticsDashboard } from '@/components/AnalyticsDashboard';
import { DatabaseTable } from './(database-table)/DatabaseTable';

// Mock data for database table
const mockGroups = ['Reject', 'Sensitive', 'Safe'];
const mockTags = ['VIN', 'NAME', 'ORG', 'ADDRESS', 'EMAIL', 'SSN', 'PHONE', 'POLICY_ID', 'MED_REC_NO', 'LICENSE', 'EMPLOYER', 'ID', 'USERNAME', 'URL', 'IP_ADDR', 'ZIP_CODE', 'ACCOUNT', 'INS_PROV', 'PROCEDURE', 'DATE', 'NATIONALITY', 'SERIAL_NO', 'CRED_CARD_NUM', 'CVV'];

const loadMoreMockObjectRecords = () => {
  return Promise.resolve([
    {
      taggedTokens: [
        ['My', 'O'] as [string, string],
        ['name', 'O'] as [string, string],
        ['is', 'O'] as [string, string],
        ['John', 'NAME'] as [string, string],
        ['Smith', 'NAME'] as [string, string],
        ['and', 'O'] as [string, string],
        ['my', 'O'] as [string, string],
        ['social', 'O'] as [string, string],
        ['is', 'O'] as [string, string],
        ['123-45-6789', 'SSN'] as [string, string],
      ],
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      taggedTokens: [
        ['Jane', 'NAME'] as [string, string],
        ['Doe', 'NAME'] as [string, string],
        ['at', 'O'] as [string, string],
        ['123', 'ADDRESS'] as [string, string],
        ['Main', 'ADDRESS'] as [string, string],
        ['St', 'ADDRESS'] as [string, string],
        ['with', 'O'] as [string, string],
        ['vehicle', 'O'] as [string, string],
        ['1HGCM82633A004352', 'VIN'] as [string, string],
      ],
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
  ]);
};

const loadMoreMockClassifiedTokenRecords = () => {
  return Promise.resolve([
    {
      token: 'John',
      tag: 'NAME',
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      token: 'Smith',
      tag: 'NAME',
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      token: '123-45-6789',
      tag: 'SSN',
      sourceObject: 'call_transcript_1.txt',
      groups: ['Sensitive'],
    },
    {
      token: 'Jane',
      tag: 'NAME',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
    {
      token: 'Doe',
      tag: 'NAME',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
    {
      token: '123 Main St',
      tag: 'ADDRESS',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
    {
      token: '1HGCM82633A004352',
      tag: 'VIN',
      sourceObject: 'call_transcript_2.txt',
      groups: ['Reject'],
    },
  ]);
};

// Source option card component
interface SourceOptionProps {
  title: string;
  description: string;
  isSelected?: boolean;
  onClick: () => void;
}

const SourceOption: React.FC<SourceOptionProps> = ({ title, description, isSelected = false, onClick }) => (
  <div 
    className={`relative p-6 border rounded-md cursor-pointer transition-all
      ${isSelected ? 'border-blue-500 border-2' : 'border-gray-200 hover:border-blue-300'}
    `}
    onClick={onClick}
  >
    <h3 className="text-base font-medium">{title}</h3>
    <p className="text-sm text-gray-500 mt-1">{description}</p>
    
    {isSelected && (
      <div className="absolute top-3 right-3">
        <Edit className="h-4 w-4 text-gray-500" />
      </div>
    )}
  </div>
);

// Tag chip component
interface TagProps {
  tag: string;
  selected?: boolean;
  onClick?: () => void;
}

const Tag: React.FC<TagProps> = ({ tag, selected = true, onClick }) => (
  <div 
    className={`px-3 py-1.5 text-sm font-medium rounded-sm cursor-pointer
      ${selected ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}
    `}
    onClick={onClick}
  >
    {tag}
  </div>
);

// Group card component
interface GroupProps {
  name: string;
  definition: string;
}

const GroupCard: React.FC<GroupProps> = ({ name, definition }) => (
  <div className="border border-gray-200 rounded-md overflow-hidden">
    <div className="p-4 border-b border-gray-200">
      <h3 className="text-base font-medium">{name}</h3>
    </div>
    <div className="p-4">
      <p className="text-sm font-mono">{definition}</p>
    </div>
  </div>
);

export default function JobDetail() {
  const params = useParams();
  const [lastUpdated, setLastUpdated] = useState(0);
  const [tabValue, setTabValue] = useState('configuration');
  const [selectedSource, setSelectedSource] = useState('s3');
  const [selectedSaveLocation, setSelectedSaveLocation] = useState('s3');
  const [selectedTags, setSelectedTags] = useState<string[]>(mockTags);

  const toggleTag = (tag: string) => {
    if (selectedTags.includes(tag)) {
      setSelectedTags(selectedTags.filter(t => t !== tag));
    } else {
      setSelectedTags([...selectedTags, tag]);
    }
  };

  const selectAllTags = () => {
    setSelectedTags(mockTags);
  };

  return (
    <div className="container px-4 py-8 mx-auto">
      {/* Breadcrumbs */}
      <div className="mb-6">
        <div className="flex items-center mb-2">
          <Link href={`/token-classification/${params.deploymentId}/jobs`} className="text-blue-500 hover:underline">
            Jobs
          </Link>
          <span className="mx-2 text-gray-400">/</span>
          <span className="text-gray-700">Customer Calls</span>
        </div>
      </div>

      {/* Title and Back Button */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-medium">Customer Calls</h1>
        
        <Button variant="outline" size="sm" asChild>
          <Link href={`/token-classification/${params.deploymentId}?tab=jobs`} className="flex items-center">
            <ArrowLeft className="mr-1 h-4 w-4" /> Back to Jobs
          </Link>
        </Button>
      </div>

      {/* Tabs and Controls */}
      <Tabs value={tabValue} onValueChange={setTabValue} className="w-full">
        <div className="flex items-center justify-between border-b mb-6">
          <TabsList className="border-0 bg-transparent p-0">
            <TabsTrigger 
              value="configuration" 
              className="data-[state=active]:border-b-2 data-[state=active]:border-blue-500 data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-3 data-[state=active]:bg-transparent"
            >
              Configuration
            </TabsTrigger>
            <TabsTrigger 
              value="analytics" 
              className="data-[state=active]:border-b-2 data-[state=active]:border-blue-500 data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-3 data-[state=active]:bg-transparent"
            >
              Analytics
            </TabsTrigger>
            <TabsTrigger 
              value="output" 
              className="data-[state=active]:border-b-2 data-[state=active]:border-blue-500 data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-3 data-[state=active]:bg-transparent"
            >
              Output
            </TabsTrigger>
          </TabsList>

          <div className="flex items-center space-x-4">
            <span className="text-sm text-gray-500">Last updated: {lastUpdated} seconds ago</span>
            <Button variant="ghost" size="icon" onClick={() => setLastUpdated(0)}>
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon">
              <Pause className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon">
              <Square className="h-4 w-4" />
            </Button>
          </div>
        </div>

        <TabsContent value="configuration" className="mt-0">
          <div className="space-y-8">
            {/* Source section */}
            <div>
              <h2 className="text-lg font-medium mb-4">Source</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <SourceOption 
                  title="S3 Bucket" 
                  description="s3://thirdai-dev/customer-calls/2025/" 
                  isSelected={selectedSource === 's3'}
                  onClick={() => setSelectedSource('s3')}
                />
                <SourceOption 
                  title="Local Storage" 
                  description="Configure now" 
                  isSelected={selectedSource === 'local'}
                  onClick={() => setSelectedSource('local')}
                />
                <SourceOption 
                  title="More options" 
                  description="coming soon" 
                  isSelected={selectedSource === 'more'}
                  onClick={() => setSelectedSource('more')}
                />
              </div>
            </div>

            {/* Tags section */}
            <div>
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-medium">Tags</h2>
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={selectAllTags}
                  className="text-sm flex items-center"
                >
                  <span className="mr-1">Select All</span>
                  <input 
                    type="checkbox" 
                    checked={selectedTags.length === mockTags.length} 
                    onChange={selectAllTags}
                    className="rounded border-gray-300"
                  />
                </Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {mockTags.map(tag => (
                  <Tag
                    key={tag}
                    tag={tag}
                    selected={selectedTags.includes(tag)}
                    onClick={() => toggleTag(tag)}
                  />
                ))}
              </div>
            </div>

            {/* Groups section */}
            <div>
              <h2 className="text-lg font-medium mb-4">Groups</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <GroupCard name="Reject" definition="COUNT(tags) > 5" />
                <GroupCard name="Sensitive" definition="COUNT(tags) > 0" />
                <GroupCard name="Safe" definition="COUNT(tags) = 0" />
                
                <div 
                  className="border border-dashed border-gray-300 rounded-md flex items-center justify-center p-6 cursor-pointer hover:border-gray-400"
                  onClick={() => console.log('Add new group')}
                >
                  <div className="flex flex-col items-center">
                    <Plus className="h-8 w-8 text-gray-400 mb-2" />
                    <span className="text-gray-600">Define new group</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Save Groups To section */}
            <div>
              <h2 className="text-lg font-medium mb-4">Save Groups To</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <SourceOption 
                  title="S3 Bucket" 
                  description="thirdai-dev/sensitive/customer-calls/2025/" 
                  isSelected={selectedSaveLocation === 's3'}
                  onClick={() => setSelectedSaveLocation('s3')}
                />
                <SourceOption 
                  title="Local Storage" 
                  description="local" 
                  isSelected={selectedSaveLocation === 'local'}
                  onClick={() => setSelectedSaveLocation('local')}
                />
                <SourceOption 
                  title="No storage location" 
                  description="You can still save groups" 
                  isSelected={selectedSaveLocation === 'none'}
                  onClick={() => setSelectedSaveLocation('none')}
                />
                <SourceOption 
                  title="More options" 
                  description="coming soon" 
                  isSelected={selectedSaveLocation === 'more'}
                  onClick={() => setSelectedSaveLocation('more')}
                />
              </div>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="analytics">
          <AnalyticsDashboard
            progress={40}
            tokensProcessed={1229000}
            latencyData={[
              { timestamp: '2024-03-10T12:00:00', latency: 0.096 },
              { timestamp: '2024-03-10T12:00:01', latency: 0.09 },
              { timestamp: '2024-03-10T12:00:02', latency: 0.082 },
              { timestamp: '2024-03-10T12:00:03', latency: 0.101 },
              { timestamp: '2024-03-10T12:00:04', latency: 0.098 },
              { timestamp: '2024-03-10T12:00:05', latency: 0.095 },
              { timestamp: '2024-03-10T12:00:06', latency: 0.097 },
              { timestamp: '2024-03-10T12:00:07', latency: 0.099 },
              { timestamp: '2024-03-10T12:00:08', latency: 0.094 },
              { timestamp: '2024-03-10T12:00:09', latency: 0.093 },
              { timestamp: '2024-03-10T12:00:10', latency: 0.088 }, 
              { timestamp: '2024-03-10T12:00:11', latency: 0.082 },
              { timestamp: '2024-03-10T12:00:12', latency: 0.079 },
              { timestamp: '2024-03-10T12:00:13', latency: 0.087 },
              { timestamp: '2024-03-10T12:00:14', latency: 0.083 },
              { timestamp: '2024-03-10T12:00:15', latency: 0.084 },
              { timestamp: '2024-03-10T12:00:16', latency: 0.086 },
              { timestamp: '2024-03-10T12:00:17', latency: 0.083 },
              { timestamp: '2024-03-10T12:00:18', latency: 0.089 },
              { timestamp: '2024-03-10T12:00:19', latency: 0.091 },
              { timestamp: '2024-03-10T12:00:20', latency: 0.083 },
              { timestamp: '2024-03-10T12:00:21', latency: 0.092 },
              { timestamp: '2024-03-10T12:00:22', latency: 0.094 },
            ]}
            tokenTypes={['NAME', 'VIN', 'ORG', 'ID', 'SSN', 'ADDRESS', 'EMAIL']}
            tokenCounts={{
              'NAME': 21200000,
              'VIN': 19800000,
              'ORG': 13300000,
              'ID': 13300000,
              'SSN': 13300000,
              'ADDRESS': 5600000,
              'EMAIL': 3800000
            }}
            clusterSpecs={{
              cpus: 48,
              vendorId: 'GenuineIntel',
              modelName: 'Intel Xeon E5-2680',
              cpuMhz: 1197.408
            }}
          />
        </TabsContent>

        <TabsContent value="output">
          <DatabaseTable 
            loadMoreObjectRecords={loadMoreMockObjectRecords}
            loadMoreClassifiedTokenRecords={loadMoreMockClassifiedTokenRecords}
            groups={mockGroups}
            tags={mockTags}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
} 