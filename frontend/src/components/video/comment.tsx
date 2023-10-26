import Picker from '@emoji-mart/react';
import styled from '@emotion/styled';
import { Typography } from '@mui/joy';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import ButtonGroup from '@mui/joy/ButtonGroup';
import Card from '@mui/joy/Card';
import CardContent from '@mui/joy/CardContent';
import CardOverflow from '@mui/joy/CardOverflow';
import IconButton from '@mui/joy/IconButton';
import Input from '@mui/joy/Input';
import Slider from '@mui/joy/Slider';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { AiFillHeart, AiOutlineClose } from 'react-icons/ai';
import { HiCurrencyYen } from 'react-icons/hi2';
import { RandomReactions } from '~/components/reaction/reaction';

const chipTable = [100, 200, 500, 1000, 2000, 5000, 10000, 20000];

export function chipColor(price: number): [string, boolean] {
  // blue
  if (price < 200) return ['#33d', true];
  // light blue
  if (price < 500) return ['#0dd', false];
  // green
  if (price < 1000) return ['#4c4', false];
  // yellow
  if (price < 2000) return ['#fb0', true];
  // orange
  if (price < 5000) return ['#f80', true];
  // magenta
  if (price < 10000) return ['#d0d', true];
  // red
  return ['#f00', true];
}

interface LiveComment {
  id: string;
  userName: string;
  text: string;
  chip?: number;
}

export default function LiveComment(): React.ReactElement {
  const [commentMode, setCommentMode] = React.useState<
    'normal' | 'chip' | 'emoji'
  >('normal');
  const [chipAmount, setChipAmount] = React.useState(2000);

  const [liveComments, setLiveComments] = React.useState<LiveComment[]>([
    {
      id: '1',
      userName: 'ユーザー名1',
      text: 'メッセージ1',
    },
    {
      id: '2',
      userName: 'ユーザー名2',
      text: 'メッセージ2',
    },
    {
      id: '3',
      userName: 'ユーザー名3',
      text: 'メッセージ3',
    },
    {
      id: '4',
      userName: 'ユーザー名4',
      text: 'メッセージ4',
      chip: 2000,
    },
  ]);

  const commentsRef = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    let counter = 4;
    const timer = setInterval(() => {
      counter++;
      let chip: number | undefined;
      if (Math.random() < 0.1) {
        chip = chipTable[Math.floor(Math.random() * chipTable.length)];
      }
      setLiveComments((comments) =>
        [
          ...comments,
          {
            id: `comment-${counter}`,
            userName: `ユーザー名${counter}`,
            text: `メッセージ${counter}`,
            chip,
          },
        ].splice(-50),
      );
    }, 300);
    return () => clearInterval(timer);
  }, []);
  React.useLayoutEffect(() => {
    if (commentsRef.current) {
      commentsRef.current.scrollTop = commentsRef.current.scrollHeight;
    }
  }, [liveComments]);

  return (
    <>
      <Card sx={{ flexBasis: '250px', flexGrow: 1, gap: 0 }}>
        <CardOverflow
          sx={{
            borderBottom: (t) =>
              `1px solid ${t.vars.palette.neutral.outlinedBorder}`,
            py: 1,
          }}
        >
          <Typography level="title-lg">Chat</Typography>
        </CardOverflow>
        <CardContent
          ref={commentsRef}
          sx={{
            overflowY: 'scroll',
            mx: '-16px',
            p: '16px',
            flex: '300px 1 0',
            '&::-webkit-scrollbar': {
              width: '10px',
              height: '10px',
              // backgroundColor: '#aaa',
            },
            '&::-webkit-scrollbar-thumb': {
              backgroundColor: '#ccc',
              borderRadius: '10px',
            },
          }}
        >
          <Stack spacing={2}>
            {liveComments.map((comment) => (
              <Comment key={comment.id} comment={comment} />
            ))}
          </Stack>
          <Stack
            sx={{
              position: 'absolute',
              bottom: commentMode === 'emoji' ? '305px' : '70px',
              right: '15px',
            }}
          >
            {commentMode !== 'chip' && (
              <IconButton
                onClick={() =>
                  setCommentMode((mode) =>
                    mode === 'emoji' ? 'normal' : 'emoji',
                  )
                }
              >
                {commentMode === 'emoji' ? (
                  <AiOutlineClose size="1.5rem" />
                ) : (
                  <AiFillHeart size="1.5rem" color="#f23d5c" />
                )}
              </IconButton>
            )}
            <RandomReactions />
          </Stack>
        </CardContent>
        <CardOverflow
          sx={{
            borderTop: (t) =>
              `1px solid ${t.vars.palette.neutral.outlinedBorder}`,
            py: 1,
            overflow: 'hidden',
          }}
        >
          {commentMode === 'emoji' && (
            <PickerWrapper>
              <Picker
                onEmojiSelect={console.log}
                previewPosition="none"
                navPosition="bottom"
                skinTonePosition="none"
                dynamicWidth={true}
                set="twitter"
              />
            </PickerWrapper>
          )}
          {commentMode === 'normal' && (
            <>
              <Stack direction="row">
                <Avatar />
                <Input
                  endDecorator={
                    <Button variant="plain" color="neutral">
                      送信
                    </Button>
                  }
                  placeholder="チャット…"
                  sx={{ flexGrow: 1, ml: 2, mr: 1 }}
                />
                <Button
                  variant="plain"
                  color="neutral"
                  sx={{ py: 0, px: 1 }}
                  onClick={() => setCommentMode('chip')}
                >
                  <HiCurrencyYen size="1.5rem" />
                </Button>
              </Stack>
            </>
          )}
          {commentMode === 'chip' && (
            <>
              <Stack spacing={2}>
                <Typography level="body-lg" sx={{ py: 1 }}>
                  チップを送る
                </Typography>
                <TipComment text="" amount={chipAmount} isEditable />
                <Slider
                  defaultValue={4}
                  step={1}
                  marks
                  min={0}
                  max={7}
                  scale={(x) => chipTable[x] ?? 0}
                  valueLabelDisplay="auto"
                  value={chipTable.indexOf(chipAmount)}
                  onChange={(e, v) => setChipAmount(chipTable[v as number])}
                />
                <ButtonGroup buttonFlex="1 0 100px">
                  <Button color="primary" variant="soft">
                    チップを送信
                  </Button>
                  <Button
                    color="neutral"
                    onClick={() => setCommentMode('normal')}
                  >
                    キャンセル
                  </Button>
                </ButtonGroup>
              </Stack>
            </>
          )}
        </CardOverflow>
      </Card>
    </>
  );
}

