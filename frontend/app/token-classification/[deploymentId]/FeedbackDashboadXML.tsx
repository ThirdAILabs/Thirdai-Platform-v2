import React, { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@mui/material';
import { Trash2 } from 'lucide-react';

// Types for selections and predictions
interface Selection {
  start: number;
  end: number;
  xpath: string;
  tag: string;
  value: string;
}

interface FeedbackDashboardProps {
  selections: Selection[];
  onDeleteSelection?: (index: number) => void;
}

function FeedbackItem({
  feedback,
  index,
  onDeleteSelection,
}: {
  feedback: any;
  index: number;
  onDeleteSelection?: (index: number) => void;
}) {
  return (
    <div className="feedback-item bg-white p-4 rounded-lg shadow-md mb-4">
      <div className="flex justify-between items-center">
        <div className="feedback-item-text text-gray-800">
          {feedback.text}
          <span className="feedback-item-label bg-green-400 text-white px-2 py-1 rounded-lg ml-2 text-sm">
            {feedback.label}
          </span>
        </div>
        <Button
          variant="contained"
          color="error"
          className="hover:bg-red-500 hover:text-white"
          onClick={() => onDeleteSelection?.(index)}
        >
          {<Trash2 className="h-6 w-5"></Trash2>}
        </Button>
      </div>
    </div>
  );
}

export function FeedbackDashboard({ selections, onDeleteSelection }: FeedbackDashboardProps) {
  return (
    <Card className="w-full max-w-4xl mx-auto" style={{ maxHeight: '41vh', overflowY: 'auto' }}>
      <CardHeader>
        <CardTitle>Feedback from this session</CardTitle>
      </CardHeader>
      <CardContent>
        {selections.map((selection, index) => (
          <FeedbackItem
            key={index}
            feedback={{
              text: selection.value,
              label: selection.tag,
            }}
            index={index}
            onDeleteSelection={onDeleteSelection}
          />
        ))}
        <Button
          variant="contained"
          style={{ width: '100%', height: '3rem', marginTop: '20px' }}
          // onClick={submitFeedbacks}
          // disabled={
          //   isXml ? feedbacks.length === 0 : charspans.length === 0
          // }
        >
          Submit Feedback
        </Button>
      </CardContent>
    </Card>
  );
}

export default FeedbackDashboard;
