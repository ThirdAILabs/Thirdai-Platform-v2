import React from "react";
import styled from "styled-components";
import PromptSVG from "../../assets/icons/prompt.svg";
import { borderRadius, color, duration } from "../../stylingConstants";

const Button = styled.button<{ $active: boolean }>`
    border: none;
    background-color: ${(props) => (props.$active ? color.accent : "white")};
    border-radius: ${borderRadius.textInput};
    transition-duration: ${duration.transition};

    &:hover {
        cursor: pointer;
        background-color: ${(props) =>
            props.$active ? color.accent : color.accentExtraLight};
    }
`;

const PromptIcon = styled(PromptSVG)<{ $active: boolean }>`
    transition-duration: ${duration.transition};
    margin-bottom: -10px;
    width: 25px;
    path {
        ${(props) => (props.$active ? "fill: white" : "fill: $color.accent")};
    }
`;

export default function PromptToggle(props: {
    active: boolean;
    onClick: () => void;
}) {
    return (
        <Button $active={props.active} onClick={props.onClick}>
            <PromptIcon $active={props.active} />
        </Button>
    );
}
