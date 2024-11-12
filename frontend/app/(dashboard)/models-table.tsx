'use client';
import React, { useEffect, useState } from 'react';
import { TableHead, TableRow, TableHeader, TableBody, Table, TableCell } from '@/components/ui/table';
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { WorkFlow } from './workflow';
import { useRouter } from 'next/navigation';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@mui/material';
import { fetchWorkflows, Workflow } from '@/lib/backend';
import { ArrowRight } from 'lucide-react';

export function ModelsTable({ searchStr, offset }: { searchStr: string; offset: number }) {
  // Hardcode the model display
  let modelsPerPage = 5;

  const [currentPage, setCurrentPage] = useState(Math.ceil(offset / modelsPerPage) + 1);

  let router = useRouter();

  function prevPage() {
    if (offset >= modelsPerPage) {
      const newOffset = offset - modelsPerPage;
      router.push(`/?q=${searchStr}&offset=${newOffset}`, { scroll: false });
    }
  }

  function nextPage() {
    if (offset + modelsPerPage < totalWorkflows) {
      const newOffset = offset + modelsPerPage;
      router.push(`/?q=${searchStr}&offset=${newOffset}`, { scroll: false });
    }
  }

  const [workflows, setWorkflows] = useState<Workflow[]>([]);

  useEffect(() => {
    async function getWorkflows() {
      try {
        const fetchedWorkflows = await fetchWorkflows();
        setWorkflows(fetchedWorkflows);
      } catch (err) {
        if (err instanceof Error) {
          console.log(err.message);
        } else {
          console.log('An unknown error occurred');
        }
      }
    }

    // Call the function immediately
    getWorkflows();

    const intervalId = setInterval(getWorkflows, 3000);

    return () => clearInterval(intervalId);
  }, []);

  const filteredWorkflows = workflows.filter(
    (workflow) =>
      workflow.model_name.toLowerCase().includes(searchStr.toLowerCase()) && workflow.type !== 'ndb'
  );
  const totalWorkflows = filteredWorkflows.length;
  const displayedWorkflows = filteredWorkflows.slice(offset, offset + modelsPerPage);

  const [isCollapsedList, setIsCollapsedList] = useState(Array(100).fill(true));

  // Toggle the value at a specific index
  const toggleIsCollapsedList = (index: number) => {
    setIsCollapsedList((prevList) =>
      prevList.map((item, idx) => (idx === index ? !item : item))
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>App Catalog</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="text-center">Name</TableHead>
              <TableHead className="text-center">Status</TableHead>
              <TableHead className="hidden md:table-cell text-center">Type</TableHead>
              <TableHead className="hidden md:table-cell text-center">Published on</TableHead>
              <TableHead className="hidden md:table-cell text-center">Action</TableHead>
              <TableHead className="hidden md:table-cell text-center">Details</TableHead>
              <TableHead className="text-center">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {displayedWorkflows
              .sort((a, b) => a.model_name.localeCompare(b.model_name)) // Sort by name alphabetically
              .map((workflow, index) => (
                <React.Fragment key={index + 200}>
                  <WorkFlow
                    workflow={workflow}
                    handleCollapse={toggleIsCollapsedList}
                    index={index}
                  />
                  {!isCollapsedList[index] && (
                    <TableRow>
                      <TableCell />  {/*To shift the dependency table slightly right*/}
                      <TableCell colSpan={3}>
                        <div
                          style={{
                            paddingLeft: '1rem', // Indentation
                            border: '1px solid #ccc', // Vertical line to show nesting
                            backgroundColor: '#f9f9f9', // Light background color for distinction
                            borderRadius: '10px', // Rounded corners
                            margin: '0.5rem 0', // Space around collapsed content
                            padding: '0.5rem',
                          }}
                        >
                          {workflow.dependencies.map((dependency) =>
                            workflows.map((item) => {
                              if (dependency.model_id === item.model_id) {
                                return (
                                  <WorkFlow
                                    key={item.model_id}
                                    workflow={item}
                                    handleCollapse={toggleIsCollapsedList}
                                    index={index}
                                  />
                                );
                              }
                              return null; // Prevents rendering undefined items
                            })
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </React.Fragment>
              ))}
          </TableBody>
        </Table>
      </CardContent>
      <CardFooter>
        <form className="flex items-center w-full justify-between">
          <div className="text-xs text-muted-foreground">
            Showing{' '}
            <strong>
              {Math.min(offset + 1, totalWorkflows)}-
              {Math.min(offset + modelsPerPage, totalWorkflows)}
            </strong>{' '}
            of <strong>{totalWorkflows}</strong> workflows
          </div>
          <div className="flex">
            <Button
              onClick={prevPage}
              variant="contained"
              color="error"
              type="button"
              disabled={offset <= 0}
            >
              <ChevronLeft className="mr-2 h-4 w-4" />
              Prev
            </Button>
            <Button
              onClick={nextPage}
              className="ml-5"
              variant="contained"
              type="button"
              disabled={offset + modelsPerPage >= totalWorkflows}
            >
              Next
              <ChevronRight className="ml-2 h-4 w-4" />
            </Button>
          </div>
        </form>
      </CardFooter>
    </Card>
  );
}
