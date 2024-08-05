import React, {
    MouseEventHandler,
    useContext,
    useEffect,
    useState,
} from "react";
import styled from "styled-components";
import {
    borderRadius,
    color,
    fontSizes,
    padding,
    shadow,
    duration,
} from "../stylingConstants";
import Arrow from "../assets/icons/read_source_arrow.svg";
import CopyButton from "./buttons/CopyButton";
import { Spacer } from "./Layout";
import { DownvoteButton, UpvoteButton } from "./buttons/VoteButtons";
import { ModelServiceContext } from "../Context";
import { ModelService, ReferenceInfo } from "../modelServices";

const Card = styled.section<{ $opacity: string }>`
    background: white;
    width: 100%;
    border-radius: ${borderRadius.card};
    height: fit-content;
    display: flex;
    flex-direction: row;
    box-shadow: ${shadow.card};
    transition-duration: ${duration.transition};
    opacity: ${(props) => props.$opacity};
`;

const Stripe = styled.section`
    background: ${color.accent};
    height: auto;
    width: 25px;
    border-radius: ${borderRadius.card} 0 0 ${borderRadius.card};
`;

const TextContainer = styled.section`
    padding: ${padding.card};
    color: black;
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
`;

const Header = styled.a`
    font-size: ${fontSizes.m};
    font-weight: bold;
    display: flex;
    text-align: left;
    align-items: center;
    margin-bottom: 10px;
    text-decoration: none;
    color: black;
    transition-duration: ${duration.transition};
    width: fit-content;

    &:hover {
        color: ${color.accent};
        transition-duration: ${duration.transition};
        cursor: pointer;
    }
`;

export const ReadSourceButton = styled.section`
    background-color: white;
    border: 1px solid ${color.accent};
    border-radius: ${borderRadius.smallButton};
    transition-duration: ${duration.transition};
    font-size: ${fontSizes.s};
    font-weight: normal;
    color: ${color.accent};
    width: fit-content;
    padding: ${padding.smallButton};

    ${Header}:hover & {
        background-color: ${color.accent};
        color: white;
    }
`;

export const StyledArrow = styled(Arrow)`
    transition-duration: ${duration.transition};

    ${Header}:hover & path {
        fill: white;
        transition-duration: ${duration.transition};
    }
`;

const Content = styled.section`
    font-size: ${fontSizes.s};
    text-align: left;
`;

const ButtonsContainer = styled.section`
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: ${padding.card};
`;

const VoteButtonsContainer = styled.section`
    display: flex;
    flex-direction: column;
`;

interface ReferenceProps {
    info: ReferenceInfo;
    onOpen: MouseEventHandler<any>;
    onUpvote: MouseEventHandler<any>;
    onDownvote: MouseEventHandler<any>;
    checked: boolean;
    onCheck: () => void;
}

export default function Reference({
    info,
    onOpen,
    onUpvote,
    onDownvote,
    checked,
    onCheck,
}: ReferenceProps) {
    const [opacity, setOpacity] = useState("0");
    useEffect(() => setOpacity("100%"), []);

    const isReadableSource = (source: string): boolean => {
        const lowerSource = source.toLowerCase();
        return lowerSource.endsWith(".pdf") || lowerSource.endsWith(".docx");
    };

    return (
        <Card $opacity={opacity}>
            <Stripe />
            <TextContainer>
                <Header onClick={onOpen} target={"_blank"}>
                    {info.sourceName}
                    <Spacer $width="10px" />
                    {isReadableSource(info.sourceName) && (
                        <ReadSourceButton>
                            {/* Read source */}
                            Read source <StyledArrow />
                        </ReadSourceButton>
                    )}
                </Header>
                <Content> {info.content} </Content>
            </TextContainer>
            <ButtonsContainer>
                <VoteButtonsContainer>
                    <input
                        type="checkbox"
                        checked={checked}
                        onChange={onCheck}
                    />
                    <Spacer $height="10px" />
                    <UpvoteButton onClick={onUpvote} />
                    <Spacer $height="10px" />
                    <DownvoteButton onClick={onDownvote} />
                </VoteButtonsContainer>
                <Spacer $height={padding.card} />
                <CopyButton toCopy={info.content} />
            </ButtonsContainer>
        </Card>
    );
}
