// app/NLPQuestions.js
import React, { useState } from 'react';

const NLPQuestions = () => {
  const [question, setQuestion] = useState('');
  const [answer, setAnswer] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuestion(e.target.value);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!question) {
      console.error('Question is not valid');
      return;
    }

    try {
      const response = await fetch('/api/which-nlp-use-case', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ question }),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const result = await response.json();
      setAnswer(result.answer);
    } catch (error) {
      console.error('Error during fetch:', error);
    }
  };

  return (
    <div>
      <form onSubmit={handleSubmit} className="intro-input-wrapper">
        <input
          type="text"
          className="intro-input"
          value={question}
          onChange={handleInputChange}
          placeholder="Please describe your use case..."
        />
        <button className="btn bg-primary" type="submit">
          <i className="bi bi-send text-white"></i>
        </button>
      </form>
      <div className="font-sm text-secondary mt-2">Example: I want to analyze my customers' reviews.</div>
      <div className="font-sm text-secondary mt-1 mb-4">Example: I want to analyze the individual tokens within a report document.</div>
      <div>
        {answer && (
          <div>
            <div>{answer}</div>
            <div className="w-100 d-flex justify-content-end mt-3">
              <button className="btn btn-outline-primary border-2" onClick={() => console.log('Build action')}>
                Sounds good. Let's start building it.
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default NLPQuestions;
