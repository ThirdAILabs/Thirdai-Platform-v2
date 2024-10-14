'use client';

import { Container, Box, CircularProgress, Typography, Switch, FormControlLabel } from '@mui/material';
import { Button } from '@/components/ui/button';
import React, { MouseEventHandler, ReactNode, useEffect, useRef, useState } from 'react';
import { Card } from '@/components/ui/card';
import * as _ from 'lodash';
import { useTokenClassificationEndpoints } from '@/lib/backend';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import Fuse from 'fuse.js';
import * as XLSX from 'xlsx';
import Papa from 'papaparse';
import FeedbackDashboard from './FeedbackDashboard'

interface Token {
  text: string;
  tag: string;
}

interface HighlightColor {
  text: string;
  tag: string;
}

interface HighlightProps {
  currentToken: Token;
  nextToken?: Token | null;
  tagColors: Record<string, HighlightColor>;
  onMouseOver: MouseEventHandler;
  onMouseDown: MouseEventHandler;
  selecting: boolean;
  selected: boolean;
}

interface TagSelectorProps {
  open: boolean;
  choices: string[];
  onSelect: (tag: string) => void;
  onNewLabel: (newLabel: string) => Promise<void>;
  currentTag: string;
}

interface ParsedData {
  type: 'csv' | 'pdf' | 'other';
  content: string;
  rows?: { label: string; content: string }[];
}

const SELECTING_COLOR = '#EFEFEF';
const SELECTED_COLOR = '#DFDFDF';

function Highlight({
  currentToken,
  nextToken,
  tagColors,
  onMouseOver,
  onMouseDown,
  selecting,
  selected,
}: HighlightProps) {
  const [hover, setHover] = useState<boolean>(false);

  return (
    <>
      <span
        style={{
          backgroundColor:
            hover || selecting
              ? SELECTING_COLOR
              : selected
              ? SELECTED_COLOR
              : tagColors[currentToken.tag]?.text || 'transparent',
          padding: '2px',
          borderRadius: '2px',
          cursor: hover ? 'pointer' : 'default',
          userSelect: 'none',
        }}
        onMouseOver={(e) => {
          setHover(true);
          onMouseOver(e);
        }}
        onMouseLeave={(e) => {
          setHover(false);
        }}
        onMouseDown={onMouseDown}
      >
        {currentToken.text}
        {tagColors[currentToken.tag] && nextToken?.tag !== currentToken.tag && (
          <span
            style={{
              backgroundColor: tagColors[currentToken.tag].tag,
              color: 'white',
              fontSize: '11px',
              fontWeight: 'bold',
              borderRadius: '2px',
              marginLeft: '4px',
              padding: '5px 3px 1px 3px',
              marginBottom: '1px',
            }}
          >
            {currentToken.tag}
          </span>
        )}
      </span>
      <span
        style={{ cursor: hover ? 'pointer' : 'default', userSelect: 'none' }}
        onMouseOver={(e) => {
          setHover(true);
          onMouseOver(e);
        }}
        onMouseLeave={(e) => {
          setHover(false);
        }}
        onMouseDown={onMouseDown}
      >
        {' '}
      </span>
    </>
  );
}

