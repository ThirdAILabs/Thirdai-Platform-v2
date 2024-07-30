import React from "react";
import styled from "styled-components";
import { ReactComponent as ChatSVG } from "../../assets/icons/chat.svg";
import { borderRadius, color, duration } from "../../stylingConstants";

const Button = styled.button<{ $active: boolean }>`
    border: none;
    background-color: ${(props) => (props.$active ? color.accent : "white")};
    padding: 5px 12px 4px 12px;
    border-radius: ${borderRadius.card};
    transition-duration: ${duration.transition};

    &:hover {
        cursor: pointer;
        background-color: ${(props) =>
            props.$active ? color.accent : color.accentExtraLight};
    }
`;

const ChatIcon = styled(ChatSVG)<{ $active: boolean }>`
    transition-duration: ${duration.transition};
    width: 25px;
    path {
        ${(props) => (props.$active ? "fill: white" : "fill: $color.accent")};
    }
`;

export default function ChatToggle(props: {
    active: boolean;
    onClick: () => void;
}) {
    return (
        <Button $active={props.active} onClick={props.onClick}>
            <ChatIcon $active={props.active} />
        </Button>
    );
}
