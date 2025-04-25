'use client';

import {
  Container,
  Box,
  CircularProgress,
  Typography,
  Switch,
  FormControlLabel,
  Alert,
} from '@mui/material';
import { Button } from '@/components/ui/button';
import React, {
  CSSProperties,
  ReactNode,
  useEffect,
  useRef,
  useState,
  ChangeEvent,
  KeyboardEvent,
} from 'react';
import { Card, CardContent } from '@/components/ui/card';
import * as _ from 'lodash';
// import { useTokenClassificationEndpoints } from '@/lib/backend';
import { useTokenClassificationEndpoints } from '@/lib/mock-backend';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import Fuse from 'fuse.js';
import FeedbackDashboard from './FeedbackDashboard';
import {
  parseCSV,
  parseExcel,
  parseTXT,
  convertCSVToPDFFormat,
  ParsedData,
} from '@/utils/fileParsingUtils';
import InferenceTimeDisplay from '@/components/ui/InferenceTimeDisplay';
import ExpandingInput from '@/components/ui/ExpandingInput';
import { useParams } from 'next/navigation';

interface Token {
  text: string;
  tag: string;
}

interface HighlightColor {
  text: string;
  tag: string;
}

interface ColumnData {
  columnName: string;
  content: Token[];
}

type CachedTagEntry = Token[] | ColumnData;

interface CachedTags {
  [key: string]: CachedTagEntry;
}

interface TagSelectorProps {
  open: boolean;
  choices: string[];
  onSelect: (tag: string) => void;
  onNewLabel: (newLabel: string) => Promise<void>;
  currentTag: string;
}

const SELECTING_COLOR = '#EFEFEF';
const SELECTED_COLOR = '#DFDFDF';

interface HighlightProps {
  currentToken: Token;
  tokenIndex: number;
  nextToken?: Token | null;
  tagColors: Record<string, HighlightColor>;
  onMouseOver: (index: number) => void;
  onMouseDown: (index: number) => void;
  selecting: boolean;
  selected: boolean;
  showDropdown: boolean;
  allLabels: string[];
  onSelectTag: (tag: string) => void;
  onNewLabel: (newLabel: string) => Promise<void>;
}

function Highlight({
  currentToken,
  tokenIndex,
  tagColors,
  onMouseOver,
  onMouseDown,
  selecting,
  selected,
  showDropdown,
  allLabels,
  onSelectTag,
  onNewLabel,
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
          display: 'inline-flex',
          alignItems: 'center',
        }}
        onMouseOver={(e) => {
          setHover(true);
          onMouseOver(tokenIndex);
        }}
        onMouseLeave={(e) => {
          setHover(false);
        }}
        onMouseDown={() => onMouseDown(tokenIndex)}
      >
        {currentToken.text}
        {currentToken.tag !== 'O' && (
          <span
            style={{
              backgroundColor: tagColors[currentToken.tag]?.tag,
              color: 'white',
              fontSize: '11px',
              fontWeight: 'bold',
              borderRadius: '2px',
              marginLeft: '4px',
              padding: '1px 3px',
            }}
          >
            {currentToken.tag}
          </span>
        )}
      </span>
      {showDropdown && (
        <TagSelector
          open={true}
          choices={allLabels}
          onSelect={onSelectTag}
          onNewLabel={onNewLabel}
          currentTag={currentToken.tag}
        />
      )}
      <span> </span>
    </>
  );
}

