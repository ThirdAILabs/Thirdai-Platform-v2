'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { fetchFeedback } from '@/lib/backend';
import { useState, useEffect } from 'react';

interface TextPairsProps {
  timestamp: string;
  label1: string;
  label2: string;
  text1: string;
  text2: string;
}

function TextPairs({ timestamp, label1, label2, text1, text2 }: TextPairsProps) {
  return (
    <div
      className="text-md"
      style={{ display: 'flex', flexDirection: 'column', marginBottom: '10px' }}
    >
      <CardDescription>{timestamp}</CardDescription>
      <div className="text-md" style={{ display: 'flex', flexDirection: 'row' }}>
        <span style={{ fontWeight: 'bold', marginRight: '5px' }}>{label1}:</span>
        <span style={{}}>{text1}</span>
      </div>
      <div className="text-md" style={{ display: 'flex', flexDirection: 'row' }}>
        <span style={{ fontWeight: 'bold', marginRight: '5px' }}>{label2}:</span>
        <span>{text2}</span>
      </div>
    </div>
  );
}

export default function RecentFeedbacks() {
  const [recentUpvotes, setRecentUpvotes] = useState([]);
  const [recentAssociations, setRecentAssociations] = useState([]);

  //This will ensure the realtime fetching of feedbacks data from backend after every 5 seconds.
  useEffect(() => {
    getFeedbackData();
    // Set up the interval
    const intervalId = setInterval(() => {
      getFeedbackData();
    }, 5000);
    // Clear the interval on component unmount
    return () => clearInterval(intervalId);
  }, []);

  const getFeedbackData = async () => {
    const data = await fetchFeedback();
    console.log('recent feedback data -> ', data);
    setRecentUpvotes(data?.upvote);
    setRecentAssociations(data?.associate);
  };

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        width: '100%',
        padding: '0 2rem',
        boxSizing: 'border-box',
      }}
    >
      <Card style={{ width: '48%', minHeight: '45rem' }} className="pb-2">
        <CardHeader className="bg-blue-900 text-white mb-3">
          <CardTitle>Recent Upvotes</CardTitle>
          <CardDescription className="text-white">The latest user-provided upvotes</CardDescription>
        </CardHeader>
        <CardContent style={{ overflowY: 'auto', maxHeight: '45rem' }}>
          {recentUpvotes ? (
            recentUpvotes.map(({ timestamp, query, reference_id, reference_text }, idx) => (
              <TextPairs
                key={idx}
                timestamp={timestamp}
                label1="Query"
                label2="Upvote"
                text1={query}
                text2={reference_text}
              />
            ))
          ) : (
            <div className="flex flex-col justify-center items-center h-full mt-20">
              <img src="/no-Data-Png.png" alt="No Data Available" className="mb-4" />
              <span className="font-mono italic">Oops! No upvotes data available.</span>
            </div>
          )}
        </CardContent>
      </Card>
      <Card style={{ width: '48%', minHeight: '45rem' }}>
        <CardHeader className="bg-blue-900 text-white mb-3">
          <CardTitle>Recent Associations</CardTitle>
          <CardDescription className="text-white">
            The latest user-provided associations
          </CardDescription>
        </CardHeader>
        <CardContent style={{ overflowY: 'auto', maxHeight: '45rem' }}>
          {recentAssociations ? (
            recentAssociations.map(({ timestamp, source, target }, idx) => (
              <TextPairs
                key={idx}
                timestamp={timestamp}
                label1="Source"
                label2="Target"
                text1={source}
                text2={target}
              />
            ))
          ) : (
            <div className="flex flex-col justify-center items-center h-full mt-20">
              <img src="/no-Data-Png.png" alt="No Data Available" className="mb-4" />
              <span className="font-mono italic">Oops! No Associations data available.</span>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
