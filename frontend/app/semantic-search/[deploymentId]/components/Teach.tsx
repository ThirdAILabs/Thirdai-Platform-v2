import React, { useCallback, useContext, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { ModelServiceContext } from '../Context';
import { ModelService } from '../modelServices';
import TeachSVG from '../assets/icons/teach.svg';
import TeachPanel from './TeachPanel';
import useClickOutside from './hooks/useClickOutside';

const Teach = () => {
  const [showPanel, setShowPanel] = useState(false);
  const [question, setQuestion] = useState('');
  const [answer, setAnswer] = useState('');
  const modelService = useContext<ModelService | null>(ModelServiceContext);
  const containerRef = useRef<HTMLDivElement>(null);

  const handleClickOutside = useCallback(() => {
    setShowPanel(false);
  }, []);

  useClickOutside(containerRef, handleClickOutside);

  const togglePanel = () => {
    setShowPanel(prev => !prev);
  };

  return (
    <div className="relative" ref={containerRef}>
      <Button 
        className="w-[60px] h-[50px] flex items-center justify-center"
        onClick={togglePanel}
      >
        <TeachSVG
          className={`w-10 transition-transform duration-300 ${
            showPanel ? 'text-primary' : ''
          }`}
        />
      </Button>
      
      {showPanel && (
        <div className="absolute top-0 right-full mr-2 w-[300px]">
          <TeachPanel
            question={question}
            answer={answer}
            canAddAnswer={true}
            setQuestion={setQuestion}
            setAnswer={setAnswer}
            onAddAnswer={(q, a) => modelService?.qna(q, a)}
            onAssociate={(q, a) => modelService?.associate(q, a)}
          />
        </div>
      )}
    </div>
  );
};

export default Teach;