// app/NLPQuestions.js
import React, { useState } from 'react';
import NERQuestions from './ner-questions';
import SCQQuestions from './sentence-classification-questions'

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
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-6">NLP Use Case Finder</h1>
        <div className="p-4 max-w-lg mx-auto bg-white shadow-lg rounded-lg">
          <form onSubmit={handleSubmit} className="flex flex-col space-y-4">

            <input
              type="text"
              className="w-full p-4 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              value={question}
              onChange={handleInputChange}
              placeholder="Please describe your use case..."
            />

            <button
              className="w-full py-3 px-6 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition duration-300"
              type="submit"
            >
              <i className="bi bi-send mr-2"></i>Submit
            </button>
          </form>
          <div className="mt-4 text-sm text-gray-600">
            <div>Example: I want to analyze my customers&apos; reviews.</div>
            <div>Example: I want to analyze the individual tokens within a report document.</div>
          </div>
          {answer && (
            <div className="mt-6">
              <div className="p-4 bg-gray-100 rounded-lg">{answer}</div>
              <div className="flex justify-end mt-4">
                <button
                  className="py-2 px-4 border-2 border-blue-600 text-blue-600 rounded-lg hover:bg-blue-600 hover:text-white transition duration-300"
                  onClick={() => console.log('Build action')}
                >
                  Sounds good. Let&apos;s start building it.
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {answer.includes('Sentence classification') ? (
              <SCQQuestions question = {question} answer = {answer}/>
            ) : answer.includes('Token classification') ? (
              <NERQuestions />
            ) : null}
    </div>
  );
};

export default NLPQuestions;
