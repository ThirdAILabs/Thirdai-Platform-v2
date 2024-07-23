"use client"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel
} from '@/components/ui/dropdown-menu';
import RAGQuestionForm from './RAGQuestionForm';
import NLPQuestionForm from './NLPQuestionForm';
import { useState } from 'react';

export default function NewModelPage() {
  const [modelType, setModelType] = useState<null|string>(null);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Model</CardTitle>
        <CardDescription>Create a new model with a few simple steps.</CardDescription>
      </CardHeader>
      <CardContent>
        <DropdownMenu>
          <div>
            {`Please select the problem type you want to address: `}
            <DropdownMenuTrigger asChild>
              <button className="dropdown-trigger-button">
              {
                modelType 
                ? 
                modelType
                :
                '...'
              }
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuLabel>Choose from following</DropdownMenuLabel>
              <DropdownMenuItem onClick={() => setModelType('NLP')}>NLP</DropdownMenuItem>
              <DropdownMenuItem onClick={() => setModelType('RAG')}>RAG</DropdownMenuItem>
            </DropdownMenuContent>
          </div>
        </DropdownMenu>

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
