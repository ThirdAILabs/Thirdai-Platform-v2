import React, { useContext } from 'react';
import NotifyingClickable from './NotifyingButton';
import Copy from '../../assets/icons/copy.svg';
import { styled } from 'styled-components';
import { color, duration } from '../../stylingConstants';
import { ModelServiceContext } from '../../Context';
import { ModelService } from '../../modelServices';

// Define interfaces for props and feedback
interface CopyButtonProps {
  toCopy: string;
  referenceID: number;
  queryText: string;
}

interface ImplicitFeedback {
  reference_id: number;
  reference_rank: number;
  query_text: string;
  event_desc: string;
}

const StyledCopy = styled(Copy)`
  transition-duration: ${duration.transition};

  &:hover rect {
    fill: ${color.accent};
    transition-duration: ${duration.transition};
  }

  &:hover {
    cursor: pointer;
  }
`;

export default function CopyButton({ toCopy, referenceID, queryText }: CopyButtonProps) {
  const modelService = useContext<ModelService | null>(ModelServiceContext);

  const copyToClipboard = async (): Promise<void> => {
    try {
      // Try modern clipboard API first
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(toCopy);
      } else {
        // Use fallback immediately if modern API is not available
        const textArea = document.createElement('textarea');
        textArea.value = toCopy;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        document.body.appendChild(textArea);
        textArea.select();
        const successful = document.execCommand('copy');
        document.body.removeChild(textArea);

        if (!successful) {
          throw new Error('Failed to copy using fallback method');
        }
      }

      // Log implicit feedback after successful copy
      const feedback: ImplicitFeedback = {
        reference_id: referenceID,
        reference_rank: 0, // TODO: fill with exact rank
        query_text: queryText,
        event_desc: 'copy_reference_text',
      };

      console.log('feedback logged', feedback);

      // Record feedback if modelService is available
      if (modelService) {
        try {
          const data = await modelService.recordImplicitFeedback(feedback);
          console.log('Implicit feedback recorded successfully:', data);
        } catch (error) {
          // Log the error but don't disrupt the user experience
          console.warn("Failed to record feedback - this won't affect the copy operation:", error);

          // Only log detailed error info in development
          if (process.env.NODE_ENV === 'development') {
            console.debug('Feedback that failed to record:', feedback);
            console.debug('Full error:', error);
          }
        }
      }
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      alert('Unable to copy to clipboard. Please try again or copy manually.');
    }
  };

  return (
    <NotifyingClickable onClick={copyToClipboard} text="Copied to clipboard!">
      <StyledCopy />
    </NotifyingClickable>
  );
}
