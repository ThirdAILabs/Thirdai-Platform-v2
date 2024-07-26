import { NextRequest, NextResponse } from 'next/server';
import { insertModel } from '@/lib/db';

export async function POST(req: NextRequest) {
  try {
    const modelData = await req.json();

    // Convert `trainedAt` to a Date object if it's not already
    modelData.trainedAt = new Date(modelData.trainedAt);

    const result = await insertModel(modelData);
    return NextResponse.json({ success: true, result });
  } catch (error) {
    console.error('Error inserting model:', error);

    // Assuming the error is an instance of Error
    if (error instanceof Error) {
      return NextResponse.json({ success: false, error: error.message }, { status: 500 });
    }

    // If the error is not an instance of Error, return a generic message
    return NextResponse.json({ success: false, error: 'An unknown error occurred' }, { status: 500 });
  }
}