const PickerWrapper = styled.div`
  & > div {
    margin: -12px -15px -9px;
  }
  & > div > em-emoji-picker {
    width: 100%;
    height: 300px;
  }
`;

interface CommentProps {
  comment: LiveComment;
}
const Comment = React.memo(function Comment({
  comment,
}: CommentProps): React.ReactElement {
  return comment.chip ? (
    <TipComment text={comment.text} amount={comment.chip} />
  ) : (
    <Stack direction="row" spacing={1} alignItems="center">
      <Avatar size="sm" />
      <Typography level="title-sm">{comment.userName}</Typography>
      <Typography level="body-md">
        <span>{comment.text}</span>
      </Typography>
    </Stack>
  );
});

interface TipCommentProps {
  amount: number;
  text: string;
  isEditable?: boolean;
  onChange?(text: string): void;
}
function TipComment(props: TipCommentProps): React.ReactElement {
  const color = chipColor(props.amount);
  return (
    <Stack
      sx={{
        p: 2,
        background: color[0],
        borderRadius: 10,
      }}
      spacing={1}
    >
      <Stack direction="row" spacing={1} alignItems="center">
        <Avatar size="sm" />
        <Typography
          level="title-sm"
          sx={{
            color: color[1] ? '#fff' : '#000',
          }}
        >
          ユーザー名
        </Typography>
        <Typography
          level="body-md"
          sx={{
            color: color[1] ? '#fff' : '#000',
          }}
        >
          <span>{props.amount} ISU</span>
        </Typography>
      </Stack>
      {props.isEditable ? (
        <Input
          sx={{
            background: 'rgba(255,255,255,0.3)',
            color: color[1] ? '#fff' : '#000',
          }}
          value={props.text}
          onChange={(e) => props.onChange?.(e.target.value)}
        />
      ) : (
        <Typography level="body-md" sx={{ color: color[1] ? '#fff' : '#000' }}>
          {props.text}
        </Typography>
      )}
    </Stack>
  );
}
