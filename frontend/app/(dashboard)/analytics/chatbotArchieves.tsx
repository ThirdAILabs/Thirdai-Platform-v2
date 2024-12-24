import React, { useState, useEffect } from 'react';
import { getAllChatHistory } from '@/lib/backend';

interface ConversationData {
  query_time: string;
  query_text: string;
  response_time: string;
  response_text: string;
}

interface Data {
  [sessionId: string]: ConversationData[]; // Maps session IDs to an array of conversations
}

function getUrlParams() {
  const url = new URL(window.location.href);
  const params = new URLSearchParams(url.search);
  const userName = params.get('username');
  const modelName = params.get('model_name');
  const model_id = params.get('old_model_id');
  const default_mode = params.get('default');
  return { userName, modelName, model_id, default_mode };
}

const Conversations: React.FC = () => {
  // State variable for storing chat history
  const [chatHistory, setChatHistory] = useState<Data | null>(null);

  // Extract parameters from the URL
  const { model_id, default_mode } = getUrlParams();

  useEffect(() => {
    if (default_mode === 'chat' && model_id !== null) {
      const getChatData = async () => {
        try {
          const data = await getAllChatHistory(model_id);
          setChatHistory(data);
        } catch (error) {
          console.error('Error fetching chat history:', error);
        }
      };
      getChatData();
    }
  }, [model_id, default_mode]);
  console.log('ChatHistory: ', chatHistory);
  // State for tracking expanded session
  const [expandedSession, setExpandedSession] = useState<string | null>(null);

  const toggleSession = (sessionId: string) => {
    setExpandedSession(expandedSession === sessionId ? null : sessionId);
  };

  if (!chatHistory) {
    return <div className="p-4 text-center">Loading chat history...</div>;
  }

  return (
    <div className="p-4">
      {Object.entries(chatHistory).map(([sessionId, conversations]) => (
        <div key={sessionId} className="mb-4 border border-gray-300 rounded-lg">
          <div
            className="p-4 bg-gray-100 cursor-pointer rounded-lg flex justify-between items-center"
            onClick={() => toggleSession(sessionId)}
          >
            <span className="font-semibold">Session ID: {sessionId}</span>
            <span>{expandedSession === sessionId ? '-' : '+'}</span>
          </div>
          {expandedSession === sessionId && (
            <div className="p-4">
              {conversations.map((conversation, index) => (
                <div key={index} className="mb-4 border border-gray-200 rounded-lg p-4 bg-white">
                  <div className="flex flex-col">
                    <div className=" border shadow-sm p-4 rounded-lg w-fit">
                      {/* <div className="font-bold">Query:</div> */}
                      <div className="text-gray-700">{conversation.query_text}</div>
                      <div className="text-xs text-gray-500">{conversation.query_time}</div>
                    </div>
                    <div className=" mt-4 bg-green-100 border shadow-sm p-4 rounded-lg w-fit">
                      {/* <div className="font-bold">Response:</div> */}
                      <div className="text-gray-700">{conversation.response_text}</div>
                      <div className="text-xs text-gray-500">{conversation.response_time}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

export default Conversations;
