'use client';

import {
  Container,
  Box,
  CircularProgress,
  Typography,
  Switch,
  FormControlLabel,
  Alert,
  Divider,
  Paper,
} from '@mui/material';
import { Button } from '@mui/material';
import React, {
  CSSProperties,
  ReactNode,
  useEffect,
  useRef,
  useState,
  ChangeEvent,
  KeyboardEvent,
} from 'react';
import { Card, CardContent } from '@mui/material';
import * as _ from 'lodash';

// Mock backend service until we implement the real one
const useTokenClassificationEndpoints = () => {
  return {
    classifyText: async (text: string) => {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Generate dummy tokens
      const words = text.split(/\s+/);
      return {
        data: {
          tokens: words.map(word => ({
            text: word,
            tag: Math.random() > 0.8 ? 'PII' : 'O'
          }))
        }
      };
    },
    getLabels: async () => {
      return {
        data: ['O', 'PII', 'NAME', 'ADDRESS', 'PHONE', 'EMAIL', 'SSN']
      };
    },
    addLabel: async (label: string) => {
      return { success: true };
    }
  };
};

// Import components
import { Input } from '@mui/material';
import { Menu, MenuItem } from '@mui/material';
import { useParams } from 'next/navigation';

interface Token {
  text: string;
  tag: string;
}

interface HighlightColor {
  text: string;
  tag: string;
}

interface TagSelectorProps {
  open: boolean;
  choices: string[];
  onSelect: (tag: string) => void;
  onNewLabel: (newLabel: string) => Promise<void>;
  currentTag: string;
  anchorEl: HTMLElement | null;
}

const SELECTING_COLOR = '#EFEFEF';
const SELECTED_COLOR = '#DFDFDF';

interface HighlightProps {
  currentToken: Token;
  tokenIndex: number;
  tagColors: Record<string, HighlightColor>;
  onMouseOver: (index: number) => void;
  onMouseDown: (index: number) => void;
  selecting: boolean;
  selected: boolean;
  showDropdown: boolean;
  allLabels: string[];
  onSelectTag: (tag: string) => void;
  onNewLabel: (newLabel: string) => Promise<void>;
  anchorEl: HTMLElement | null;
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
  anchorEl
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
        onMouseDown={(e) => {
          onMouseDown(tokenIndex);
        }}
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
          anchorEl={anchorEl}
        />
      )}
      <span> </span>
    </>
  );
}

function TagSelector({ open, choices, onSelect, onNewLabel, currentTag, anchorEl }: TagSelectorProps) {
  const [query, setQuery] = useState('');
  const [searchableChoices, setSearchableChoices] = useState<string[]>([]);
  const [newLabel, setNewLabel] = useState('');

  useEffect(() => {
    const updatedChoices = choices.filter((choice) => choice !== 'O');
    if (currentTag !== 'O') {
      updatedChoices.unshift('Delete TAG');
    }
    setSearchableChoices(updatedChoices);
  }, [choices, currentTag]);

  const filteredChoices = query
    ? searchableChoices.filter(choice => 
        choice.toLowerCase().includes(query.toLowerCase()))
    : searchableChoices;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && query && !filteredChoices.includes(query)) {
      e.preventDefault();
      onNewLabel(query).then(() => {
        onSelect(query);
        setQuery('');
      });
    }
  };

  return (
    <Menu
      open={open}
      anchorEl={anchorEl}
      onClose={() => onSelect(currentTag)}
      anchorOrigin={{
        vertical: 'bottom',
        horizontal: 'left',
      }}
    >
      <Box sx={{ p: 1, width: 200 }}>
        <Input
          autoFocus
          fullWidth
          value={query}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder="Search or add label..."
          sx={{ mb: 1 }}
        />
      </Box>
      
      {filteredChoices.map((choice, index) => (
        <MenuItem 
          key={index}
          onClick={() => {
            const selectedTag = choice === 'Delete TAG' ? 'O' : choice;
            onSelect(selectedTag);
          }}
        >
          {choice}
        </MenuItem>
      ))}
      
      {query && !filteredChoices.includes(query) && (
        <MenuItem 
          onClick={() => {
            onNewLabel(query).then(() => {
              onSelect(query);
              setQuery('');
            });
          }}
        >
          <Typography sx={{ fontStyle: 'italic' }}>
            Add "{query}" as new label
          </Typography>
        </MenuItem>
      )}
    </Menu>
  );
}

