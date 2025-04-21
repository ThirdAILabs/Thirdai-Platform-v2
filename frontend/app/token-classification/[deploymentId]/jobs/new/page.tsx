'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import * as Label from '@radix-ui/react-label';
import { createReport } from '@/lib/backend';
import { toast } from 'sonner';

export default function NewJob() {
  const router = useRouter();
  const params = useParams();
  const [file, setFile] = useState<File | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const selectedFile = e.target.files[0];
      if (selectedFile.type === 'application/pdf') {
        setFile(selectedFile);
      } else {
        toast.error('Please upload a PDF file');
        e.target.value = '';
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      toast.error('Please select a file');
      return;
    }

    setIsSubmitting(true);
    try {
      const deploymentId = params.deploymentId as string;
      await createReport(deploymentId, file);
      toast.success('Report created successfully');
      router.back();
    } catch (error) {
      console.error('Error creating report:', error);
      toast.error('Failed to create report');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="container mx-auto py-8">
      <Card className="p-6">
        <h1 className="text-2xl font-semibold mb-6">Create New Job</h1>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <Label.Root htmlFor="file" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
              Upload PDF File
            </Label.Root>
            <Input
              id="file"
              type="file"
              accept=".pdf"
              onChange={handleFileChange}
              className="cursor-pointer"
            />
            <p className="text-sm text-muted-foreground">
              Please upload a PDF file for processing
            </p>
          </div>
          <div className="flex justify-end space-x-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.back()}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={!file || isSubmitting}>
              {isSubmitting ? 'Creating...' : 'Create Job'}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
} 