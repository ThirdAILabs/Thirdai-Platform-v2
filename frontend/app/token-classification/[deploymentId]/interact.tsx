'use client';

import { Container, TextField, Box } from '@mui/material';
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
}

function TagSelector({ open, choices, onSelect, onNewLabel }: TagSelectorProps) {
  const [fuse, setFuse] = useState(new Fuse(choices));
  const [query, setQuery] = useState('');
  useEffect(() => {
    setFuse(new Fuse(choices));
  }, [choices]);
  const searchResults = query !== '' ? fuse.search(query).map((val) => val.item) : choices;
  const makeDropdownMenuItem = (key: number, value: string, child: ReactNode) => (
    <DropdownMenuItem className="font-medium" key={key}>
      <button
        style={{ width: '100%', height: '100%', textAlign: 'left' }}
        onClick={async () => {
          if (!choices.includes(value)) {
            await onNewLabel(value);
          }
          onSelect(value);
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
          !searchResults.map((val) => val).includes(query) &&
          makeDropdownMenuItem(
            /* key= */ 0,
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

  const handleRun = async () => {
    try {
      const result = await predict(inputText);
      updateTagColors(result.predicted_tags);
      setAnnotations(
        _.zip(result.tokens, result.predicted_tags).map(([text, tag]) => ({
          text: text as string,
          tag: tag![0] as string,
        }))
      );

      // Fetch labels after prediction
      const labels = await getLabels();
      setAllLabels(labels);
      updateTagColors([labels]);
    } catch (error) {
      console.error('Error during prediction or fetching labels:', error);
    }
  };

  // New state for caching manual tags
  const [cachedTags, setCachedTags] = useState<{ [sentence: string]: Token[] }>({});

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

    setCachedTags(prev => ({
      ...prev,
      [normalizedSentence]: updatedTags
    }));

    setAnnotations(updatedTags);
    updateTagColors([[newTag]]);
    setSelectedRange(null);
    setMouseDownIndex(null);
    setMouseUpIndex(null);
    setSelecting(false);
  };

  const submitFeedback = async () => {
    try {
      for (const [sentence, tags] of Object.entries(cachedTags)) {
        await insertSample({
          tokens: sentence.split(' '),
          tags: tags.map(t => t.tag),
        });
      }
      console.log('All samples inserted successfully');
      setCachedTags({});  // Clear the cache after successful submission
    } catch (error) {
      console.error('Error inserting samples:', error);
    }
  };

  return (
  <Container style={{ display: 'flex', paddingTop: '20vh', width: '90%', maxWidth: '1200px' }}>
    <div style={{ flex: 2, marginRight: '20px' }}>
      <Box display="flex" justifyContent="center" alignItems="center" width="100%">
        <Input
          autoFocus
          className="text-md"
          style={{ height: '3rem' }}
          value={inputText}
          onChange={handleInputChange}
          placeholder="Enter your text..."
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              handleRun();
            }
          }}
        />
        <Button
          size="sm"
          style={{ height: '3rem', marginLeft: '10px', padding: '0 20px' }}
          onClick={handleRun}
        >
          Run
        </Button>
      </Box>
      {annotations.length > 0 && (
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
            {annotations.map((token, index) => {
              const nextToken = index === annotations.length - 1 ? null : annotations[index + 1];
              return (
                <React.Fragment key={index}>
                  <Highlight
                    currentToken={token}
                    nextToken={nextToken}
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
                  />
                </React.Fragment>
              );
            })}
          </Card>
        </Box>
      )}
    </div>
    <div style={{ flex: 1 }}>
      <Card className="p-7 text-start" style={{ marginTop: '3rem' }}>
        <h3 className="text-lg font-semibold mb-4">Feedback from this session</h3>
        {Object.entries(cachedTags).map(([sentence, tags], index) => (
          <div key={index} className="mb-4">
            <div style={{ lineHeight: 2 }}>
              {tags.map((token, tokenIndex) => (
                <Highlight
                  key={tokenIndex}
                  currentToken={token}
                  nextToken={tokenIndex === tags.length - 1 ? null : tags[tokenIndex + 1]}
                  tagColors={tagColors}
                  onMouseOver={() => {}}
                  onMouseDown={() => {}}
                  selecting={false}
                  selected={false}
                />
              ))}
            </div>
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