export default function Interact() {
  const { deploymentId } = useParams();
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [tokens, setTokens] = useState<Token[]>([]);
  const [allLabels, setAllLabels] = useState<string[]>(['O']);
  const [tagColors, setTagColors] = useState<Record<string, HighlightColor>>({});
  const [selectionStart, setSelectionStart] = useState<number | null>(null);
  const [selectionEnd, setSelectionEnd] = useState<number | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [dropdownPosition, setDropdownPosition] = useState<number | null>(null);
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const cardRef = useRef<HTMLDivElement>(null);
  
  const API = useTokenClassificationEndpoints();

  // Load available labels on component mount
  useEffect(() => {
    async function fetchLabels() {
      try {
        const response = await API.getLabels();
        setAllLabels(response.data);
        updateTagColors(response.data);
      } catch (error) {
        console.error('Failed to fetch labels', error);
      }
    }
    
    fetchLabels();
  }, []);

  const handleInputChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    setInput(event.target.value);
  };

  const updateTagColors = (tags: string[]) => {
    // Generate colors for each tag
    const colors: Record<string, HighlightColor> = {};
    const tagColorPairs = [
      { tag: '#E57373', text: '#FFEBEE' }, // red
      { tag: '#7986CB', text: '#E8EAF6' }, // indigo
      { tag: '#4FC3F7', text: '#E1F5FE' }, // light blue
      { tag: '#81C784', text: '#E8F5E9' }, // green
      { tag: '#FFF176', text: '#FFFDE7' }, // yellow
      { tag: '#FFB74D', text: '#FFF3E0' }, // orange
      { tag: '#BA68C8', text: '#F3E5F5' }, // purple
    ];

    tags.forEach((tag, index) => {
      if (tag !== 'O') {
        const colorIndex = index % tagColorPairs.length;
        colors[tag] = tagColorPairs[colorIndex];
      }
    });

    setTagColors(colors);
  };

  const handleRun = async () => {
    if (!input.trim()) return;
    
    try {
      setIsLoading(true);
      const response = await API.classifyText(input);
      setTokens(response.data.tokens);
    } catch (error) {
      console.error('Error classifying text:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileUpload = () => {
    // This would be implemented to handle file uploads
    console.log('File upload clicked');
  };

  const handleNewLabel = async (newLabel: string) => {
    try {
      await API.addLabel(newLabel);
      setAllLabels([...allLabels, newLabel]);
      
      // Update colors for the new label
      const newTagColors = { ...tagColors };
      const colorIndex = allLabels.length % 7;
      const tagColorPairs = [
        { tag: '#E57373', text: '#FFEBEE' }, // red
        { tag: '#7986CB', text: '#E8EAF6' }, // indigo
        { tag: '#4FC3F7', text: '#E1F5FE' }, // light blue
        { tag: '#81C784', text: '#E8F5E9' }, // green
        { tag: '#FFF176', text: '#FFFDE7' }, // yellow
        { tag: '#FFB74D', text: '#FFF3E0' }, // orange
        { tag: '#BA68C8', text: '#F3E5F5' }, // purple
      ];
      
      newTagColors[newLabel] = tagColorPairs[colorIndex];
      setTagColors(newTagColors);
    } catch (error) {
      console.error('Error adding new label:', error);
    }
  };

  const handleMouseDown = (index: number) => {
    setSelectionStart(index);
    setSelectionEnd(index);
    setShowDropdown(false);
  };

  const handleMouseOver = (index: number) => {
    if (selectionStart !== null) {
      setSelectionEnd(index);
    }
  };

  const handleCardMouseUp = (e: React.MouseEvent) => {
    if (selectionStart !== null && selectionEnd !== null) {
      // Sort selection range
      const start = Math.min(selectionStart, selectionEnd);
      const end = Math.max(selectionStart, selectionEnd);
      
      // Show dropdown menu for tag selection
      setShowDropdown(true);
      setDropdownPosition(end);
      setAnchorEl(e.currentTarget as HTMLElement);
    }
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleDocumentClick = (e: MouseEvent) => {
      if (showDropdown && cardRef.current && !cardRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
        setSelectionStart(null);
        setSelectionEnd(null);
      }
    };

    document.addEventListener('mousedown', handleDocumentClick);
    return () => {
      document.removeEventListener('mousedown', handleDocumentClick);
    };
  }, [showDropdown]);

  const handleSelectTag = (tag: string) => {
    if (selectionStart === null || selectionEnd === null) return;

    // Sort selection range
    const start = Math.min(selectionStart, selectionEnd);
    const end = Math.max(selectionStart, selectionEnd);

    // Update tokens with new tag
    const newTokens = [...tokens];
    for (let i = start; i <= end; i++) {
      newTokens[i] = { ...newTokens[i], tag };
    }

    setTokens(newTokens);
    setShowDropdown(false);
    setSelectionStart(null);
    setSelectionEnd(null);
  };

  const renderHighlightedContent = () => {
    if (tokens.length === 0) return null;

    return (
      <div ref={cardRef} onMouseUp={handleCardMouseUp}>
        {tokens.map((token, index) => (
          <Highlight
            key={index}
            currentToken={token}
            tokenIndex={index}
            tagColors={tagColors}
            onMouseOver={handleMouseOver}
            onMouseDown={handleMouseDown}
            selecting={
              selectionStart !== null &&
              selectionEnd !== null &&
              ((index >= Math.min(selectionStart, selectionEnd) &&
                index <= Math.max(selectionStart, selectionEnd)))
            }
            selected={false}
            showDropdown={showDropdown && dropdownPosition === index}
            allLabels={allLabels}
            onSelectTag={handleSelectTag}
            onNewLabel={handleNewLabel}
            anchorEl={anchorEl}
          />
        ))}
      </div>
    );
  };

  const handleSubmitFeedback = () => {
    console.log('Submit feedback clicked');
  };

  return (
    <Box sx={{ display: 'flex', gap: 3 }}>
      <Box sx={{ flex: '1' }}>
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <textarea
              className="w-full p-3 border border-gray-300 rounded min-h-32 mb-4"
              placeholder="Enter text here..."
              value={input}
              onChange={handleInputChange}
              style={{ 
                resize: 'none', 
                minHeight: '120px',
                backgroundColor: '#fff',
                borderRadius: '4px',
                border: '1px solid #ddd',
                padding: '12px'
              }}
            />
            
            <Box 
              sx={{ 
                display: 'flex', 
                alignItems: 'center', 
                my: 2,
                '&::before, &::after': {
                  content: '""',
                  flex: 1,
                  borderBottom: '1px solid #ddd'
                }
              }}
            >
              <Typography 
                variant="body2" 
                color="text.secondary" 
                sx={{ mx: 2 }}
              >
                OR
              </Typography>
            </Box>
            
            <Box 
              sx={{ 
                border: '1px dashed #ccc', 
                borderRadius: '4px', 
                p: 3, 
                textAlign: 'center',
                cursor: 'pointer',
                my: 2,
                '&:hover': {
                  borderColor: '#aaa'
                }
              }}
              onClick={handleFileUpload}
            >
              <Typography color="text.secondary">
                Upload Document here
              </Typography>
            </Box>
            
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
              <Button
                variant="contained"
                onClick={handleRun}
                disabled={isLoading || !input.trim()}
                sx={{ 
                  backgroundColor: '#9E9E9E',
                  color: 'white',
                  '&:hover': {
                    backgroundColor: '#757575'
                  },
                  minWidth: '100px',
                  borderRadius: '4px',
                  boxShadow: 'none',
                  textTransform: 'none'
                }}
              >
                {isLoading ? <CircularProgress size={24} color="inherit" /> : 'Run'}
              </Button>
            </Box>
          </CardContent>
        </Card>

        {tokens.length > 0 && (
          <Card>
            <CardContent>
              <Box ref={cardRef} onMouseUp={handleCardMouseUp}>
                {renderHighlightedContent()}
              </Box>
            </CardContent>
          </Card>
        )}
      </Box>
      
      <Box sx={{ width: '320px' }}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Feedback from this session
            </Typography>
            
            <Box sx={{ mt: 4, display: 'flex', justifyContent: 'flex-start' }}>
              <Button 
                variant="contained" 
                sx={{ 
                  backgroundColor: '#EEEEEE', 
                  color: '#666',
                  '&:hover': {
                    backgroundColor: '#E0E0E0'
                  },
                  textTransform: 'none',
                  boxShadow: 'none'
                }}
                onClick={handleSubmitFeedback}
              >
                Submit Feedback
              </Button>
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Box>
  );
} 