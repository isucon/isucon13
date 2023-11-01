import { Emoji } from '@emoji-mart/data';
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
import Skeleton from '@mui/joy/Skeleton';
import Slider from '@mui/joy/Slider';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { Controller, useForm } from 'react-hook-form';
import { AiFillHeart, AiOutlineClose } from 'react-icons/ai';
import { HiCurrencyYen } from 'react-icons/hi2';
import { apiClient } from '~/api/client';
import { useLiveStreamComment, useLiveStreamReaction } from '~/api/hooks';
import { Schemas } from '~/api/types';
import {
  RandomReactions,
  Reaction,
  ReactionView,
} from '~/components/reaction/reaction';

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

// interface LiveComment {
//   id: string;
//   userName: string;
//   text: string;
//   chip?: number;
// }

export interface LiveCommentProps {
  type: 'real' | 'random'; // real: 実際のAPIコールを伴う, random: ランダムコメントでのデモ
  livestream_id: number;
  is_loading?: boolean;
}

export default function LiveComment(
  props: LiveCommentProps,
): React.ReactElement {
  const [commentMode, setCommentMode] = React.useState<
    'normal' | 'chip' | 'emoji'
  >('normal');
  const [chipAmount, setChipAmount] = React.useState(2000);

  const [localLiveComments, setLocalLiveComments] = React.useState<
    Schemas.Livecomment[]
  >([]);
  const remoteLiveComment = useLiveStreamComment(
    props.type === 'real' ? props.livestream_id.toString() : null,
  );
  const liveComments: Schemas.Livecomment[] =
    props.type === 'real' ? remoteLiveComment.data ?? [] : localLiveComments;

  const counterRef = React.useRef(0);
  React.useEffect(() => {
    if (props.type === 'real') {
      let timer: number | undefined = undefined;
      const cb = async () => {
        await Promise.all([
          remoteLiveComment.mutate(),
          remoteReaction.mutate(),
        ]);
        timer = setTimeout(cb, 3000); // fetch interval
      };
      cb();
      return () => {
        if (timer) {
          clearTimeout(timer);
        }
      };
    } else {
      const timer = setInterval(() => {
        counterRef.current++;
        let tip: number | undefined;
        if (Math.random() < 0.1) {
          tip = chipTable[Math.floor(Math.random() * chipTable.length)];
        }
        setLocalLiveComments((comments) => {
          const newList = [
            ...comments,
            {
              id: counterRef.current,
              user: {
                id: counterRef.current,
                name: `user-${counterRef.current}`,
                display_name: `ユーザー名${counterRef.current}`,
                description: '',
                is_popular: false,
                theme: '',
              },
              comment: `メッセージ${counterRef.current}`,
              tip: tip,
            } satisfies Schemas.Livecomment,
          ];
          if (newList.length > 40) {
            return newList.splice(-30);
          } else {
            return newList;
          }
        });
      }, 300);
      return () => clearInterval(timer);
    }
  }, [props.type]);

  const commentsRef = React.useRef<HTMLDivElement>(null);
  React.useLayoutEffect(() => {
    if (commentsRef.current) {
      commentsRef.current.scrollTop = commentsRef.current.scrollHeight;
    }
  }, [liveComments]);

  const form = useForm<{ normalComment: string; tipComment: string }>({
    defaultValues: {
      normalComment: '',
      tipComment: '',
    },
  });
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const submitNormal = React.useCallback(async () => {
    if (props.type === 'real') {
      const comment = form.getValues('normalComment');
      setIsSubmitting(true);
      try {
        await apiClient.post$livestream$livestreamid$livecomment({
          parameter: {
            livestreamid: props.livestream_id.toString(),
          },
          requestBody: {
            comment: comment,
          },
        });
        remoteLiveComment.mutate();
        form.setValue('normalComment', '');
      } finally {
        setIsSubmitting(false);
      }
    } else {
      const comment = form.getValues('normalComment');
      setLocalLiveComments((comments) => {
        counterRef.current++;
        return [
          ...comments,
          {
            id: counterRef.current,
            user: {
              id: counterRef.current,
              name: `user-${counterRef.current}`,
              display_name: `ユーザー名${counterRef.current}`,
              description: '',
              is_popular: false,
              theme: '',
            },
            comment: comment,
            tip: undefined,
          } satisfies Schemas.Livecomment,
        ];
      });
      form.setValue('normalComment', '');
    }
  }, [form, props.type]);

  const submitTip = React.useCallback(async () => {
    if (props.type === 'real') {
      const comment = form.getValues('tipComment');
      setIsSubmitting(true);
      try {
        await apiClient.post$livestream$livestreamid$livecomment({
          parameter: {
            livestreamid: props.livestream_id.toString(),
          },
          requestBody: {
            comment: comment,
            tip: chipAmount,
          },
        });
        remoteLiveComment.mutate();
        form.setValue('tipComment', '');
      } finally {
        setIsSubmitting(false);
      }
    } else {
      const comment = form.getValues('tipComment');
      setLocalLiveComments((comments) => {
        counterRef.current++;
        return [
          ...comments,
          {
            id: counterRef.current,
            user: {
              id: counterRef.current,
              name: `user-${counterRef.current}`,
              display_name: `ユーザー名${counterRef.current}`,
              description: '',
              is_popular: false,
              theme: '',
            },
            comment: comment,
            tip: chipAmount,
          } satisfies Schemas.Livecomment,
        ];
      });
      form.setValue('tipComment', '');
    }
  }, [form, props.type, chipAmount]);

  const remoteReaction = useLiveStreamReaction(
    props.type === 'real' ? props.livestream_id.toString() : null,
  );
  const reactions: Reaction[] =
    remoteReaction.data?.map(
      (r) => ({ id: r.id, shortcodes: r.emoji_name }) satisfies Reaction,
    ) ?? [];
  const onEmojiSelect = React.useCallback(
    async (emoji: Emoji & { shortcodes: string }) => {
      if (props.type === 'real') {
        await apiClient.post$livestream$livestreamid$reaction({
          parameter: {
            livestreamid: props.livestream_id.toString(),
          },
          requestBody: {
            emoji_name: emoji.shortcodes,
          },
        });
        await remoteReaction.mutate();
      }
    },
    [],
  );

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
          <Stack spacing={2} sx={{ wordBreak: 'break-all' }}>
            {props.is_loading ||
            (props.type === 'real' && remoteLiveComment.isLoading) ? (
              <>
                <Skeleton variant="text" />
                <Skeleton variant="text" />
                <Skeleton variant="text" />
              </>
            ) : (
              liveComments.map((comment) => (
                <Comment key={comment.id} comment={comment} />
              ))
            )}
          </Stack>
          <Stack
            sx={{
              position: 'absolute',
              bottom:
                commentMode === 'emoji'
                  ? '305px'
                  : commentMode === 'normal'
                  ? '70px'
                  : '300px',
              right: commentMode === 'chip' ? '45px' : '15px',
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
            {props.type === 'real' ? (
              <ReactionView reactions={reactions} />
            ) : (
              <RandomReactions />
            )}
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
                onEmojiSelect={onEmojiSelect}
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
                  {...form.register('normalComment')}
                  endDecorator={
                    <Button
                      variant="plain"
                      color="neutral"
                      onClick={() => submitNormal()}
                      loading={isSubmitting}
                    >
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
                <Controller
                  control={form.control}
                  name="tipComment"
                  render={({ field }) => (
                    <TipComment
                      text={field.value}
                      onChange={(text) => field.onChange(text)}
                      amount={chipAmount}
                      isEditable
                    />
                  )}
                />
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
                  <Button
                    color="primary"
                    variant="soft"
                    onClick={() => submitTip()}
                    loading={isSubmitting}
                  >
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
  comment: Schemas.Livecomment;
}
const Comment = React.memo(function Comment({
  comment,
}: CommentProps): React.ReactElement {
  return comment.tip ? (
    <TipComment text={comment.comment ?? ''} amount={comment.tip ?? 0} />
  ) : (
    <CommentWrapper>
      <Avatar size="sm" sx={{ position: 'absolute' }} />
      <Typography component="div" sx={{ ml: '40px', pt: '2px' }}>
        <Typography level="title-sm" component="span">
          {comment.user?.display_name ?? ''}
        </Typography>
        <Typography level="body-md" component="span" sx={{ ml: 1 }}>
          <span>{comment.comment ?? ''}</span>
        </Typography>
      </Typography>
    </CommentWrapper>
    // <Stack direction="row" spacing={1} alignItems="center">
    //   <Avatar size="sm" />
    //   <Typography level="title-sm">{comment.userName}</Typography>
    //   <Typography level="body-md">
    //     <span>{comment.text}</span>
    //   </Typography>
    // </Stack>
  );
});

const CommentWrapper = styled.div`
  position: relative;
`;

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