function TagSelector({ open, choices, onSelect, onNewLabel, currentTag }: TagSelectorProps) {
  const [fuse, setFuse] = useState<Fuse<string>>(new Fuse([]));
  const [query, setQuery] = useState('');
  const [searchableChoices, setSearchableChoices] = useState<string[]>([]);

  useEffect(() => {
    const updatedChoices = choices.filter(choice => choice !== 'O');
    if (currentTag !== 'O') {
      updatedChoices.unshift('Delete TAG');
    }
    setSearchableChoices(updatedChoices);
    setFuse(new Fuse(updatedChoices));
  }, [choices, currentTag]);

  const searchResults = query !== '' 
    ? fuse.search(query).map((result) => result.item) 
    : searchableChoices;

  const makeDropdownMenuItem = (key: number, value: string, child: ReactNode) => (
    <DropdownMenuItem className="font-medium" key={key}>
      <button
        style={{ width: '100%', height: '100%', textAlign: 'left' }}
        onClick={async () => {
          const selectedTag = value === 'Delete TAG' ? 'O' : value;
          if (!choices.includes(selectedTag) && selectedTag !== 'O') {
            await onNewLabel(selectedTag);
          }
          onSelect(selectedTag);
          setQuery('');
        }}
      >
        {child}
      </button>
    </DropdownMenuItem>
  );

  return (
    <DropdownMenu open={open} modal={false}>
      <DropdownMenuTrigger>
        <span />
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <Input
          autoFocus
          className="font-medium"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          style={{ marginBottom: '5px' }}
          onKeyDown={(e) => {
            e.stopPropagation();
          }}
        />
        {searchResults.map((val, index) => makeDropdownMenuItem(index, val, val))}
        {query !== '' &&
          !searchResults.includes(query) &&
          query !== 'Delete TAG' &&
          makeDropdownMenuItem(
            searchResults.length,
            query,
            <>
              <span
                className="bg-accent font-medium"
                style={{
                  padding: '0 3px',
                  marginRight: '5px',
                  borderRadius: '2px',
                  fontWeight: 'bold',
                }}
              >
                New{' '}
              </span>{' '}
              {query}
            </>
          )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default function Interact() {
  const { predict, insertSample, addLabel, getLabels, getTextFromFile } = useTokenClassificationEndpoints();

  const [inputText, setInputText] = useState<string>('');
  const [annotations, setAnnotations] = useState<Token[]>([]);
  const [tagColors, setTagColors] = useState<Record<string, HighlightColor>>({});
  const [allLabels, setAllLabels] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [parsedData, setParsedData] = useState<ParsedData | null>(null);
  const [showHighlightedOnly, setShowHighlightedOnly] = useState(false);

  const [mouseDownIndex, setMouseDownIndex] = useState<number | null>(null);
  const [mouseUpIndex, setMouseUpIndex] = useState<number | null>(null);
  const [selecting, setSelecting] = useState<boolean>(false);
  const startIndex =
    mouseDownIndex !== null && mouseUpIndex !== null
      ? Math.min(mouseDownIndex, mouseUpIndex)
      : null;
  const endIndex =
    mouseDownIndex !== null && mouseUpIndex !== null
      ? Math.max(mouseDownIndex, mouseUpIndex)
      : null;
  const [selectedRange, setSelectedRange] = useState<[number, number] | null>(null);

  const triggers = useRef<(HTMLElement | null)[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [cachedTags, setCachedTags] = useState<{ [sentence: string]: Token[] }>({});

  useEffect(() => {
    const stopSelectingOnOutsideClick = () => {
      setSelecting(false);
      setSelectedRange(null);
    };
    window.addEventListener('mousedown', stopSelectingOnOutsideClick);
    return () => window.removeEventListener('mousedown', stopSelectingOnOutsideClick);
  }, []);

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputText(event.target.value);
    setParsedData(null);
    setAnnotations([]);
  };

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const fileExtension = file.name.split('.').pop()?.toLowerCase();
      setIsLoading(true);

      let parsed: ParsedData;
      if (fileExtension === 'csv') {
        parsed = await parseCSV(file);
      } else if (['pdf', 'docx'].includes(fileExtension ?? '')) {
        try {
          const content = await getTextFromFile(file);
          console.log('content', content);
          parsed = { 
            type: fileExtension === 'pdf' ? 'pdf' : 'other', 
            content: content.join('\n') 
          };
        } catch (error) {
          console.error('Error parsing file:', error);
          setIsLoading(false);
          return;
        }
      } else if (['xls', 'xlsx'].includes(fileExtension ?? '')) {
        const excelRows = await parseExcel(file);
        parsed = {
          type: 'csv',
          content: excelRows.map((row) => row.content).join('\n\n'),
          rows: excelRows,
        };
      } else {
        parsed = { type: 'other', content: await parseTXT(file) };
      }

      setInputText(parsed.content);
      setParsedData(parsed);
      handleRun(parsed.content, true);
      setIsLoading(false);
    }

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const rowsPerPage = 10;

  const parseCSV = (file: File): Promise<ParsedData> => {
    return new Promise((resolve, reject) => {
      Papa.parse(file, {
        complete: (results) => {
          const data = results.data as string[][];
          if (data.length < 2) {
            resolve({ type: 'csv', content: '', rows: [] });
            return;
          }
          const headers = data[0];
          const rows = data.slice(1);
          let parsedRows = rows
            .filter((row) => row.some((cell) => cell.trim() !== ''))
            .map((row, rowIndex) => {
              let content = headers
                .map((header, index) => `${header}: ${row[index] || ''}`)
                .join('\n');
              return {
                label: `Row ${rowIndex + 1}`,
                content: content,
              };
            });
          const fullContent = parsedRows.map((row) => row.content).join('\n\n');
          setTotalPages(Math.ceil(parsedRows.length / rowsPerPage));
          resolve({ type: 'csv', content: fullContent, rows: parsedRows });
        },
        error: reject,
      });
    });
  };

  const parseExcel = (file: File): Promise<{ label: string; content: string }[]> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (e) => {
        const data = new Uint8Array(e.target?.result as ArrayBuffer);
        const workbook = XLSX.read(data, { type: 'array' });
        const firstSheetName = workbook.SheetNames[0];
        const worksheet = workbook.Sheets[firstSheetName];
        const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 }) as (string | number | null)[][];

        if (jsonData.length < 2) {
          resolve([]);
          return;
        }

        const headers = jsonData[0].map(String);
        const rows = jsonData.slice(1);

        let parsedRows = rows
          .map((row, rowIndex) => {
            if (row.some((cell) => cell !== null && cell !== '')) {
              let content = headers
                .map((header, index) => {
                  const cellValue = row[index];
                  return `${header}: ${cellValue !== null && cellValue !== undefined ? cellValue : ''}`;
                })
                .join('\n');
              return {
                label: `Row ${rowIndex + 1}`,
                content: content,
              };
            }
            return null;
          })
          .filter((row): row is { label: string; content: string } => row !== null);

        setTotalPages(Math.ceil(parsedRows.length / rowsPerPage));
        resolve(parsedRows);
      };
      reader.onerror = reject;
      reader.readAsArrayBuffer(file);
    });
  };

  const parseTXT = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (e) => {
        resolve(e.target?.result as string);
      };
      reader.onerror = reject;
      reader.readAsText(file);
    });
  };

  const updateTagColors = (tags: string[][]) => {
    const pastels = ['#E5A49C', '#F6C886', '#FBE7AA', '#99E3B5', '#A6E6E7', '#A5A1E1', '#D8A4E2'];
    const darkers = ['#D34F3E', '#F09336', '#F7CF5F', '#5CC96E', '#65CFD0', '#597CE2', '#B64DC8'];

    setTagColors((existingColors) => {
      const colors = { ...existingColors };
      const newTags = Array.from(new Set(tags.flatMap((tokenTags) => tokenTags))).filter(
        (tag) => !existingColors[tag] && tag !== 'O'
      );
      newTags.forEach((tag, index) => {
        const i = Object.keys(existingColors).length + index;
        colors[tag] = {
          text: pastels[i % pastels.length],
          tag: darkers[i % darkers.length],
        };
      });
      return colors;
    });
  };

  const handleRun = async (text: string, isFileUpload: boolean = false) => {
    setIsLoading(true);
    try {
      const result = await predict(text);
      updateTagColors(result.predicted_tags);
      setAnnotations(
        _.zip(result.tokens, result.predicted_tags).map(([text, tag]) => ({
          text: text as string,
          tag: (tag as string[])[0],
        }))
      );

      if (!isFileUpload) {
        setParsedData({ type: 'other', content: text });
      }

      const labels = await getLabels();
      const filteredLabels = labels.filter(label => label !== 'O');
      setAllLabels(filteredLabels);
      updateTagColors([filteredLabels]);
    } catch (error) {
      console.error('Error during prediction or fetching labels:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const normalizeSentence = (sentence: string): string => {
    return sentence.replace(/[.,]$/, '').trim();
  };

  const handleNewLabel = async (newLabel: string) => {
    try {
      await addLabel({
        tags: [{ name: newLabel, description: `Description for ${newLabel}` }]
      });
      setAllLabels(prevLabels => [...prevLabels, newLabel]);
      console.log('New label added successfully');
    } catch (error) {
      console.error('Error adding new label:', error);
    }
  };

  const cacheNewTag = async (newTag: string) => {
    if (!selectedRange) return;

    const updatedTags = annotations.map((token, index) => ({
      text: token.text,
      tag: (selectedRange && index >= selectedRange[0] && index <= selectedRange[1]) ? newTag : token.tag
    }));

    const normalizedSentence = normalizeSentence(inputText);

    setCachedTags(prev => {
      const updatedCachedTags = {
        ...prev,
        [normalizedSentence]: updatedTags
      };

      if (!updatedTags.some(token => token.tag !== 'O')) {
        delete updatedCachedTags[normalizedSentence];
      }

      return updatedCachedTags;
    });

    setAnnotations(updatedTags);
    updateTagColors([[newTag]]);
    setSelectedRange(null);
    setMouseDownIndex(null);
    setMouseUpIndex(null);
    setSelecting(false);
  };

  const updateFeedbackDashboard = () => {
    setCachedTags(prev => {
      const updatedCachedTags = { ...prev };
      
      Object.keys(updatedCachedTags).forEach(sentence => {
        if (!updatedCachedTags[sentence].some(token => token.tag !== 'O')) {
          delete updatedCachedTags[sentence];
        }
      });

      return updatedCachedTags;
    });
  };

  useEffect(() => {
    updateFeedbackDashboard();
  }, [annotations]);

  const submitFeedback = async () => {
    try {
      for (const [sentence, tags] of Object.entries(cachedTags)) {
        // Debug statement to log what's being sent to the backend
        console.log('Submitting feedback for sentence:', sentence);
        console.log('Tags:', tags);
        
        const tokens = tags.map(t => t.text);
        console.log('Tokens:', tokens);
        
        const submission = {
          tokens: tokens,
          tags: tags.map(t => t.tag),
        };
        console.log('Submission to backend:', submission);

        await insertSample(submission);
      }
      console.log('All samples inserted successfully');
      setCachedTags({});
    } catch (error) {
      console.error('Error inserting samples:', error);
    }
  };

  const deleteFeedbackExample = (normalizedSentence: string) => {
    setCachedTags(prev => {
      const updatedCachedTags = { ...prev };
      Object.keys(updatedCachedTags).forEach(key => {
        if (updatedCachedTags[key].map(t => t.text).join(' ') === normalizedSentence) {
          delete updatedCachedTags[key];
        }
      });
      return updatedCachedTags;
    });
  };

  const toggleHighlightedOnly = (event: React.ChangeEvent<HTMLInputElement>) => {
    setShowHighlightedOnly(event.target.checked);
  };

  const isWordHighlighted = (word: string) => {
    return annotations.some(
      (token) => token.text.toLowerCase() === word.toLowerCase() && token.tag !== 'O'
    );
  };

  const renderCSVContent = (rows: { label: string; content: string }[]) => {
    const startIndex = (currentPage - 1) * rowsPerPage;
    const endIndex = startIndex + rowsPerPage;
    const visibleRows = rows.slice(startIndex, endIndex);

    return (
      <>
        {visibleRows.map((row, rowIndex) => {
          const columns = row.content.split('\n');
          const visibleColumns = columns.filter((column) => {
            const [columnName, ...columnContent] = column.split(':');
            const content = columnContent.join(':').trim();
            return !showHighlightedOnly || content.split(' ').some(isWordHighlighted);
          });

          if (visibleColumns.length === 0) {
            return null;
          }

          return (
            <div
              key={rowIndex}
              style={{
                marginBottom: '20px',
                padding: '10px',
                border: '1px solid #ccc',
                borderRadius: '5px',
              }}
            >
              <strong>{row.label}:</strong>
              {visibleColumns.map((column, columnIndex) => {
                const [columnName, ...columnContent] = column.split(':');
                const content = columnContent.join(':').trim();
                return (
                  <p key={columnIndex}>
                    <strong>{columnName}:</strong> {renderHighlightedContent(content)}
                  </p>
                );
              })}
            </div>
          );
        })}
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '20px' }}>
          <Button
            size="sm"
            onClick={() => setCurrentPage((prev) => Math.max(1, prev - 1))}
            disabled={currentPage === 1}
          >
            Previous
          </Button>
          <span>
            Page {currentPage} of {totalPages}
          </span>
          <Button
            size="sm"
            onClick={() => setCurrentPage((prev) => Math.min(totalPages, prev + 1))}
            disabled={currentPage === totalPages}
          >
            Next
          </Button>
        </div>
      </>
    );
  };

  const renderPDFContent = (content: string) => {
    return content.split('\n').map((paragraph, index) => (
      <React.Fragment key={index}>
        {paragraph === '' ? <br /> : renderHighlightedContent(paragraph)}
        <br />
      </React.Fragment>
    ));
  };

  const renderHighlightedContent = (content: string) => {
    const words = content.split(/\s+/);
    return words.map((word, wordIndex) => {
      const tokenIndex = annotations.findIndex(
        (token) => token.text.toLowerCase() === word.toLowerCase()
      );

      if (tokenIndex !== -1) {
        return (
          <React.Fragment key={wordIndex}>
            <Highlight
              currentToken={annotations[tokenIndex]}
              nextToken={annotations[tokenIndex + 1] || null}
              tagColors={tagColors}
              onMouseOver={(e) => {
                if (selecting) {
                  setMouseUpIndex(tokenIndex);
                }
              }}
              onMouseDown={(e) => {
                e.stopPropagation();
                setSelecting(true);
                setMouseDownIndex(tokenIndex);
                setMouseUpIndex(tokenIndex);
                setSelectedRange(null);
              }}
              selecting={
                selecting &&
                startIndex !== null &&
                endIndex !== null &&
                tokenIndex >= startIndex &&
                tokenIndex <= endIndex
              }
              selected={
                selectedRange !== null &&
                tokenIndex >= selectedRange[0] &&
                tokenIndex <= selectedRange[1]
              }
            />
            <TagSelector
              open={!!selectedRange && tokenIndex === selectedRange[1]}
              choices={allLabels}
              onSelect={cacheNewTag}
              onNewLabel={handleNewLabel}
              currentTag={annotations[tokenIndex].tag}
            />
          </React.Fragment>
        );
      }
      return showHighlightedOnly && annotations[tokenIndex]?.tag === 'O' ? null : (
        <span key={wordIndex}>{word} </span>
      );
    });
  };

  const renderContent = () => {
    if (!parsedData) return null;

    if (parsedData.type === 'csv' && parsedData.rows) {
      return renderCSVContent(parsedData.rows);
    } else if (parsedData.type === 'pdf') {
      return renderPDFContent(parsedData.content);
    } else {
      return renderHighlightedContent(parsedData.content);
    }
  };

  return (
    <Container style={{ 
      display: 'flex', 
      width: '90%', 
      maxWidth: '1200px', 
      marginTop: '20vh',
      paddingBottom: '100vh' // Add extra space at the bottom
    }}>
      <div style={{ flex: 2, marginRight: '20px' }}>
        <Box display="flex" justifyContent="center" alignItems="center" width="100%">
          <label htmlFor="file-upload" style={{ marginRight: '10px' }}>
            <Button size="sm" asChild>
              <span>Upload File</span>
            </Button>
          </label>
          <input
            ref={fileInputRef}
            id="file-upload"
            type="file"
            accept=".txt,.pdf,.docx,.csv,.xls,.xlsx"
            onChange={handleFileChange}
            style={{ display: 'none' }}
          />
          <Input
            autoFocus
            className="text-md"
            style={{ height: '3rem', flex: 1 }}
            value={inputText}
            onChange={handleInputChange}
            placeholder="Enter your text..."
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleRun(inputText);
              }
            }}
          />
          <Button
            size="sm"
            style={{ height: '3rem', marginLeft: '10px', padding: '0 20px' }}
            onClick={() => handleRun(inputText)}
          >
            Run
          </Button>
        </Box>

        <Typography variant="caption" display="block" mt={1}>
          Supported file types: .txt, .pdf, .docx, .csv, .xls, .xlsx
        </Typography>

        {annotations.length > 0 && (
          <Box mt={4} mb={2} display="flex" alignItems="center" justifyContent="flex-end">
            <FormControlLabel
              control={
                <Switch
                  checked={showHighlightedOnly}
                  onChange={toggleHighlightedOnly}
                  color="primary"
                />
              }
              label="Tagged Text Only"
            />
          </Box>
        )}

        {isLoading ? (
          <Box mt={4} display="flex" justifyContent="center">
            <CircularProgress />
          </Box>
        ) : (
          annotations.length > 0 && (
            <Box mt={4}>
              <Card
                className="p-7 text-start"
                style={{ lineHeight: 2 }}
                onMouseUp={(e) => {
                  setSelecting(false);
                  if (startIndex !== null && endIndex !== null) {
                    setSelectedRange([startIndex, endIndex]);
                    triggers.current[endIndex]?.click();
                  }
                }}
              >
                {renderContent()}
              </Card>
            </Box>
          )
        )}
      </div>
      <div style={{ flex: 1 }}>
        <Card className="p-7 text-start">
          <FeedbackDashboard
            cachedTags={cachedTags}
            tagColors={tagColors}
            deleteFeedbackExample={deleteFeedbackExample}
            submitFeedback={submitFeedback}
          />
        </Card>
      </div>
    </Container>
  );
}