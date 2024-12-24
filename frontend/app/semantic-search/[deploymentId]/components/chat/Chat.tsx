import React, { useContext, useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import styled from 'styled-components';
import { borderRadius, color, duration, fontSizes, padding } from '../../stylingConstants';
import { ModelServiceContext } from '../../Context';
import { ChatMessage, ModelService, ReferenceInfo, PdfInfo } from '../../modelServices';
import { Chunk } from '../pdf_viewer/interfaces';
import PdfViewer from '../pdf_viewer/PdfViewer';
import TypingAnimation from '../TypingAnimation';
import { piiDetect, useSentimentClassification } from '@/lib/backend'; // Import for sentiment classification
// Import FontAwesomeIcon and faPause
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faStop } from '@fortawesome/free-solid-svg-icons';

const PdfViewerWrapper = styled.section`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
  width: 100%;
  height: 100%;
  padding: ${padding.card};
  box-sizing: border-box;
  background: rgba(255, 255, 255, 0.95);
  display: flex;
  justify-content: center;
  align-items: center;
`;

const ChatContainer = styled.section`
  position: fixed;
  width: 60%;
  left: 10%;
  display: flex;
  flex-direction: column;
  justify-content: end;
  z-index: 100;
  height: 100%;
  font-family: Helvetica, Arial, sans-serif;
`;

// Styled component for the pause button
const PauseButton = styled.button`
  width: 40px;
  height: 40px;
  background-color: black;
  border: none;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s ease-in-out;
  box-sizing: border-box;
  padding: 0;

  &:hover {
    background-color: #333;
  }
`;

const ChatBarContainer = styled.div`
  display: flex;
  align-items: center;
  width: 100%; // Ensure it takes the full width
  margin: 10px 0 50px 0%; // Adjust margins as needed
`;

const ChatBoxContainer = styled.section`
  padding: 15px 15px 0 15px;
`;

const ChatBoxSender = styled.section`
  font-size: ${fontSizes.l};
  font-weight: bold;
  color: ${color.accent};
`;

const ChatBoxContent = styled.section`
  font-size: ${fontSizes.m};
  padding-bottom: 15px;
`;

const TypingAnimationContainer = styled.section`
  padding: ${padding.card} 0 7px 0;
`;

const ChatBar = styled.textarea`
  background-color: ${color.textInput};
  font-size: ${fontSizes.m};
  padding: 20px;
  border-radius: ${borderRadius.textInput};
  outline: none;
  border: none;
  transition-duration: ${duration.transition};
  height: ${fontSizes.xl};
  resize: none;
  font-family: Helvetica, Arial, sans-serif;
  flex: 1; // Add this line to make it expand
  margin-right: 10px; // Add some space between the ChatBar and the PauseButton

  &:focus {
    height: 100px;
    transition-duration: ${duration.transition};
  }
`;

const ScrollableArea = styled.section`
  overflow-y: scroll;
  display: flex;
  flex-direction: column-reverse;
  height: 80%;
`;

const AILoadingWrapper = styled.div`
  margin-left: 10px;
`;

const AllChatBoxes = styled.section``;

const Placeholder = styled.section`
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  font-size: ${fontSizes.l};
  height: 80%;
`;

const labels = [
  {
    id: 1,
    name: 'BRAND',
    color: 'blue',
    amount: '217,323',
    checked: true,
    description:
      'The format of a US phone number is (XXX) XXX-XXXX, where "X" represents a digit from 0 to 9. It consists of a three-digit area code, followed by a three-digit exchange code, and a four-digit line number.',
  },
  {
    id: 2,
    name: 'MODEL_NUMBER',
    color: 'orange',
    amount: '8,979',
    checked: true,
    description:
      'The format of a US Social Security Number (SSN) is XXX-XX-XXXX, where "X" represents a digit from 0 to 9. It consists of three parts: area, group, and serial numbers.',
  },
];

// ChatBox component to display human/AI message with sentiment
const sentimentColor = (sentiment: string) => {
  switch (sentiment) {
    case 'positive': return 'green';
    case 'neutral': return 'orange';
    case 'negative': return 'red';
    default: return '#888';
  }
};

