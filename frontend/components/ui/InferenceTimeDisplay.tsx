import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { TimerIcon, ZapIcon, Hash } from 'lucide-react';

interface InferenceTimeDisplayProps {
  processingTime: number;
  tokenCount?: number;
}

const InferenceTimeDisplay: React.FC<InferenceTimeDisplayProps> = ({
  processingTime,
  tokenCount,
}) => {
  let timePerToken = 0;
  if (tokenCount) {
    timePerToken = processingTime / tokenCount;
  }

  return (
    <Card className="bg-white hover:bg-gray-50 transition-colors">
      <CardContent className="p-6">
        <div className="space-y-4">
          {/* Total Inference Time */}
          <div className="flex items-center gap-4">
            <div className="p-3 bg-primary/10 rounded-full">
              <TimerIcon className="w-6 h-6 text-primary" />
            </div>
            <div className="flex flex-col">
              <span className="text-sm text-muted-foreground">Total Inference Time</span>
              <div className="flex items-baseline gap-1">
                <span className="text-2xl font-bold">{(processingTime * 1000).toFixed(2)}</span>
                <span className="text-sm text-muted-foreground">ms</span>
              </div>
            </div>
          </div>

          {tokenCount && (
            <>
              {/* Divider */}
              <div className="border-t border-gray-200" />

              {/* Total Tokens */}
              <div className="flex items-center gap-4">
                <div className="p-3 bg-primary/10 rounded-full">
                  <Hash className="w-6 h-6 text-primary" />
                </div>
                <div className="flex flex-col">
                  <span className="text-sm text-muted-foreground">Total Tokens</span>
                  <div className="flex items-baseline gap-1">
                    <span className="text-2xl font-bold">{tokenCount}</span>
                    <span className="text-sm text-muted-foreground">tokens</span>
                  </div>
                </div>
              </div>

              {/* Divider */}
              <div className="border-t border-gray-200" />

              {/* Time per Token */}
              <div className="flex items-center gap-4">
                <div className="p-3 bg-primary/10 rounded-full">
                  <ZapIcon className="w-6 h-6 text-primary" />
                </div>
                <div className="flex flex-col">
                  <span className="text-sm text-muted-foreground">Time per Token</span>
                  <div className="flex items-baseline gap-1">
                    <span className="text-2xl font-bold">{(timePerToken * 1000).toFixed(3)}</span>
                    <span className="text-sm text-muted-foreground">ms/token</span>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  );
};

export default InferenceTimeDisplay;
