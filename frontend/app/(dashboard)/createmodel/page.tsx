"use client"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import RAGQuestionForm from './RAGQuestionForm';
import NLPQuestionForm from './NLPQuestionForm';
import { useState } from 'react';

export default function NewModelPage() {
  const [modelType, setModelType] = useState('RAG');

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Model</CardTitle>
        <CardDescription>Create a new model with a few simple steps.</CardDescription>
      </CardHeader>
      <CardContent>
        <span className="block text-lg font-semibold mb-2">Use case</span>
        <div className="mb-4">
          <label htmlFor="modelType" className="block text-sm font-medium text-gray-700">Please select the model type based on your use case:</label>
          <select
            id="modelType"
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
            value={modelType}
            onChange={(e)=>setModelType(e.target.value)}
          >
            <option value="">-- Please choose an option --</option>
            <option value="NLP">NLP (Natural Language Processing)</option>
            <option value="RAG">RAG (Retrieval Augmented Generation)</option>
          </select>
        </div>

        {modelType && (
          <div>
            {modelType === 'RAG' && <RAGQuestionForm />}
            {modelType === 'NLP' && <NLPQuestionForm />}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