function ChatBox({
  message,
  transformedMessage,
  sentiment,
  context,
  modelService,
  onOpenPdf
}: {
  message: ChatMessage;
  transformedMessage?: string[][];
  sentiment?: string;
  context?: Array<{
    chunk_id: number;
    query: string;
    sourceURL: string;
    sourceName: string;
    content: string;
    metadata: any;
  }>;
  modelService: ModelService | null;
  onOpenPdf: (pdfInfo: PdfInfo) => void;
}) {
  const handleReferenceClick = async (chunkInfo: any) => {
    if (!modelService) return;

    const ref: ReferenceInfo = {
      id: chunkInfo.chunk_id,
      sourceURL: chunkInfo.sourceURL,
      sourceName: chunkInfo.sourceName,
      content: chunkInfo.content,
      metadata: chunkInfo.metadata
    };

    try {
      if (!ref.sourceURL.toLowerCase().endsWith('.pdf')) {
        modelService.openReferenceSource(ref);
        return;
      }

      // Append the full path prefix for PDF sources
      const pdfPrefix = '/home/peter/share/model_bazaar_cache/models/c53ea287-5622-4dd5-be17-70d805368737/model.ndb/documents/';
      ref.sourceURL = pdfPrefix + ref.sourceURL;

      const pdfInfo = await modelService.getPdfInfo(ref);
      onOpenPdf(pdfInfo);
    } catch (error) {
      console.error('Failed to open reference:', error);
      alert('Failed to open reference. Please try again.');
    }
  };

  return (
    <ChatBoxContainer>
      <ChatBoxSender>{message.sender === 'human' ? '👋 You' : '🤖 AI'}</ChatBoxSender>
      <ChatBoxContent>
        <div>
          {transformedMessage && transformedMessage.length > 0 ? (
            transformedMessage.map(([sentence, tag], index) => {
              const label = labels.find((label) => label.name === tag);
              return (
                <span key={index} style={{ color: label?.checked ? label.color : 'inherit' }}>
                  {sentence} {label?.checked && `(${tag}) `}
                </span>
              );
            })
          ) : (
            <ReactMarkdown>{message.content}</ReactMarkdown>
          )}

          {message.sender === 'human' && sentiment && (
            <span style={{
              fontSize: '0.85rem',
              marginLeft: '8px',
              color: sentimentColor(sentiment),
              whiteSpace: 'nowrap',
            }}>
              [sentiment: {sentiment}]
            </span>
          )}
        </div>
        
        {context && message.sender === 'AI' && (
          <div className="mt-2 text-sm text-gray-600">
            <div className="font-medium mb-1">References:</div>
            <div className="flex flex-wrap gap-2">
              {context.map((ref, i) => (
                <button
                  key={i}
                  onClick={() => handleReferenceClick(ref)}
                  className="px-3 py-1 bg-gray-100 hover:bg-gray-200 rounded-full text-sm transition-colors"
                  title={`From ${ref.sourceName}`}
                >
                  Reference {i + 1}
                </button>
              ))}
            </div>
          </div>
        )}
      </ChatBoxContent>
    </ChatBoxContainer>
  );
}


// AI typing animation while the response is being processed
function AILoadingChatBox() {
  return <TypingAnimation />;
}

export interface SearchConstraint {
  constraint_type: "EqualTo";
  value: string;
}

// Using Record directly instead of a custom interface
type SearchConstraints = Record<string, SearchConstraint>;

