import React, { MouseEventHandler, useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';

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

interface FeedbackDashboardProps {
    cachedTags: Record<string, Token[]>;
    tagColors: Record<string, HighlightColor>;
    deleteFeedbackExample: (sentence: string) => void;
    submitFeedback: () => void;
  }
  
  interface Token {
    text: string;
    tag: string;
  }
  
  interface HighlightColor {
    text: string;
    tag: string;
  }
  
  const FeedbackDashboard: React.FC<FeedbackDashboardProps> = ({
    cachedTags,
    tagColors,
    deleteFeedbackExample,
    submitFeedback,
  }) => {
    const deduplicatedTags = useMemo(() => {
      const uniqueSentences: Record<string, Token[]> = {};
  
      Object.entries(cachedTags).forEach(([sentence, tokens]) => {
        const normalizedSentence = tokens.map(t => t.text).join(' ');
        
        if (!uniqueSentences[normalizedSentence]) {
          uniqueSentences[normalizedSentence] = tokens;
        } else {
          uniqueSentences[normalizedSentence] = uniqueSentences[normalizedSentence].map((token, index) => ({
            ...token,
            tag: token.tag !== 'O' ? token.tag : tokens[index].tag
          }));
        }
      });
  
      return uniqueSentences;
    }, [cachedTags]);
  
    const renderFeedbackContent = (tags: Token[]) => {
      let result: React.ReactNode[] = [];
      let columnNameBuffer: Token[] = [];
      let isInColumnName = false;
  
      const renderBuffer = () => {
        if (columnNameBuffer.length > 0) {
          if (isInColumnName) {
            result.push(
              <React.Fragment key={`col-${result.length}`}>
                <br />
                <strong>
                  {columnNameBuffer.map(t => t.text).join(' ')}
                </strong>
                {' '}
              </React.Fragment>
            );
          } else {
            result.push(...renderTokens(columnNameBuffer));
          }
          columnNameBuffer = [];
        }
      };
  
      const renderTokens = (tokens: Token[]) => {
        const mergedTokens: Token[] = [];
        let currentMerge: Token | null = null;
  
        tokens.forEach((token) => {
          if (currentMerge && token.tag === currentMerge.tag) {
            currentMerge.text += ' ' + token.text;
          } else {
            if (currentMerge) {
              mergedTokens.push(currentMerge);
            }
            currentMerge = { ...token };
          }
        });
  
        if (currentMerge) {
          mergedTokens.push(currentMerge);
        }
  
        return mergedTokens.map((token, index) => renderToken(token, `content-${index}`));
      };
  
      const renderToken = (token: Token, key: string | number) => {
        if (token.tag !== 'O') {
          return (
            <Highlight
              key={key}
              currentToken={token}
              nextToken={null}
              tagColors={tagColors}
              onMouseOver={() => {}}
              onMouseDown={() => {}}
              selecting={false}
              selected={false}
            />
          );
        } else {
          return <React.Fragment key={key}>{token.text} </React.Fragment>;
        }
      };
  
      for (let i = 0; i < tags.length; i++) {
        const token = tags[i];
  
        if (token.text.trim().endsWith(':')) {
          if (isInColumnName) {
            columnNameBuffer.push(token);
            renderBuffer();
            isInColumnName = false;
          } else {
            renderBuffer();
            isInColumnName = true;
            columnNameBuffer.push(token);
          }
        } else if (isInColumnName) {
          if (token.text.trim() === '') {
            renderBuffer();
            isInColumnName = false;
          } else {
            columnNameBuffer.push(token);
          }
        } else {
          columnNameBuffer.push(token);
        }
      }
  
      renderBuffer();
  
      return result;
    };
  
    return (
      <div>
        <h3 className="text-lg font-semibold mb-4">Feedback from this session</h3>
        {Object.entries(deduplicatedTags).map(([sentence, tags], index) => (
          <div key={index} className="mb-4 flex items-start">
            <div style={{ flex: 1, lineHeight: 2 }}>{renderFeedbackContent(tags)}</div>
            <Button
              size="sm"
              variant="destructive"
              onClick={() => deleteFeedbackExample(sentence)}
              style={{ marginLeft: '10px', padding: '0 10px', height: '24px', fontSize: '12px' }}
            >
              DELETE
            </Button>
          </div>
        ))}
        <Button
          size="sm"
          style={{ marginTop: '20px' }}
          onClick={submitFeedback}
          disabled={Object.keys(deduplicatedTags).length === 0}
        >
          Submit Feedback
        </Button>
      </div>
    );
  };
  
  export default FeedbackDashboard;