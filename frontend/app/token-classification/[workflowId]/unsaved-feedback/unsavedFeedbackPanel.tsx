import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { UpdatedEntity } from "./useUnsavedFeedback";
import { Trash } from "lucide-react";
import styled from 'styled-components';
import { Button } from "@/components/ui/button";

const RedHoverTrash = styled(Trash)`
  transition: 0.2s;
  &:hover {
    color: red;
    cursor: pointer;
  }
`;

interface UnsavedFeedbackPanelProps {
  feedbackEntries: UpdatedEntity[];
  tagColors: Record<string, string>;
};

export default function UnsavedFeedback({feedbackEntries, tagColors}: UnsavedFeedbackPanelProps) {

  return <Card style={{textAlign: "left"}}>
    <CardHeader>
      <CardTitle>
        Review Feedback
      </CardTitle>
    </CardHeader>
    <CardContent>
      {
        feedbackEntries.map((feedback, index) => (
          <div
            className="text-md text-gray-500"
            key={index}
            style={{
              display: 'flex',
              flexDirection: 'column',
              marginBottom: '20px',
              wordBreak: 'break-word'
            }}
          >
            <div
              className="border-b"
              style={{
                display: 'block',
                width: '100%',
                marginBottom: '20px'
              }}
            />
            <div style={{display: 'flex', flexDirection: 'row', justifyContent: 'space-between', alignItems: 'start'}}>
                {feedback.text}
              <RedHoverTrash onClick={feedback.remove}/>
            </div>
            <span
              style={{
                backgroundColor: tagColors[feedback.tag] ?? '#5CC96E',
                width: 'fit-content',
                padding: '4px 10px',
                marginTop: '10px',
                borderRadius: '4px',
                color: 'white'
              }}
            >
              {feedback.tag}
            </span>
          </div>
        ))
      }
      <Button>Submit</Button>
    </CardContent>
  </Card>
}