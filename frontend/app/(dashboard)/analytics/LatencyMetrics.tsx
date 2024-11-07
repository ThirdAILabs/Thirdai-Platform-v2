import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tag, Clock, Database } from 'lucide-react';

const LatencyMetrics = () => {
  return (
    <div className="container mx-auto px-4">
      <Card>
        <CardHeader>
          <CardTitle>Model Performance Metrics</CardTitle>
          <CardDescription>
            Real-time latency metrics based on production inference
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-6">
            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Tag className="h-5 w-5" />
                <span className="text-sm font-medium">Entity Types</span>
              </div>
              <div className="text-3xl font-semibold">24</div>
              <div className="text-sm text-gray-500 mt-1">distinct entities recognized</div>
            </div>

            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Clock className="h-5 w-5" />
                <span className="text-sm font-medium">Average Latency</span>
              </div>
              <div className="text-3xl font-semibold">52ms</div>
              <div className="text-sm text-gray-500 mt-1">per inference request</div>
            </div>

            <div className="flex flex-col p-6 bg-white rounded-lg border">
              <div className="flex items-center gap-2 text-gray-600 mb-2">
                <Database className="h-5 w-5" />
                <span className="text-sm font-medium">Sample Size</span>
              </div>
              <div className="text-3xl font-semibold">10.2K</div>
              <div className="text-sm text-gray-500 mt-1">inference requests measured</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default LatencyMetrics;