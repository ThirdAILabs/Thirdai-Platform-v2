import React from 'react';
import styled from 'styled-components';
import SaveButton from './buttons/SaveButton';
import Teach from './Teach';

const PanelContainer = styled.div`
  position: fixed;
  right: -300px;
  top: 0;
  width: 300px;
  height: 100%;
  background-color: white;
  box-shadow: -2px 0 5px rgba(0,0,0,0.1);
  transition: right 0.3s ease-in-out;
  padding: 20px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 25px;

  &:hover {
    right: 0;
  }
`;

const PanelTitle = styled.h2`
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 10px 0;
  padding-bottom: 15px;
  border-bottom: 1px solid #e0e0e0;
`;

const SectionContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: 15px;
`;

const SectionTitle = styled.h3`
  font-size: 20px;
  font-weight: 600;
  margin: 0;
`;

const SectionContent = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const SectionText = styled.span`
  font-size: 16px;
  flex: 1;
  margin-right: 10px;
`;

const ButtonContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  width: 48px;
  height: 48px;
`;

interface SidePanelProps {
  cacheEnabled: boolean;
  setCacheEnabled: (enabled: boolean) => void;
  onSaveClick: () => void;
}

const SidePanel: React.FC<SidePanelProps> = ({ cacheEnabled, setCacheEnabled, onSaveClick }) => {
  return (
    <PanelContainer>
      <PanelTitle>Advanced Configuration</PanelTitle>
      
      <SectionContainer>
        <SectionTitle>Cache Control</SectionTitle>
        <SectionContent>
          <SectionText>Enable cache for faster responses</SectionText>
          <div className="flex items-center">
            <button
              onClick={() => setCacheEnabled(!cacheEnabled)}
              className={`relative inline-flex items-center h-6 rounded-full w-11 transition-colors duration-300 focus:outline-none ${
                cacheEnabled ? 'bg-blue-500' : 'bg-gray-300'
              }`}
            >
              <span
                className={`transform transition-transform duration-300 inline-block w-4 h-4 bg-white rounded-full ${
                  cacheEnabled ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
          </div>
        </SectionContent>
      </SectionContainer>

      <SectionContainer>
        <SectionTitle>Save Model</SectionTitle>
        <SectionContent>
          <SectionText>Save current model configuration</SectionText>
          <ButtonContainer>
            <SaveButton onClick={onSaveClick} />
          </ButtonContainer>
        </SectionContent>
      </SectionContainer>

      <SectionContainer>
        <SectionTitle>Teaching Mode</SectionTitle>
        <SectionContent>
          <SectionText>Enter teaching mode to train the model</SectionText>
          <ButtonContainer>
            <Teach />
          </ButtonContainer>
        </SectionContent>
      </SectionContainer>
    </PanelContainer>
  );
};

export default SidePanel;