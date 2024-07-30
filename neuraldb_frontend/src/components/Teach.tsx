import React, { useCallback, useContext, useRef, useState } from "react";
import styled from "styled-components";
import { borderRadius, color, duration } from "../stylingConstants";
import { ReactComponent as TeachSVG } from "../assets/icons/teach.svg";
import TeachPanel from "./TeachPanel";
import useClickOutside from "./hooks/useClickOutside";
import { ModelServiceContext } from "../Context";
import { ModelService } from "../modelServices";

const Container = styled.section`
    width: 50px;
    height: 30px;
    display: flex;
    flex-direction: column;
    overflow: visible;
    padding-right: ${borderRadius.card};
    align-items: flex-end;
`;

function buttonBorderRadius(props: { $active: boolean }) {
    return props.$active
        ? `${borderRadius.card} ${borderRadius.card} 0 0`
        : borderRadius.card;
}

const Button = styled.button<{ $active: boolean }>`
    border: none;
    background-color: ${(props) => (props.$active ? color.accent : "white")};
    padding: 5px 10px;
    border-radius: ${buttonBorderRadius};
    transition-duration: ${duration.transition};

    &:hover {
        cursor: pointer;
        background-color: ${(props) =>
            props.$active ? color.accent : color.accentExtraLight};
    }
`;

const TeachIcon = styled(TeachSVG)<{ $active: boolean }>`
    transition-duration: ${duration.transition};
    width: 30px;
    path {
        stroke: ${(props) => (props.$active ? "white" : color.accent)};
    }
`;

const PanelContainer = styled.section`
    position: relative;
    right: -${borderRadius.card};
    width: 300px;
`;

export default function Teach() {
    const [showPanel, setShowPanel] = useState(false);
    const [question, setQuestion] = useState("");
    const [answer, setAnswer] = useState("");
    const modelService = useContext<ModelService>(ModelServiceContext);

    const containerRef = useRef<HTMLElement>();

    const handleClickOutside = useCallback(() => {
        setShowPanel(false);
    }, []);

    useClickOutside(containerRef, handleClickOutside);

    function togglePanel() {
        setShowPanel((prev) => !prev);
    }

    return (
        <Container ref={containerRef}>
            <Button onClick={togglePanel} $active={showPanel}>
                <TeachIcon $active={showPanel} />
            </Button>
            {showPanel && (
                <PanelContainer>
                    <TeachPanel
                        question={question}
                        answer={answer}
                        canAddAnswer={true}
                        setQuestion={setQuestion}
                        setAnswer={setAnswer}
                        onAddAnswer={(q, a) => modelService.qna(q, a)}
                        onAssociate={(q, a) => modelService.associate(q, a)}
                    />
                </PanelContainer>
            )}
        </Container>
    );
}
