import React, { useState } from 'react';
import styled from 'styled-components';
import { borderRadius, color, fontSizes, shadow } from '../stylingConstants';
import { Pad, Spacer } from './Layout';
import PillButton from './buttons/PillButton';
import NotifyingClickable from './buttons/NotifyingButton';

const Panel = styled.section`
  display: flex;
  flex-direction: column;
  padding: 10px;
  box-shadow: ${shadow.card};
  border-radius: ${borderRadius.card};
  background-color: white;
`;

const Label = styled.section`
  font-weight: bold;
  font-size: ${fontSizes.s};
`;

const TextInput = styled.textarea`
  background-color: ${color.textInput};
  font-size: ${fontSizes.s};
  padding: 10px 10px 13px 10px;
  border-radius: ${borderRadius.textInput};
  border: none;
  height: 100px;
  resize: none;
  font-family: Helvetica, Arial, sans-serif;
`;

const ButtonContainer = styled.section`
  width: 100%;
  display: flex;
  flex-direction: row;
  justify-content: center;
`;

const ErrorMessage = styled.div`
  color: red;
  font-size: ${fontSizes.s};
  text-align: center;
`;

interface TeachPanelProps {
  question: string;
  answer: string;
  canAddAnswer: boolean;
  setQuestion: (question: string) => void;
  setAnswer: (answer: string) => void;
  onAssociate: (question: string, answer: string) => void;
  onAddAnswer: (question: string, answer: string) => void;
}

export default function TeachPanel({
  question,
  answer,
  canAddAnswer,
  setQuestion,
  setAnswer,
  onAssociate,
  onAddAnswer,
}: TeachPanelProps) {
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  function handleButtonClick(onSubmit: (question: string, answer: string) => void) {
    if (question.trim() === '' || answer.trim() === '') {
      setErrorMessage('Question and answer fields cannot be empty.');
    } else {
      setErrorMessage(null);
      onSubmit(question, answer);
    }
  }

  return (
    <Panel>
      <Pad $left="2px">
        <Label>Teach</Label>
      </Pad>
      <Spacer $height="10px" />
      <TextInput
        placeholder="Question goes here"
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
      />
      <Spacer $height="10px" />
      <TextInput
        placeholder="Answer goes here"
        value={answer}
        onChange={(e) => setAnswer(e.target.value)}
      />
      <Spacer $height="15px" />
      {errorMessage && <ErrorMessage>{errorMessage}</ErrorMessage>}
      <ButtonContainer>
        {question && answer ? (
          <NotifyingClickable
            text="Feedback received!"
            onClick={() => handleButtonClick(onAssociate)}
          >
            <PillButton>Associate</PillButton>
          </NotifyingClickable>
        ) : (
          <PillButton
            onClick={() => setErrorMessage('Question and answer fields cannot be empty.')}
          >
            Associate
          </PillButton>
        )}
      </ButtonContainer>

      <Spacer $height="15px" />
    </Panel>
  );
}
