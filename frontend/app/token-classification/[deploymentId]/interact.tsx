'use client';

import React, { MouseEventHandler, ReactNode, useEffect, useRef, useState } from 'react';
import {
  Container,
  Box,
  CircularProgress,
  Typography,
  Switch,
  FormControlLabel,
} from '@mui/material';
import { Button } from '@/components/ui/button';
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
import mammoth from 'mammoth'; // DOCX Parsing
import * as XLSX from 'xlsx';
import Papa from 'papaparse';
// @ts-ignore
import * as pdfjsLib from 'pdfjs-dist/build/pdf'; // PDF Parsing
// @ts-ignore
import pdfjsWorker from 'pdfjs-dist/build/pdf.worker.entry'; // PDF Worker
pdfjsLib.GlobalWorkerOptions.workerSrc = pdfjsWorker;

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

interface TagSelectorProps {
  open: boolean;
  choices: string[];
  onSelect: (tag: string) => void;
  onNewLabel: (newLabel: string) => Promise<void>;
  currentTag: string;
}

function TagSelector({ open, choices, onSelect, onNewLabel, currentTag }: TagSelectorProps) {
  const [fuse, setFuse] = useState<Fuse<string>>(new Fuse([]));
  const [query, setQuery] = useState('');
  const [searchableChoices, setSearchableChoices] = useState<string[]>([]);

  useEffect(() => {
    // Filter out 'O' from choices and add 'Delete TAG' if currentTag is not 'O'
    const updatedChoices = choices.filter((choice) => choice !== 'O');
    if (currentTag !== 'O') {
      updatedChoices.unshift('Delete TAG');
    }
    setSearchableChoices(updatedChoices);
    setFuse(new Fuse(updatedChoices));
  }, [choices, currentTag]);

  const searchResults =
    query !== '' ? fuse.search(query).map((result) => result.item) : searchableChoices;

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
            /* key= */ searchResults.length,
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
  const { predict, insertSample, addLabel, getLabels } = useTokenClassificationEndpoints();

  const [inputText, setInputText] = useState<string>('');
  const [annotations, setAnnotations] = useState<Token[]>([]);

  const [tagColors, setTagColors] = useState<Record<string, HighlightColor>>({});
  const [allLabels, setAllLabels] = useState<string[]>([]);

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

  const [parsedData, setParsedData] = useState<ParsedData | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [parsedRows, setParsedRows] = useState<{ label: string; content: string }[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [showHighlightedOnly, setShowHighlightedOnly] = useState(false);

  useEffect(() => {
    const stopSelectingOnOutsideClick = () => {
      setSelecting(false);
      setSelectedRange(null);
    };
    window.addEventListener('mousedown', stopSelectingOnOutsideClick);
    return stopSelectingOnOutsideClick;
  }, []);

  const handleInputChange = (event: any) => {
    setInputText(event.target.value);
    setParsedData(null);
    setAnnotations([]);
  };

  // Parsing functions
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
        const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 }) as (
          | string
          | number
          | null
        )[][];

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
                  return `${header}: ${
                    cellValue !== null && cellValue !== undefined ? cellValue : ''
                  }`;
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

  interface PDFTextItem {
    text: string;
    x: number;
    y: number;
    fontName: string;
    height: number;
  }

  const parsePDF = async (file: File): Promise<ParsedData> => {
    const loadingTask = pdfjsLib.getDocument(URL.createObjectURL(file));
    const pdf = await loadingTask.promise;
    let fullText = '';

    for (let pageNum = 1; pageNum <= pdf.numPages; pageNum++) {
      const page = await pdf.getPage(pageNum);
      const textContent = await page.getTextContent();
      let lastY = Infinity;
      let lastHeight = 0;
      let currentParagraph = '';
      let paragraphs: string[] = [];

      const pageItems = textContent.items
        .map((item: any) => ({
          text: item.str,
          x: item.transform[4],
          y: item.transform[5],
          fontName: item.fontName,
          height: item.height,
        }))
        .sort((a: PDFTextItem, b: PDFTextItem) => b.y - a.y || a.x - b.x);

      pageItems.forEach((curr: PDFTextItem, index: number) => {
        const verticalGap = lastY - (curr.y + curr.height);

        if (index > 0) {
          if (verticalGap > Math.max(lastHeight, curr.height) * 1.5) {
            if (currentParagraph.trim() !== '') {
              paragraphs.push(currentParagraph.trim());
              currentParagraph = '';
            }
            if (paragraphs.length > 0) {
              paragraphs.push('');
            }
          }
          if (curr.x - (pageItems[index - 1].x + pageItems[index - 1].text.length * 5) > 10) {
            currentParagraph += ' ';
          }
        }

        currentParagraph += curr.text;
        lastY = curr.y;
        lastHeight = curr.height;
      });

      if (currentParagraph.trim() !== '') {
        paragraphs.push(currentParagraph.trim());
      }

      fullText += paragraphs.join('\n') + '\n';
    }

    fullText = fullText.replace(/^PDF:\s*/i, '').trim();
    return { type: 'pdf', content: fullText };
  };

  const parseDOCX = async (file: File) => {
    const arrayBuffer = await file.arrayBuffer();
    const result = await mammoth.extractRawText({ arrayBuffer });
    return result.value;
  };

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const fileExtension = file.name.split('.').pop()?.toLowerCase();
      setIsLoading(true);

      let parsed: ParsedData;
      if (fileExtension === 'csv') {
        parsed = await parseCSV(file);
      } else if (fileExtension === 'pdf') {
        parsed = await parsePDF(file);
      } else if (['xls', 'xlsx'].includes(fileExtension ?? '')) {
        const excelRows = await parseExcel(file);
        parsed = {
          type: 'csv',
          content: excelRows.map((row) => row.content).join('\n\n'),
          rows: excelRows,
        };
      } else {
        const content =
          fileExtension === 'docx' ? await parseDOCX(file) : await parseTXT(file);
        parsed = { type: 'other', content };
      }

      setInputText(parsed.content);
      setParsedData(parsed);
      await handleRun(parsed.content, true); // Pass true to indicate it's a file upload
      setIsLoading(false);
    }

    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
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
    try {
      const result = await predict(text);
      updateTagColors(result.predicted_tags);
      setAnnotations(
        _.zip(result.tokens, result.predicted_tags).map(([text, tag]) => ({
          text: text as string,
          tag: (tag as string[])[0],
        }))
      );

      // Fetch labels after prediction
      const labels = await getLabels();
      const filteredLabels = labels.filter((label) => label !== 'O');
      setAllLabels(filteredLabels);
      updateTagColors([filteredLabels]);

      if (!isFileUpload && !parsedData) {
        setParsedData({ type: 'other', content: text });
      }

      setIsLoading(false);
    } catch (error) {
      console.error('Error during prediction or fetching labels:', error);
      setIsLoading(false);
    }
  };

  const handleNewLabel = async (newLabel: string) => {
    try {
      await addLabel({
        tags: [{ name: newLabel, description: `Description for ${newLabel}` }],
      });
      setAllLabels((prevLabels) => [...prevLabels, newLabel]);
      console.log('New label added successfully');
    } catch (error) {
      console.error('Error adding new label:', error);
    }
  };

  const finetuneTags = async (newTag: string) => {
    if (!selectedRange) return;

    const updatedTags = annotations.map((token, index) => ({
      text: token.text,
      tag:
        selectedRange && index >= selectedRange[0] && index <= selectedRange[1]
          ? newTag
          : token.tag,
    }));

    setAnnotations(updatedTags);
    updateTagColors([[newTag]]);
    setSelectedRange(null);
    setMouseDownIndex(null);
    setMouseUpIndex(null);
    setSelecting(false);
  };

  // Modify the renderContent function to use renderHighlightedContent directly
  const renderContent = () => {
    if (!parsedData) return null;

    if (parsedData.type === 'csv' && parsedData.rows) {
      return renderCSVContent(parsedData.rows);
    } else if (parsedData.type === 'pdf') {
      return renderPDFContent(parsedData.content);
    } else {
      return (
        <div style={{ whiteSpace: 'pre-wrap' }}>
          {renderHighlightedContent(parsedData.content)}
        </div>
      );
    }
  };

  const toggleHighlightedOnly = (event: React.ChangeEvent<HTMLInputElement>) => {
    setShowHighlightedOnly(event.target.checked);
  };

  const isWordHighlighted = (word: string) => {
    return annotations.some(
      (token) => token.text.toLowerCase() === word.toLowerCase() && token.tag !== 'O'
    );
  };

  // Adjust renderCSVContent and renderPDFContent to use renderHighlightedContent appropriately
  const renderCSVContent = (rows: { label: string; content: string }[]) => {
    return rows.map((row, rowIndex) => (
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
        <div style={{ whiteSpace: 'pre-wrap' }}>
          {renderHighlightedContent(row.content)}
        </div>
      </div>
    ));
  };

  const renderPDFContent = (content: string) => {
    return (
      <div style={{ whiteSpace: 'pre-wrap' }}>
        {renderHighlightedContent(content)}
      </div>
    );
  };

  const renderHighlightedContent = (content: string) => {
    // Instead of splitting content and trying to match words,
    // we can directly map over the annotations.

    return annotations.map((token, index) => (
      <React.Fragment key={index}>
        <Highlight
          currentToken={token}
          nextToken={annotations[index + 1] || null}
          tagColors={tagColors}
          onMouseOver={(e) => {
            if (selecting) {
              setMouseUpIndex(index);
            }
          }}
          onMouseDown={(e) => {
            e.stopPropagation();
            setSelecting(true);
            setMouseDownIndex(index);
            setMouseUpIndex(index);
            setSelectedRange(null);
          }}
          selecting={
            selecting &&
            startIndex !== null &&
            endIndex !== null &&
            index >= startIndex &&
            index <= endIndex
          }
          selected={
            selectedRange !== null &&
            index >= selectedRange[0] &&
            index <= selectedRange[1]
          }
        />
        <TagSelector
          open={!!selectedRange && index === selectedRange[1]}
          choices={allLabels}
          onSelect={cacheNewTag}
          onNewLabel={handleNewLabel}
          currentTag={token.tag}
        />
      </React.Fragment>
    ));
  };

  // Feedback Dashboard Logic
  const [cachedTags, setCachedTags] = useState<{ [sentence: string]: Token[] }>({});

  const normalizeSentence = (sentence: string): string => {
    return sentence.replace(/[.,]$/, '').trim();
  };

  const cacheNewTag = async (newTag: string) => {
    if (!selectedRange) return;

    const updatedTags = annotations.map((token, index) => ({
      text: token.text,
      tag:
        selectedRange && index >= selectedRange[0] && index <= selectedRange[1]
          ? newTag
          : token.tag,
    }));

    const normalizedSentence = normalizeSentence(inputText);

    setCachedTags((prev) => {
      const updatedCachedTags = {
        ...prev,
        [normalizedSentence]: updatedTags,
      };

      if (!updatedTags.some((token) => token.tag !== 'O')) {
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
    setCachedTags((prev) => {
      const updatedCachedTags = { ...prev };

      Object.keys(updatedCachedTags).forEach((sentence) => {
        if (!updatedCachedTags[sentence].some((token) => token.tag !== 'O')) {
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
        await insertSample({
          tokens: sentence.split(' '),
          tags: tags.map((t) => t.tag),
        });
      }
      console.log('All samples inserted successfully');
      setCachedTags({});
    } catch (error) {
      console.error('Error inserting samples:', error);
    }
  };

  const deleteFeedbackExample = (sentenceToDelete: string) => {
    setCachedTags((prev) => {
      const updatedCachedTags = { ...prev };
      delete updatedCachedTags[sentenceToDelete];
      return updatedCachedTags;
    });
  };

  return (
    <Container style={{ display: 'flex', paddingTop: '20vh', width: '90%', maxWidth: '1200px' }}>
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
              style={{
                lineHeight: 1.6,
                fontWeight: 'normal',
                whiteSpace: 'pre-wrap',
                wordWrap: 'break-word',
              }}
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
      {/* Feedback Dashboard */}
      <div style={{ flex: 1 }}>
        <Card className="p-7 text-start" style={{ marginTop: '3rem' }}>
          <h3 className="text-lg font-semibold mb-4">Feedback from this session</h3>
          {Object.entries(cachedTags).map(([sentence, tags], index) => (
            <div key={index} className="mb-4 flex items-start">
              <div style={{ flex: 1, lineHeight: 2 }}>
                {tags.map((token, tokenIndex) => (
                  <Highlight
                    key={tokenIndex}
                    currentToken={token}
                    nextToken={
                      tokenIndex === tags.length - 1 ? null : tags[tokenIndex + 1]
                    }
                    tagColors={tagColors}
                    onMouseOver={() => {}}
                    onMouseDown={() => {}}
                    selecting={false}
                    selected={false}
                  />
                ))}
              </div>
              <Button
                size="sm"
                variant="destructive"
                onClick={() => deleteFeedbackExample(sentence)}
                style={{
                  marginLeft: '10px',
                  padding: '0 10px',
                  height: '24px',
                  fontSize: '12px',
                }}
              >
                DELETE
              </Button>
            </div>
          ))}
          <Button
            size="sm"
            style={{ marginTop: '20px' }}
            onClick={submitFeedback}
            disabled={Object.keys(cachedTags).length === 0}
          >
            Submit Feedback
          </Button>
        </Card>
      </div>
    </Container>
  );
}
