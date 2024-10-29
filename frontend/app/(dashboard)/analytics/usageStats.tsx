'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import Link from 'next/link';
import { Button } from '@mui/material';
import _ from 'lodash';

export default function UsageStats() {

    const thirdaiPlatformBaseUrl = _.trim(process.env.THIRDAI_PLATFORM_BASE_URL!, '/');
    const grafanaUrl = `${thirdaiPlatformBaseUrl}/grafana`;

    const panelURL = grafanaUrl;
    return (
        <>
            <Card>
                <CardHeader>
                    <CardTitle>System Status</CardTitle>
                    <CardDescription>Monitor real-time usage and system improvements.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {/* <UsageDurationChart data={usageDurationData} />
            <UsageFrequencyChart data={usageFrequencyData} />
            <ReformulatedQueriesChart data={reformulatedQueriesData} /> */}
                        <iframe src="http://localhost/grafana/d-solo/be1n22yidbdvkc/analytics?orgId=1&var-Model_ID=01204a39-f729-4b43-8385-89415e870fc8&from=1730106790723&to=1730121155204&theme=light&panelId=2"
                            width="450" height="200"></iframe>
                    </div>

                    <div className="mt-4 flex justify-center items-center">
                        <Link href={grafanaUrl} passHref legacyBehavior>
                            <a target="_blank" rel="noopener noreferrer">
                                <Button variant="contained">See more system stats</Button>
                            </a>
                        </Link>
                    </div>
                </CardContent>
            </Card>

            <div className="mt-6">
                <h2 className="text-lg font-semibold mb-2">Usage Duration Trend</h2>
                <p className="mb-4">
                    Over the past few months, we&apos;ve observed a steady increase in the duration users
                    spend on the system. This indicates that users find the search system increasingly
                    valuable and are willing to engage with it for longer periods.
                </p>

                <h2 className="text-lg font-semibold mb-2">User Usage Frequency</h2>
                <p className="mb-4">
                    The frequency with which users interact with the system has remained relatively stable,
                    with a slight upward trend. This consistency shows that users continue to rely on the
                    search system regularly.
                </p>

                <h2 className="text-lg font-semibold mb-2">Reformulated Queries</h2>
                <p className="mb-4">
                    While the number of reformulated queries fluctuates, there&apos;s an overall downward
                    trend. This suggests that users are becoming more adept at finding the information they
                    need on the first attempt, indicating improved search accuracy and user satisfaction.
                </p>
            </div>
        </>
    );
}
