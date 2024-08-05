import React from "react";
import NotifyingClickable from "./NotifyingButton";
import Copy from "../../assets/icons/copy.svg";
import { styled } from "styled-components";
import { color, duration } from "../../stylingConstants";

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

export default function CopyButton({ toCopy }: { toCopy: string }) {
    function copyToClipboard() {
        navigator.clipboard.writeText(toCopy);
    }
    return (
        <NotifyingClickable
            onClick={copyToClipboard}
            text="Copied to clipboard!"
        >
            <StyledCopy />
        </NotifyingClickable>
    );
}
