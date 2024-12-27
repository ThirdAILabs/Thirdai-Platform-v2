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
function convertAndSortDataByQueryTime(data: Data): ConversationData[] {
  const conversationList: ConversationData[] = [];
  for (const [sessionId, conversations] of Object.entries(data)) {
    conversations.forEach(conversation => {
      conversationList.push({ ...conversation });
    });
  }
  // Sort the list by query_time
  conversationList.sort((a, b) => new Date(a.query_time).getTime() - new Date(b.query_time).getTime());
  return conversationList;
}

const Conversations: React.FC = () => {
  // State variable for storing chat history
  const [chatHistory, setChatHistory] = useState<ConversationData[]>([]);

  // Extract parameters from the URL
  const { model_id, default_mode } = getUrlParams();

  useEffect(() => {
    if (default_mode === 'chat' && model_id !== null) {
      const getChatData = async () => {
        try {
          const data = await getAllChatHistory(model_id);
          const convertedData = convertAndSortDataByQueryTime(data)
          setChatHistory(convertedData);
        } catch (error) {
          console.error('Error fetching chat history:', error);
        }
      };
      getChatData();
    }
  }, [model_id, default_mode]);
  console.log('ChatHistory: ', chatHistory);
  // State for tracking expanded session

  if (!chatHistory) {
    return <div className="p-4 text-center">Loading chat history...</div>;
  }

  return (
    <div className="mb-4 border border-gray-300 rounded-lg">
      {(chatHistory).map((conversation) => (
        <div key={`${conversation.query_time}-${conversation.response_time}`} >
          <div className="p-4">
            <div className="flex flex-col">
              {/* Query Section */}
              <div className="flex justify-start">
                <div className="border shadow-sm p-4 rounded-lg w-fit max-w-[70%]">
                  <div className="text-gray-700">{conversation.query_text}</div>
                  <div className="text-xs text-gray-500">{conversation.query_time}</div>
                </div>
              </div>

              {/* Response Section */}
              <div className="flex justify-end my-4">
                <div className="bg-green-100 border shadow-sm p-4 rounded-lg w-fit max-w-[70%]">
                  <div className="text-gray-700">{conversation.response_text}</div>
                  <div className="text-xs text-gray-500">{conversation.response_time}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};

export default Conversations;