export default function Chat({
  piiWorkflowId, // Workflow ID for pii detection
  sentimentWorkflowId, // Workflow ID for sentiment classification
  provider,
}: {
  piiWorkflowId: string | null;
  sentimentWorkflowId: string | null;
  provider: string;
}) {
  const modelService = useContext<ModelService | null>(ModelServiceContext);
  const { predictSentiment } = useSentimentClassification(sentimentWorkflowId); // Use new hook for sentiment classification

  const [chatHistory, setChatHistory] = useState<ChatMessage[]>([]);
  const [textInput, setTextInput] = useState('');
  const [transformedMessages, setTransformedMessages] = useState<Record<number, string[][]>>({});
  const [aiLoading, setAiLoading] = useState(false);
  const scrollableAreaRef = useRef<HTMLElement | null>(null);
  const responseBuffer = useRef<string>('');
  const contextReceived = useRef<boolean>(false);
  const [contextData, setContextData] = useState<Record<number, any>>({});

  useEffect(() => {
    if (modelService && provider) {
      console.log('print the provider', provider)

      // Set the chat settings based on the provider
      modelService
        .setChat(provider)
        .then(() => {
          // After setting chat settings, fetch chat history
          modelService.getChatHistory(provider).then(setChatHistory);
        })
        .catch((e) => {
          console.error('Failed to update chat settings:', e);
        });
    }
  }, [modelService, provider]);

  const performPIIDetection = (messageContent: string): Promise<string[][]> => {
    if (!piiWorkflowId) {
      return Promise.resolve([]);
    }

    return piiDetect(messageContent, piiWorkflowId)
      .then((result) => {
        const { tokens, predicted_tags } = result.data.prediction_results;

        let transformed: string[][] = [];
        let currentSentence = '';
        let currentTag = '';

        console.log('tokens', tokens)
        console.log('predicted_tags', predicted_tags)

        for (let i = 0; i < tokens.length; i++) {
          const word = tokens[i];
          const tag = predicted_tags[i] && predicted_tags[i][0];

          if (tag !== currentTag) {
            if (currentSentence) {
              transformed.push([currentSentence.trim(), currentTag]);
            }
            currentSentence = word;
            currentTag = tag;
          } else {
            currentSentence += ` ${word}`;
          }
        }

        if (currentSentence) {
          transformed.push([currentSentence.trim(), currentTag]);
        }

        return transformed;
      })
      .catch((error) => {
        console.error('Error detecting PII:', error);
        return [];
      });
  };

  const [sentiments, setSentiments] = useState<Record<number, string>>({}); // Store sentiment for human messages

  // Function to classify sentiment and store the highest sentiment score
  const classifySentiment = async (messageContent: string, messageIndex: number) => {
    if (!sentimentWorkflowId) {
      return;
    }

    try {
      const result = await predictSentiment(messageContent);
      const predictions = result.predicted_classes; // Array of [sentiment, score]
      console.log('Sentiment Prediction:', result);

      // Find the sentiment with the highest score
      const [maxSentiment, maxScore] = predictions.reduce((prev, current) => {
        return current[1] > prev[1] ? current : prev;
      });

      // Special case: if sentiment is 'positive', 'neutral', or 'negative', apply the 0.7 threshold
      let finalSentiment = maxSentiment;
      if (['positive', 'negative', 'neutral'].includes(maxSentiment)) {
        if ((maxSentiment === 'positive' || maxSentiment === 'negative') && maxScore < 0.7) {
          finalSentiment = 'neutral'; // Override to 'neutral' if score is below 0.7
        }
      }

      // For any other sentiment labels (not positive/negative/neutral), use the highest score directly
      setSentiments((prev) => ({
        ...prev,
        [messageIndex]: finalSentiment, // Save the final sentiment for this message
      }));
    } catch (error) {
      console.error('Error classifying sentiment:', error);
    }
  };

  const [abortController, setAbortController] = useState<AbortController | null>(null);

  const handleEnterPress = async (e: any) => {
    if (e.keyCode === 13 && e.shiftKey === false) {
      e.preventDefault();
      if (!textInput.trim()) return;

      // If a generation is already in progress, abort it
      if (abortController) {
        abortController.abort();
        setAiLoading(false);
      }

      const controller = new AbortController();
      setAbortController(controller);

      const lastTextInput = textInput;
      const lastChatHistory = chatHistory;
      const currentIndex = chatHistory.length;
      const aiIndex = chatHistory.length + 1;

      setAiLoading(true);
      setChatHistory(history => [...history, { sender: 'human', content: lastTextInput }]);
      setTextInput('');

      // Trigger sentiment classification if classifier exists
      if (sentimentWorkflowId) {
        classifySentiment(lastTextInput, currentIndex);
      }
      responseBuffer.current = '';
      contextReceived.current = false;

      // Perform PII detection on the human's message
      if (piiWorkflowId) {
        const humanTransformed = await performPIIDetection(lastTextInput);
        setTransformedMessages((prev) => ({
          ...prev,
          [currentIndex]: humanTransformed,
        }));
      }

      try {
        const detectedPII = await performPIIDetection(lastTextInput);
        const searchConstraints: SearchConstraints = {};
        detectedPII.forEach(([text, tag]) => {
          if (tag === 'BRAND' || tag === 'MODEL_NUMBER') {
            searchConstraints[tag] = {
              constraint_type: "EqualTo",
              value: text.trim()
            };
          }
        });

        await modelService?.chat(
          lastTextInput,
          provider,
          searchConstraints,
          (newData: string) => {
            if (newData.startsWith('context:')) {
              try {
                const contextJson = JSON.parse(newData.substring(9));
                setContextData(prev => ({
                  ...prev,
                  [aiIndex]: contextJson
                }));
                contextReceived.current = true;
              } catch (e) {
                console.error('Error parsing context:', e);
              }
            } else {
              if (!contextReceived.current) {
                responseBuffer.current += newData;
              } else {
                if (responseBuffer.current) {
                  setChatHistory(history => {
                    const newHistory = [...history];
                    newHistory.push({ sender: 'AI', content: responseBuffer.current + newData });
                    return newHistory;
                  });
                  responseBuffer.current = '';
                } else {
                  setChatHistory(history => {
                    const newHistory = [...history];
                    if (newHistory[newHistory.length - 1].sender === 'AI') {
                      newHistory[newHistory.length - 1].content += newData;
                    } else {
                      newHistory.push({ sender: 'AI', content: newData });
                    }
                    return newHistory;
                  });
                }
              }
            }
          },
          async (finalResponse: string) => {
            if (piiWorkflowId) {
              const cleanResponse = finalResponse.replace(/^context:.*?\]/, '').trim();
              const aiTransformed = await performPIIDetection(cleanResponse);
              setTransformedMessages(prev => ({
                ...prev,
                [aiIndex]: aiTransformed,
              }));
            }
            setAiLoading(false);
            setAbortController(null);
            responseBuffer.current = '';
            contextReceived.current = false;
          },
          controller.signal
        );
      } catch (error) {
        console.error(error, lastChatHistory, lastTextInput);
      }
    }
  };

  const [pdfInfo, setPdfInfo] = useState<PdfInfo | null>(null);
  const [selectedPdfChunk, setSelectedPdfChunk] = useState<Chunk | null>(null);
  
  const handleOpenPdf = (info: PdfInfo) => {
    setPdfInfo(info);
    setSelectedPdfChunk(info.highlighted);
  };

  return (
    <ChatContainer>
      {pdfInfo && (
        <PdfViewerWrapper>
          <PdfViewer
            name={pdfInfo.filename}
            src={pdfInfo.source}
            chunks={pdfInfo.docChunks}
            initialChunk={pdfInfo.highlighted}
            onSelect={setSelectedPdfChunk}
            onClose={() => {
              setSelectedPdfChunk(null);
              setPdfInfo(null);
            }}
          />
        </PdfViewerWrapper>
      )}
      <ScrollableArea ref={scrollableAreaRef}>
        {chatHistory && chatHistory.length ? (
          <AllChatBoxes>
            {chatHistory.map((message, i) => (
              <ChatBox
                key={i}
                modelService={modelService}
                message={message}
                transformedMessage={piiWorkflowId ? transformedMessages[i] : undefined}
                sentiment={sentiments[i]}
                context={contextData[i]}
                onOpenPdf={handleOpenPdf}
              />
            ))}
            {aiLoading && (
              <AILoadingWrapper>
                <AILoadingChatBox />
              </AILoadingWrapper>
            )}
          </AllChatBoxes>
        ) : (
          <Placeholder> Ask anything to start chatting! </Placeholder>
        )}
      </ScrollableArea>
      <ChatBarContainer>
        <ChatBar
          placeholder="Ask anything..."
          onKeyDown={handleEnterPress}
          value={textInput}
          onChange={(e) => setTextInput(e.target.value)}
        />
        {abortController && aiLoading && (
          <PauseButton
            onClick={() => {
              abortController.abort();
              setAbortController(null);
              setAiLoading(false);
            }}
          >
            <FontAwesomeIcon icon={faStop} style={{ color: 'white', fontSize: '16px' }} />
          </PauseButton>
        )}
      </ChatBarContainer>
    </ChatContainer>
  );
}