function TagSelector({ open, choices, onSelect, onNewLabel, currentTag }: TagSelectorProps) {
  const [fuse, setFuse] = useState<Fuse<string>>(new Fuse([]));
  const [query, setQuery] = useState('');
  const [searchableChoices, setSearchableChoices] = useState<string[]>([]);

  useEffect(() => {
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
      <DropdownMenuContent className="tag-selector">
        <Input
          autoFocus
          className="font-medium"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="find existing label or create a new one"
          style={{ marginBottom: '5px', width: '300px' }}
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
  const params = useParams();
  const deploymentId = params.deploymentId as string;
  const { predict, insertSample, addLabel, getLabels, getTextFromFile } =
    useTokenClassificationEndpoints();

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
  const [selectedTokenIndex, setSelectedTokenIndex] = useState<number | null>(null);
  const [selectedRange, setSelectedRange] = useState<[number, number] | null>(null);
  const [processingTime, setProcessingTime] = useState<number | undefined>();
  const startIndex =
    mouseDownIndex !== null && mouseUpIndex !== null
      ? Math.min(mouseDownIndex, mouseUpIndex)
      : null;
  const endIndex =
    mouseDownIndex !== null && mouseUpIndex !== null
      ? Math.max(mouseDownIndex, mouseUpIndex)
      : null;

  const triggers = useRef<(HTMLElement | null)[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [cachedTags, setCachedTags] = useState<CachedTags>({});

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInputText(event.target.value);
    setParsedData(null);
    setAnnotations([]);
    setProcessingTime(undefined); // make time display disppears as typying begins
  };

  const [fileError, setFileError] = useState<string | null>(null);

  const MAX_FILE_SIZE = 1024 * 1024; // 1MB in bytes

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    setFileError(null); // Reset error on new file selection

    if (file) {
      if (file.size > MAX_FILE_SIZE) {
        setFileError('File size exceeds 1MB. Please use the API for larger files.');
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
        return;
      }

      // Reset relevant state
      setAnnotations([]);
      setCachedTags({});
      setParsedData(null);
      setInputText('');

      const fileExtension = file.name.split('.').pop()?.toLowerCase();
      setIsLoading(true);

      try {
        let parsed: ParsedData;
        if (fileExtension === 'csv') {
          parsed = await parseCSV(file);
        } else if (['pdf', 'docx', 'html'].includes(fileExtension ?? '')) {
          try {
            const content = await getTextFromFile(file);
            parsed = {
              type: 'pdf',
              content: content.join('\n'),
              pdfParagraphs: content,
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
      } catch (error) {
        console.error('Error processing file:', error);
        setFileError('Error processing file. Please try again.');
      } finally {
        setIsLoading(false);
      }
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
    // Check for empty text or only whitespace
    if (!text?.trim()) {
      return;
    }

    setIsLoading(true);
    try {
      const result = await predict(text);
      console.log('Prediction result:', result);
      updateTagColors(result.prediction_results.predicted_tags);
      setProcessingTime(result.time_taken);
      setAnnotations(
        _.zip(result.prediction_results.tokens, result.prediction_results.predicted_tags).map(
          ([text, tag]) => ({
            text: text as string,
            tag: (tag as string[])[0],
          })
        )
      );

      if (!isFileUpload) {
        setParsedData({ type: 'other', content: text });
      }

      const labels = await getLabels();
      const filteredLabels = labels.filter((label) => label !== 'O');
      setAllLabels(filteredLabels);
      updateTagColors([filteredLabels]);
    } catch (error) {
      console.error('Error during prediction or fetching labels:', error);
    } finally {
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

  function isColumnData(entry: CachedTagEntry): entry is ColumnData {
    return (entry as ColumnData).content !== undefined;
  }

  function mergeTokens(existingTokens: Token[], newTokens: Token[]): Token[] {
    return existingTokens.map((token, index) => {
      const newToken = newTokens[index];
      if (newToken) {
        // Always update the tag, even if it's 'O'
        return { ...token, tag: newToken.tag };
      }
      return token;
    });
  }

  const tokenizeParagraph = (paragraph: string): string[] => {
    return paragraph.split(/\s+/);
  };

  const findParagraphIndex = (selectedRange: [number, number], paragraphs: string[]): number => {
    let tokenCount = 0;
    for (let i = 0; i < paragraphs.length; i++) {
      const paragraphTokens = tokenizeParagraph(paragraphs[i]);
      tokenCount += paragraphTokens.length;
      if (tokenCount > selectedRange[0]) {
        return i;
      }
    }
    return -1; // This should never happen if selectedRange is valid
  };

  // New function to find the correct row index for CSV data
  const findCSVRowIndex = (selectedRange: [number, number], rowIndices: number[]): number => {
    for (let i = 0; i < rowIndices.length - 1; i++) {
      if (selectedRange[0] >= rowIndices[i] && selectedRange[0] < rowIndices[i + 1]) {
        return i;
      }
    }
    return -1; // This should never happen if selectedRange is valid
  };

  const handleMouseDown = (index: number) => {
    console.log('Mouse down:', { index });
    setSelecting(true);
    setMouseDownIndex(index);
    setMouseUpIndex(index);
    setSelectedTokenIndex(index);
    setSelectedRange([index, index]);
  };

  const handleMouseOver = (index: number) => {
    if (selecting) {
      console.log('Mouse over while selecting:', { index, mouseDownIndex });
      setMouseUpIndex(index);
      if (mouseDownIndex !== null) {
        const start = Math.min(mouseDownIndex, index);
        const end = Math.max(mouseDownIndex, index);
        setSelectedRange([start, end]);
      }
    }
  };

  const handleCardMouseUp = (e: React.MouseEvent) => {
    console.log('Card mouse up:', { startIndex, endIndex });
    e.stopPropagation();
    if (startIndex !== null && endIndex !== null) {
      setSelectedRange([startIndex, endIndex]);
    }
  };

  // Modify the mouse up handler to maintain selection
  useEffect(() => {
    const handleMouseUp = (e: MouseEvent) => {
      console.log('Mouse up:', { selecting, mouseDownIndex, mouseUpIndex, selectedRange });
      if (selecting && mouseDownIndex !== null && mouseUpIndex !== null) {
        const start = Math.min(mouseDownIndex, mouseUpIndex);
        const end = Math.max(mouseDownIndex, mouseUpIndex);
        setSelectedRange([start, end]);
        // Only clear the selecting state, keep other states
        setSelecting(false);
      }
    };

    window.addEventListener('mouseup', handleMouseUp);
    return () => window.removeEventListener('mouseup', handleMouseUp);
  }, [selecting, mouseDownIndex, mouseUpIndex]);

  const cacheNewTag = async (newTag: string) => {
    console.log('Cache new tag called:', { newTag, selectedRange, annotations });
    if (!selectedRange) {
      console.log('No selected range, returning');
      return;
    }

    // Clear dropdown
    setSelectedTokenIndex(null);

    // Update annotations with new tag
    const updatedTags = annotations.map((token, index) => {
      const isInRange = index >= selectedRange[0] && index <= selectedRange[1];
      console.log('Processing token:', { index, token, isInRange, newTag });
      return {
        text: token.text,
        tag: isInRange ? newTag : token.tag
      };
    });

    // Update annotations state
    console.log('Setting new annotations:', updatedTags);
    setAnnotations(updatedTags);
    
    // Update tag colors
    updateTagColors([[newTag]]);

    // Update cached tags
    if (parsedData) {
      let paragraphs: string[];
      let rowIndices: number[] | null = null;
      let isCSV = false;
      let isDirectQuery = false;

      if (parsedData.type === 'pdf' && parsedData.pdfParagraphs) {
        paragraphs = parsedData.pdfParagraphs;
      } else if ((parsedData.type === 'csv' || parsedData.type === 'other') && parsedData.rows) {
        const csvData = convertCSVToPDFFormat(parsedData.rows);
        paragraphs = parsedData.rows.map((row) => row.content.replace(/\n/g, ' '));
        rowIndices = csvData.rowIndices;
        isCSV = true;
      } else {
        isDirectQuery = true;
        paragraphs = [parsedData.content];
      }

      let relevantParagraphIndex: number;
      if (isDirectQuery) {
        relevantParagraphIndex = 0;
      } else if (isCSV && rowIndices) {
        relevantParagraphIndex = findCSVRowIndex(selectedRange, rowIndices);
      } else {
        relevantParagraphIndex = findParagraphIndex(selectedRange, paragraphs);
      }

      console.log('Found relevant paragraph:', { relevantParagraphIndex, isDirectQuery, isCSV });

      if (relevantParagraphIndex !== -1) {
        const relevantParagraph = paragraphs[relevantParagraphIndex];
        const feedbackKey = isDirectQuery
          ? 'direct-query'
          : isCSV
            ? `row-${relevantParagraphIndex}`
            : `paragraph-${relevantParagraphIndex}`;

        const paragraphTokens = tokenizeParagraph(relevantParagraph);
        let paragraphStartIndex = 0;
        
        if (isDirectQuery) {
          paragraphStartIndex = 0;
        } else if (isCSV && rowIndices) {
          paragraphStartIndex = rowIndices[relevantParagraphIndex];
        } else {
          let tokenCount = 0;
          for (let i = 0; i < relevantParagraphIndex; i++) {
            tokenCount += tokenizeParagraph(paragraphs[i]).length;
          }
          paragraphStartIndex = tokenCount;
        }

        const newContent = paragraphTokens.map((word, index) => ({
          text: word,
          tag: updatedTags[paragraphStartIndex + index]?.tag || 'O',
        }));

        console.log('Updating cached tags:', { feedbackKey, newContent });

        setCachedTags((prev) => {
          const updatedCachedTags = { ...prev };
          
          if (updatedCachedTags[feedbackKey]) {
            if (isColumnData(updatedCachedTags[feedbackKey])) {
              const existingContent = (updatedCachedTags[feedbackKey] as ColumnData).content;
              const mergedContent = mergeTokens(existingContent, newContent);
              (updatedCachedTags[feedbackKey] as ColumnData).content = mergedContent;
            }
          } else {
            updatedCachedTags[feedbackKey] = {
              columnName: isDirectQuery
                ? 'Direct Query'
                : isCSV
                  ? `Row ${relevantParagraphIndex + 1}`
                  : `Paragraph ${relevantParagraphIndex + 1}`,
              content: newContent,
            };
          }
          
          console.log('New cached tags state:', updatedCachedTags);
          return updatedCachedTags;
        });
      }
    }

    // Reset selection state after the tag has been applied
    setSelectedRange(null);
    setMouseDownIndex(null);
    setMouseUpIndex(null);
    setSelecting(false);
  };

  const updateFeedbackDashboard = () => {
    setCachedTags((prev) => {
      const updatedCachedTags: CachedTags = { ...prev };

      Object.keys(updatedCachedTags).forEach((sentence) => {
        const entry = updatedCachedTags[sentence];

        if (Array.isArray(entry)) {
          // Handle Token[] case
          if (!entry.some((token) => token.tag !== 'O')) {
            delete updatedCachedTags[sentence];
          }
        } else if ('content' in entry) {
          // Handle ColumnData case
          if (!entry.content.some((token) => token.tag !== 'O')) {
            delete updatedCachedTags[sentence];
          }
        }
      });

      return updatedCachedTags;
    });
  };

  useEffect(() => {
    updateFeedbackDashboard();
    // Add debug logging
    console.log('Current cached tags:', cachedTags);
  }, [annotations]);

  const submitFeedback = async () => {
    console.log('Submitting feedback with cached tags:', cachedTags);
    try {
      for (const [sentence, tags] of Object.entries(cachedTags)) {
        let submission: { tokens: string[]; tags: string[] };

        if (Array.isArray(tags)) {
          submission = {
            tokens: tags.map((t) => t.text),
            tags: tags.map((t) => t.tag),
          };
        } else {
          submission = {
            tokens: tags.content.map((t) => t.text),
            tags: tags.content.map((t) => t.tag),
          };
        }

        await insertSample(submission);
      }
      console.log('All samples inserted successfully');
      setCachedTags({});
    } catch (error) {
      console.error('Error inserting samples:', error);
    }
  };

  const deleteFeedbackExample = (feedbackKey: string) => {
    setCachedTags((prev) => {
      const updatedCachedTags = { ...prev };
      delete updatedCachedTags[feedbackKey];
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

  const renderHighlightedCSVContent = (
    content: string,
    rowWords: { text: string; originalIndex: number }[]
  ) => {
    const words = content.split(/\s+/);
    let currentRowWordIndex = 0;

    return words.map((word, wordIndex) => {
      while (
        currentRowWordIndex < rowWords.length &&
        rowWords[currentRowWordIndex].text.toLowerCase() !== word.toLowerCase()
      ) {
        currentRowWordIndex++;
      }

      if (currentRowWordIndex < rowWords.length) {
        const { text, originalIndex } = rowWords[currentRowWordIndex];
        currentRowWordIndex++;

        return (
          <Highlight
            key={`${wordIndex}-${originalIndex}`}
            currentToken={annotations[originalIndex]}
            tokenIndex={originalIndex}
            nextToken={annotations[originalIndex + 1] || null}
            tagColors={tagColors}
            onMouseOver={handleMouseOver}
            onMouseDown={handleMouseDown}
            selecting={
              selecting &&
              startIndex !== null &&
              endIndex !== null &&
              originalIndex >= startIndex &&
              originalIndex <= endIndex
            }
            selected={
              selectedRange !== null &&
              originalIndex >= selectedRange[0] &&
              originalIndex <= selectedRange[1]
            }
            showDropdown={originalIndex === selectedTokenIndex}
            allLabels={allLabels}
            onSelectTag={cacheNewTag}
            onNewLabel={handleNewLabel}
          />
        );
      }

      return <span key={`${wordIndex}-text`}>{word} </span>;
    });
  };

  const renderCSVContent = (rows: { label: string; content: string }[]) => {
    const { words, rowIndices } = convertCSVToPDFFormat(rows);

    return (
      <>
        {rows.map((row, rowIndex) => {
          const columns = row.content.split('\n');
          const visibleColumns = columns.filter((column) => {
            const [columnName, ...columnContent] = column.split(':');
            const content = columnContent.join(':').trim();
            return !showHighlightedOnly || content.split(' ').some(isWordHighlighted);
          });

          if (visibleColumns.length === 0) {
            return null;
          }

          const rowStartIndex = rowIndices[rowIndex];
          const rowEndIndex = rowIndices[rowIndex + 1];
          const rowWords = words.slice(rowStartIndex, rowEndIndex);

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
                    <strong>{columnName}:</strong> {renderHighlightedCSVContent(content, rowWords)}
                  </p>
                );
              })}
            </div>
          );
        })}
      </>
    );
  };

  const renderPDFContent = (paragraphs: string[]) => {
    return paragraphs.map((paragraph, index) => (
      <React.Fragment key={index}>
        {paragraph.split('\n').map((line, lineIndex) => (
          <>
            {renderHighlightedContent(line)}
            <br />
          </>
        ))}
        <br />
      </React.Fragment>
    ));
  };

  const renderHighlightedContent = (content: string) => {
    const words = content.split(/\s+/);
    let currentIndex = 0;

    return words
      .map((word, wordIndex) => {
        const tokenIndex = annotations.findIndex(
          (token, index) => token.text.toLowerCase() === word.toLowerCase() && index >= currentIndex
        );

        if (tokenIndex !== -1) {
          currentIndex = tokenIndex + 1;
          return (
            <Highlight
              key={`${wordIndex}-${tokenIndex}`}
              currentToken={annotations[tokenIndex]}
              tokenIndex={tokenIndex}
              nextToken={annotations[tokenIndex + 1] || null}
              tagColors={tagColors}
              onMouseOver={handleMouseOver}
              onMouseDown={handleMouseDown}
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
              showDropdown={tokenIndex === selectedTokenIndex}
              allLabels={allLabels}
              onSelectTag={cacheNewTag}
              onNewLabel={handleNewLabel}
            />
          );
        }

        // Render non-highlighted words
        if (!showHighlightedOnly || (tokenIndex !== -1 && annotations[tokenIndex].tag !== 'O')) {
          return <span key={`${wordIndex}-text`}>{word} </span>;
        }

        // If showHighlightedOnly is true and this word is not tagged, return null
        return null;
      })
      .filter(Boolean); // Remove null elements
  };

  // Add this effect to close the dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (selectedTokenIndex !== null && !(event.target as Element).closest('.tag-selector')) {
        setSelectedTokenIndex(null);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [selectedTokenIndex]);

  const renderContent = () => {
    if (!parsedData) return null;

    const contentStyle: CSSProperties = {
      lineHeight: 2, // Adjusted line height
      position: 'relative',
    };

    if (parsedData.type === 'csv' && parsedData.rows) {
      return <div style={contentStyle}>{renderCSVContent(parsedData.rows)}</div>;
    } else if (parsedData.type === 'pdf' && parsedData.pdfParagraphs) {
      return <div style={contentStyle}>{renderPDFContent(parsedData.pdfParagraphs)}</div>;
    } else {
      console.log("parsedData.content:", parsedData.content);
      return <div style={contentStyle}>{renderHighlightedContent(parsedData.content)}</div>;
    }
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {/* Left Column: User Input Area */}
      <div className="space-y-4">
        <Card>
          <CardContent className="p-6">
            <textarea
              placeholder="Enter text here..."
              value={inputText}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setInputText(e.target.value)}
              className="w-full min-h-[100px] p-3 border rounded-md resize-y"
            />
            
            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  or
                </span>
              </div>
            </div>

            <div 
              className="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer hover:border-primary transition-colors"
              onClick={() => fileInputRef.current?.click()}
            >
              <input
                type="file"
                ref={fileInputRef}
                className="hidden"
                onChange={handleFileChange}
                accept=".txt,.csv,.xlsx,.pdf"
              />
              <p className="text-muted-foreground">Upload Document here</p>
              {fileInputRef.current?.files?.[0] && (
                <p className="mt-2 text-sm text-primary">{fileInputRef.current.files[0].name}</p>
              )}
            </div>

            <Button 
              className="w-full mt-4" 
              onClick={() => handleRun(inputText)}
              disabled={isLoading || (!inputText && !fileInputRef.current?.files?.[0])}
            >
              {isLoading ? "Processing..." : "Run"}
            </Button>
          </CardContent>
        </Card>

        {/* Results Section */}
        {annotations.length > 0 && (
          <Card>
            <CardContent className="p-6">
              {renderContent()}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Right Column: Feedback Dashboard */}
      <div>
        {processingTime !== undefined && annotations.length > 0 && (
          <div className="mb-4">
            <InferenceTimeDisplay processingTime={processingTime} tokenCount={annotations.length} />
          </div>
        )}

        <Card className="p-6">
          <FeedbackDashboard
            cachedTags={cachedTags}
            tagColors={tagColors}
            deleteFeedbackExample={deleteFeedbackExample}
            submitFeedback={submitFeedback}
          />
        </Card>
      </div>
    </div>
  );
}
