import React, { useContext, useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import styled from 'styled-components';
import { borderRadius, color, duration, fontSizes, padding } from '../../stylingConstants';
import { ModelServiceContext } from '../../Context';
import { ChatMessage, ModelService } from '../../modelServices';
import TypingAnimation from '../TypingAnimation';

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

const ChatBoxContainer = styled.section`
  /* border-bottom: solid 1px black; */
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

function ChatBox({ message, transformedMessage }: { message: ChatMessage, transformedMessage?: string[][] }) {
  return (
    <ChatBoxContainer>
      <ChatBoxSender>{message.sender === 'human' ? '👋 You' : '🤖 AI'}</ChatBoxSender>
      <ChatBoxContent>
        {transformedMessage && transformedMessage.length > 0
          ? transformedMessage.map(([sentence, tag], index) => {
              const label = labels.find((label) => label.name === tag);
              return (
                <span
                  key={index}
                  style={{
                    color: label?.checked ? label.color : 'inherit',
                  }}
                >
                  {sentence} {label?.checked && `(${tag}) `}
                </span>
              );
            })
          : <ReactMarkdown>{message.content}</ReactMarkdown> // Render without PII highlighting if no transformation is available
        }
      </ChatBoxContent>
    </ChatBoxContainer>
  );
}

function AILoadingChatBox() {
  return (
    <ChatBoxContainer>
      <ChatBoxSender>🤖 AI</ChatBoxSender>
      <TypingAnimationContainer>
        <TypingAnimation />
      </TypingAnimationContainer>
    </ChatBoxContainer>
  );
}

const ChatBar = styled.textarea`
  background-color: ${color.textInput};
  /* width: 100%; */
  font-size: ${fontSizes.m};
  padding: 20px 20px 27px 20px;
  margin: 10px 0 50px 0%;
  border-radius: ${borderRadius.textInput};
  outline: none;
  border: none;
  transition-duration: ${duration.transition};
  height: ${fontSizes.xl};
  resize: none;
  font-family: Helvetica, Arial, sans-serif;

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
    name: 'PHONENUMBER',
    color: 'blue',
    amount: '217,323',
    checked: true,
    description:
      'The format of a US phone number is (XXX) XXX-XXXX, where "X" represents a digit from 0 to 9. It consists of a three-digit area code, followed by a three-digit exchange code, and a four-digit line number.',
  },
  {
    id: 2,
    name: 'SSN',
    color: 'orange',
    amount: '8,979',
    checked: true,
    description:
      'The format of a US Social Security Number (SSN) is XXX-XX-XXXX, where "X" represents a digit from 0 to 9. It consists of three parts: area, group, and serial numbers.',
  },
  {
    id: 3,
    name: 'CREDITCARDNUMBER',
    color: 'red',
    amount: '13,272',
    checked: true,
    description:
      'A US credit card number is a 16-digit number typically formatted as XXXX XXXX XXXX XXXX, where "X" represents a digit from 0 to 9. It includes the Issuer Identifier, account number, and a check digit.',
  },
  {
    id: 4,
    name: 'LOCATION',
    color: 'green',
    amount: '2,576,904',
    checked: true,
    description: `A US address format includes the recipient's name, street address (number and name), city, state abbreviation, and ZIP code, for example: John Doe 123 Main St Springfield, IL 62701`,
  },
  {
    id: 5,
    name: 'NAME',
    color: 'purple',
    amount: '1,758,131',
    checked: true,
    description: `An English name format typically consists of a first name, middle name(s), and last name (surname), for example: John Michael Smith. Titles and suffixes, like Mr. or Jr., may also be included.`,
  },
];

export default function Chat(props: any) {
  const modelService = useContext<ModelService | null>(ModelServiceContext);

  const [chatHistory, setChatHistory] = useState<ChatMessage[]>([]);
  const [textInput, setTextInput] = useState('');
  const [transformedMessages, setTransformedMessages] = useState<Record<number, string[][]>>({}); // Store transformed messages for both human and AI
  const [aiLoading, setAiLoading] = useState(false);
  const scrollableAreaRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    modelService?.getChatHistory().then(setChatHistory);
  }, []);

  const performPIIDetection = (messageContent: string): Promise<string[][]> => {
    if (!modelService) {
      return Promise.resolve([]);
    }

    return modelService.piiDetect(messageContent)
      .then((result) => {
        const { tokens, predicted_tags } = result;
        let transformed: string[][] = [];
        let currentSentence = '';
        let currentTag = '';

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

  const handleEnterPress = async (e: any) => {
    if (e.keyCode === 13 && e.shiftKey === false) {
      e.preventDefault();
      if (aiLoading || !textInput.trim()) return;

      const lastTextInput = textInput;
      const lastChatHistory = chatHistory;
      const currentIndex = chatHistory.length;

      setAiLoading(true);
      setChatHistory((history) => [...history, { sender: 'human', content: textInput }]);
      setTextInput('');

      // Perform PII detection on the human's message
      const humanTransformed = await performPIIDetection(lastTextInput);
      setTransformedMessages((prev) => ({
        ...prev,
        [currentIndex]: humanTransformed, // Store human's PII-detected message
      }));

      // Simulate AI response
      modelService?.chat(lastTextInput)
        .then(async ({ response }) => {
          const aiIndex = chatHistory.length + 1;
          setChatHistory((history) => [...history, { sender: 'AI', content: response }]);

          // Perform PII detection on the AI's response
          const aiTransformed = await performPIIDetection(response);
          setTransformedMessages((prev) => ({
            ...prev,
            [aiIndex]: aiTransformed, // Store AI's PII-detected message
          }));

          setAiLoading(false);
        })
        .catch((error) => {
          alert('Failed to send chat. Please try again.');
          setChatHistory(lastChatHistory);
          setTextInput(lastTextInput);
          setAiLoading(false);
        });
    }
  };

  return (
    <ChatContainer>
      <ScrollableArea ref={scrollableAreaRef}>
        {chatHistory.length ? (
          <AllChatBoxes>
            {chatHistory.map((message, i) => (
              <ChatBox
                key={i}
                message={message}
                transformedMessage={transformedMessages[i]} // Pass PII-transformed message for human and AI
              />
            ))}
            {aiLoading && <AILoadingChatBox />}
          </AllChatBoxes>
        ) : (
          <Placeholder> Ask anything to start chatting! </Placeholder>
        )}
      </ScrollableArea>
      <ChatBar
        placeholder="Ask anything..."
        onKeyDown={handleEnterPress}
        value={textInput}
        onChange={(e) => setTextInput(e.target.value)}
      />
    </ChatContainer>
  );
}
