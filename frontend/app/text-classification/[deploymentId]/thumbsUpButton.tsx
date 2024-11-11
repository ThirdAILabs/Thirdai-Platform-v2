import { ThumbsUp } from 'lucide-react';
import { Typography } from '@mui/material';
import { useState } from 'react';
import axios from 'axios';
import { deploymentBaseUrl } from '@/lib/backend';
import { useParams } from 'next/navigation';

interface ThumbsUpButtonProps {
  inputText: string;
  prediction: string;
}

const ThumbsUpButton: React.FC<ThumbsUpButtonProps> = ({ inputText, prediction }) => {
  const [isActive, setIsActive] = useState(false);
  const [feedbackMessage, setFeedbackMessage] = useState('');
  const params = useParams();
  const accessToken = localStorage.getItem('accessToken');

  const handleFeedback = async () => {
    setIsActive(true);
    setFeedbackMessage('Feedback received');

    // Send feedback to backend
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
    try {
      await axios.post(`${deploymentBaseUrl}/add_sample`, {
        text: inputText,
        label: prediction
      });
    } catch (error) {
      console.error('Error sending feedback:', error);
      setFeedbackMessage('Error sending feedback');
    }

    // Reset the feedback message and thumbs-up icon after 5 seconds
    setTimeout(() => {
      setIsActive(false);
      setFeedbackMessage('');
    }, 3000);
  };

  return (
    <div className="flex items-center gap-2">
      <button
        onClick={handleFeedback}
        className={`p-2 rounded-full transition-all duration-200 ${isActive
          ? 'bg-blue-100 text-blue-600'
          : 'hover:bg-gray-100 text-gray-600'
          }`}
      >
        <ThumbsUp
          size={24}
          className={`transition-transform duration-200 ${isActive ? 'scale-110 fill-current' : ''
            }`}
        />
      </button>
      {feedbackMessage && (
        <Typography variant="caption" color="textSecondary">
          {feedbackMessage}
        </Typography>
      )}
    </div>
  );
};

export default ThumbsUpButton;