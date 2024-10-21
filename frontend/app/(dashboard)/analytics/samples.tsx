'use client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { reformulations } from './mock_samples';
import useRollingSamples from './rolling';
import { fetchFeedback } from '@/lib/backend';
import { useEffect, useState } from 'react';


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


interface ReformulationProps {
  timestamp: string;
  original: string;
  reformulations: string[];
}


function Reformulation({ timestamp, original, reformulations }: ReformulationProps) {
  return (
    <div
      className="text-md"
      style={{ display: 'flex', flexDirection: 'column', marginBottom: '10px' }}
    >
      <CardDescription>{timestamp}</CardDescription>
      <span className="text-md" style={{ fontWeight: 'bold' }}>
        {original}
      </span>
      {reformulations.map((r, i) => (
        <span key={i} className="text-md" style={{ marginLeft: '10px' }}>
          {r}
        </span>
      ))}
    </div>
  );
}




export default function RecentSamples() {
  const [upvotes, setUpvotes] = useState([]);
  const [associates, setAssociates] = useState([]);


  useEffect(() => {
    getFeedbackData();
  }, [])


  const getFeedbackData = async () => {
    const data = await fetchFeedback();
    setUpvotes(data.upvote);
    setAssociates(data.associate);
  }


  const recentReformulations = useRollingSamples(
   /* samples= */ reformulations,
   /* numSamples= */ 3,
   /* maxNewSamples= */ 1,
   /* probabilityNewSamples= */ 0.4,
   /* intervalSeconds= */ 2
  );


  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'space-between',
        width: '100%',
      }}
    >
      <Card style={{ width: '32.5%', height: '45rem' }}>
        <CardHeader>
          <CardTitle>Recent Upvotes</CardTitle>
          <CardDescription>The latest user-provided upvotes</CardDescription>
        </CardHeader>
        {(upvotes.length !== 0) && <CardContent>
          {upvotes.map(({ timestamp, query, reference_text }, idx) => (
            <TextPairs
              key={idx}
              timestamp={timestamp}
              label1="Query"
              label2="Upvote"
              text1={query}
              text2={reference_text}
            />
          ))}
        </CardContent>}


      </Card>
      <Card style={{ width: '32.5%', height: '45rem' }}>
        <CardHeader>
          <CardTitle>Recent Associations</CardTitle>
          <CardDescription>The latest user-provided associations</CardDescription>
        </CardHeader>
        <CardContent>
          {associates.map(({ timestamp, source, target }, idx) => (
            <TextPairs
              key={idx}
              timestamp={timestamp}
              label1="Source"
              label2="Target"
              text1={source}
              text2={target}
            />
          ))}
        </CardContent>
      </Card>
      <Card style={{ width: '32.5%', height: '45rem' }}>
        <CardHeader>
          <CardTitle>Recent Query Reformulations</CardTitle>
          <CardDescription>The latest queries that required reformulation</CardDescription>
        </CardHeader>
        <CardContent>
          {recentReformulations.map(({ timestamp, original, reformulations }, idx) => (
            <Reformulation
              key={idx}
              timestamp={timestamp}
              original={original}
              reformulations={reformulations}
            />
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
