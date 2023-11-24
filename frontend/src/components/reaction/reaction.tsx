import { keyframes } from '@emotion/react';
import styled from '@emotion/styled';
import React from 'react';

export interface ReactionProps {
  children: React.ReactNode;
  onFinished?: () => void;
}
export const Reaction = React.memo(function Reaction(
  props: ReactionProps,
): React.ReactNode {
  const elem = React.useRef<HTMLSpanElement>(null);
  const [isFinished, setIsFinished] = React.useState(false);

  const startPos = React.useMemo(() => {
    const offset = 15;
    return Math.random() * offset - offset / 2;
  }, []);
  const endPos = React.useMemo(() => {
    const offset = 30;
    return Math.random() * offset - offset / 2;
  }, []);
  const rotation = React.useMemo(() => {
    const offset = 30;
    return Math.random() * offset - offset / 2;
  }, []);

  React.useEffect(() => {
    const e = elem.current;
    if (e) {
      const handler = () => {
        setIsFinished(true);
        props.onFinished?.();
      };
      e.addEventListener('animationend', handler);
      return () => e.removeEventListener('animationend', handler);
    }
  }, [props.onFinished]);

  if (isFinished) {
    return null;
  }
  return (
    <Emoji ref={elem} style={{ left: startPos }} rotate={rotation} end={endPos}>
      {props.children}
    </Emoji>
  );
});

const shortcoes = [':+1:', ':heart:', ':clap:', ':tada:', ':smile:'];
export function RandomReactions(): React.ReactElement {
  const [reactions, setReactions] = React.useState<React.ReactNode[]>([]);
  const count = React.useRef(0);
  React.useEffect(() => {
    const timer = setInterval(() => {
      setReactions((prev) =>
        [
          ...prev,
          <Reaction key={count.current++}>
            <em-emoji
              shortcodes={shortcoes[(Math.random() * shortcoes.length) | 0]}
              set="twitter"
            ></em-emoji>
          </Reaction>,
        ].slice(-50),
      );
    }, 300);
    return () => clearInterval(timer);
  }, []);
  return <RandomReactionsWrapper>{reactions}</RandomReactionsWrapper>;
}

export interface Reaction {
  id: string;
  shortcodes: string;
}
export interface ReactionViewProps {
  reactions: Reaction[];
  onFinished?(id: string): void;
}
export function ReactionView({
  reactions,
  onFinished,
}: ReactionViewProps): React.ReactElement {
  return (
    <>
      {reactions.map((r) => (
        <Reaction key={r.id} onFinished={() => onFinished?.(r.id)}>
          <em-emoji shortcodes={r.shortcodes} set="twitter"></em-emoji>
        </Reaction>
      ))}
    </>
  );
}

const RandomReactionsWrapper = styled.div`
  position: absolute;
`;

const floatAnimation = (rotate: number, end: number) => keyframes`
  0% {
    transform: translateY(0);
    opacity: 1;
  }
  100% {
    transform: translateY(-100px) translateX(${end}px) rotate(${rotate}deg) scale(0.5);
    opacity: 0;
  }
`;

const Emoji = styled.span<{ rotate: number; end: number }>`
  font-size: 24px;
  position: absolute;
  bottom: 0;
  opacity: 0;
  animation: ${(props) => floatAnimation(props.rotate, props.end)} 3s linear;
`;
