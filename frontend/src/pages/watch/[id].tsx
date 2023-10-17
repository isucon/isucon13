import Picker from '@emoji-mart/react';
import styled from '@emotion/styled';
import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
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
import { Link } from 'react-router-dom';
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

export default function WatchPage(): React.ReactElement {
  const [commentMode, setCommentMode] = React.useState<
    'normal' | 'chip' | 'emoji'
  >('normal');
  const [chipAmount, setChipAmount] = React.useState(2000);

  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={2}>
      <Stack direction="row" gap={2}>
        <AspectRatio ratio={16 / 9} sx={{ flexBasis: '600px', flexGrow: 3 }}>
          <video />
        </AspectRatio>
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
              {Array(50)
                .fill(0)
                .map((_, i) => (
                  <Stack
                    direction="row"
                    spacing={1}
                    key={i}
                    alignItems="center"
                  >
                    <Avatar size="sm" />
                    <Typography level="title-sm">ユーザー名{i + 1}</Typography>
                    <Typography level="body-md">
                      <span>メッセージ{i + 1}</span>
                    </Typography>
                  </Stack>
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
                  {/* todo convert component */}
                  <Stack
                    sx={{
                      p: 2,
                      background: chipColor(chipAmount)[0],
                      borderRadius: 10,
                    }}
                    spacing={1}
                  >
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Avatar size="sm" />
                      <Typography
                        level="title-sm"
                        sx={{
                          color: chipColor(chipAmount)[1] ? '#fff' : '#000',
                        }}
                      >
                        ユーザー名
                      </Typography>
                      <Typography
                        level="body-md"
                        sx={{
                          color: chipColor(chipAmount)[1] ? '#fff' : '#000',
                        }}
                      >
                        <span>{chipAmount} ISU</span>
                      </Typography>
                    </Stack>
                    <Input
                      sx={{
                        background: 'rgba(255,255,255,0.3)',
                        color: chipColor(chipAmount)[1] ? '#fff' : '#000',
                      }}
                    />
                  </Stack>
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
      </Stack>

      <Stack direction="row" gap={2}>
        <Stack sx={{ flexBasis: '600px', flexGrow: 3 }}>
          <Typography level="h3">動画タイトル</Typography>
          <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
            <Link to="/user">
              <Avatar />
            </Link>
            <div>
              <Link to="/user" style={{ textDecoration: 'none' }}>
                <Typography level="title-sm">チャンネル名</Typography>
              </Link>
              <Typography level="body-sm">
                <Stack direction="row" spacing={2}>
                  <span>チャンネル登録者数1234人</span>
                </Stack>
              </Typography>
            </div>
            <div>
              <Button variant="outlined" color="neutral" sx={{ marginLeft: 3 }}>
                チャンネル登録
              </Button>
            </div>
          </Stack>
          <Card variant="plain" sx={{ my: 2 }}>
            <Stack direction="row" spacing={2}>
              <Typography level="title-sm">1,234人が視聴中</Typography>
              <Typography level="title-sm">2時間前にライブ配信開始</Typography>
            </Stack>
            <Typography
              level="body-md"
              sx={{ whiteSpace: 'pre-wrap' }}
            >{`説明文\n2行目\n三行目`}</Typography>
          </Card>
        </Stack>
        <Stack sx={{ flexBasis: '250px', flexGrow: 1, px: '16px' }}>
          <Typography level="h3">Related Live</Typography>
          <Stack spacing={2} sx={{ my: 1 }}>
            {Array(5)
              .fill(0)
              .map((_, i) => (
                <Link key={i} to="/" style={{ textDecoration: 'none' }}>
                  <Stack direction="row" spacing={2}>
                    <AspectRatio sx={{ borderRadius: 10, flexBasis: '35%' }}>
                      <img
                        src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
                        loading="lazy"
                      />
                    </AspectRatio>
                    <Stack sx={{ marginTop: 1 }}>
                      <Typography level="title-md">
                        動画タイトル{i + 1}
                      </Typography>
                      <Typography level="body-sm">チャンネル名</Typography>
                      <Typography level="body-sm">1,234人が視聴中</Typography>
                    </Stack>
                  </Stack>
                </Link>
              ))}
          </Stack>
        </Stack>
      </Stack>
    </Stack>
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
