'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { reformulations } from './mock_samples';
import useRollingSamples from './rolling';
import { fetchFeedback } from '@/lib/backend'
import { useState, useEffect } from 'react';
import { grey } from '@mui/material/colors';

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
        console.log("recent feedback data -> ", data);
        setRecentUpvotes(data?.upvote);
        setRecentAssociations(data?.associate);
    };


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
            <Card style={{ width: '32.5%', minHeight: '45rem' }}>
                <CardHeader className='bg-blue-900 text-white'>
                    <CardTitle>Recent Upvotes</CardTitle>
                    <CardDescription className='text-white'>The latest user-provided upvotes</CardDescription>
                </CardHeader>
                <CardContent style={{ overflowY: 'auto', maxHeight: '45rem' }}>
                    {recentUpvotes ? recentUpvotes.map(({ timestamp, query, reference_id, reference_text }, idx) => (
                        <TextPairs
                            key={idx}
                            timestamp={timestamp}
                            label1="Query"
                            label2="Upvote"
                            text1={query}
                            text2={reference_text}
                        />
                    )) : (
                        <div className="flex flex-col justify-center items-center h-full mt-20">
                            <img src="/no-Data-Png.png" alt="No Data Available" className="mb-4" />
                            <span className='font-mono italic'>Oops! No upvotes data available.</span>
                        </div>
                    )}
                </CardContent>
            </Card>
            <Card style={{ width: '32.5%', minHeight: '45rem' }}>
                <CardHeader className='bg-blue-900 text-white'>
                    <CardTitle>Recent Associations</CardTitle>
                    <CardDescription className='text-white'>The latest user-provided associations</CardDescription>
                </CardHeader>
                <CardContent style={{ overflowY: 'auto', maxHeight: '45rem' }}>
                    {recentAssociations ? recentAssociations.map(({ timestamp, source, target }, idx) => (
                        <TextPairs
                            key={idx}
                            timestamp={timestamp}
                            label1="Source"
                            label2="Target"
                            text1={source}
                            text2={target}
                        />
                    )) : (
                        <div className="flex flex-col justify-center items-center h-full mt-20">
                            <img src="/no-Data-Png.png" alt="No Data Available" className="mb-4" />
                            <span className='font-mono italic'>Oops! No Associations data available.</span>
                        </div>
                    )}
                </CardContent>
            </Card>
            <Card style={{ width: '32.5%', minHeight: '45rem' }}>
                <CardHeader className='bg-blue-900 text-white'>
                    <CardTitle>Recent Query Reformulations</CardTitle>
                    <CardDescription className='text-white'>The latest queries that required reformulation</CardDescription>
                </CardHeader>
                <CardContent style={{ overflowY: 'auto', maxHeight: '45rem' }}>
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