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
  const [files, setFiles] = useState<File[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files);
      const pdfFiles = selectedFiles.filter(file => file.type === 'application/pdf');
      
      if (pdfFiles.length === 0) {
        toast.error('Please upload PDF files only');
        e.target.value = '';
        return;
      }

      setFiles(pdfFiles);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (files.length === 0) {
      toast.error('Please select at least one file');
      return;
    }

    setIsSubmitting(true);
    try {
      const deploymentId = params.deploymentId as string;
      // Process each file
      for (const file of files) {
        await createReport(deploymentId, file);
      }
      toast.success(`Successfully created ${files.length} report${files.length > 1 ? 's' : ''}`);
      router.push(`/token-classification/${deploymentId}?tab=jobs`);
    } catch (error) {
      console.error('Error creating reports:', error);
      toast.error('Failed to create reports');
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
              Upload PDF Files
            </Label.Root>
            <Input
              id="file"
              type="file"
              accept=".pdf"
              multiple
              onChange={handleFileChange}
              className="cursor-pointer"
            />
            <p className="text-sm text-muted-foreground">
              You can upload one or more PDF files for processing
            </p>
            {files.length > 0 && (
              <div className="mt-2">
                <p className="text-sm font-medium">Selected files:</p>
                <ul className="text-sm text-muted-foreground list-disc list-inside">
                  {files.map((file, index) => (
                    <li key={index}>{file.name}</li>
                  ))}
                </ul>
              </div>
            )}
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
            <Button type="submit" disabled={files.length === 0 || isSubmitting}>
              {isSubmitting ? 'Creating...' : `Create Job${files.length > 1 ? 's' : ''}`}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
} 