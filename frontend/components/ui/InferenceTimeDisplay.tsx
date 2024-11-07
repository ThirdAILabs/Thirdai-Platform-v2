import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { TimerIcon } from 'lucide-react';

interface InferenceTimeDisplayProps {
  processingTime: number;
}

const InferenceTimeDisplay: React.FC<InferenceTimeDisplayProps> = ({ processingTime }) => {
  return (
    <Card className="bg-white hover:bg-gray-50 transition-colors mb-5">
      <CardContent className="p-6">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-primary/10 rounded-full">
            <TimerIcon className="w-6 h-6 text-primary" />
          </div>
          <div className="flex flex-col">
            <span className="text-sm text-muted-foreground">Inference Time</span>
            <div className="flex items-baseline gap-1">
              <span className="text-2xl font-bold">{(processingTime * 1000).toFixed(2)}</span>
              <span className="text-sm text-muted-foreground">ms</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default InferenceTimeDisplay;