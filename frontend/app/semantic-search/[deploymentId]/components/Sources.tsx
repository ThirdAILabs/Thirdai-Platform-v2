import React, { useContext, useEffect, useState } from 'react';
import { DropdownMenuContent, DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Button } from '@mui/material';
import Fuse from 'fuse.js';
import { ModelService, Source, PdfInfo } from '../modelServices';
import { ModelServiceContext } from '../Context';
import FileUploadModal from './FileUploadModal';

interface SourcesProps {
  sources: Source[];
  visible: boolean;
  setSources: (sources: Source[]) => void;
}

interface FuseSource {
  source: string;
  source_id: string;
  originalSource?: Source;
}

interface CloudUrl {
  type: 's3' | 'azure' | 'gcp';
  url: string;
}

const PAGE_SIZE = 10;

const Sources: React.FC<SourcesProps> = ({ sources, visible, setSources }) => {
  const [fuse, setFuse] = useState<Fuse<FuseSource> | null>(null);
  const [matches, setMatches] = useState<FuseSource[]>([]);
  const [open, setOpen] = useState<boolean>(false);
  const [currentPage, setCurrentPage] = useState<number>(0);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [lastSourcesLength, setLastSourcesLength] = useState<number>(sources.length);
  
  const modelService = useContext<ModelService | null>(ModelServiceContext);

  const totalPages = Math.ceil(matches.length / PAGE_SIZE);
  const startIndex = currentPage * PAGE_SIZE;
  const endIndex = Math.min(startIndex + PAGE_SIZE, matches.length);
  const currentDocs = matches.slice(startIndex, endIndex);

  function formatSource(source: string): string {
    const lowerSource = source.toLowerCase();
    if (
      lowerSource.endsWith('.pdf') ||
      lowerSource.endsWith('.docx') ||
      lowerSource.endsWith('.csv') ||
      lowerSource.endsWith('.txt') ||
      lowerSource.endsWith('.pptx') ||
      lowerSource.endsWith('.eml')
    ) {
      return source.split('/').pop() || source;
    }
    return source;
  }

  // Update matches when sources change
  useEffect(() => {
    const currentLength = sources.length;
    
    if (currentLength !== lastSourcesLength) {
      const fuseData = sources.map((source) => ({
        source: formatSource(source.source),
        source_id: source.source_id,
        originalSource: source
      }));
      
      setFuse(
        new Fuse(fuseData, {
          keys: ['source'],
          threshold: 0.3,
        })
      );
      
      setMatches(fuseData);
      setLastSourcesLength(currentLength);
    }
  }, [sources, lastSourcesLength]);

  // Update matches when visibility changes
  useEffect(() => {
    if (!visible) return;
    
    const fuseData = sources.map((source) => ({
      source: formatSource(source.source),
      source_id: source.source_id,
      originalSource: source
    }));
    setMatches(fuseData);
  }, [visible]);

  const handleSearchBarChangeEvent = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const value = e.target.value;
    setSearchTerm(value);
    setCurrentPage(0);

    if (!fuse || value.trim() === '') {
      const fuseData = sources.map((source) => ({
        source: formatSource(source.source),
        source_id: source.source_id,
        originalSource: source
      }));
      setMatches(fuseData);
      return;
    }
    
    const searchResults = fuse.search(value).map((res) => res.item);
    setMatches(searchResults);
  };

  const handleDeleteSource = (e: React.MouseEvent<HTMLButtonElement>, sourceId: string): void => {
    e.stopPropagation();
    if (modelService) {
      modelService.deleteSources([sourceId]);
      setSources(sources.filter(x => x.source_id !== sourceId));
    }
  };

  const handlePageChange = (newPage: number): void => {
    setCurrentPage(newPage);
  };

  const refreshSources = (): void => {
    if (modelService) {
      modelService.sources().then(setSources);
    }
  };

  const handleAddSources = async (
    selectedFiles: FileList | null,
    cloudUrls: CloudUrl[]
  ): Promise<void> => {
    if (!modelService) return;
    
    const filesArray = selectedFiles ? Array.from(selectedFiles) : [];
    await modelService.addSources(filesArray, cloudUrls);
    refreshSources();
  };

  const renderPageNumbers = (): React.ReactNode[] => {
    const maxVisiblePages = 5;
    let pageNumbers: number[] = [];

    if (totalPages <= maxVisiblePages) {
      pageNumbers = Array.from({ length: totalPages }, (_, i) => i);
    } else if (currentPage < 2) {
      pageNumbers = [0, 1, 2, 3, 4];
    } else if (currentPage > totalPages - 3) {
      pageNumbers = Array.from({ length: 5 }, (_, i) => totalPages - 5 + i);
    } else {
      pageNumbers = [
        currentPage - 2,
        currentPage - 1,
        currentPage,
        currentPage + 1,
        currentPage + 2,
      ];
    }

    return pageNumbers.map((pageNum) => (
      <Button 
        key={pageNum}
        size="small"
        variant={currentPage === pageNum ? "contained" : "outlined"}
        onClick={() => handlePageChange(pageNum)}
        className="min-w-[32px]"
      >
        {pageNum + 1}
      </Button>
    ));
  };

  return (
    <DropdownMenuContent
      className="w-[400px] max-h-[500px] overflow-hidden flex flex-col"
      align="start"
      side="bottom"
    >
      <div className="p-2 border-b">
        <Input
          autoFocus
          className="font-medium mb-2"
          placeholder="Filter documents by name..."
          value={searchTerm}
          onChange={handleSearchBarChangeEvent}
          onKeyDown={(e: React.KeyboardEvent) => e.stopPropagation()}
        />
        <Button
          className="w-full mt-2"
          onClick={() => setOpen(true)}
          variant="contained"
        >
          Add Documents
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto min-h-[200px]">
        {currentDocs.map((source, i) => (
          <DropdownMenuItem
            key={`${source.source_id}-${i}`}
            className="flex justify-between items-center p-2 hover:bg-gray-100 cursor-pointer"
            onClick={() => {
              if (source.originalSource?.source) {
                const fileExtension = source.originalSource.source.split('.').pop()?.toLowerCase();
                if (fileExtension === 'pdf' || fileExtension === 'docx') {
                  console.log('Opening source:', source.originalSource.source);
                  modelService?.openSource(source.originalSource.source);
                }
              }
            }}
          >
            <span className="truncate flex-1">
              {source.source}
            </span>
            <Button
              className="min-w-[40px] h-8 ml-2 bg-transparent hover:bg-red-500 text-red-500 hover:text-white border border-red-500"
              onClick={(e) => handleDeleteSource(e, source.source_id)}
            >
              ✕
            </Button>
          </DropdownMenuItem>
        ))}
      </div>

      {totalPages > 1 && (
        <div className="p-2 border-t flex items-center justify-between bg-white">
          <div className="text-sm text-gray-500">
            {`${startIndex + 1}-${endIndex} of ${matches.length}`}
          </div>
          <div className="flex gap-2">
            <Button
              size="small"
              disabled={currentPage === 0}
              onClick={() => handlePageChange(currentPage - 1)}
              variant="outlined"
            >
              Previous
            </Button>
            <div className="flex items-center gap-1">
              {renderPageNumbers()}
            </div>
            <Button
              size="small"
              disabled={currentPage >= totalPages - 1}
              onClick={() => handlePageChange(currentPage + 1)}
              variant="outlined"
            >
              Next
            </Button>
          </div>
        </div>
      )}

      <FileUploadModal
        isOpen={open}
        handleCloseModal={() => setOpen(false)}
        addSources={handleAddSources}
        refreshSources={refreshSources}
      />
    </DropdownMenuContent>
  );
};

export default Sources;